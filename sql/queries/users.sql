-- name: GetUserByID :one
SELECT id, email, name, is_admin, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, name, is_admin, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByEmailWithPassword :one
SELECT id, email, name, password_hash, is_admin, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, name, is_admin)
VALUES ($1, $2, $3)
RETURNING id, email, name, is_admin, created_at, updated_at;

-- name: CreateUserWithPassword :one
INSERT INTO users (email, name, password_hash, is_admin)
VALUES ($1, $2, $3, $4)
RETURNING id, email, name, is_admin, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET name = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, name, is_admin, created_at, updated_at;

-- name: SetUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserSettings :exec
UPDATE users
SET user_settings = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserSettings :one
SELECT user_settings FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserLocale :exec
UPDATE users
SET locale = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CountUsers :one
SELECT count(*) FROM users WHERE deleted_at IS NULL;
