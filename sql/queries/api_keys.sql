-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, name, key_hash)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, created_at;

-- name: ListAPIKeysByUserID :many
SELECT id, user_id, name, last_used_at, created_at
FROM api_keys
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetAPIKeyByHash :one
SELECT ak.id, ak.user_id, ak.name, ak.key_hash, ak.last_used_at, ak.created_at,
       u.email, u.name AS user_name, u.is_admin
FROM api_keys ak
JOIN users u ON u.id = ak.user_id AND u.deleted_at IS NULL
WHERE ak.key_hash = $1 AND ak.deleted_at IS NULL;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = now() WHERE id = $1;

-- name: DeleteAPIKey :exec
UPDATE api_keys SET deleted_at = now() WHERE id = $1 AND user_id = $2;
