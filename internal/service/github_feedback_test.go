package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/model"
)

func TestSanitizeFeedbackMessageRedactsEmails(t *testing.T) {
	got := sanitizeFeedbackMessage("contact me at reporter@example.com")
	if strings.Contains(got, "reporter@example.com") {
		t.Fatalf("expected email to be redacted, got %q", got)
	}
	if !strings.Contains(got, "[redacted-email]") {
		t.Fatalf("expected redaction marker, got %q", got)
	}
}

func TestSanitizeFeedbackMessageRedactsCommonSensitivePatterns(t *testing.T) {
	input := strings.Join([]string{
		"https://example.com/private?token=abc",
		"Call me at +81 90-1234-5678",
		"My IP is 192.168.0.1",
		"Bearer abcdefghijklmnopqrstuv",
		"sk-abcdefghijklmnopqrstuvwxyz123456",
	}, "\n")

	got := sanitizeFeedbackMessage(input)
	for _, raw := range []string{
		"https://example.com/private?token=abc",
		"+81 90-1234-5678",
		"192.168.0.1",
		"Bearer abcdefghijklmnopqrstuv",
		"sk-abcdefghijklmnopqrstuvwxyz123456",
	} {
		if strings.Contains(got, raw) {
			t.Fatalf("expected pattern to be redacted: %q in %q", raw, got)
		}
	}
	for _, marker := range []string{"[redacted-url]", "[redacted-phone]", "[redacted-ip]", "[redacted-secret]"} {
		if !strings.Contains(got, marker) {
			t.Fatalf("expected marker %q in %q", marker, got)
		}
	}
}

func TestSanitizeFeedbackMessageTruncatesLongText(t *testing.T) {
	got := sanitizeFeedbackMessage(strings.Repeat("a", 1500))
	if len(got) > 1220 {
		t.Fatalf("expected message to be truncated, length=%d", len(got))
	}
	if !strings.Contains(got, "[truncated]") {
		t.Fatalf("expected truncated marker, got %q", got)
	}
}

func TestGitHubFeedbackReporterReportFeedback(t *testing.T) {
	submission := &model.FeedbackSubmission{
		ID:         uuid.MustParse("3a6c53d7-57ea-4c9d-8cf4-03d8b0f07b68"),
		Email:      "reporter@example.com",
		Category:   "bug",
		Message:    "Observed issue while testing. Contact reporter@example.com if needed. See https://example.com/details and call +81 90-1234-5678.",
		SourcePage: "/settings",
		CreatedAt:  time.Date(2026, 3, 18, 8, 0, 0, 0, time.UTC),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/repos/krtw00/konbu/issues" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization: %s", got)
		}

		var payload struct {
			Title  string   `json:"title"`
			Body   string   `json:"body"`
			Labels []string `json:"labels"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}

		if payload.Title != "[feedback/bug] 3a6c53d7" {
			t.Fatalf("unexpected title: %q", payload.Title)
		}
		if strings.Contains(payload.Body, "reporter@example.com") {
			t.Fatalf("issue body should not contain email: %q", payload.Body)
		}
		if strings.Contains(payload.Body, "/settings") {
			t.Fatalf("issue body should not contain source page: %q", payload.Body)
		}
		if !strings.Contains(payload.Body, "[redacted-email]") {
			t.Fatalf("issue body should redact email: %q", payload.Body)
		}
		if strings.Contains(payload.Body, "https://example.com/details") || strings.Contains(payload.Body, "+81 90-1234-5678") {
			t.Fatalf("issue body should not contain raw contact detail or url: %q", payload.Body)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"html_url":"https://github.com/krtw00/konbu/issues/1"}`))
	}))
	defer server.Close()

	reporter := &GitHubFeedbackReporter{
		client:  server.Client(),
		baseURL: server.URL,
		token:   "test-token",
		owner:   "krtw00",
		repo:    "konbu",
	}
	if err := reporter.ReportFeedback(context.Background(), submission); err != nil {
		t.Fatalf("ReportFeedback: %v", err)
	}
}
