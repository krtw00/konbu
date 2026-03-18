package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CalendarFeedToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (q *Queries) UpsertCalendarFeedToken(ctx context.Context, userID uuid.UUID, tokenHash string) (CalendarFeedToken, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO calendar_feed_tokens (user_id, token_hash)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id)
		 DO UPDATE SET token_hash = EXCLUDED.token_hash, updated_at = now(), deleted_at = NULL
		 RETURNING id, user_id, token_hash, last_used_at, created_at, updated_at`,
		userID, tokenHash)
	var token CalendarFeedToken
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.LastUsedAt, &token.CreatedAt, &token.UpdatedAt)
	return token, err
}

func (q *Queries) GetCalendarFeedTokenByUserID(ctx context.Context, userID uuid.UUID) (CalendarFeedToken, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, last_used_at, created_at, updated_at
		 FROM calendar_feed_tokens
		 WHERE user_id = $1 AND deleted_at IS NULL`,
		userID)
	var token CalendarFeedToken
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.LastUsedAt, &token.CreatedAt, &token.UpdatedAt)
	return token, err
}

func (q *Queries) GetCalendarFeedTokenByHash(ctx context.Context, tokenHash string) (CalendarFeedToken, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, last_used_at, created_at, updated_at
		 FROM calendar_feed_tokens
		 WHERE token_hash = $1 AND deleted_at IS NULL`,
		tokenHash)
	var token CalendarFeedToken
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.LastUsedAt, &token.CreatedAt, &token.UpdatedAt)
	return token, err
}

func (q *Queries) UpdateCalendarFeedTokenLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_feed_tokens SET last_used_at = now() WHERE id = $1`,
		id)
	return err
}

func (q *Queries) DeleteCalendarFeedToken(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_feed_tokens SET deleted_at = now(), updated_at = now() WHERE user_id = $1 AND deleted_at IS NULL`,
		userID)
	return err
}
