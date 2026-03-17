package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type EventHandler struct {
	eventSvc *service.EventService
}

func NewEventHandler(eventSvc *service.EventService) *EventHandler {
	return &EventHandler{eventSvc: eventSvc}
}

func (h *EventHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	params := parseListParams(r)

	var result *model.PaginatedResult
	var err error

	if calIDStr := r.URL.Query().Get("calendar_id"); calIDStr != "" {
		calID, parseErr := uuid.Parse(calIDStr)
		if parseErr != nil {
			writeError(w, parseErr)
			return
		}
		result, err = h.eventSvc.ListEventsByCalendar(r.Context(), user.ID, calID, params)
	} else {
		result, err = h.eventSvc.ListEvents(r.Context(), user.ID, params)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *EventHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	event, err := h.eventSvc.GetEvent(r.Context(), id, user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, event)
}

func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	event, err := h.eventSvc.CreateEvent(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, event)
}

func (h *EventHandler) update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.UpdateEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	event, err := h.eventSvc.UpdateEvent(r.Context(), id, user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, event)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.eventSvc.DeleteEvent(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
