-- name: CreateEvent :one
INSERT INTO calendar_events (user_id, title, description, start_at, end_at, all_day)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at;

-- name: GetEventByID :one
SELECT id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at
FROM calendar_events
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListEventsByUserID :many
SELECT id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at
FROM calendar_events
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY start_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEventsByUserID :one
SELECT count(*) FROM calendar_events WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateEvent :one
UPDATE calendar_events
SET title = $3, description = $4, start_at = $5, end_at = $6, all_day = $7, updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at;

-- name: SoftDeleteEvent :exec
UPDATE calendar_events SET deleted_at = now() WHERE id = $1 AND user_id = $2;

-- name: SetEventTags :exec
DELETE FROM calendar_event_tags WHERE event_id = $1;

-- name: AddEventTag :exec
INSERT INTO calendar_event_tags (event_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: GetEventTags :many
SELECT t.id, t.name
FROM tags t
JOIN calendar_event_tags et ON et.tag_id = t.id
WHERE et.event_id = $1 AND t.deleted_at IS NULL;
