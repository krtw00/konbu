-- name: CreateTool :one
INSERT INTO tools (user_id, name, url, icon, icon_checked_at, sort_order)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, name, url, icon, icon_checked_at, sort_order, created_at;

-- name: GetToolByID :one
SELECT id, user_id, name, url, icon, icon_checked_at, sort_order, created_at
FROM tools
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListToolsByUserID :many
SELECT id, user_id, name, url, icon, icon_checked_at, sort_order, created_at
FROM tools
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY sort_order, created_at;

-- name: UpdateTool :one
UPDATE tools
SET name = $3, url = $4, icon = $5, icon_checked_at = $6, sort_order = $7
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, name, url, icon, icon_checked_at, sort_order, created_at;

-- name: SoftDeleteTool :exec
UPDATE tools SET deleted_at = now() WHERE id = $1 AND user_id = $2;

-- name: UpdateToolSortOrder :exec
UPDATE tools SET sort_order = $3 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: MaxToolSortOrder :one
SELECT COALESCE(MAX(sort_order), 0)::int AS max_sort FROM tools WHERE user_id = $1 AND deleted_at IS NULL;
