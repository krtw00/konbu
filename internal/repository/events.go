package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EventRow struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	StartAt     time.Time
	EndAt       *time.Time
	AllDay      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (q *Queries) CreateEvent(ctx context.Context, userID uuid.UUID, title, description string, startAt time.Time, endAt *time.Time, allDay bool) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, start_at, end_at, all_day)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at`,
		userID, title, description, startAt, endAt, allDay)
	var e EventRow
	err := row.Scan(&e.ID, &e.UserID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &e.AllDay, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (q *Queries) GetEventByID(ctx context.Context, id, userID uuid.UUID) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at
		 FROM calendar_events WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	var e EventRow
	err := row.Scan(&e.ID, &e.UserID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &e.AllDay, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (q *Queries) ListEventsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]EventRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at
		 FROM calendar_events WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY start_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRow
	for rows.Next() {
		var e EventRow
		if err := rows.Scan(&e.ID, &e.UserID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &e.AllDay, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) CountEventsByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT count(*) FROM calendar_events WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) UpdateEvent(ctx context.Context, id, userID uuid.UUID, title, description string, startAt time.Time, endAt *time.Time, allDay bool) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE calendar_events SET title = $3, description = $4, start_at = $5, end_at = $6, all_day = $7, updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at`,
		id, userID, title, description, startAt, endAt, allDay)
	var e EventRow
	err := row.Scan(&e.ID, &e.UserID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &e.AllDay, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (q *Queries) SoftDeleteEvent(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_events SET deleted_at = now() WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (q *Queries) ClearEventTags(ctx context.Context, eventID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM calendar_event_tags WHERE event_id = $1`, eventID)
	return err
}

func (q *Queries) AddEventTag(ctx context.Context, eventID, tagID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO calendar_event_tags (event_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		eventID, tagID)
	return err
}

func (q *Queries) GetEventTags(ctx context.Context, eventID uuid.UUID) ([]MemoTag, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT t.id, t.name FROM tags t
		 JOIN calendar_event_tags et ON et.tag_id = t.id
		 WHERE et.event_id = $1 AND t.deleted_at IS NULL`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []MemoTag
	for rows.Next() {
		var t MemoTag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
