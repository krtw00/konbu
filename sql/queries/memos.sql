-- name: CreateMemo :one
INSERT INTO memos (user_id, title, type, content, table_columns)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, title, type, content, table_columns, created_at, updated_at;

-- name: GetMemoByID :one
SELECT id, user_id, title, type, content, table_columns, created_at, updated_at
FROM memos
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListMemosByUserID :many
SELECT id, user_id, title, type, created_at, updated_at
FROM memos
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMemosByUserID :one
SELECT count(*) FROM memos WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateMemo :one
UPDATE memos
SET title = $3, content = $4, table_columns = $5, updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, title, type, content, table_columns, created_at, updated_at;

-- name: SoftDeleteMemo :exec
UPDATE memos SET deleted_at = now() WHERE id = $1 AND user_id = $2;

-- name: SetMemoTags :exec
DELETE FROM memo_tags WHERE memo_id = $1;

-- name: AddMemoTag :exec
INSERT INTO memo_tags (memo_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: GetMemoTags :many
SELECT t.id, t.name
FROM tags t
JOIN memo_tags mt ON mt.tag_id = t.id
WHERE mt.memo_id = $1 AND t.deleted_at IS NULL;
