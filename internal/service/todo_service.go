package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type TodoService struct {
	queries *repository.Queries
	db      *sql.DB
	tagSvc  *TagService
}

func NewTodoService(db *sql.DB, tagSvc *TagService) *TodoService {
	return &TodoService{
		queries: repository.New(db),
		db:      db,
		tagSvc:  tagSvc,
	}
}

func (s *TodoService) ListTodos(ctx context.Context, userID uuid.UUID, params model.ListParams) (*model.PaginatedResult, error) {
	rows, err := s.queries.ListTodosByUserID(ctx, userID, params.Limit, params.Offset)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	total, err := s.queries.CountTodosByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	items := make([]model.Todo, len(rows))
	for i, r := range rows {
		items[i] = toModelTodo(r)
	}
	return &model.PaginatedResult{Data: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

func (s *TodoService) GetTodo(ctx context.Context, id, userID uuid.UUID) (*model.Todo, error) {
	r, err := s.queries.GetTodoByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("todo")
		}
		return nil, apperror.Internal(err)
	}
	todo := toModelTodo(r)

	tags, err := s.queries.GetTodoTags(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	todo.Tags = make([]model.Tag, len(tags))
	for i, t := range tags {
		todo.Tags[i] = model.Tag{ID: t.ID, Name: t.Name}
	}
	return &todo, nil
}

func (s *TodoService) CreateTodo(ctx context.Context, userID uuid.UUID, req model.CreateTodoRequest) (*model.Todo, error) {
	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, apperror.BadRequest("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &t
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	r, err := q.CreateTodo(ctx, userID, req.Title, req.Description, "open", dueDate)
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
			if err := q.AddTodoTag(ctx, r.ID, t.ID); err != nil {
				return nil, apperror.Internal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	todo := toModelTodo(r)
	todo.Tags = modelTags
	return &todo, nil
}

func (s *TodoService) UpdateTodo(ctx context.Context, id, userID uuid.UUID, req model.UpdateTodoRequest) (*model.Todo, error) {
	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, apperror.BadRequest("invalid due_date format")
		}
		dueDate = &t
	}

	status := req.Status
	if status == "" {
		status = "open"
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	r, err := q.UpdateTodo(ctx, id, userID, req.Title, req.Description, status, dueDate)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("todo")
		}
		return nil, apperror.Internal(err)
	}

	var modelTags []model.Tag
	if req.Tags != nil {
		if err := q.ClearTodoTags(ctx, id); err != nil {
			return nil, apperror.Internal(err)
		}
		modelTags, err = s.tagSvc.EnsureTags(ctx, userID, req.Tags)
		if err != nil {
			return nil, err
		}
		for _, t := range modelTags {
			if err := q.AddTodoTag(ctx, id, t.ID); err != nil {
				return nil, apperror.Internal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal(err)
	}

	todo := toModelTodo(r)
	todo.Tags = modelTags
	return &todo, nil
}

func (s *TodoService) MarkDone(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.UpdateTodoStatus(ctx, id, userID, "done")
}

func (s *TodoService) Reopen(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.UpdateTodoStatus(ctx, id, userID, "open")
}

func (s *TodoService) DeleteTodo(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.SoftDeleteTodo(ctx, id, userID)
}

func toModelTodo(r repository.TodoRow) model.Todo {
	todo := model.Todo{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.DueDate != nil {
		d := r.DueDate.Format("2006-01-02")
		todo.DueDate = &d
	}
	return todo
}
