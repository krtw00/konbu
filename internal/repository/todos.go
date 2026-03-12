package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TodoRow struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Status      string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (q *Queries) CreateTodo(ctx context.Context, userID uuid.UUID, title, description, status string, dueDate *time.Time) (TodoRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO todos (user_id, title, description, status, due_date)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, title, description, status, due_date, created_at, updated_at`,
		userID, title, description, status, dueDate)
	var t TodoRow
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (q *Queries) GetTodoByID(ctx context.Context, id, userID uuid.UUID) (TodoRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, description, status, due_date, created_at, updated_at
		 FROM todos WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	var t TodoRow
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (q *Queries) ListTodosByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]TodoRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, title, description, status, due_date, created_at, updated_at
		 FROM todos WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []TodoRow
	for rows.Next() {
		var t TodoRow
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (q *Queries) CountTodosByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT count(*) FROM todos WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) UpdateTodo(ctx context.Context, id, userID uuid.UUID, title, description, status string, dueDate *time.Time) (TodoRow, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE todos SET title = $3, description = $4, status = $5, due_date = $6, updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING id, user_id, title, description, status, due_date, created_at, updated_at`,
		id, userID, title, description, status, dueDate)
	var t TodoRow
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (q *Queries) UpdateTodoStatus(ctx context.Context, id, userID uuid.UUID, status string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE todos SET status = $3, updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID, status)
	return err
}

func (q *Queries) SoftDeleteTodo(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE todos SET deleted_at = now() WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (q *Queries) ClearTodoTags(ctx context.Context, todoID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM todo_tags WHERE todo_id = $1`, todoID)
	return err
}

func (q *Queries) AddTodoTag(ctx context.Context, todoID, tagID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO todo_tags (todo_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		todoID, tagID)
	return err
}

func (q *Queries) GetTodoTags(ctx context.Context, todoID uuid.UUID) ([]MemoTag, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT t.id, t.name FROM tags t
		 JOIN todo_tags tt ON tt.tag_id = t.id
		 WHERE tt.todo_id = $1 AND t.deleted_at IS NULL`, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []MemoTag
	for rows.Next() {
		var t MemoTag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
