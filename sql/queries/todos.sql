-- name: CreateTodo :one
INSERT INTO todos (user_id, title, description, status, due_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, title, description, status, due_date, created_at, updated_at;

-- name: GetTodoByID :one
SELECT id, user_id, title, description, status, due_date, created_at, updated_at
FROM todos
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListTodosByUserID :many
SELECT id, user_id, title, description, status, due_date, created_at, updated_at
FROM todos
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTodosByUserID :one
SELECT count(*) FROM todos WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateTodo :one
UPDATE todos
SET title = $3, description = $4, status = $5, due_date = $6, updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, title, description, status, due_date, created_at, updated_at;

-- name: UpdateTodoStatus :exec
UPDATE todos SET status = $3, updated_at = now() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: SoftDeleteTodo :exec
UPDATE todos SET deleted_at = now() WHERE id = $1 AND user_id = $2;

-- name: SetTodoTags :exec
DELETE FROM todo_tags WHERE todo_id = $1;

-- name: AddTodoTag :exec
INSERT INTO todo_tags (todo_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: GetTodoTags :many
SELECT t.id, t.name
FROM tags t
JOIN todo_tags tt ON tt.tag_id = t.id
WHERE tt.todo_id = $1 AND t.deleted_at IS NULL;
