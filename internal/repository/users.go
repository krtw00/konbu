package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

type UserWithPassword struct {
	ID           uuid.UUID
	Email        string
	Name         string
	PasswordHash *string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

func (q *Queries) GetUserByEmailWithPassword(ctx context.Context, email string) (UserWithPassword, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, is_admin, created_at, updated_at
		 FROM users WHERE email = $1 AND deleted_at IS NULL`, email)
	var u UserWithPassword
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
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

func (q *Queries) CreateUserWithPassword(ctx context.Context, email, name, passwordHash string, isAdmin bool) (User, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO users (email, name, password_hash, is_admin)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, name, is_admin, created_at, updated_at`,
		email, name, passwordHash, isAdmin)
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

func (q *Queries) SetUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL`, id, passwordHash)
	return err
}

func (q *Queries) UpdateUserSettings(ctx context.Context, id uuid.UUID, settings json.RawMessage) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET user_settings = $2, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL`, id, settings)
	return err
}

func (q *Queries) GetUserSettings(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT user_settings FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	var settings json.RawMessage
	err := row.Scan(&settings)
	return settings, err
}

func (q *Queries) UpdateUserLocale(ctx context.Context, id uuid.UUID, locale string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET locale = $2, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL`, id, locale)
	return err
}

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT count(*) FROM users WHERE deleted_at IS NULL`)
	var count int64
	err := row.Scan(&count)
	return count, err
}

var ErrNoRows = sql.ErrNoRows
