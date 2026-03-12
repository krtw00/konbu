package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type TagHandler struct {
	tagSvc *service.TagService
}

func NewTagHandler(tagSvc *service.TagService) *TagHandler {
	return &TagHandler{tagSvc: tagSvc}
}

func (h *TagHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *TagHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	tags, err := h.tagSvc.ListTags(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, tags)
}

func (h *TagHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateTagRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tag, err := h.tagSvc.CreateTag(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, tag)
}

func (h *TagHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.UpdateTagRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tag, err := h.tagSvc.UpdateTag(r.Context(), id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, tag)
}

func (h *TagHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.tagSvc.DeleteTag(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
