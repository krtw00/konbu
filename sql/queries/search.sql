-- name: SearchMemos :many
SELECT id, title, type, created_at, updated_at
FROM memos
WHERE user_id = $1 AND deleted_at IS NULL
  AND (title LIKE '%' || $2 || '%' OR content LIKE '%' || $2 || '%')
ORDER BY updated_at DESC
LIMIT $3;

-- name: SearchTodos :many
SELECT id, title, status, due_date, created_at, updated_at
FROM todos
WHERE user_id = $1 AND deleted_at IS NULL
  AND (title LIKE '%' || $2 || '%' OR description LIKE '%' || $2 || '%')
ORDER BY updated_at DESC
LIMIT $3;

-- name: SearchEvents :many
SELECT id, title, start_at, end_at, all_day, created_at, updated_at
FROM calendar_events
WHERE user_id = $1 AND deleted_at IS NULL
  AND (title LIKE '%' || $2 || '%' OR description LIKE '%' || $2 || '%')
ORDER BY start_at DESC
LIMIT $3;
