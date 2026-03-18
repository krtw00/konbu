package service

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type FeedbackService struct {
	queries  *repository.Queries
	reporter FeedbackReporter
}

func NewFeedbackService(db *sql.DB, reporter FeedbackReporter) *FeedbackService {
	return &FeedbackService{queries: repository.New(db), reporter: reporter}
}

func (s *FeedbackService) Submit(ctx context.Context, req model.CreateFeedbackSubmissionRequest, userID *uuid.UUID, userAgent string) (*model.FeedbackSubmission, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	category := strings.TrimSpace(req.Category)
	message := strings.TrimSpace(req.Message)
	sourcePage := strings.TrimSpace(req.SourcePage)

	if email == "" || !strings.Contains(email, "@") {
		return nil, apperror.BadRequest("valid email is required")
	}
	if message == "" {
		return nil, apperror.BadRequest("message is required")
	}
	if len(message) > 5000 {
		return nil, apperror.BadRequest("message is too long")
	}
	if len(sourcePage) > 300 {
		return nil, apperror.BadRequest("source page is too long")
	}

	switch category {
	case "bug", "feature", "question", "other":
	default:
		category = "other"
	}

	row, err := s.queries.CreateFeedbackSubmission(ctx, userID, email, category, message, sourcePage, userAgent)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	feedback := &model.FeedbackSubmission{
		ID:         row.ID,
		UserID:     row.UserID,
		Email:      row.Email,
		Category:   row.Category,
		Message:    row.Message,
		SourcePage: row.SourcePage,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}

	if s.reporter != nil {
		if err := s.reporter.ReportFeedback(ctx, feedback); err != nil {
			log.Printf("feedback reporter error: %v", err)
		}
	}

	return feedback, nil
}
