package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

type DailyHandler struct {
	dailySvc *service.DailyService
}

func NewDailyHandler(dailySvc *service.DailyService) *DailyHandler {
	return &DailyHandler{dailySvc: dailySvc}
}

func (h *DailyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.get)
	return r
}

func (h *DailyHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	q := r.URL.Query()

	from, err := time.Parse(time.RFC3339, q.Get("from"))
	if err != nil {
		writeError(w, apperror.BadRequest("from is required and must be RFC3339"))
		return
	}
	to, err := time.Parse(time.RFC3339, q.Get("to"))
	if err != nil {
		writeError(w, apperror.BadRequest("to is required and must be RFC3339"))
		return
	}

	events, todos, memos, err := h.dailySvc.GetDaily(r.Context(), user.ID, from, to)
	if err != nil {
		writeError(w, err)
		return
	}

	writeData(w, map[string]any{
		"events": events,
		"todos":  todos,
		"memos":  memos,
	})
}
