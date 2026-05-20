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

type CalendarService struct {
	queries *repository.Queries
	db      *sql.DB
}

func NewCalendarService(db *sql.DB) *CalendarService {
	return &CalendarService{
		queries: repository.New(db),
		db:      db,
	}
}

func (s *CalendarService) ListCalendars(ctx context.Context, userID uuid.UUID) ([]model.Calendar, error) {
	rows, err := s.queries.ListCalendarsByUser(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	calendars := make([]model.Calendar, len(rows))
	for i, r := range rows {
		calendars[i] = toModelCalendar(r)
	}
	return calendars, nil
}

func (s *CalendarService) GetCalendar(ctx context.Context, userID, calendarID uuid.UUID) (*model.CalendarDetail, error) {
	r, err := s.queries.GetCalendarByID(ctx, calendarID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}
	if r.OwnerID != userID {
		return nil, apperror.NotFound("calendar")
	}

	cal := toModelCalendar(r)
	detail := model.CalendarDetail{Calendar: cal}
	return &detail, nil
}

func (s *CalendarService) CreateCalendar(ctx context.Context, userID uuid.UUID, req model.CreateCalendarRequest) (*model.Calendar, error) {
	if req.Name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	color := req.Color
	if color == "" {
		color = "#4F46E5"
	}

	r, err := s.queries.CreateCalendar(ctx, userID, req.Name, color, false)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	cal := toModelCalendar(r)
	return &cal, nil
}

func (s *CalendarService) UpdateCalendar(ctx context.Context, userID, calendarID uuid.UUID, req model.UpdateCalendarRequest) (*model.Calendar, error) {
	existing, err := s.queries.GetCalendarByID(ctx, calendarID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}
	if existing.OwnerID != userID {
		return nil, apperror.NotFound("calendar")
	}

	color := req.Color
	if color == "" {
		color = "#4F46E5"
	}

	r, err := s.queries.UpdateCalendar(ctx, calendarID, req.Name, color)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}
	cal := toModelCalendar(r)
	return &cal, nil
}

func (s *CalendarService) DeleteCalendar(ctx context.Context, userID, calendarID uuid.UUID) error {
	r, err := s.queries.GetCalendarByID(ctx, calendarID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("calendar")
		}
		return apperror.Internal(err)
	}
	if r.OwnerID != userID {
		return apperror.NotFound("calendar")
	}
	if r.IsDefault {
		return apperror.BadRequest("cannot delete default calendar")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	if err := q.SoftDeleteEventsByCalendar(ctx, calendarID); err != nil {
		return apperror.Internal(err)
	}
	if err := q.SoftDeleteCalendar(ctx, calendarID); err != nil {
		return apperror.Internal(err)
	}

	if err := tx.Commit(); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *CalendarService) EnsureDefaultCalendar(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	r, err := s.queries.GetDefaultCalendar(ctx, userID)
	if err == nil {
		return r.ID, nil
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return uuid.Nil, apperror.Internal(err)
	}

	r, err = s.queries.CreateCalendar(ctx, userID, "My Calendar", "#4F46E5", true)
	if err != nil {
		return uuid.Nil, apperror.Internal(err)
	}

	return r.ID, nil
}

func toModelCalendar(r repository.CalendarRow) model.Calendar {
	return model.Calendar{
		ID:        r.ID,
		OwnerID:   r.OwnerID,
		Name:      r.Name,
		IsDefault: r.IsDefault,
		Color:     r.Color,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
