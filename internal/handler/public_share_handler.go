package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

type PublicShareHandler struct {
	svc *service.PublicShareService
}

func NewPublicShareHandler(svc *service.PublicShareService) *PublicShareHandler {
	return &PublicShareHandler{svc: svc}
}

func (h *PublicShareHandler) AuthRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{resourceType}/{id}", h.get)
	r.Post("/{resourceType}/{id}", h.create)
	r.Delete("/{resourceType}/{id}", h.delete)
	return r
}

func (h *PublicShareHandler) PublicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{token}", h.view)
	return r
}

func (h *PublicShareHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	share, err := h.svc.GetShare(r.Context(), user.ID, chi.URLParam(r, "resourceType"), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, share)
}

func (h *PublicShareHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	share, err := h.svc.CreateShare(r.Context(), user.ID, chi.URLParam(r, "resourceType"), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, share)
}

func (h *PublicShareHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.DeleteShare(r.Context(), user.ID, chi.URLParam(r, "resourceType"), id); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *PublicShareHandler) view(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.GetPublicView(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, view)
}
