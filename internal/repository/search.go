package repository

import (
	"context"
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

func (q *Queries) SearchEvents(ctx context.Context, userID uuid.UUID, pattern string, limit int) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, title, COALESCE(description, ''), updated_at
		 FROM calendar_events
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
