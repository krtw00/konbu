package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type TagService struct {
	queries *repository.Queries
	db      *sql.DB
}

func NewTagService(db *sql.DB) *TagService {
	return &TagService{
		queries: repository.New(db),
		db:      db,
	}
}

func (s *TagService) EnsureTags(ctx context.Context, userID uuid.UUID, names []string) ([]model.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	tags := make([]model.Tag, 0, len(names))
	for _, name := range names {
		t, err := s.queries.CreateTag(ctx, userID, name)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		tags = append(tags, model.Tag{ID: t.ID, Name: t.Name})
	}
	return tags, nil
}

func (s *TagService) ListTags(ctx context.Context, userID uuid.UUID) ([]model.Tag, error) {
	rows, err := s.queries.ListTagsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	tags := make([]model.Tag, len(rows))
	for i, r := range rows {
		tags[i] = model.Tag{ID: r.ID, Name: r.Name}
	}
	return tags, nil
}

func (s *TagService) CreateTag(ctx context.Context, userID uuid.UUID, req model.CreateTagRequest) (*model.Tag, error) {
	t, err := s.queries.CreateTag(ctx, userID, req.Name)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &model.Tag{ID: t.ID, Name: t.Name}, nil
}

func (s *TagService) UpdateTag(ctx context.Context, id uuid.UUID, req model.UpdateTagRequest) (*model.Tag, error) {
	t, err := s.queries.UpdateTag(ctx, id, req.Name)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &model.Tag{ID: t.ID, Name: t.Name}, nil
}

func (s *TagService) DeleteTag(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.SoftDeleteTag(ctx, id, userID)
}
