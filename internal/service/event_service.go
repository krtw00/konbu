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

type EventService struct {
	queries *repository.Queries
	db      *sql.DB
	tagSvc  *TagService
}

func NewEventService(db *sql.DB, tagSvc *TagService) *EventService {
	return &EventService{
		queries: repository.New(db),
		db:      db,
		tagSvc:  tagSvc,
	}
}

func (s *EventService) ListEvents(ctx context.Context, userID uuid.UUID, params model.ListParams) (*model.PaginatedResult, error) {
	rows, err := s.queries.ListEventsByUserID(ctx, userID, params.Limit, params.Offset)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	total, err := s.queries.CountEventsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	items := make([]model.CalendarEvent, len(rows))
	for i, r := range rows {
		items[i] = toModelEvent(r)
	}
	return &model.PaginatedResult{Data: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

func (s *EventService) GetEvent(ctx context.Context, id, userID uuid.UUID) (*model.CalendarEvent, error) {
	r, err := s.queries.GetEventByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("event")
		}
		return nil, apperror.Internal(err)
	}
	event := toModelEvent(r)

	tags, err := s.queries.GetEventTags(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	event.Tags = make([]model.Tag, len(tags))
	for i, t := range tags {
		event.Tags[i] = model.Tag{ID: t.ID, Name: t.Name}
	}
	return &event, nil
}

func (s *EventService) CreateEvent(ctx context.Context, userID uuid.UUID, req model.CreateEventRequest) (*model.CalendarEvent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	r, err := q.CreateEvent(ctx, userID, req.Title, req.Description, req.StartAt, req.EndAt, req.AllDay, req.RecurrenceRule, req.RecurrenceEnd)
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
			if err := q.AddEventTag(ctx, r.ID, t.ID); err != nil {
				return nil, apperror.Internal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	event := toModelEvent(r)
	event.Tags = modelTags
	return &event, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, id, userID uuid.UUID, req model.UpdateEventRequest) (*model.CalendarEvent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	r, err := q.UpdateEvent(ctx, id, userID, req.Title, req.Description, req.StartAt, req.EndAt, req.AllDay, req.RecurrenceRule, req.RecurrenceEnd)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("event")
		}
		return nil, apperror.Internal(err)
	}

	var modelTags []model.Tag
	if req.Tags != nil {
		if err := q.ClearEventTags(ctx, id); err != nil {
			return nil, apperror.Internal(err)
		}
		modelTags, err = s.tagSvc.EnsureTags(ctx, userID, req.Tags)
		if err != nil {
			return nil, err
		}
		for _, t := range modelTags {
			if err := q.AddEventTag(ctx, id, t.ID); err != nil {
				return nil, apperror.Internal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	event := toModelEvent(r)
	event.Tags = modelTags
	return &event, nil
}

func (s *EventService) ListAllEvents(ctx context.Context, userID uuid.UUID) ([]model.CalendarEvent, error) {
	rows, err := s.queries.ListAllEventsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	items := make([]model.CalendarEvent, len(rows))
	for i, r := range rows {
		items[i] = toModelEvent(r)
	}
	return items, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.SoftDeleteEvent(ctx, id, userID)
}

func toModelEvent(r repository.EventRow) model.CalendarEvent {
	return model.CalendarEvent{
		ID:             r.ID,
		Title:          r.Title,
		Description:    r.Description,
		StartAt:        r.StartAt,
		EndAt:          r.EndAt,
		AllDay:         r.AllDay,
		RecurrenceRule: r.RecurrenceRule,
		RecurrenceEnd:  r.RecurrenceEnd,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
