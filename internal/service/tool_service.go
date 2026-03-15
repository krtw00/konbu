package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

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
	if req.URL != "" {
		icon = FetchFavicon(req.URL)
	}

	maxSort, err := s.queries.MaxToolSortOrder(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	t, err := s.queries.CreateTool(ctx, userID, req.Name, req.URL, icon, req.Category, maxSort+1)
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

	// Re-fetch favicon if URL changed or icon is missing
	icon := existing.Icon
	if req.URL != existing.URL || icon == "" {
		if req.URL != "" {
			icon = FetchFavicon(req.URL)
		}
	}

	t, err := s.queries.UpdateTool(ctx, id, userID, req.Name, req.URL, icon, req.Category, existing.SortOrder)
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
	count := 0
	for _, r := range rows {
		if r.Icon != "" || r.URL == "" {
			continue
		}
		icon := FetchFavicon(r.URL)
		if icon == "" {
			continue
		}
		s.queries.UpdateTool(ctx, r.ID, userID, r.Name, r.URL, icon, r.Category, r.SortOrder)
		count++
	}
	return count, nil
}

func (s *ToolService) RefreshEmptyIcons(ctx context.Context) (int, error) {
	rows, err := s.queries.ListToolsWithEmptyIcon(ctx)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, r := range rows {
		icon := FetchFavicon(r.URL)
		if icon == "" {
			continue
		}
		s.queries.UpdateTool(ctx, r.ID, r.UserID, r.Name, r.URL, icon, r.Category, r.SortOrder)
		count++
	}
	return count, nil
}

func (s *ToolService) StartIconRefreshLoop(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			count, err := s.RefreshEmptyIcons(ctx)
			if err != nil {
				log.Printf("icon refresh error: %v", err)
			} else if count > 0 {
				log.Printf("icon refresh: updated %d tools", count)
			}
			cancel()
		}
	}()
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

type HealthResult struct {
	ID     uuid.UUID `json:"id"`
	URL    string    `json:"url"`
	Alive  bool      `json:"alive"`
	Status int       `json:"status"`
}

func (s *ToolService) HealthCheck(ctx context.Context, userID uuid.UUID) ([]HealthResult, error) {
	tools, err := s.queries.ListToolsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	results := make([]HealthResult, len(tools))
	for i, t := range tools {
		results[i] = HealthResult{ID: t.ID, URL: t.URL}
		resp, err := client.Head(t.URL)
		if err != nil {
			continue
		}
		resp.Body.Close()
		results[i].Status = resp.StatusCode
		results[i].Alive = resp.StatusCode >= 200 && resp.StatusCode < 400
	}
	return results, nil
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
