package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type MemoRowService struct {
	queries *repository.Queries
	db      *sql.DB
}

func NewMemoRowService(db *sql.DB) *MemoRowService {
	return &MemoRowService{
		queries: repository.New(db),
		db:      db,
	}
}

func (s *MemoRowService) ListRows(ctx context.Context, userID, memoID uuid.UUID, sortCol, sortOrder string, limit, offset int) (*model.PaginatedResult, error) {
	if err := s.checkMemoOwner(ctx, userID, memoID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	rows, err := s.queries.ListMemoRows(ctx, memoID, sortCol, sortOrder, limit, offset)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	total, err := s.queries.CountMemoRows(ctx, memoID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	data := make([]model.MemoRow, 0, len(rows))
	for _, r := range rows {
		data = append(data, toModelRow(r))
	}

	return &model.PaginatedResult{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *MemoRowService) CreateRow(ctx context.Context, userID, memoID uuid.UUID, req model.CreateMemoRowRequest) (*model.MemoRow, error) {
	if err := s.checkMemoOwner(ctx, userID, memoID); err != nil {
		return nil, err
	}

	maxOrder, err := s.queries.MaxMemoRowSortOrder(ctx, memoID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	r, err := s.queries.CreateMemoRow(ctx, memoID, req.RowData, maxOrder+1)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	row := toModelRow(r)
	return &row, nil
}

func (s *MemoRowService) UpdateRow(ctx context.Context, userID, memoID, rowID uuid.UUID, req model.UpdateMemoRowRequest) error {
	if err := s.checkMemoOwner(ctx, userID, memoID); err != nil {
		return err
	}
	if err := s.queries.UpdateMemoRow(ctx, rowID, memoID, req.RowData); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *MemoRowService) DeleteRow(ctx context.Context, userID, memoID, rowID uuid.UUID) error {
	if err := s.checkMemoOwner(ctx, userID, memoID); err != nil {
		return err
	}
	return s.queries.SoftDeleteMemoRow(ctx, rowID, memoID)
}

func (s *MemoRowService) BatchCreateRows(ctx context.Context, userID, memoID uuid.UUID, req model.BatchCreateMemoRowsRequest) ([]model.MemoRow, error) {
	if err := s.checkMemoOwner(ctx, userID, memoID); err != nil {
		return nil, err
	}
	if len(req.Rows) > 500 {
		return nil, apperror.BadRequest("batch limit is 500 rows")
	}

	maxOrder, err := s.queries.MaxMemoRowSortOrder(ctx, memoID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	rows, err := s.queries.BatchCreateMemoRows(ctx, memoID, req.Rows, maxOrder+1)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	result := make([]model.MemoRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, toModelRow(r))
	}
	return result, nil
}

func (s *MemoRowService) ExportCSV(ctx context.Context, userID, memoID uuid.UUID, w io.Writer) error {
	if err := s.checkMemoOwner(ctx, userID, memoID); err != nil {
		return err
	}

	memo, err := s.queries.GetMemoByID(ctx, memoID, userID)
	if err != nil {
		return apperror.NotFound("memo not found")
	}

	var columns []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if memo.TableColumns != nil {
		if err := json.Unmarshal(*memo.TableColumns, &columns); err != nil {
			return apperror.Internal(fmt.Errorf("invalid table_columns: %w", err))
		}
	}

	rows, err := s.queries.ListAllMemoRowsForExport(ctx, memoID)
	if err != nil {
		return apperror.Internal(err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	header := make([]string, len(columns))
	for i, col := range columns {
		header[i] = col.Name
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Data rows
	for _, row := range rows {
		var data map[string]string
		if err := json.Unmarshal(row.RowData, &data); err != nil {
			continue
		}
		record := make([]string, len(columns))
		for i, col := range columns {
			record[i] = data[col.ID]
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func (s *MemoRowService) checkMemoOwner(ctx context.Context, userID, memoID uuid.UUID) error {
	_, err := s.queries.GetMemoByID(ctx, memoID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return apperror.NotFound("memo not found")
		}
		return apperror.Internal(err)
	}
	return nil
}

func toModelRow(r repository.MemoRowRow) model.MemoRow {
	return model.MemoRow{
		ID:        r.ID,
		MemoID:    r.MemoID,
		RowData:   r.RowData,
		SortOrder: r.SortOrder,
		CreatedAt: r.CreatedAt,
	}
}
