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

// Search performs the legacy search (used by CommandPalette, CLI, MCP).
// Returns flat []SearchResult for backward compat.
func (s *SearchService) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]model.SearchResult, error) {
	resp, err := s.SearchAdvanced(ctx, userID, model.SearchParams{
		Query:  query,
		Limit:  limit,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SearchAdvanced performs extended search with filters, tag search, tool search, and fuzzy suggestions.
func (s *SearchService) SearchAdvanced(ctx context.Context, userID uuid.UUID, params model.SearchParams) (*model.SearchResponse, error) {
	query := strings.TrimSpace(params.Query)
	if len(query) < 2 {
		return nil, apperror.BadRequest("query must be at least 2 characters")
	}
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 20
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	pattern := "%" + query + "%"
	seen := make(map[uuid.UUID]struct{})
	var allResults []model.SearchResult
	typeSet := makeTypeSet(params.Types)

	// 1. Content search (ILIKE) + tag-based search
	if typeSet["memo"] {
		items, err := s.searchMemosAll(ctx, userID, pattern, query, params)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if _, ok := seen[item.ID]; !ok {
				seen[item.ID] = struct{}{}
				allResults = append(allResults, item)
			}
		}
	}

	if typeSet["todo"] {
		items, err := s.searchTodosAll(ctx, userID, pattern, query, params)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if _, ok := seen[item.ID]; !ok {
				seen[item.ID] = struct{}{}
				allResults = append(allResults, item)
			}
		}
	}

	if typeSet["event"] {
		items, err := s.searchEventsAll(ctx, userID, pattern, query, params)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if _, ok := seen[item.ID]; !ok {
				seen[item.ID] = struct{}{}
				allResults = append(allResults, item)
			}
		}
	}

	// 2. Tool search
	if typeSet["tool"] && params.Tag == "" {
		tools, err := s.queries.SearchTools(ctx, userID, pattern, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, t := range tools {
			if _, ok := seen[t.ID]; !ok {
				seen[t.ID] = struct{}{}
				allResults = append(allResults, model.SearchResult{
					Type:      "tool",
					ID:        t.ID,
					Title:     t.Name,
					Snippet:   t.URL,
					Tags:      []string{},
					UpdatedAt: t.CreatedAt,
				})
			}
		}
	}

	// 3. Sort by updated_at desc
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].UpdatedAt.After(allResults[j].UpdatedAt)
	})

	// 4. Count total (approximate)
	total := len(allResults)

	// 5. Paginate
	data := paginate(allResults, params.Offset, params.Limit)

	// 6. Fuzzy suggestions
	excludeIDs := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		excludeIDs = append(excludeIDs, id)
	}
	suggestions := s.fuzzySearch(ctx, userID, query, excludeIDs, typeSet)

	return &model.SearchResponse{
		Data:        data,
		Total:       total,
		Suggestions: suggestions,
	}, nil
}

func (s *SearchService) searchMemosAll(ctx context.Context, userID uuid.UUID, pattern, query string, params model.SearchParams) ([]model.SearchResult, error) {
	var results []model.SearchResult

	// Content search
	if params.Tag == "" {
		memos, err := s.queries.SearchMemosFiltered(ctx, userID, pattern, params.From, params.To, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, m := range memos {
			tags := s.getTagNames(ctx, "memo", m.ID)
			results = append(results, model.SearchResult{
				Type:      "memo",
				ID:        m.ID,
				Title:     m.Title,
				Snippet:   truncate(m.Content, 120),
				Tags:      tags,
				UpdatedAt: m.UpdatedAt,
			})
		}
	} else {
		memos, err := s.queries.SearchMemosWithTagFilter(ctx, userID, pattern, params.Tag, params.From, params.To, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, m := range memos {
			tags := s.getTagNames(ctx, "memo", m.ID)
			results = append(results, model.SearchResult{
				Type:      "memo",
				ID:        m.ID,
				Title:     m.Title,
				Snippet:   truncate(m.Content, 120),
				Tags:      tags,
				UpdatedAt: m.UpdatedAt,
			})
		}
	}

	// Tag-name search (q matches tag name → include tagged items)
	if params.Tag == "" {
		tagPattern := "%" + query + "%"
		tagMemos, err := s.queries.SearchMemosByTag(ctx, userID, tagPattern, params.Limit, 0)
		if err == nil {
			for _, m := range tagMemos {
				tags := s.getTagNames(ctx, "memo", m.ID)
				results = append(results, model.SearchResult{
					Type:      "memo",
					ID:        m.ID,
					Title:     m.Title,
					Snippet:   truncate(m.Content, 120),
					Tags:      tags,
					UpdatedAt: m.UpdatedAt,
				})
			}
		}
	}

	return results, nil
}

func (s *SearchService) searchTodosAll(ctx context.Context, userID uuid.UUID, pattern, query string, params model.SearchParams) ([]model.SearchResult, error) {
	var results []model.SearchResult

	if params.Tag == "" {
		todos, err := s.queries.SearchTodosFiltered(ctx, userID, pattern, params.From, params.To, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, t := range todos {
			tags := s.getTagNames(ctx, "todo", t.ID)
			results = append(results, model.SearchResult{
				Type:      "todo",
				ID:        t.ID,
				Title:     t.Title,
				Snippet:   truncate(t.Description, 120),
				Tags:      tags,
				UpdatedAt: t.UpdatedAt,
			})
		}
	} else {
		todos, err := s.queries.SearchTodosWithTagFilter(ctx, userID, pattern, params.Tag, params.From, params.To, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, t := range todos {
			tags := s.getTagNames(ctx, "todo", t.ID)
			results = append(results, model.SearchResult{
				Type:      "todo",
				ID:        t.ID,
				Title:     t.Title,
				Snippet:   truncate(t.Description, 120),
				Tags:      tags,
				UpdatedAt: t.UpdatedAt,
			})
		}
	}

	// Tag-name search
	if params.Tag == "" {
		tagPattern := "%" + query + "%"
		tagTodos, err := s.queries.SearchTodosByTag(ctx, userID, tagPattern, params.Limit, 0)
		if err == nil {
			for _, t := range tagTodos {
				tags := s.getTagNames(ctx, "todo", t.ID)
				results = append(results, model.SearchResult{
					Type:      "todo",
					ID:        t.ID,
					Title:     t.Title,
					Snippet:   truncate(t.Description, 120),
					Tags:      tags,
					UpdatedAt: t.UpdatedAt,
				})
			}
		}
	}

	return results, nil
}

func (s *SearchService) searchEventsAll(ctx context.Context, userID uuid.UUID, pattern, query string, params model.SearchParams) ([]model.SearchResult, error) {
	var results []model.SearchResult

	if params.Tag == "" {
		events, err := s.queries.SearchEventsFiltered(ctx, userID, pattern, params.From, params.To, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, e := range events {
			tags := s.getTagNames(ctx, "event", e.ID)
			results = append(results, model.SearchResult{
				Type:      "event",
				ID:        e.ID,
				Title:     e.Title,
				Snippet:   truncate(e.Description, 120),
				Tags:      tags,
				UpdatedAt: e.UpdatedAt,
			})
		}
	} else {
		events, err := s.queries.SearchEventsWithTagFilter(ctx, userID, pattern, params.Tag, params.From, params.To, params.Limit, 0)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		for _, e := range events {
			tags := s.getTagNames(ctx, "event", e.ID)
			results = append(results, model.SearchResult{
				Type:      "event",
				ID:        e.ID,
				Title:     e.Title,
				Snippet:   truncate(e.Description, 120),
				Tags:      tags,
				UpdatedAt: e.UpdatedAt,
			})
		}
	}

	// Tag-name search
	if params.Tag == "" {
		tagPattern := "%" + query + "%"
		tagEvents, err := s.queries.SearchEventsByTag(ctx, userID, tagPattern, params.Limit, 0)
		if err == nil {
			for _, e := range tagEvents {
				tags := s.getTagNames(ctx, "event", e.ID)
				results = append(results, model.SearchResult{
					Type:      "event",
					ID:        e.ID,
					Title:     e.Title,
					Snippet:   truncate(e.Description, 120),
					Tags:      tags,
					UpdatedAt: e.UpdatedAt,
				})
			}
		}
	}

	return results, nil
}

func (s *SearchService) fuzzySearch(ctx context.Context, userID uuid.UUID, query string, excludeIDs []uuid.UUID, typeSet map[string]bool) []model.SearchResult {
	var suggestions []model.SearchResult
	limit := 5

	if typeSet["memo"] {
		rows, err := s.queries.FuzzySearchMemos(ctx, userID, query, excludeIDs, limit)
		if err == nil {
			for _, r := range rows {
				suggestions = append(suggestions, model.SearchResult{
					Type:       r.Type,
					ID:         r.ID,
					Title:      r.Title,
					Tags:       []string{},
					Similarity: r.Similarity,
				})
			}
		}
	}
	if typeSet["todo"] {
		rows, err := s.queries.FuzzySearchTodos(ctx, userID, query, excludeIDs, limit)
		if err == nil {
			for _, r := range rows {
				suggestions = append(suggestions, model.SearchResult{
					Type:       r.Type,
					ID:         r.ID,
					Title:      r.Title,
					Tags:       []string{},
					Similarity: r.Similarity,
				})
			}
		}
	}
	if typeSet["event"] {
		rows, err := s.queries.FuzzySearchEvents(ctx, userID, query, excludeIDs, limit)
		if err == nil {
			for _, r := range rows {
				suggestions = append(suggestions, model.SearchResult{
					Type:       r.Type,
					ID:         r.ID,
					Title:      r.Title,
					Tags:       []string{},
					Similarity: r.Similarity,
				})
			}
		}
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Similarity > suggestions[j].Similarity
	})
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return suggestions
}

func (s *SearchService) getTagNames(ctx context.Context, itemType string, itemID uuid.UUID) []string {
	var tags []string
	switch itemType {
	case "memo":
		mt, err := s.queries.GetMemoTags(ctx, itemID)
		if err == nil {
			for _, t := range mt {
				tags = append(tags, t.Name)
			}
		}
	case "todo":
		tt, err := s.queries.GetTodoTags(ctx, itemID)
		if err == nil {
			for _, t := range tt {
				tags = append(tags, t.Name)
			}
		}
	case "event":
		et, err := s.queries.GetEventTags(ctx, itemID)
		if err == nil {
			for _, t := range et {
				tags = append(tags, t.Name)
			}
		}
	}
	if tags == nil {
		tags = []string{}
	}
	return tags
}

func makeTypeSet(types []string) map[string]bool {
	if len(types) == 0 {
		return map[string]bool{"memo": true, "todo": true, "event": true, "tool": true}
	}
	set := make(map[string]bool)
	for _, t := range types {
		set[t] = true
	}
	return set
}

func paginate(items []model.SearchResult, offset, limit int) []model.SearchResult {
	if offset >= len(items) {
		return []model.SearchResult{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
