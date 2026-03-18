package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PublishedResourceRow struct {
	ID           uuid.UUID
	ResourceType string
	ResourceID   uuid.UUID
	CreatedBy    uuid.UUID
	Slug         string
	Title        string
	Description  *string
	Visibility   string
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func scanPublishedResource(row interface{ Scan(...any) error }) (PublishedResourceRow, error) {
	var r PublishedResourceRow
	err := row.Scan(
		&r.ID,
		&r.ResourceType,
		&r.ResourceID,
		&r.CreatedBy,
		&r.Slug,
		&r.Title,
		&r.Description,
		&r.Visibility,
		&r.PublishedAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	return r, err
}

func (q *Queries) UpsertPublishedResource(ctx context.Context, resourceType string, resourceID, createdBy uuid.UUID, slug, title string, description *string, visibility string) (PublishedResourceRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO published_resources (resource_type, resource_id, created_by, slug, title, description, visibility, published_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, CASE WHEN $7 = 'private' THEN NULL ELSE now() END)
		 ON CONFLICT (resource_type, resource_id)
		 DO UPDATE SET
		   created_by = EXCLUDED.created_by,
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   visibility = EXCLUDED.visibility,
		   published_at = CASE
		     WHEN EXCLUDED.visibility = 'private' THEN NULL
		     WHEN published_resources.published_at IS NULL THEN now()
		     ELSE published_resources.published_at
		   END,
		   updated_at = now()
		 RETURNING id, resource_type, resource_id, created_by, slug, title, description, visibility, published_at, created_at, updated_at`,
		resourceType, resourceID, createdBy, slug, title, description, visibility,
	)
	return scanPublishedResource(row)
}

func (q *Queries) GetPublishedResourceByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) (PublishedResourceRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, resource_type, resource_id, created_by, slug, title, description, visibility, published_at, created_at, updated_at
		 FROM published_resources
		 WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID,
	)
	return scanPublishedResource(row)
}

func (q *Queries) GetPublishedResourceBySlug(ctx context.Context, resourceType, slug string) (PublishedResourceRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, resource_type, resource_id, created_by, slug, title, description, visibility, published_at, created_at, updated_at
		 FROM published_resources
		 WHERE resource_type = $1 AND slug = $2`,
		resourceType, slug,
	)
	return scanPublishedResource(row)
}

func (q *Queries) DeletePublishedResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM published_resources WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID,
	)
	return err
}
