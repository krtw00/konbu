package service

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type SearchService struct {
	queries *repository.Queries
	db      *sql.DB
}

func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{
		queries: repository.New(db),
		db:      db,
	}
}

func (s *SearchService) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]model.SearchResult, error) {
	if len(strings.TrimSpace(query)) < 2 {
		return nil, apperror.BadRequest("query must be at least 2 characters")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	pattern := "%" + query + "%"
	var results []model.SearchResult

	// Search memos
	memos, err := s.queries.SearchMemos(ctx, userID, pattern, limit)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	for _, m := range memos {
		results = append(results, model.SearchResult{
			Type:      "memo",
			ID:        m.ID,
			Title:     m.Title,
			Snippet:   truncate(m.Content, 120),
			UpdatedAt: m.UpdatedAt,
		})
	}

	// Search todos
	todos, err := s.queries.SearchTodos(ctx, userID, pattern, limit)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	for _, t := range todos {
		results = append(results, model.SearchResult{
			Type:      "todo",
			ID:        t.ID,
			Title:     t.Title,
			Snippet:   truncate(t.Description, 120),
			UpdatedAt: t.UpdatedAt,
		})
	}

	// Search events
	events, err := s.queries.SearchEvents(ctx, userID, pattern, limit)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	for _, e := range events {
		results = append(results, model.SearchResult{
			Type:      "event",
			ID:        e.ID,
			Title:     e.Title,
			Snippet:   truncate(e.Description, 120),
			UpdatedAt: e.UpdatedAt,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt.After(results[j].UpdatedAt)
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
