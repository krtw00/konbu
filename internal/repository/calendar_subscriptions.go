package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CalendarSubscriptionRow struct {
	ID            uuid.UUID
	OwnerID       uuid.UUID
	CalendarID    uuid.UUID
	ICalURL       string
	LastFetchedAt *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

const subscriptionCols = `id, owner_id, calendar_id, ical_url, last_fetched_at, last_error, created_at, updated_at`

func scanSubscription(row interface{ Scan(...any) error }) (CalendarSubscriptionRow, error) {
	var s CalendarSubscriptionRow
	err := row.Scan(&s.ID, &s.OwnerID, &s.CalendarID, &s.ICalURL, &s.LastFetchedAt, &s.LastError, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (q *Queries) CreateSubscription(ctx context.Context, ownerID, calendarID uuid.UUID, icalURL string) (CalendarSubscriptionRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO calendar_subscriptions (owner_id, calendar_id, ical_url)
		 VALUES ($1, $2, $3)
		 RETURNING `+subscriptionCols,
		ownerID, calendarID, icalURL)
	return scanSubscription(row)
}

func (q *Queries) ListSubscriptionsByOwner(ctx context.Context, ownerID uuid.UUID) ([]CalendarSubscriptionRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+subscriptionCols+`
		 FROM calendar_subscriptions
		 WHERE owner_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at ASC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []CalendarSubscriptionRow
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// ListAllSubscriptions returns every non-deleted subscription across all
// owners. Used by the sync loop.
func (q *Queries) ListAllSubscriptions(ctx context.Context) ([]CalendarSubscriptionRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+subscriptionCols+`
		 FROM calendar_subscriptions
		 WHERE deleted_at IS NULL
		 ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []CalendarSubscriptionRow
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// GetSubscriptionByID returns the subscription scoped to the given owner.
func (q *Queries) GetSubscriptionByID(ctx context.Context, id, ownerID uuid.UUID) (CalendarSubscriptionRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT `+subscriptionCols+`
		 FROM calendar_subscriptions
		 WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`, id, ownerID)
	return scanSubscription(row)
}

// SoftDeleteSubscription logically deletes the subscription scoped to the owner.
func (q *Queries) SoftDeleteSubscription(ctx context.Context, id, ownerID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_subscriptions SET deleted_at = now(), updated_at = now()
		 WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`,
		id, ownerID)
	return err
}

// UpdateFetchResult records the outcome of a sync run. lastError is nil on
// success.
func (q *Queries) UpdateSubscriptionFetchResult(ctx context.Context, id uuid.UUID, fetchedAt time.Time, lastError *string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_subscriptions
		 SET last_fetched_at = $2, last_error = $3, updated_at = now()
		 WHERE id = $1`,
		id, fetchedAt, lastError)
	return err
}
