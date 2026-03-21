package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type PublishHandler struct {
	svc *service.PublishService
}

func NewPublishHandler(svc *service.PublishService) *PublishHandler {
	return &PublishHandler{svc: svc}
}

func (h *PublishHandler) AuthRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{resourceType}/{id}", h.get)
	r.Put("/{resourceType}/{id}", h.upsert)
	r.Delete("/{resourceType}/{id}", h.delete)
	return r
}

func (h *PublishHandler) PublicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/memo/{slug}/view", h.memoView)
	r.Get("/{resourceType}/{slug}", h.view)
	return r
}

func (h *PublishHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	pub, err := h.svc.Get(r.Context(), user.ID, chi.URLParam(r, "resourceType"), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, pub)
}

func (h *PublishHandler) upsert(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req model.UpsertPublishedResourceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	pub, err := h.svc.Upsert(r.Context(), user.ID, chi.URLParam(r, "resourceType"), id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, pub)
}

func (h *PublishHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Delete(r.Context(), user.ID, chi.URLParam(r, "resourceType"), id); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *PublishHandler) view(w http.ResponseWriter, r *http.Request) {
	pub, err := h.svc.GetPublicMetadata(r.Context(), chi.URLParam(r, "resourceType"), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, pub)
}

func (h *PublishHandler) memoView(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.GetPublicMemoView(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, view)
}
