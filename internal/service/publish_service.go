package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
	"github.com/lib/pq"
)

const (
	PublishedResourceMemo     = "memo"
	PublishedResourceEvent    = "event"
	PublishedResourceCalendar = "calendar"

	PublishVisibilityPrivate  = "private"
	PublishVisibilityUnlisted = "unlisted"
	PublishVisibilityPublic   = "public"
)

var publishSlugSanitizer = regexp.MustCompile(`[^a-z0-9-]+`)

type PublishService struct {
	queries *repository.Queries
}

func NewPublishService(db *sql.DB) *PublishService {
	return &PublishService{queries: repository.New(db)}
}

func (s *PublishService) Get(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (*model.PublishedResource, error) {
	defaultTitle, err := s.ensurePublishable(ctx, userID, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	row, err := s.queries.GetPublishedResourceByResource(ctx, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, nil
		}
		return nil, apperror.Internal(err)
	}
	pub := toModelPublishedResource(row)
	if pub.Title == "" {
		pub.Title = defaultTitle
	}
	return &pub, nil
}

func (s *PublishService) Upsert(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, req model.UpsertPublishedResourceRequest) (*model.PublishedResource, error) {
	defaultTitle, err := s.ensurePublishable(ctx, userID, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	visibility := normalizePublishVisibility(req.Visibility)
	if visibility == "" {
		return nil, apperror.BadRequest("visibility must be private, unlisted, or public")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = defaultTitle
	}
	if title == "" {
		return nil, apperror.BadRequest("title is required")
	}

	slug := normalizePublishSlug(req.Slug)
	if slug == "" {
		slug = fallbackPublishSlug(resourceType, resourceID, title)
	}

	description := strings.TrimSpace(req.Description)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	row, err := s.queries.UpsertPublishedResource(ctx, resourceType, resourceID, userID, slug, title, descriptionPtr, visibility)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, apperror.BadRequest("slug already exists for this resource type")
		}
		return nil, apperror.Internal(err)
	}
	pub := toModelPublishedResource(row)
	return &pub, nil
}

func (s *PublishService) Delete(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	if _, err := s.ensurePublishable(ctx, userID, resourceType, resourceID); err != nil {
		return err
	}
	if err := s.queries.DeletePublishedResource(ctx, resourceType, resourceID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *PublishService) GetPublicMetadata(ctx context.Context, resourceType, slug string) (*model.PublishedResource, error) {
	resourceType = normalizePublishedResourceType(resourceType)
	if resourceType == "" {
		return nil, apperror.BadRequest("unsupported resource type")
	}
	slug = normalizePublishSlug(slug)
	if slug == "" {
		return nil, apperror.BadRequest("invalid slug")
	}

	row, err := s.queries.GetPublishedResourceBySlug(ctx, resourceType, slug)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("published resource")
		}
		return nil, apperror.Internal(err)
	}
	if row.Visibility == PublishVisibilityPrivate {
		return nil, apperror.NotFound("published resource")
	}
	pub := toModelPublishedResource(row)
	return &pub, nil
}

func (s *PublishService) ensurePublishable(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (string, error) {
	switch normalizePublishedResourceType(resourceType) {
	case PublishedResourceMemo:
		memo, err := s.queries.GetMemoByID(ctx, resourceID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return "", apperror.NotFound("memo")
			}
			return "", apperror.Internal(err)
		}
		return memo.Title, nil
	case PublishedResourceEvent:
		event, err := s.queries.GetEventByID(ctx, resourceID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return "", apperror.NotFound("event")
			}
			return "", apperror.Internal(err)
		}
		return event.Title, nil
	case PublishedResourceCalendar:
		cal, err := s.queries.GetCalendarByID(ctx, resourceID)
		if err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return "", apperror.NotFound("calendar")
			}
			return "", apperror.Internal(err)
		}
		if cal.OwnerID != userID {
			return "", apperror.NotFound("calendar")
		}
		return cal.Name, nil
	default:
		return "", apperror.BadRequest("unsupported resource type")
	}
}

func normalizePublishedResourceType(resourceType string) string {
	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case PublishedResourceMemo, PublishedResourceEvent, PublishedResourceCalendar:
		return strings.ToLower(strings.TrimSpace(resourceType))
	default:
		return ""
	}
}

func normalizePublishVisibility(visibility string) string {
	switch strings.ToLower(strings.TrimSpace(visibility)) {
	case PublishVisibilityPrivate, PublishVisibilityUnlisted, PublishVisibilityPublic:
		return strings.ToLower(strings.TrimSpace(visibility))
	default:
		return ""
	}
}

func normalizePublishSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = publishSlugSanitizer.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

func fallbackPublishSlug(resourceType string, resourceID uuid.UUID, title string) string {
	if normalized := normalizePublishSlug(title); normalized != "" {
		return normalized
	}
	return fmt.Sprintf("%s-%s", normalizePublishedResourceType(resourceType), resourceID.String()[:8])
}

func toModelPublishedResource(row repository.PublishedResourceRow) model.PublishedResource {
	pub := model.PublishedResource{
		ResourceType: row.ResourceType,
		ResourceID:   row.ResourceID,
		Slug:         row.Slug,
		Title:        row.Title,
		Visibility:   row.Visibility,
		PublishedAt:  row.PublishedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.Description != nil {
		pub.Description = *row.Description
	}
	return pub
}
