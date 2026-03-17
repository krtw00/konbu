package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Memo struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Type         string
	Content      *string
	TableColumns *json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MemoTag struct {
	ID   uuid.UUID
	Name string
}

func (q *Queries) CreateMemo(ctx context.Context, userID uuid.UUID, title, memoType string, content *string, tableCols *json.RawMessage) (Memo, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO memos (user_id, title, type, content, table_columns)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, title, type, content, table_columns, created_at, updated_at`,
		userID, title, memoType, content, tableCols)
	var m Memo
	err := row.Scan(&m.ID, &m.UserID, &m.Title, &m.Type, &m.Content, &m.TableColumns, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (q *Queries) GetMemoByID(ctx context.Context, id, userID uuid.UUID) (Memo, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, type, content, table_columns, created_at, updated_at
		 FROM memos WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID)
	var m Memo
	err := row.Scan(&m.ID, &m.UserID, &m.Title, &m.Type, &m.Content, &m.TableColumns, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (q *Queries) GetMemoByIDPublic(ctx context.Context, id uuid.UUID) (Memo, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, type, content, table_columns, created_at, updated_at
		 FROM memos WHERE id = $1 AND deleted_at IS NULL`,
		id)
	var m Memo
	err := row.Scan(&m.ID, &m.UserID, &m.Title, &m.Type, &m.Content, &m.TableColumns, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (q *Queries) ListMemosByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Memo, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, title, type, created_at, updated_at
		 FROM memos WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memos []Memo
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.ID, &m.UserID, &m.Title, &m.Type, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		memos = append(memos, m)
	}
	return memos, rows.Err()
}

func (q *Queries) CountMemosByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT count(*) FROM memos WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) UpdateMemo(ctx context.Context, id, userID uuid.UUID, title string, content *string, tableCols *json.RawMessage) (Memo, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE memos SET title = $3, content = $4, table_columns = $5, updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING id, user_id, title, type, content, table_columns, created_at, updated_at`,
		id, userID, title, content, tableCols)
	var m Memo
	err := row.Scan(&m.ID, &m.UserID, &m.Title, &m.Type, &m.Content, &m.TableColumns, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (q *Queries) SoftDeleteMemo(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE memos SET deleted_at = now() WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (q *Queries) ClearMemoTags(ctx context.Context, memoID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM memo_tags WHERE memo_id = $1`, memoID)
	return err
}

func (q *Queries) AddMemoTag(ctx context.Context, memoID, tagID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO memo_tags (memo_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		memoID, tagID)
	return err
}

func (q *Queries) GetMemoTags(ctx context.Context, memoID uuid.UUID) ([]MemoTag, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT t.id, t.name FROM tags t
		 JOIN memo_tags mt ON mt.tag_id = t.id
		 WHERE mt.memo_id = $1 AND t.deleted_at IS NULL`, memoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []MemoTag
	for rows.Next() {
		var t MemoTag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
