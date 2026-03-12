-- name: GetUserByID :one
SELECT id, email, name, is_admin, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, name, is_admin, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, name, is_admin)
VALUES ($1, $2, $3)
RETURNING id, email, name, is_admin, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET name = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, name, is_admin, created_at, updated_at;

-- name: CountUsers :one
SELECT count(*) FROM users WHERE deleted_at IS NULL;
