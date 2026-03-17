package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SearchRow struct {
	ID          uuid.UUID
	Title       string
	Content     string
	Description string
	UpdatedAt   time.Time
}

type SuggestionRow struct {
	ID         uuid.UUID
	Title      string
	Similarity float64
	Type       string
}

type ToolSearchRow struct {
	ID        uuid.UUID
	Name      string
	URL       string
	Icon      string
	CreatedAt time.Time
}

// --- Content search (ILIKE) with optional date filter ---

func (q *Queries) SearchMemos(ctx context.Context, userID uuid.UUID, pattern string, limit int) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, title, COALESCE(content, ''), created_at AS updated_at
		 FROM memos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR content ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		userID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchRow
	for rows.Next() {
		var r SearchRow
		if err := rows.Scan(&r.ID, &r.Title, &r.Content, &r.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (q *Queries) SearchMemosFiltered(ctx context.Context, userID uuid.UUID, pattern string, from, to *time.Time, limit, offset int) ([]SearchRow, error) {
	query := `SELECT id, title, COALESCE(content, ''), created_at AS updated_at
		 FROM memos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR content ILIKE $2)`
	args := []any{userID, pattern}
	idx := 3

	if from != nil {
		query += " AND created_at >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND created_at <= " + ph(idx)
		args = append(args, *to)
		idx++
	}

	query += " ORDER BY updated_at DESC"
	query += " LIMIT " + ph(idx)
	args = append(args, limit)
	idx++
	query += " OFFSET " + ph(idx)
	args = append(args, offset)

	return q.scanSearchRows(ctx, query, args)
}

func (q *Queries) SearchTodos(ctx context.Context, userID uuid.UUID, pattern string, limit int) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, title, COALESCE(description, ''), updated_at
		 FROM todos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR description ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		userID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchRow
	for rows.Next() {
		var r SearchRow
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (q *Queries) SearchTodosFiltered(ctx context.Context, userID uuid.UUID, pattern string, from, to *time.Time, limit, offset int) ([]SearchRow, error) {
	query := `SELECT id, title, COALESCE(description, ''), updated_at
		 FROM todos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR description ILIKE $2)`
	args := []any{userID, pattern}
	idx := 3

	if from != nil {
		query += " AND due_date >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND due_date <= " + ph(idx)
		args = append(args, *to)
		idx++
	}

	query += " ORDER BY updated_at DESC"
	query += " LIMIT " + ph(idx)
	args = append(args, limit)
	idx++
	query += " OFFSET " + ph(idx)
	args = append(args, offset)

	return q.scanSearchRows(ctx, query, args)
}

func (q *Queries) SearchEvents(ctx context.Context, userID uuid.UUID, pattern string, limit int) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, title, COALESCE(description, ''), updated_at
		 FROM calendar_events
		 WHERE created_by = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR description ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		userID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchRow
	for rows.Next() {
		var r SearchRow
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (q *Queries) SearchEventsFiltered(ctx context.Context, userID uuid.UUID, pattern string, from, to *time.Time, limit, offset int) ([]SearchRow, error) {
	query := `SELECT id, title, COALESCE(description, ''), updated_at
		 FROM calendar_events
		 WHERE created_by = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR description ILIKE $2)`
	args := []any{userID, pattern}
	idx := 3

	if from != nil {
		query += " AND start_at >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND start_at <= " + ph(idx)
		args = append(args, *to)
		idx++
	}

	query += " ORDER BY updated_at DESC"
	query += " LIMIT " + ph(idx)
	args = append(args, limit)
	idx++
	query += " OFFSET " + ph(idx)
	args = append(args, offset)

	return q.scanSearchRows(ctx, query, args)
}

// --- Tag-based search ---

func (q *Queries) SearchMemosByTag(ctx context.Context, userID uuid.UUID, pattern string, limit, offset int) ([]SearchRow, error) {
	return q.scanSearchRows(ctx,
		`SELECT DISTINCT m.id, m.title, COALESCE(m.content, ''), m.created_at AS updated_at
		 FROM memos m
		 JOIN memo_tags mt ON mt.memo_id = m.id
		 JOIN tags t ON t.id = mt.tag_id
		 WHERE m.user_id = $1 AND m.deleted_at IS NULL AND t.deleted_at IS NULL
		   AND t.name ILIKE $2
		 ORDER BY updated_at DESC LIMIT $3 OFFSET $4`,
		[]any{userID, pattern, limit, offset})
}

func (q *Queries) SearchTodosByTag(ctx context.Context, userID uuid.UUID, pattern string, limit, offset int) ([]SearchRow, error) {
	return q.scanSearchRows(ctx,
		`SELECT DISTINCT td.id, td.title, COALESCE(td.description, ''), td.updated_at
		 FROM todos td
		 JOIN todo_tags tt ON tt.todo_id = td.id
		 JOIN tags t ON t.id = tt.tag_id
		 WHERE td.user_id = $1 AND td.deleted_at IS NULL AND t.deleted_at IS NULL
		   AND t.name ILIKE $2
		 ORDER BY td.updated_at DESC LIMIT $3 OFFSET $4`,
		[]any{userID, pattern, limit, offset})
}

func (q *Queries) SearchEventsByTag(ctx context.Context, userID uuid.UUID, pattern string, limit, offset int) ([]SearchRow, error) {
	return q.scanSearchRows(ctx,
		`SELECT DISTINCT e.id, e.title, COALESCE(e.description, ''), e.updated_at
		 FROM calendar_events e
		 JOIN calendar_event_tags et ON et.event_id = e.id
		 JOIN tags t ON t.id = et.tag_id
		 WHERE e.created_by = $1 AND e.deleted_at IS NULL AND t.deleted_at IS NULL
		   AND t.name ILIKE $2
		 ORDER BY e.updated_at DESC LIMIT $3 OFFSET $4`,
		[]any{userID, pattern, limit, offset})
}

// --- Tool search ---

func (q *Queries) SearchTools(ctx context.Context, userID uuid.UUID, pattern string, limit, offset int) ([]ToolSearchRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, url, icon, created_at
		 FROM tools
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (name ILIKE $2 OR url ILIKE $2)
		 ORDER BY sort_order LIMIT $3 OFFSET $4`,
		userID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ToolSearchRow
	for rows.Next() {
		var r ToolSearchRow
		if err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Icon, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// --- Fuzzy search (similarity) ---

func (q *Queries) FuzzySearchMemos(ctx context.Context, userID uuid.UUID, query string, excludeIDs []uuid.UUID, limit int) ([]SuggestionRow, error) {
	return q.fuzzySearch(ctx,
		`SELECT id, title, similarity(title, $2) AS sim
		 FROM memos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND similarity(title, $2) > 0.3`,
		"memo", userID, query, excludeIDs, limit)
}

func (q *Queries) FuzzySearchTodos(ctx context.Context, userID uuid.UUID, query string, excludeIDs []uuid.UUID, limit int) ([]SuggestionRow, error) {
	return q.fuzzySearch(ctx,
		`SELECT id, title, similarity(title, $2) AS sim
		 FROM todos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND similarity(title, $2) > 0.3`,
		"todo", userID, query, excludeIDs, limit)
}

func (q *Queries) FuzzySearchEvents(ctx context.Context, userID uuid.UUID, query string, excludeIDs []uuid.UUID, limit int) ([]SuggestionRow, error) {
	return q.fuzzySearch(ctx,
		`SELECT id, title, similarity(title, $2) AS sim
		 FROM calendar_events
		 WHERE created_by = $1 AND deleted_at IS NULL
		   AND similarity(title, $2) > 0.3`,
		"event", userID, query, excludeIDs, limit)
}

func (q *Queries) fuzzySearch(ctx context.Context, baseQuery, typeName string, userID uuid.UUID, query string, excludeIDs []uuid.UUID, limit int) ([]SuggestionRow, error) {
	args := []any{userID, query}
	idx := 3

	if len(excludeIDs) > 0 {
		placeholders := make([]string, len(excludeIDs))
		for i, id := range excludeIDs {
			placeholders[i] = ph(idx)
			args = append(args, id)
			idx++
		}
		baseQuery += " AND id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}

	baseQuery += " ORDER BY sim DESC"
	baseQuery += " LIMIT " + ph(idx)
	args = append(args, limit)

	rows, err := q.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SuggestionRow
	for rows.Next() {
		var r SuggestionRow
		if err := rows.Scan(&r.ID, &r.Title, &r.Similarity); err != nil {
			return nil, err
		}
		r.Type = typeName
		results = append(results, r)
	}
	return results, rows.Err()
}

// --- Count queries ---

func (q *Queries) CountSearchMemos(ctx context.Context, userID uuid.UUID, pattern string, from, to *time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM memos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR content ILIKE $2)`
	args := []any{userID, pattern}
	idx := 3

	if from != nil {
		query += " AND created_at >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND created_at <= " + ph(idx)
		args = append(args, *to)
	}

	var count int
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (q *Queries) CountSearchTodos(ctx context.Context, userID uuid.UUID, pattern string, from, to *time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM todos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR description ILIKE $2)`
	args := []any{userID, pattern}
	idx := 3

	if from != nil {
		query += " AND due_date >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND due_date <= " + ph(idx)
		args = append(args, *to)
	}

	var count int
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (q *Queries) CountSearchEvents(ctx context.Context, userID uuid.UUID, pattern string, from, to *time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM calendar_events
		 WHERE created_by = $1 AND deleted_at IS NULL
		   AND (title ILIKE $2 OR description ILIKE $2)`
	args := []any{userID, pattern}
	idx := 3

	if from != nil {
		query += " AND start_at >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND start_at <= " + ph(idx)
		args = append(args, *to)
	}

	var count int
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (q *Queries) CountSearchTools(ctx context.Context, userID uuid.UUID, pattern string) (int, error) {
	var count int
	err := q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tools
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND (name ILIKE $2 OR url ILIKE $2)`,
		userID, pattern).Scan(&count)
	return count, err
}

// --- Tag filter queries ---

func (q *Queries) SearchMemosWithTagFilter(ctx context.Context, userID uuid.UUID, pattern, tagName string, from, to *time.Time, limit, offset int) ([]SearchRow, error) {
	query := `SELECT DISTINCT m.id, m.title, COALESCE(m.content, ''), m.created_at AS updated_at
		 FROM memos m
		 JOIN memo_tags mt ON mt.memo_id = m.id
		 JOIN tags t ON t.id = mt.tag_id
		 WHERE m.user_id = $1 AND m.deleted_at IS NULL AND t.deleted_at IS NULL
		   AND (m.title ILIKE $2 OR m.content ILIKE $2)
		   AND t.name = $3`
	args := []any{userID, pattern, tagName}
	idx := 4

	if from != nil {
		query += " AND m.created_at >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND m.created_at <= " + ph(idx)
		args = append(args, *to)
		idx++
	}

	query += " ORDER BY updated_at DESC"
	query += " LIMIT " + ph(idx)
	args = append(args, limit)
	idx++
	query += " OFFSET " + ph(idx)
	args = append(args, offset)

	return q.scanSearchRows(ctx, query, args)
}

func (q *Queries) SearchTodosWithTagFilter(ctx context.Context, userID uuid.UUID, pattern, tagName string, from, to *time.Time, limit, offset int) ([]SearchRow, error) {
	query := `SELECT DISTINCT td.id, td.title, COALESCE(td.description, ''), td.updated_at
		 FROM todos td
		 JOIN todo_tags tt ON tt.todo_id = td.id
		 JOIN tags t ON t.id = tt.tag_id
		 WHERE td.user_id = $1 AND td.deleted_at IS NULL AND t.deleted_at IS NULL
		   AND (td.title ILIKE $2 OR td.description ILIKE $2)
		   AND t.name = $3`
	args := []any{userID, pattern, tagName}
	idx := 4

	if from != nil {
		query += " AND td.due_date >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND td.due_date <= " + ph(idx)
		args = append(args, *to)
		idx++
	}

	query += " ORDER BY td.updated_at DESC"
	query += " LIMIT " + ph(idx)
	args = append(args, limit)
	idx++
	query += " OFFSET " + ph(idx)
	args = append(args, offset)

	return q.scanSearchRows(ctx, query, args)
}

func (q *Queries) SearchEventsWithTagFilter(ctx context.Context, userID uuid.UUID, pattern, tagName string, from, to *time.Time, limit, offset int) ([]SearchRow, error) {
	query := `SELECT DISTINCT e.id, e.title, COALESCE(e.description, ''), e.updated_at
		 FROM calendar_events e
		 JOIN calendar_event_tags et ON et.event_id = e.id
		 JOIN tags t ON t.id = et.tag_id
		 WHERE e.created_by = $1 AND e.deleted_at IS NULL AND t.deleted_at IS NULL
		   AND (e.title ILIKE $2 OR e.description ILIKE $2)
		   AND t.name = $3`
	args := []any{userID, pattern, tagName}
	idx := 4

	if from != nil {
		query += " AND e.start_at >= " + ph(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND e.start_at <= " + ph(idx)
		args = append(args, *to)
		idx++
	}

	query += " ORDER BY e.updated_at DESC"
	query += " LIMIT " + ph(idx)
	args = append(args, limit)
	idx++
	query += " OFFSET " + ph(idx)
	args = append(args, offset)

	return q.scanSearchRows(ctx, query, args)
}

// --- Helpers ---

func (q *Queries) scanSearchRows(ctx context.Context, query string, args []any) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchRow
	for rows.Next() {
		var r SearchRow
		if err := rows.Scan(&r.ID, &r.Title, &r.Content, &r.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func ph(idx int) string {
	return fmt.Sprintf("$%d", idx)
}
