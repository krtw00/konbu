package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationUser is a sweep target user with their settings + email.
type NotificationUser struct {
	ID           uuid.UUID
	Email        string
	UserSettings json.RawMessage
}

// ListUsersForNotifications returns every non-deleted user with their email
// and user_settings JSON. Used by the notification sweep loop. The loop is
// single-instance so a full scan per tick is acceptable for the foreseeable
// user count.
func (q *Queries) ListUsersForNotifications(ctx context.Context) ([]NotificationUser, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, email, COALESCE(user_settings, '{}'::jsonb)
		 FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []NotificationUser
	for rows.Next() {
		var u NotificationUser
		if err := rows.Scan(&u.ID, &u.Email, &u.UserSettings); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListDueTodosForUser returns open todos for the given user whose due_date is
// not null and falls on or before `until`. Used by the notification sweep to
// build the candidate set for "todo due" notifications.
func (q *Queries) ListDueTodosForUser(ctx context.Context, userID uuid.UUID, until time.Time) ([]TodoRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, title, description, status, due_date, created_at, updated_at
		 FROM todos
		 WHERE user_id = $1 AND deleted_at IS NULL
		   AND status <> 'done'
		   AND due_date IS NOT NULL
		   AND due_date <= $2`,
		userID, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []TodoRow
	for rows.Next() {
		var t TodoRow
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

// ListUpcomingEventsForUser returns events the user owns or is a member of
// whose start_at falls within the [from, to] window. Used by the notification
// sweep to build the candidate set for "event lead" notifications.
func (q *Queries) ListUpcomingEventsForUser(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]EventRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+eventCols+` FROM calendar_events
		 WHERE deleted_at IS NULL
		   AND start_at >= $2 AND start_at <= $3
		   AND (created_by = $1
		        OR calendar_id IN (SELECT calendar_id FROM calendar_members WHERE user_id = $1))`,
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

// MarkNotificationSent inserts a notification_sent_log row.
// Returns true if the row was newly inserted (i.e. notification has not been
// sent before), false if the unique constraint already had a matching row.
func (q *Queries) MarkNotificationSent(ctx context.Context, userID, resourceID uuid.UUID, resourceType, kind string) (bool, error) {
	res, err := q.db.ExecContext(ctx,
		`INSERT INTO notification_sent_log (user_id, resource_type, resource_id, kind)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, resource_type, resource_id, kind) DO NOTHING`,
		userID, resourceType, resourceID, kind)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
