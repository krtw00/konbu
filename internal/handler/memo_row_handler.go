package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type MemoRowHandler struct {
	rowSvc *service.MemoRowService
}

func NewMemoRowHandler(rowSvc *service.MemoRowService) *MemoRowHandler {
	return &MemoRowHandler{rowSvc: rowSvc}
}

func (h *MemoRowHandler) Routes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Post("/batch", h.batchCreate)
	r.Get("/export", h.export)
	r.Put("/{rowId}", h.update)
	r.Delete("/{rowId}", h.delete)
}

func (h *MemoRowHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	memoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	q := r.URL.Query()
	limit := 100
	offset := 0
	if l, e := strconv.Atoi(q.Get("limit")); e == nil && l > 0 {
		limit = l
	}
	if o, e := strconv.Atoi(q.Get("offset")); e == nil && o >= 0 {
		offset = o
	}

	result, err := h.rowSvc.ListRows(r.Context(), user.ID, memoID, q.Get("sort"), q.Get("order"), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MemoRowHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	memoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.CreateMemoRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
		return
	}

	row, err := h.rowSvc.CreateRow(r.Context(), user.ID, memoID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, row)
}

func (h *MemoRowHandler) update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	memoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	rowID, err := uuid.Parse(chi.URLParam(r, "rowId"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.UpdateMemoRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.rowSvc.UpdateRow(r.Context(), user.ID, memoID, rowID, req); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MemoRowHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	memoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	rowID, err := uuid.Parse(chi.URLParam(r, "rowId"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.rowSvc.DeleteRow(r.Context(), user.ID, memoID, rowID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MemoRowHandler) batchCreate(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	memoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.BatchCreateMemoRowsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
		return
	}

	rows, err := h.rowSvc.BatchCreateRows(r.Context(), user.ID, memoID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, rows)
}

func (h *MemoRowHandler) export(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	memoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=export.csv")

	if err := h.rowSvc.ExportCSV(r.Context(), user.ID, memoID, w); err != nil {
		writeError(w, err)
	}
}
