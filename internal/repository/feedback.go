package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FeedbackSubmission struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	Email      string
	Category   string
	Message    string
	SourcePage string
	UserAgent  string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (q *Queries) CreateFeedbackSubmission(ctx context.Context, userID *uuid.UUID, email, category, message, sourcePage, userAgent string) (FeedbackSubmission, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO feedback_submissions (user_id, email, category, message, source_page, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, email, category, message, source_page, user_agent, status, created_at, updated_at`,
		userID, email, category, message, sourcePage, userAgent)
	var fb FeedbackSubmission
	err := row.Scan(&fb.ID, &fb.UserID, &fb.Email, &fb.Category, &fb.Message, &fb.SourcePage, &fb.UserAgent, &fb.Status, &fb.CreatedAt, &fb.UpdatedAt)
	return fb, err
}
