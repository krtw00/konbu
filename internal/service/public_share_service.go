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

const (
	PublicShareMemo     = "memo"
	PublicShareTodo     = "todo"
	PublicShareCalendar = "calendar"
	PublicShareEvent    = "event"
	PublicShareTool     = "tool" // legacy read-only support
)

type PublicShareService struct {
	queries *repository.Queries
}

func NewPublicShareService(db *sql.DB) *PublicShareService {
	return &PublicShareService{queries: repository.New(db)}
}

func (s *PublicShareService) GetShare(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (*model.PublicShare, error) {
	if err := s.ensurePublishable(ctx, userID, resourceType, resourceID); err != nil {
		return nil, err
	}

	row, err := s.queries.GetPublicShareByResource(ctx, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, nil
		}
		return nil, apperror.Internal(err)
	}
	share := toModelPublicShare(row)
	return &share, nil
}

func (s *PublicShareService) CreateShare(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (*model.PublicShare, error) {
	if err := s.ensurePublishable(ctx, userID, resourceType, resourceID); err != nil {
		return nil, err
	}

	if existing, err := s.queries.GetPublicShareByResource(ctx, resourceType, resourceID); err == nil {
		share := toModelPublicShare(existing)
		return &share, nil
	} else if !errors.Is(err, repository.ErrNoRows) {
		return nil, apperror.Internal(err)
	}

	token, err := randomToken()
	if err != nil {
		return nil, apperror.Internal(err)
	}

	row, err := s.queries.UpsertPublicShare(ctx, resourceType, resourceID, userID, token)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	share := toModelPublicShare(row)
	return &share, nil
}

func (s *PublicShareService) DeleteShare(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	if err := s.ensurePublishable(ctx, userID, resourceType, resourceID); err != nil {
		return err
	}
	if err := s.queries.DeletePublicShare(ctx, resourceType, resourceID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *PublicShareService) GetPublicView(ctx context.Context, token string) (*model.PublicShareView, error) {
	row, err := s.queries.GetPublicShareByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("public share")
		}
		return nil, apperror.Internal(err)
	}

	view := &model.PublicShareView{
		Token:        row.Token,
		ResourceType: row.ResourceType,
	}

	switch row.ResourceType {
	case PublicShareMemo:
		memo, err := s.queries.GetMemoByIDPublic(ctx, row.ResourceID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, apperror.NotFound("memo")
			}
			return nil, apperror.Internal(err)
		}
		tags, err := s.queries.GetMemoTags(ctx, memo.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		modelTags := make([]model.Tag, len(tags))
		for i, t := range tags {
			modelTags[i] = model.Tag{ID: t.ID, Name: t.Name}
		}
		pubMemo := model.PublicMemoView{
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
			rows, err := s.queries.ListAllMemoRowsForExport(ctx, memo.ID)
			if err != nil {
				return nil, apperror.Internal(err)
			}
			pubMemo.Rows = make([]model.MemoRow, len(rows))
			for i, r := range rows {
				pubMemo.Rows[i] = model.MemoRow{
					ID:        r.ID,
					MemoID:    r.MemoID,
					RowData:   r.RowData,
					SortOrder: r.SortOrder,
					CreatedAt: r.CreatedAt,
				}
			}
		}
		view.Memo = &pubMemo
	case PublicShareTodo:
		todo, err := s.queries.GetTodoByIDPublic(ctx, row.ResourceID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, apperror.NotFound("todo")
			}
			return nil, apperror.Internal(err)
		}
		modelTodo := toModelTodo(todo)
		tags, err := s.queries.GetTodoTags(ctx, todo.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		modelTodo.Tags = make([]model.Tag, len(tags))
		for i, t := range tags {
			modelTodo.Tags[i] = model.Tag{ID: t.ID, Name: t.Name}
		}
		view.Todo = &modelTodo
	case PublicShareEvent:
		event, err := s.queries.GetEventByIDPublic(ctx, row.ResourceID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, apperror.NotFound("event")
			}
			return nil, apperror.Internal(err)
		}
		modelEvent := toModelEvent(event)
		tags, err := s.queries.GetEventTags(ctx, event.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		modelEvent.Tags = make([]model.Tag, len(tags))
		for i, t := range tags {
			modelEvent.Tags[i] = model.Tag{ID: t.ID, Name: t.Name}
		}
		view.Event = &modelEvent
	case PublicShareCalendar:
		cal, err := s.queries.GetCalendarByID(ctx, row.ResourceID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, apperror.NotFound("calendar")
			}
			return nil, apperror.Internal(err)
		}
		rows, err := s.queries.ListEventsByCalendarPublic(ctx, cal.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		events := make([]model.CalendarEvent, len(rows))
		for i, r := range rows {
			events[i] = toModelEvent(r)
		}
		view.Calendar = &model.PublicCalendarView{
			ID:        cal.ID,
			Name:      cal.Name,
			Color:     cal.Color,
			Events:    events,
			CreatedAt: cal.CreatedAt,
			UpdatedAt: cal.UpdatedAt,
		}
	case PublicShareTool:
		tool, err := s.queries.GetToolByIDPublic(ctx, row.ResourceID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, apperror.NotFound("tool")
			}
			return nil, apperror.Internal(err)
		}
		modelTool := toModelTool(tool)
		view.Tool = &modelTool
	default:
		return nil, apperror.BadRequest("unsupported public share type")
	}

	return view, nil
}

func (s *PublicShareService) ensurePublishable(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	switch resourceType {
	case PublicShareMemo:
		_, err := s.queries.GetMemoByID(ctx, resourceID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return apperror.NotFound("memo")
			}
			return apperror.Internal(err)
		}
	case PublicShareTodo:
		_, err := s.queries.GetTodoByID(ctx, resourceID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return apperror.NotFound("todo")
			}
			return apperror.Internal(err)
		}
	case PublicShareEvent:
		_, err := s.queries.GetEventByID(ctx, resourceID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return apperror.NotFound("event")
			}
			return apperror.Internal(err)
		}
	case PublicShareCalendar:
		member, err := s.queries.GetCalendarMember(ctx, resourceID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return apperror.NotFound("calendar")
			}
			return apperror.Internal(err)
		}
		if member.Role != "admin" {
			return apperror.Forbidden("admin role required")
		}
	default:
		return apperror.BadRequest("unsupported public share type")
	}
	return nil
}

func randomToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toModelPublicShare(row repository.PublicShareRow) model.PublicShare {
	return model.PublicShare{
		ResourceType: row.ResourceType,
		ResourceID:   row.ResourceID,
		Token:        row.Token,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
