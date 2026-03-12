package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type MemoService struct {
	queries *repository.Queries
	db      *sql.DB
	tagSvc  *TagService
}

func NewMemoService(db *sql.DB, tagSvc *TagService) *MemoService {
	return &MemoService{
		queries: repository.New(db),
		db:      db,
		tagSvc:  tagSvc,
	}
}

func (s *MemoService) ListMemos(ctx context.Context, userID uuid.UUID, params model.ListParams) (*model.PaginatedResult, error) {
	memos, err := s.queries.ListMemosByUserID(ctx, userID, params.Limit, params.Offset)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	total, err := s.queries.CountMemosByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	items := make([]model.Memo, len(memos))
	for i, m := range memos {
		items[i] = model.Memo{
			ID:        m.ID,
			Title:     m.Title,
			Type:      m.Type,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}

	return &model.PaginatedResult{
		Data:   items,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (s *MemoService) GetMemo(ctx context.Context, id, userID uuid.UUID) (*model.Memo, error) {
	m, err := s.queries.GetMemoByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("memo")
		}
		return nil, apperror.Internal(err)
	}

	tags, _ := s.queries.GetMemoTags(ctx, id)
	modelTags := make([]model.Tag, len(tags))
	for i, t := range tags {
		modelTags[i] = model.Tag{ID: t.ID, Name: t.Name}
	}

	return &model.Memo{
		ID:           m.ID,
		Title:        m.Title,
		Type:         m.Type,
		Content:      m.Content,
		TableColumns: m.TableColumns,
		Tags:         modelTags,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func (s *MemoService) CreateMemo(ctx context.Context, userID uuid.UUID, req model.CreateMemoRequest) (*model.Memo, error) {
	if req.Type == "" {
		req.Type = "markdown"
	}
	if req.Type != "markdown" && req.Type != "table" {
		return nil, apperror.BadRequest("type must be 'markdown' or 'table'")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	m, err := q.CreateMemo(ctx, userID, req.Title, req.Type, req.Content, req.TableColumns)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	var modelTags []model.Tag
	if len(req.Tags) > 0 {
		modelTags, err = s.tagSvc.EnsureTags(ctx, userID, req.Tags)
		if err != nil {
			return nil, err
		}
		for _, t := range modelTags {
			if err := q.AddMemoTag(ctx, m.ID, t.ID); err != nil {
				return nil, apperror.Internal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	return &model.Memo{
		ID:           m.ID,
		Title:        m.Title,
		Type:         m.Type,
		Content:      m.Content,
		TableColumns: m.TableColumns,
		Tags:         modelTags,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func (s *MemoService) UpdateMemo(ctx context.Context, id, userID uuid.UUID, req model.UpdateMemoRequest) (*model.Memo, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	m, err := q.UpdateMemo(ctx, id, userID, req.Title, req.Content, req.TableColumns)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("memo")
		}
		return nil, apperror.Internal(err)
	}

	var modelTags []model.Tag
	if req.Tags != nil {
		if err := q.ClearMemoTags(ctx, id); err != nil {
			return nil, apperror.Internal(err)
		}
		modelTags, err = s.tagSvc.EnsureTags(ctx, userID, req.Tags)
		if err != nil {
			return nil, err
		}
		for _, t := range modelTags {
			if err := q.AddMemoTag(ctx, id, t.ID); err != nil {
				return nil, apperror.Internal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	return &model.Memo{
		ID:           m.ID,
		Title:        m.Title,
		Type:         m.Type,
		Content:      m.Content,
		TableColumns: m.TableColumns,
		Tags:         modelTags,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func (s *MemoService) DeleteMemo(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.SoftDeleteMemo(ctx, id, userID)
}
