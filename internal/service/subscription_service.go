package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

const (
	subscriptionFetchTimeout = 15 * time.Second
	subscriptionMaxBodyBytes = 5 << 20 // 5 MB
	subscriptionSyncTimeout  = 10 * time.Minute
	defaultExternalColor     = "#4F46E5"
)

// SubscriptionService manages iCal subscriptions: creation, periodic sync, and
// deletion of the dedicated read-only calendars they back.
type SubscriptionService struct {
	queries    *repository.Queries
	db         *sql.DB
	httpClient *http.Client
}

func NewSubscriptionService(db *sql.DB) *SubscriptionService {
	return &SubscriptionService{
		queries:    repository.New(db),
		db:         db,
		httpClient: &http.Client{Timeout: subscriptionFetchTimeout},
	}
}

// Create validates the URL, provisions a dedicated external calendar, persists
// the subscription, then performs an initial sync. A failed initial fetch does
// not roll back the subscription; the error is recorded in last_error.
func (s *SubscriptionService) Create(ctx context.Context, ownerID uuid.UUID, name, icalURL, color string) (*model.CalendarSubscription, error) {
	if strings.TrimSpace(name) == "" {
		return nil, apperror.BadRequest("name is required")
	}
	icalURL = strings.TrimSpace(icalURL)
	if err := validateICalURL(icalURL); err != nil {
		return nil, err
	}
	if color == "" {
		color = defaultExternalColor
	}

	cal, err := s.queries.CreateExternalCalendar(ctx, ownerID, name, color)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	subRow, err := s.queries.CreateSubscription(ctx, ownerID, cal.ID, icalURL)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Initial sync. Failure is recorded in last_error but does not fail Create.
	s.syncOne(ctx, subRow)

	// Re-read to pick up last_fetched_at / last_error from the initial sync.
	refreshed, err := s.queries.GetSubscriptionByID(ctx, subRow.ID, ownerID)
	if err != nil {
		// Fall back to the pre-sync row rather than failing the create.
		sub := toModelSubscription(subRow)
		return &sub, nil
	}
	sub := toModelSubscription(refreshed)
	return &sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, ownerID uuid.UUID) ([]model.CalendarSubscription, error) {
	rows, err := s.queries.ListSubscriptionsByOwner(ctx, ownerID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	subs := make([]model.CalendarSubscription, len(rows))
	for i, r := range rows {
		subs[i] = toModelSubscription(r)
	}
	return subs, nil
}

// Delete removes the subscription, its dedicated calendar, and the imported
// events under it. The calendar's ON DELETE CASCADE is not used because all
// deletes are logical (soft delete).
func (s *SubscriptionService) Delete(ctx context.Context, ownerID, subID uuid.UUID) error {
	subRow, err := s.queries.GetSubscriptionByID(ctx, subID, ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("subscription")
		}
		return apperror.Internal(err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.Internal(err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)
	if err := q.SoftDeleteSubscription(ctx, subID, ownerID); err != nil {
		return apperror.Internal(err)
	}
	if err := q.SoftDeleteEventsByCalendar(ctx, subRow.CalendarID); err != nil {
		return apperror.Internal(err)
	}
	if err := q.SoftDeleteCalendar(ctx, subRow.CalendarID); err != nil {
		return apperror.Internal(err)
	}

	if err := tx.Commit(); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

// SyncByID performs a manual sync of a single subscription scoped to the owner.
func (s *SubscriptionService) SyncByID(ctx context.Context, ownerID, subID uuid.UUID) (*model.CalendarSubscription, error) {
	subRow, err := s.queries.GetSubscriptionByID(ctx, subID, ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("subscription")
		}
		return nil, apperror.Internal(err)
	}

	s.syncOne(ctx, subRow)

	refreshed, err := s.queries.GetSubscriptionByID(ctx, subID, ownerID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	sub := toModelSubscription(refreshed)
	return &sub, nil
}

// SyncAll syncs every non-deleted subscription. Called from the loop. A failure
// on one subscription does not stop the others.
func (s *SubscriptionService) SyncAll(ctx context.Context) error {
	subs, err := s.queries.ListAllSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}
	for _, sub := range subs {
		s.syncOne(ctx, sub)
	}
	return nil
}

// syncOne fetches the iCal feed, upserts events keyed by UID, removes events
// dropped upstream, and records the fetch outcome. All errors are captured in
// last_error so the caller (loop / handler) never has to handle them.
func (s *SubscriptionService) syncOne(ctx context.Context, sub repository.CalendarSubscriptionRow) {
	now := time.Now().UTC()
	if err := s.doSync(ctx, sub); err != nil {
		msg := err.Error()
		if uErr := s.queries.UpdateSubscriptionFetchResult(ctx, sub.ID, now, &msg); uErr != nil {
			log.Printf("subscription sync: update fetch result (id=%s): %v", sub.ID, uErr)
		}
		log.Printf("subscription sync failed (id=%s): %v", sub.ID, err)
		return
	}
	if err := s.queries.UpdateSubscriptionFetchResult(ctx, sub.ID, now, nil); err != nil {
		log.Printf("subscription sync: update fetch result (id=%s): %v", sub.ID, err)
	}
}

func (s *SubscriptionService) doSync(ctx context.Context, sub repository.CalendarSubscriptionRow) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sub.ICalURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body := io.LimitReader(resp.Body, subscriptionMaxBodyBytes)
	events, err := parseICal(body)
	if err != nil {
		return fmt.Errorf("parse ical: %w", err)
	}

	keepUIDs := make([]string, 0, len(events))
	for _, ev := range events {
		if ev.uid == "" {
			// No UID: cannot dedup safely, skip.
			continue
		}
		req, err := toCreateEventRequest(ev)
		if err != nil {
			// Skip individual malformed events but keep syncing the rest.
			log.Printf("subscription sync: skip event (id=%s uid=%s): %v", sub.ID, ev.uid, err)
			continue
		}
		if _, err := s.queries.UpsertExternalEvent(ctx, sub.OwnerID, sub.CalendarID, ev.uid, req.Title, req.Description, req.StartAt, req.EndAt, req.AllDay, req.RecurrenceRule); err != nil {
			return fmt.Errorf("upsert event (uid=%s): %w", ev.uid, err)
		}
		keepUIDs = append(keepUIDs, ev.uid)
	}

	if err := s.queries.DeleteStaleExternalEvents(ctx, sub.CalendarID, keepUIDs); err != nil {
		return fmt.Errorf("delete stale events: %w", err)
	}
	return nil
}

// StartLoop spawns a goroutine that runs SyncAll every `interval`. The first
// sync runs immediately so deploys see a quick signal in logs.
func (s *SubscriptionService) StartLoop(interval time.Duration) {
	go func() {
		s.runSyncAll()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.runSyncAll()
		}
	}()
}

func (s *SubscriptionService) runSyncAll() {
	ctx, cancel := context.WithTimeout(context.Background(), subscriptionSyncTimeout)
	defer cancel()
	if err := s.SyncAll(ctx); err != nil {
		log.Printf("subscription sync sweep error: %v", err)
	}
}

func validateICalURL(raw string) error {
	if raw == "" {
		return apperror.BadRequest("ical_url is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return apperror.BadRequest("invalid ical_url")
	}
	if u.Scheme != "https" {
		return apperror.BadRequest("ical_url must use https")
	}
	if u.Host == "" {
		return apperror.BadRequest("invalid ical_url")
	}
	return nil
}

func toModelSubscription(r repository.CalendarSubscriptionRow) model.CalendarSubscription {
	return model.CalendarSubscription{
		ID:            r.ID,
		OwnerID:       r.OwnerID,
		CalendarID:    r.CalendarID,
		ICalURL:       r.ICalURL,
		LastFetchedAt: r.LastFetchedAt,
		LastError:     r.LastError,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
