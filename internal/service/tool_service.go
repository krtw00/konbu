package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

const automaticIconRefreshMaxAge = 24 * time.Hour

type ToolService struct {
	queries *repository.Queries
	db      *sql.DB
}

func NewToolService(db *sql.DB) *ToolService {
	return &ToolService{
		queries: repository.New(db),
		db:      db,
	}
}

func (s *ToolService) ListTools(ctx context.Context, userID uuid.UUID) ([]model.Tool, error) {
	rows, err := s.queries.ListToolsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	tools := make([]model.Tool, len(rows))
	for i, r := range rows {
		tools[i] = toModelTool(r)
	}
	return tools, nil
}

func (s *ToolService) CreateTool(ctx context.Context, userID uuid.UUID, req model.CreateToolRequest) (*model.Tool, error) {
	if req.Name == "" || req.URL == "" {
		return nil, apperror.BadRequest("name and url are required")
	}

	// Auto-fetch favicon
	icon := ""
	iconCheckedAt := (*time.Time)(nil)
	if req.URL != "" {
		now := time.Now().UTC()
		iconCheckedAt = &now
		icon = FetchFavicon(req.URL)
	}

	maxSort, err := s.queries.MaxToolSortOrder(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	t, err := s.queries.CreateTool(ctx, userID, req.Name, req.URL, icon, iconCheckedAt, req.Category, maxSort+1)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	result := toModelTool(t)
	return &result, nil
}

func (s *ToolService) UpdateTool(ctx context.Context, id, userID uuid.UUID, req model.UpdateToolRequest) (*model.Tool, error) {
	existing, err := s.queries.GetToolByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("tool")
		}
		return nil, apperror.Internal(err)
	}

	// Merge: keep existing values for empty fields
	name := req.Name
	if name == "" {
		name = existing.Name
	}
	url := req.URL
	if url == "" {
		url = existing.URL
	}
	category := req.Category
	if category == "" {
		category = existing.Category
	}

	// Re-fetch favicon if URL changed or icon is missing
	icon := existing.Icon
	iconCheckedAt := existing.IconCheckedAt
	if url != existing.URL {
		icon = ""
		iconCheckedAt = nil
	}
	if url != "" && (url != existing.URL || icon == "") {
		now := time.Now().UTC()
		iconCheckedAt = &now
		fetched := FetchFavicon(url)
		if fetched != "" {
			icon = fetched
		}
	}
	if url == "" {
		iconCheckedAt = nil
	}

	t, err := s.queries.UpdateTool(ctx, id, userID, name, url, icon, iconCheckedAt, category, existing.SortOrder)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	result := toModelTool(t)
	return &result, nil
}

func (s *ToolService) RefreshToolIcons(ctx context.Context, userID uuid.UUID) (int, error) {
	rows, err := s.queries.ListToolsByUserID(ctx, userID)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	now := time.Now().UTC()
	count := 0
	for _, r := range rows {
		updated, err := s.refreshToolIcon(ctx, r, now, true)
		if err != nil {
			return count, apperror.Internal(err)
		}
		if updated {
			count++
		}
	}
	return count, nil
}

func (s *ToolService) RefreshStaleIcons(ctx context.Context, now time.Time) (int, error) {
	rows, err := s.queries.ListToolsNeedingIconRefresh(ctx, now.Add(-automaticIconRefreshMaxAge))
	if err != nil {
		return 0, err
	}
	log.Printf("icon refresh: found %d stale tools", len(rows))
	count := 0
	for _, r := range rows {
		updated, err := s.refreshToolIcon(ctx, r, now, false)
		if err != nil {
			log.Printf("icon refresh: failed to refresh %s: %v", r.URL, err)
			continue
		}
		if updated {
			count++
		}
	}
	return count, nil
}

func (s *ToolService) StartIconRefreshLoop(interval time.Duration) {
	go func() {
		s.runIconRefresh()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.runIconRefresh()
		}
	}()
}

func (s *ToolService) runIconRefresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	count, err := s.RefreshStaleIcons(ctx, time.Now().UTC())
	if err != nil {
		log.Printf("icon refresh error: %v", err)
	} else {
		log.Printf("icon refresh: updated %d tools", count)
	}
}

func (s *ToolService) DeleteTool(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.SoftDeleteTool(ctx, id, userID)
}

func (s *ToolService) ReorderTools(ctx context.Context, userID uuid.UUID, order []uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	for i, id := range order {
		if err := q.UpdateToolSortOrder(ctx, id, userID, i); err != nil {
			return apperror.Internal(err)
		}
	}
	return tx.Commit()
}

func (s *ToolService) refreshToolIcon(ctx context.Context, tool repository.Tool, checkedAt time.Time, force bool) (bool, error) {
	if !force && !toolNeedsIconRefresh(tool, checkedAt) {
		return false, nil
	}
	if tool.URL == "" {
		return false, nil
	}

	icon := tool.Icon
	if fetched := FetchFavicon(tool.URL); fetched != "" {
		icon = fetched
	}

	refreshedAt := checkedAt.UTC()
	_, err := s.queries.UpdateTool(ctx, tool.ID, tool.UserID, tool.Name, tool.URL, icon, &refreshedAt, tool.Category, tool.SortOrder)
	if err != nil {
		return false, err
	}
	return true, nil
}

func toolNeedsIconRefresh(tool repository.Tool, now time.Time) bool {
	if tool.URL == "" {
		return false
	}
	if tool.IconCheckedAt == nil {
		return true
	}
	return !tool.IconCheckedAt.After(now.Add(-automaticIconRefreshMaxAge))
}

func toModelTool(t repository.Tool) model.Tool {
	return model.Tool{
		ID:        t.ID,
		Name:      t.Name,
		URL:       t.URL,
		Icon:      t.Icon,
		Category:  t.Category,
		SortOrder: t.SortOrder,
		CreatedAt: t.CreatedAt,
	}
}
