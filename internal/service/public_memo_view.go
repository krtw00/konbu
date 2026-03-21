package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

func loadPublicMemoView(ctx context.Context, queries *repository.Queries, memoID uuid.UUID) (*model.PublicMemoView, error) {
	memo, err := queries.GetMemoByIDPublic(ctx, memoID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("memo")
		}
		return nil, apperror.Internal(err)
	}

	tags, err := queries.GetMemoTags(ctx, memo.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	modelTags := make([]model.Tag, len(tags))
	for i, tag := range tags {
		modelTags[i] = model.Tag{ID: tag.ID, Name: tag.Name}
	}

	pubMemo := &model.PublicMemoView{
		Memo: model.Memo{
			ID:           memo.ID,
			Title:        memo.Title,
			Type:         memo.Type,
			Content:      memo.Content,
			TableColumns: memo.TableColumns,
			Tags:         modelTags,
			CreatedAt:    memo.CreatedAt,
			UpdatedAt:    memo.UpdatedAt,
		},
	}

	if memo.Type == "table" {
		rows, err := queries.ListAllMemoRowsForExport(ctx, memo.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		pubMemo.Rows = make([]model.MemoRow, len(rows))
		for i, row := range rows {
			pubMemo.Rows[i] = model.MemoRow{
				ID:        row.ID,
				MemoID:    row.MemoID,
				RowData:   row.RowData,
				SortOrder: row.SortOrder,
				CreatedAt: row.CreatedAt,
			}
		}
	}

	return pubMemo, nil
}
