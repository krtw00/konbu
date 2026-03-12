package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tool struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	URL       string
	Icon      string
	SortOrder int
	CreatedAt time.Time
}

func (q *Queries) CreateTool(ctx context.Context, userID uuid.UUID, name, url, icon string, sortOrder int) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO tools (user_id, name, url, icon, sort_order)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, url, icon, sort_order, created_at`,
		userID, name, url, icon, sortOrder)
	var t Tool
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.URL, &t.Icon, &t.SortOrder, &t.CreatedAt)
	return t, err
}

func (q *Queries) GetToolByID(ctx context.Context, id, userID uuid.UUID) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, url, icon, sort_order, created_at
		 FROM tools WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID)
	var t Tool
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.URL, &t.Icon, &t.SortOrder, &t.CreatedAt)
	return t, err
}

func (q *Queries) ListToolsByUserID(ctx context.Context, userID uuid.UUID) ([]Tool, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, name, url, icon, sort_order, created_at
		 FROM tools WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY sort_order, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.URL, &t.Icon, &t.SortOrder, &t.CreatedAt); err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (q *Queries) UpdateTool(ctx context.Context, id, userID uuid.UUID, name, url, icon string, sortOrder int) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE tools SET name = $3, url = $4, icon = $5, sort_order = $6
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING id, user_id, name, url, icon, sort_order, created_at`,
		id, userID, name, url, icon, sortOrder)
	var t Tool
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.URL, &t.Icon, &t.SortOrder, &t.CreatedAt)
	return t, err
}

func (q *Queries) SoftDeleteTool(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tools SET deleted_at = now() WHERE id = $1 AND user_id = $2`,
		id, userID)
	return err
}

func (q *Queries) UpdateToolSortOrder(ctx context.Context, id, userID uuid.UUID, sortOrder int) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tools SET sort_order = $3 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID, sortOrder)
	return err
}

func (q *Queries) MaxToolSortOrder(ctx context.Context, userID uuid.UUID) (int, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order), 0) FROM tools WHERE user_id = $1 AND deleted_at IS NULL`,
		userID)
	var max int
	err := row.Scan(&max)
	return max, err
}
