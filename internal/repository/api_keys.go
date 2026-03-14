package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Name       string
	KeyHash    string
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

type APIKeyWithUser struct {
	APIKey
	Email    string
	UserName string
	IsAdmin  bool
	Plan     string
}

func (q *Queries) CreateAPIKey(ctx context.Context, userID uuid.UUID, name, keyHash string) (APIKey, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO api_keys (user_id, name, key_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, name, key_hash, created_at`,
		userID, name, keyHash)
	var k APIKey
	err := row.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.CreatedAt)
	return k, err
}

func (q *Queries) ListAPIKeysByUserID(ctx context.Context, userID uuid.UUID) ([]APIKey, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, name, key_hash, last_used_at, created_at
		 FROM api_keys WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (q *Queries) GetAPIKeyByHash(ctx context.Context, keyHash string) (APIKeyWithUser, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT ak.id, ak.user_id, ak.name, ak.key_hash, ak.last_used_at, ak.created_at,
		        u.email, u.name, u.is_admin, u.plan
		 FROM api_keys ak
		 JOIN users u ON u.id = ak.user_id AND u.deleted_at IS NULL
		 WHERE ak.key_hash = $1 AND ak.deleted_at IS NULL`, keyHash)
	var k APIKeyWithUser
	err := row.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.LastUsedAt, &k.CreatedAt,
		&k.Email, &k.UserName, &k.IsAdmin, &k.Plan)
	return k, err
}

func (q *Queries) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = now() WHERE id = $1`, id)
	return err
}

func (q *Queries) DeleteAPIKey(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE api_keys SET deleted_at = now() WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
