package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, email, name, is_admin, created_at, updated_at
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, email, name, is_admin, created_at, updated_at
		 FROM users WHERE email = $1 AND deleted_at IS NULL`, email)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) CreateUser(ctx context.Context, email, name string, isAdmin bool) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO users (email, name, is_admin)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, name, is_admin, created_at, updated_at`,
		email, name, isAdmin)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) UpdateUser(ctx context.Context, id uuid.UUID, name string) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE users SET name = $2, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING id, email, name, is_admin, created_at, updated_at`,
		id, name)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT count(*) FROM users WHERE deleted_at IS NULL`)
	var count int64
	err := row.Scan(&count)
	return count, err
}

// ErrNoRows is re-exported for convenience.
var ErrNoRows = sql.ErrNoRows
