package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type MemoHandler struct {
	memoSvc *service.MemoService
}

func NewMemoHandler(memoSvc *service.MemoService) *MemoHandler {
	return &MemoHandler{memoSvc: memoSvc}
}

func (h *MemoHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *MemoHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	params := parseListParams(r)

	result, err := h.memoSvc.ListMemos(r.Context(), user.ID, params)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MemoHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	memo, err := h.memoSvc.GetMemo(r.Context(), id, user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, memo)
}

func (h *MemoHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateMemoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	memo, err := h.memoSvc.CreateMemo(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, memo)
}

func (h *MemoHandler) update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.UpdateMemoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	memo, err := h.memoSvc.UpdateMemo(r.Context(), id, user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, memo)
}

func (h *MemoHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.memoSvc.DeleteMemo(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
