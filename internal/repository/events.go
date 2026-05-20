package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type EventRow struct {
	ID             uuid.UUID
	CreatedBy      uuid.UUID
	CalendarID     *uuid.UUID
	Title          string
	Description    string
	StartAt        time.Time
	EndAt          *time.Time
	AllDay         bool
	RecurrenceRule *string
	RecurrenceEnd  *string
	ExternalUID    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var eventCols = `id, created_by, calendar_id, title, description, start_at, end_at, all_day, recurrence_rule, recurrence_end, external_uid, created_at, updated_at`

func scanEvent(row interface{ Scan(...any) error }) (EventRow, error) {
	var e EventRow
	err := row.Scan(&e.ID, &e.CreatedBy, &e.CalendarID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &e.AllDay, &e.RecurrenceRule, &e.RecurrenceEnd, &e.ExternalUID, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (q *Queries) CreateEvent(ctx context.Context, userID uuid.UUID, calendarID *uuid.UUID, title, description string, startAt time.Time, endAt *time.Time, allDay bool, recurrenceRule, recurrenceEnd *string) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (created_by, calendar_id, title, description, start_at, end_at, all_day, recurrence_rule, recurrence_end)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+eventCols,
		userID, calendarID, title, description, startAt, endAt, allDay, recurrenceRule, recurrenceEnd)
	return scanEvent(row)
}

func (q *Queries) GetEventByID(ctx context.Context, id, userID uuid.UUID) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT `+eventCols+` FROM calendar_events
		 WHERE id = $1 AND deleted_at IS NULL AND created_by = $2`,
		id, userID)
	return scanEvent(row)
}

func (q *Queries) ListEventsByUserID(ctx context.Context, userID uuid.UUID, calendarID *uuid.UUID, limit, offset int) ([]EventRow, error) {
	var query string
	var args []any

	if calendarID != nil {
		query = `SELECT ` + eventCols + ` FROM calendar_events
			 WHERE calendar_id = $1 AND deleted_at IS NULL AND created_by = $2
			 ORDER BY start_at DESC LIMIT $3 OFFSET $4`
		args = []any{*calendarID, userID, limit, offset}
	} else {
		query = `SELECT ` + eventCols + ` FROM calendar_events
			 WHERE deleted_at IS NULL AND created_by = $1
			 ORDER BY start_at DESC LIMIT $2 OFFSET $3`
		args = []any{userID, limit, offset}
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRow
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) ListEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]EventRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+eventCols+` FROM calendar_events
		 WHERE created_by = $1 AND deleted_at IS NULL AND start_at >= $2 AND start_at < $3
		 ORDER BY start_at`,
		userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRow
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) ListAllEventsByUserID(ctx context.Context, userID uuid.UUID) ([]EventRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+eventCols+` FROM calendar_events
		 WHERE deleted_at IS NULL AND created_by = $1
		 ORDER BY start_at ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRow
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) CountEventsByUserID(ctx context.Context, userID uuid.UUID, calendarID *uuid.UUID) (int64, error) {
	var query string
	var args []any

	if calendarID != nil {
		query = `SELECT count(*) FROM calendar_events
			 WHERE calendar_id = $1 AND deleted_at IS NULL AND created_by = $2`
		args = []any{*calendarID, userID}
	} else {
		query = `SELECT count(*) FROM calendar_events
			 WHERE deleted_at IS NULL AND created_by = $1`
		args = []any{userID}
	}

	row := q.db.QueryRowContext(ctx, query, args...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) UpdateEvent(ctx context.Context, id, userID uuid.UUID, title, description string, startAt time.Time, endAt *time.Time, allDay bool, recurrenceRule, recurrenceEnd *string) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE calendar_events SET title = $3, description = $4, start_at = $5, end_at = $6, all_day = $7, recurrence_rule = $8, recurrence_end = $9, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL AND created_by = $2
		 RETURNING `+eventCols,
		id, userID, title, description, startAt, endAt, allDay, recurrenceRule, recurrenceEnd)
	return scanEvent(row)
}

func (q *Queries) SoftDeleteEvent(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_events SET deleted_at = now()
		 WHERE id = $1 AND created_by = $2`,
		id, userID)
	return err
}

func (q *Queries) SoftDeleteEventsByCalendar(ctx context.Context, calendarID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_events
		 SET deleted_at = now()
		 WHERE calendar_id = $1 AND deleted_at IS NULL`,
		calendarID)
	return err
}

// UpsertExternalEvent inserts or updates an event imported from an external
// iCal subscription, keyed by (calendar_id, external_uid). On conflict it
// refreshes the mutable fields. The conflict target matches the partial unique
// index idx_calendar_events_external.
func (q *Queries) UpsertExternalEvent(ctx context.Context, userID uuid.UUID, calendarID uuid.UUID, externalUID, title, description string, startAt time.Time, endAt *time.Time, allDay bool, recurrenceRule *string) (EventRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (created_by, calendar_id, title, description, start_at, end_at, all_day, recurrence_rule, external_uid)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (calendar_id, external_uid) WHERE external_uid IS NOT NULL AND deleted_at IS NULL
		 DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   start_at = EXCLUDED.start_at,
		   end_at = EXCLUDED.end_at,
		   all_day = EXCLUDED.all_day,
		   recurrence_rule = EXCLUDED.recurrence_rule,
		   updated_at = now()
		 RETURNING `+eventCols,
		userID, calendarID, title, description, startAt, endAt, allDay, recurrenceRule, externalUID)
	return scanEvent(row)
}

// DeleteStaleExternalEvents soft-deletes external events under the given
// calendar whose external_uid is not in keepUIDs. Self-authored events
// (external_uid IS NULL) are never touched. When keepUIDs is empty all external
// events under the calendar are removed.
func (q *Queries) DeleteStaleExternalEvents(ctx context.Context, calendarID uuid.UUID, keepUIDs []string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_events
		 SET deleted_at = now()
		 WHERE calendar_id = $1
		   AND deleted_at IS NULL
		   AND external_uid IS NOT NULL
		   AND NOT (external_uid = ANY($2))`,
		calendarID, pq.Array(keepUIDs))
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
