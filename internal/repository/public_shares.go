package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PublicShareRow struct {
	ID           uuid.UUID
	ResourceType string
	ResourceID   uuid.UUID
	CreatedBy    uuid.UUID
	Token        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func scanPublicShare(row interface{ Scan(...any) error }) (PublicShareRow, error) {
	var s PublicShareRow
	err := row.Scan(&s.ID, &s.ResourceType, &s.ResourceID, &s.CreatedBy, &s.Token, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (q *Queries) UpsertPublicShare(ctx context.Context, resourceType string, resourceID, createdBy uuid.UUID, token string) (PublicShareRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO public_shares (resource_type, resource_id, created_by, token)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (resource_type, resource_id)
		 DO UPDATE SET token = EXCLUDED.token, created_by = EXCLUDED.created_by, updated_at = now()
		 RETURNING id, resource_type, resource_id, created_by, token, created_at, updated_at`,
		resourceType, resourceID, createdBy, token)
	return scanPublicShare(row)
}

func (q *Queries) GetPublicShareByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) (PublicShareRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, resource_type, resource_id, created_by, token, created_at, updated_at
		 FROM public_shares
		 WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID)
	return scanPublicShare(row)
}

func (q *Queries) GetPublicShareByToken(ctx context.Context, token string) (PublicShareRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, resource_type, resource_id, created_by, token, created_at, updated_at
		 FROM public_shares
		 WHERE token = $1`,
		token)
	return scanPublicShare(row)
}

func (q *Queries) DeletePublicShare(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM public_shares WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID)
	return err
}
