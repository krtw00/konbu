package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type DailyService struct {
	queries *repository.Queries
}

func NewDailyService(db *sql.DB) *DailyService {
	return &DailyService{queries: repository.New(db)}
}

// GetDaily returns the events, todos and memos that fall within the half-open
// range [from, to). Events are matched on start_at, todos on due_date and memos
// on created_at.
func (s *DailyService) GetDaily(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*model.CalendarEvent, []*model.Todo, []*model.Memo, error) {
	eventRows, err := s.queries.ListEventsByRange(ctx, userID, from, to)
	if err != nil {
		return nil, nil, nil, apperror.Internal(err)
	}
	events := make([]*model.CalendarEvent, len(eventRows))
	for i, r := range eventRows {
		e := toModelEvent(r)
		events[i] = &e
	}

	todoRows, err := s.queries.ListTodosByDueRange(ctx, userID, from, to)
	if err != nil {
		return nil, nil, nil, apperror.Internal(err)
	}
	todos := make([]*model.Todo, len(todoRows))
	for i, r := range todoRows {
		t := toModelTodo(r)
		todos[i] = &t
	}

	memoRows, err := s.queries.ListMemosByCreatedRange(ctx, userID, from, to)
	if err != nil {
		return nil, nil, nil, apperror.Internal(err)
	}
	memos := make([]*model.Memo, len(memoRows))
	for i, m := range memoRows {
		memos[i] = &model.Memo{
			ID:        m.ID,
			Title:     m.Title,
			Type:      m.Type,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}

	return events, todos, memos, nil
}
