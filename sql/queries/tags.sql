-- name: CreateTag :one
INSERT INTO tags (user_id, name)
VALUES ($1, $2)
ON CONFLICT (user_id, name) DO UPDATE SET deleted_at = NULL
RETURNING id, user_id, name, created_at;

-- name: GetTagByID :one
SELECT id, user_id, name, created_at
FROM tags
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListTagsByUserID :many
SELECT id, user_id, name, created_at
FROM tags
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: UpdateTag :one
UPDATE tags SET name = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, user_id, name, created_at;

-- name: SoftDeleteTag :exec
UPDATE tags SET deleted_at = now() WHERE id = $1 AND user_id = $2;

-- name: GetTagsByNames :many
SELECT id, user_id, name, created_at
FROM tags
WHERE user_id = $1 AND name = ANY($2::text[]) AND deleted_at IS NULL;
