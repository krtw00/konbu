package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tool struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Name          string
	URL           string
	Icon          string
	IconCheckedAt *time.Time
	Category      string
	SortOrder     int
	CreatedAt     time.Time
}

var toolCols = `id, user_id, name, url, icon, icon_checked_at, COALESCE(category, ''), sort_order, created_at`

func scanTool(row interface{ Scan(...any) error }) (Tool, error) {
	var t Tool
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.URL, &t.Icon, &t.IconCheckedAt, &t.Category, &t.SortOrder, &t.CreatedAt)
	return t, err
}

func (q *Queries) CreateTool(ctx context.Context, userID uuid.UUID, name, url, icon string, iconCheckedAt *time.Time, category string, sortOrder int) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO tools (user_id, name, url, icon, icon_checked_at, category, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING `+toolCols,
		userID, name, url, icon, iconCheckedAt, category, sortOrder)
	return scanTool(row)
}

func (q *Queries) GetToolByID(ctx context.Context, id, userID uuid.UUID) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT `+toolCols+` FROM tools WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID)
	return scanTool(row)
}

func (q *Queries) GetToolByIDPublic(ctx context.Context, id uuid.UUID) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT `+toolCols+` FROM tools WHERE id = $1 AND deleted_at IS NULL`,
		id)
	return scanTool(row)
}

func (q *Queries) ListToolsByUserID(ctx context.Context, userID uuid.UUID) ([]Tool, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+toolCols+` FROM tools WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY COALESCE(category, ''), sort_order, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		t, err := scanTool(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (q *Queries) UpdateTool(ctx context.Context, id, userID uuid.UUID, name, url, icon string, iconCheckedAt *time.Time, category string, sortOrder int) (Tool, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE tools SET name = $3, url = $4, icon = $5, icon_checked_at = $6, category = $7, sort_order = $8
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING `+toolCols,
		id, userID, name, url, icon, iconCheckedAt, category, sortOrder)
	return scanTool(row)
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

func (q *Queries) ListToolsWithEmptyIcon(ctx context.Context) ([]Tool, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+toolCols+` FROM tools WHERE icon = '' AND url != '' AND deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		t, err := scanTool(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (q *Queries) ListToolsNeedingIconRefresh(ctx context.Context, cutoff time.Time) ([]Tool, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+toolCols+` FROM tools
		 WHERE url != '' AND deleted_at IS NULL
		   AND (icon_checked_at IS NULL OR icon_checked_at <= $1)`,
		cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		t, err := scanTool(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (q *Queries) MaxToolSortOrder(ctx context.Context, userID uuid.UUID) (int, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order), 0) FROM tools WHERE user_id = $1 AND deleted_at IS NULL`,
		userID)
	var max int
	err := row.Scan(&max)
	return max, err
}
