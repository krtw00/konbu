package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Tag struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	CreatedAt time.Time
}

func (q *Queries) CreateTag(ctx context.Context, userID uuid.UUID, name string) (Tag, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO tags (user_id, name)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, name) DO UPDATE SET deleted_at = NULL
		 RETURNING id, user_id, name, created_at`,
		userID, name)
	var t Tag
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt)
	return t, err
}

func (q *Queries) GetTagByID(ctx context.Context, id, userID uuid.UUID) (Tag, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, created_at
		 FROM tags WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID)
	var t Tag
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt)
	return t, err
}

func (q *Queries) ListTagsByUserID(ctx context.Context, userID uuid.UUID) ([]Tag, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, name, created_at
		 FROM tags WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (q *Queries) UpdateTag(ctx context.Context, id, userID uuid.UUID, name string) (Tag, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE tags SET name = $3
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING id, user_id, name, created_at`,
		id, userID, name)
	var t Tag
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt)
	return t, err
}

func (q *Queries) SoftDeleteTag(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tags SET deleted_at = now() WHERE id = $1 AND user_id = $2`,
		id, userID)
	return err
}

func (q *Queries) GetTagsByNames(ctx context.Context, userID uuid.UUID, names []string) ([]Tag, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, name, created_at
		 FROM tags WHERE user_id = $1 AND name = ANY($2) AND deleted_at IS NULL`,
		userID, pq.Array(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
