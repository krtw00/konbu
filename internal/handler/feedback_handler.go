package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type FeedbackHandler struct {
	svc *service.FeedbackService
}

func NewFeedbackHandler(svc *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{svc: svc}
}

func (h *FeedbackHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.create)
	return r
}

func (h *FeedbackHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateFeedbackSubmissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	feedback, err := h.svc.Submit(r.Context(), req, nil, r.UserAgent())
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, feedback)
}
