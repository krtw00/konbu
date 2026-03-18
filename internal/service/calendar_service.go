package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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
	_, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}

	r, err := s.queries.GetCalendarByID(ctx, calendarID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}

	members, err := s.queries.ListCalendarMembers(ctx, calendarID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	cal := toModelCalendar(r)
	detail := model.CalendarDetail{Calendar: cal}
	detail.Members = make([]model.CalendarMember, len(members))
	for i, m := range members {
		detail.Members[i] = toModelCalendarMember(m)
	}
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	r, err := q.CreateCalendar(ctx, userID, req.Name, color, false)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	if err := q.AddCalendarMember(ctx, r.ID, userID, "admin", color); err != nil {
		return nil, apperror.Internal(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	r.MemberCount = 1
	cal := toModelCalendar(r)
	return &cal, nil
}

func (s *CalendarService) UpdateCalendar(ctx context.Context, userID, calendarID uuid.UUID, req model.UpdateCalendarRequest) (*model.Calendar, error) {
	member, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}
	if member.Role != "admin" {
		return nil, apperror.Forbidden("admin role required")
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
		return apperror.Forbidden("only owner can delete calendar")
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

func (s *CalendarService) CreateShareLink(ctx context.Context, userID, calendarID uuid.UUID) (string, error) {
	member, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return "", apperror.NotFound("calendar")
		}
		return "", apperror.Internal(err)
	}
	if member.Role != "admin" {
		return "", apperror.Forbidden("admin role required")
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", apperror.Internal(err)
	}
	token := hex.EncodeToString(b)

	if err := s.queries.SetShareToken(ctx, calendarID, token); err != nil {
		return "", apperror.Internal(err)
	}
	return token, nil
}

func (s *CalendarService) DeleteShareLink(ctx context.Context, userID, calendarID uuid.UUID) error {
	member, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("calendar")
		}
		return apperror.Internal(err)
	}
	if member.Role != "admin" {
		return apperror.Forbidden("admin role required")
	}
	return s.queries.ClearShareToken(ctx, calendarID)
}

func (s *CalendarService) JoinByToken(ctx context.Context, userID uuid.UUID, token string) (*model.Calendar, error) {
	r, err := s.queries.GetCalendarByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}

	if err := s.queries.AddCalendarMember(ctx, r.ID, userID, "editor", r.Color); err != nil {
		return nil, apperror.Internal(err)
	}

	cal := toModelCalendar(r)
	return &cal, nil
}

func (s *CalendarService) AddMember(ctx context.Context, userID, calendarID uuid.UUID, req model.AddMemberRequest) (*model.CalendarMember, error) {
	member, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("calendar")
		}
		return nil, apperror.Internal(err)
	}
	if member.Role != "admin" {
		return nil, apperror.Forbidden("admin role required")
	}

	targetUser, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("user")
		}
		return nil, apperror.Internal(err)
	}

	role := req.Role
	if role == "" {
		role = "editor"
	}

	cal, err := s.queries.GetCalendarByID(ctx, calendarID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	if err := s.queries.AddCalendarMember(ctx, calendarID, targetUser.ID, role, cal.Color); err != nil {
		return nil, apperror.Internal(err)
	}

	m, err := s.queries.GetCalendarMember(ctx, calendarID, targetUser.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	result := toModelCalendarMember(m)
	return &result, nil
}

func (s *CalendarService) UpdateMember(ctx context.Context, userID, calendarID, targetUserID uuid.UUID, req model.UpdateMemberRequest) error {
	member, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("calendar")
		}
		return apperror.Internal(err)
	}
	if member.Role != "admin" {
		return apperror.Forbidden("admin role required")
	}

	_, err = s.queries.GetCalendarMember(ctx, calendarID, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("member")
		}
		return apperror.Internal(err)
	}

	return s.queries.UpdateCalendarMember(ctx, calendarID, targetUserID, req.Role, req.Color)
}

func (s *CalendarService) RemoveMember(ctx context.Context, userID, calendarID, targetUserID uuid.UUID) error {
	member, err := s.queries.GetCalendarMember(ctx, calendarID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("calendar")
		}
		return apperror.Internal(err)
	}
	if member.Role != "admin" && userID != targetUserID {
		return apperror.Forbidden("admin role required")
	}

	cal, err := s.queries.GetCalendarByID(ctx, calendarID)
	if err != nil {
		return apperror.Internal(err)
	}
	if cal.OwnerID == targetUserID {
		return apperror.BadRequest("cannot remove calendar owner")
	}

	return s.queries.RemoveCalendarMember(ctx, calendarID, targetUserID)
}

func (s *CalendarService) EnsureDefaultCalendar(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	r, err := s.queries.GetDefaultCalendar(ctx, userID)
	if err == nil {
		return r.ID, nil
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return uuid.Nil, apperror.Internal(err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	r, err = q.CreateCalendar(ctx, userID, "My Calendar", "#4F46E5", true)
	if err != nil {
		return uuid.Nil, apperror.Internal(err)
	}

	if err := q.AddCalendarMember(ctx, r.ID, userID, "admin", "#4F46E5"); err != nil {
		return uuid.Nil, apperror.Internal(err)
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, apperror.Internal(err)
	}

	return r.ID, nil
}

func toModelCalendar(r repository.CalendarRow) model.Calendar {
	return model.Calendar{
		ID:          r.ID,
		OwnerID:     r.OwnerID,
		Name:        r.Name,
		IsDefault:   r.IsDefault,
		ShareToken:  r.ShareToken,
		Color:       r.Color,
		MemberCount: r.MemberCount,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toModelCalendarMember(r repository.CalendarMemberRow) model.CalendarMember {
	return model.CalendarMember{
		CalendarID: r.CalendarID,
		UserID:     r.UserID,
		UserName:   r.UserName,
		UserEmail:  r.UserEmail,
		Role:       r.Role,
		Color:      r.Color,
		JoinedAt:   r.JoinedAt,
	}
}
