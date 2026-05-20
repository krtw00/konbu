package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type SubscriptionHandler struct {
	subSvc *service.SubscriptionService
}

func NewSubscriptionHandler(subSvc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subSvc: subSvc}
}

func (h *SubscriptionHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/sync", h.sync)

	return r
}

func (h *SubscriptionHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	subs, err := h.subSvc.List(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, subs)
}

func (h *SubscriptionHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	sub, err := h.subSvc.Create(r.Context(), user.ID, req.Name, req.ICalURL, req.Color)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, sub)
}

func (h *SubscriptionHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.subSvc.Delete(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *SubscriptionHandler) sync(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	sub, err := h.subSvc.SyncByID(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, sub)
}
