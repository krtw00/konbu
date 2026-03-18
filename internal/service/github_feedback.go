package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/model"
)

type FeedbackReporter interface {
	ReportFeedback(ctx context.Context, submission *model.FeedbackSubmission) error
}

type GitHubFeedbackReporter struct {
	client  *http.Client
	baseURL string
	token   string
	owner   string
	repo    string
	labels  []string
}

var feedbackEmailPattern = regexp.MustCompile(`(?i)[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}`)
var feedbackURLPattern = regexp.MustCompile(`(?i)\b(?:https?://|www\.)\S+`)
var feedbackPhonePattern = regexp.MustCompile(`(?m)(?:\+?\d[\d().\-\s]{8,}\d)`)
var feedbackIPv4Pattern = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
var feedbackSecretPattern = regexp.MustCompile(`(?i)\b(?:sk-[A-Z0-9_-]{12,}|gh[pousr]_[A-Z0-9_]{12,}|github_pat_[A-Z0-9_]{12,}|pat_[A-Z0-9_]{12,}|xox[baprs]-[A-Z0-9-]{12,}|bearer\s+[A-Z0-9._-]{12,})\b`)

func NewGitHubFeedbackReporter(cfg *config.Config) *GitHubFeedbackReporter {
	if cfg == nil || cfg.GitHubFeedbackToken == "" || cfg.GitHubFeedbackRepo == "" {
		return nil
	}

	owner, repo, ok := strings.Cut(strings.TrimSpace(cfg.GitHubFeedbackRepo), "/")
	if !ok || owner == "" || repo == "" {
		log.Printf("feedback github reporter disabled: invalid repo %q", cfg.GitHubFeedbackRepo)
		return nil
	}

	var labels []string
	for _, part := range strings.Split(cfg.GitHubFeedbackLabels, ",") {
		label := strings.TrimSpace(part)
		if label != "" {
			labels = append(labels, label)
		}
	}

	return &GitHubFeedbackReporter{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://api.github.com",
		token:   cfg.GitHubFeedbackToken,
		owner:   owner,
		repo:    repo,
		labels:  labels,
	}
}

func (r *GitHubFeedbackReporter) ReportFeedback(ctx context.Context, submission *model.FeedbackSubmission) error {
	if r == nil || submission == nil {
		return nil
	}

	payload := map[string]any{
		"title": githubFeedbackIssueTitle(submission),
		"body":  githubFeedbackIssueBody(submission),
	}
	if len(r.labels) > 0 {
		payload["labels"] = r.labels
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/repos/%s/%s/issues", r.baseURL, r.owner, r.repo), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		resp, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("github issue create failed: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(resp)))
	}

	return nil
}

func githubFeedbackIssueTitle(submission *model.FeedbackSubmission) string {
	return fmt.Sprintf("[feedback/%s] %s", submission.Category, submission.ID.String()[:8])
}

func githubFeedbackIssueBody(submission *model.FeedbackSubmission) string {
	return strings.TrimSpace(fmt.Sprintf(`
A new anonymized feedback submission was received from konbu.

- Category: %s
- Submitted at: %s
- Internal ID: %s

## Message

%s

---
Reporter metadata such as email address, user ID, user agent, and source page was intentionally omitted.
`, submission.Category, submission.CreatedAt.UTC().Format(time.RFC3339), submission.ID.String(), sanitizeFeedbackMessage(submission.Message)))
}

func sanitizeFeedbackMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "(empty)"
	}

	trimmed = strings.ReplaceAll(trimmed, "\r\n", "\n")
	trimmed = feedbackEmailPattern.ReplaceAllString(trimmed, "[redacted-email]")
	trimmed = feedbackURLPattern.ReplaceAllString(trimmed, "[redacted-url]")
	trimmed = feedbackIPv4Pattern.ReplaceAllString(trimmed, "[redacted-ip]")
	trimmed = feedbackPhonePattern.ReplaceAllString(trimmed, "[redacted-phone]")
	trimmed = feedbackSecretPattern.ReplaceAllString(trimmed, "[redacted-secret]")
	trimmed = strings.TrimSpace(trimmed)
	if len(trimmed) > 1200 {
		trimmed = strings.TrimSpace(trimmed[:1200]) + "\n\n[truncated]"
	}
	return trimmed
}
