package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MemoRowRow struct {
	ID        uuid.UUID
	MemoID    uuid.UUID
	RowData   json.RawMessage
	SortOrder int
	CreatedAt time.Time
}

func (q *Queries) CreateMemoRow(ctx context.Context, memoID uuid.UUID, rowData json.RawMessage, sortOrder int) (MemoRowRow, error) {
	var r MemoRowRow
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO memo_rows (id, memo_id, row_data, sort_order, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, now())
		 RETURNING id, memo_id, row_data, sort_order, created_at`,
		memoID, rowData, sortOrder).Scan(&r.ID, &r.MemoID, &r.RowData, &r.SortOrder, &r.CreatedAt)
	return r, err
}

func (q *Queries) ListMemoRows(ctx context.Context, memoID uuid.UUID, sortCol, sortOrder string, limit, offset int) ([]MemoRowRow, error) {
	orderClause := "sort_order ASC"
	if sortCol != "" {
		dir := "ASC"
		if sortOrder == "desc" {
			dir = "DESC"
		}
		orderClause = fmt.Sprintf(
			"CASE WHEN row_data->>%s ~ '^-?[0-9.]+$' THEN (row_data->>%s)::numeric ELSE NULL END %s NULLS LAST, (row_data->>%s)::text %s",
			quote(sortCol), quote(sortCol), dir, quote(sortCol), dir,
		)
	}

	query := fmt.Sprintf(
		`SELECT id, memo_id, row_data, sort_order, created_at
		 FROM memo_rows
		 WHERE memo_id = $1 AND deleted_at IS NULL
		 ORDER BY %s
		 LIMIT $2 OFFSET $3`, orderClause)

	rows, err := q.db.QueryContext(ctx, query, memoID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MemoRowRow
	for rows.Next() {
		var r MemoRowRow
		if err := rows.Scan(&r.ID, &r.MemoID, &r.RowData, &r.SortOrder, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (q *Queries) CountMemoRows(ctx context.Context, memoID uuid.UUID) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM memo_rows WHERE memo_id = $1 AND deleted_at IS NULL`,
		memoID).Scan(&count)
	return count, err
}

func (q *Queries) GetMemoRow(ctx context.Context, rowID, memoID uuid.UUID) (MemoRowRow, error) {
	var r MemoRowRow
	err := q.db.QueryRowContext(ctx,
		`SELECT id, memo_id, row_data, sort_order, created_at
		 FROM memo_rows
		 WHERE id = $1 AND memo_id = $2 AND deleted_at IS NULL`,
		rowID, memoID).Scan(&r.ID, &r.MemoID, &r.RowData, &r.SortOrder, &r.CreatedAt)
	return r, err
}

func (q *Queries) UpdateMemoRow(ctx context.Context, rowID, memoID uuid.UUID, rowData json.RawMessage) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE memo_rows SET row_data = $3 WHERE id = $1 AND memo_id = $2 AND deleted_at IS NULL`,
		rowID, memoID, rowData)
	return err
}

func (q *Queries) SoftDeleteMemoRow(ctx context.Context, rowID, memoID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE memo_rows SET deleted_at = now() WHERE id = $1 AND memo_id = $2`,
		rowID, memoID)
	return err
}

func (q *Queries) MaxMemoRowSortOrder(ctx context.Context, memoID uuid.UUID) (int, error) {
	var maxOrder int
	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order), -1) FROM memo_rows WHERE memo_id = $1 AND deleted_at IS NULL`,
		memoID).Scan(&maxOrder)
	return maxOrder, err
}

func (q *Queries) BatchCreateMemoRows(ctx context.Context, memoID uuid.UUID, rows []json.RawMessage, startOrder int) ([]MemoRowRow, error) {
	var results []MemoRowRow
	for i, data := range rows {
		r, err := q.CreateMemoRow(ctx, memoID, data, startOrder+i)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (q *Queries) ListAllMemoRowsForExport(ctx context.Context, memoID uuid.UUID) ([]MemoRowRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, memo_id, row_data, sort_order, created_at
		 FROM memo_rows
		 WHERE memo_id = $1 AND deleted_at IS NULL
		 ORDER BY sort_order ASC`,
		memoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MemoRowRow
	for rows.Next() {
		var r MemoRowRow
		if err := rows.Scan(&r.ID, &r.MemoID, &r.RowData, &r.SortOrder, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func quote(s string) string {
	return "'" + s + "'"
}
