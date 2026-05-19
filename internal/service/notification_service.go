package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/krtw00/konbu/internal/repository"
)

// Notification kinds stored in notification_sent_log.kind.
const (
	notificationKindTodoDue    = "due"
	notificationKindEventLead  = "lead"
	resourceTypeTodoDue        = "todo_due"
	resourceTypeEventStart     = "event_start"
	defaultEventLeadMinutes    = 30
	defaultTodoDueTime         = "09:00"
	defaultTimezone            = "Asia/Tokyo"
	notificationSweepTimeout   = 5 * time.Minute
)

// notificationSettings is the user-facing sub-schema persisted under the
// "notifications" key of users.user_settings JSONB.
type notificationSettings struct {
	Enabled          bool   `json:"enabled"`
	Email            string `json:"email"`
	EventLeadMinutes int    `json:"event_lead_minutes"`
	TodoDueTime      string `json:"todo_due_time"`
	Timezone         string `json:"timezone"`
}

// resolve applies opt-in defaults: enabled=false, lead=30m, due_time=09:00,
// tz=Asia/Tokyo. Returns the resolved settings and the recipient email
// (settings.Email, falling back to the user account email).
func resolveNotificationSettings(raw json.RawMessage, accountEmail string) (notificationSettings, string) {
	out := notificationSettings{
		Enabled:          false,
		EventLeadMinutes: defaultEventLeadMinutes,
		TodoDueTime:      defaultTodoDueTime,
		Timezone:         defaultTimezone,
	}
	if len(raw) > 0 {
		var envelope struct {
			Notifications *notificationSettings `json:"notifications"`
		}
		if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Notifications != nil {
			n := *envelope.Notifications
			out.Enabled = n.Enabled
			if n.EventLeadMinutes > 0 {
				out.EventLeadMinutes = n.EventLeadMinutes
			}
			if n.TodoDueTime != "" {
				out.TodoDueTime = n.TodoDueTime
			}
			if n.Timezone != "" {
				out.Timezone = n.Timezone
			}
			out.Email = n.Email
		}
	}
	to := out.Email
	if to == "" {
		to = accountEmail
	}
	return out, to
}

// NotificationService periodically sweeps todos / events for users with
// notifications enabled and emails reminders via SMTP.
type NotificationService struct {
	queries *repository.Queries
	mailer  Mailer
}

// NewNotificationService returns a service. Callers should call StartLoop only
// when mailer != nil.
func NewNotificationService(db *sql.DB, mailer Mailer) *NotificationService {
	return &NotificationService{
		queries: repository.New(db),
		mailer:  mailer,
	}
}

// StartLoop spawns a goroutine that runs sweepOnce every `tick`. The first
// sweep runs immediately so deploys see a quick signal in logs.
func (s *NotificationService) StartLoop(tick time.Duration) {
	go func() {
		s.runSweep()
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		for range ticker.C {
			s.runSweep()
		}
	}()
}

func (s *NotificationService) runSweep() {
	ctx, cancel := context.WithTimeout(context.Background(), notificationSweepTimeout)
	defer cancel()
	now := time.Now().UTC()
	if err := s.sweepOnce(ctx, now); err != nil {
		log.Printf("notification sweep error: %v", err)
	}
}

// sweepOnce iterates all users, picks those who opted in, and emits any
// pending reminders. Called from the ticker and from tests directly.
func (s *NotificationService) sweepOnce(ctx context.Context, now time.Time) error {
	users, err := s.queries.ListUsersForNotifications(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	for _, u := range users {
		settings, to := resolveNotificationSettings(u.UserSettings, u.Email)
		if !settings.Enabled || to == "" {
			continue
		}
		loc, err := time.LoadLocation(settings.Timezone)
		if err != nil {
			log.Printf("notification: invalid timezone %q for user %s; falling back to UTC", settings.Timezone, u.ID)
			loc = time.UTC
		}

		s.sweepTodos(ctx, u.ID, to, settings, loc, now)
		s.sweepEvents(ctx, u.ID, to, settings, now)
	}
	return nil
}

// sweepTodos sends one "due" reminder per open todo whose due_date (interpreted
// at the user's todo_due_time in their timezone) has already arrived.
func (s *NotificationService) sweepTodos(ctx context.Context, userID uuid.UUID, to string, settings notificationSettings, loc *time.Location, now time.Time) {
	todos, err := s.queries.ListDueTodosForUser(ctx, userID, now)
	if err != nil {
		log.Printf("notification: list todos for %s: %v", userID, err)
		return
	}
	hour, minute, ok := parseHHMM(settings.TodoDueTime)
	if !ok {
		hour, minute = 9, 0
	}
	for _, t := range todos {
		if t.DueDate == nil {
			continue
		}
		// due_date is a DATE (postgres) — interpret it as YYYY-MM-DD in the
		// user's timezone at todo_due_time.
		due := *t.DueDate
		fireAt := time.Date(due.Year(), due.Month(), due.Day(), hour, minute, 0, 0, loc)
		if now.Before(fireAt) {
			continue
		}
		s.maybeSend(ctx, userID, t.ID, resourceTypeTodoDue, notificationKindTodoDue, to,
			fmt.Sprintf("[konbu] ToDo の期日: %s", t.Title),
			fmt.Sprintf("ToDo 「%s」 の期日になりました。\n\n期限: %s", t.Title, fireAt.Format("2006-01-02 15:04 MST")))
	}
}

// sweepEvents sends one "lead" reminder per event whose start_at is within the
// configured lead window. The kind is `lead_<N>m` so changing lead_minutes
// later doesn't suppress new windows.
func (s *NotificationService) sweepEvents(ctx context.Context, userID uuid.UUID, to string, settings notificationSettings, now time.Time) {
	lead := time.Duration(settings.EventLeadMinutes) * time.Minute
	from := now
	until := now.Add(lead)
	events, err := s.queries.ListUpcomingEventsForUser(ctx, userID, from, until)
	if err != nil {
		log.Printf("notification: list events for %s: %v", userID, err)
		return
	}
	kind := fmt.Sprintf("%s_%dm", notificationKindEventLead, settings.EventLeadMinutes)
	for _, e := range events {
		s.maybeSend(ctx, userID, e.ID, resourceTypeEventStart, kind, to,
			fmt.Sprintf("[konbu] 予定がもうすぐ始まります: %s", e.Title),
			fmt.Sprintf("予定 「%s」 が %d 分後に始まります。\n\n開始: %s", e.Title, settings.EventLeadMinutes, e.StartAt.Format("2006-01-02 15:04 MST")))
	}
}

// maybeSend records the (user, resource, kind) tuple first. If the row is
// newly inserted, the email is dispatched. On send failure we log and leave
// the log row in place (intentional: we don't want a transient SMTP outage to
// cause repeated retries that may eventually flood the user).
func (s *NotificationService) maybeSend(ctx context.Context, userID, resourceID uuid.UUID, resourceType, kind, to, subject, body string) {
	fresh, err := s.queries.MarkNotificationSent(ctx, userID, resourceID, resourceType, kind)
	if err != nil {
		log.Printf("notification: mark sent failed (user=%s resource=%s/%s kind=%s): %v", userID, resourceType, resourceID, kind, err)
		return
	}
	if !fresh {
		return
	}
	if err := s.mailer.Send(to, subject, body); err != nil {
		log.Printf("notification: send failed (user=%s resource=%s/%s kind=%s): %v", userID, resourceType, resourceID, kind, err)
		return
	}
	log.Printf("notification: sent (user=%s resource=%s/%s kind=%s)", userID, resourceType, resourceID, kind)
}

// parseHHMM parses "HH:MM" strings such as "09:00" into hours/minutes.
func parseHHMM(s string) (int, int, bool) {
	var h, m int
	n, err := fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil || n != 2 {
		return 0, 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}
