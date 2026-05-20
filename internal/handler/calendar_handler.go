package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type CalendarHandler struct {
	calSvc *service.CalendarService
}

func NewCalendarHandler(calSvc *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{calSvc: calSvc}
}

func (h *CalendarHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *CalendarHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	calendars, err := h.calSvc.ListCalendars(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, calendars)
}

func (h *CalendarHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateCalendarRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	cal, err := h.calSvc.CreateCalendar(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, cal)
}

func (h *CalendarHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	detail, err := h.calSvc.GetCalendar(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, detail)
}

func (h *CalendarHandler) update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req model.UpdateCalendarRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	cal, err := h.calSvc.UpdateCalendar(r.Context(), user.ID, id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, cal)
}

func (h *CalendarHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.calSvc.DeleteCalendar(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
