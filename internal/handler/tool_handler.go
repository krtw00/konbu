package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type ToolHandler struct {
	toolSvc *service.ToolService
}

func NewToolHandler(toolSvc *service.ToolService) *ToolHandler {
	return &ToolHandler{toolSvc: toolSvc}
}

func (h *ToolHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Post("/refresh-icons", h.refreshIcons)
	r.Put("/reorder", h.reorder)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *ToolHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	tools, err := h.toolSvc.ListTools(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, tools)
}

func (h *ToolHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateToolRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tool, err := h.toolSvc.CreateTool(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, tool)
}

func (h *ToolHandler) update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.UpdateToolRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	tool, err := h.toolSvc.UpdateTool(r.Context(), id, user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, tool)
}

func (h *ToolHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.toolSvc.DeleteTool(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *ToolHandler) refreshIcons(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	count, err := h.toolSvc.RefreshToolIcons(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, map[string]int{"updated": count})
}

func (h *ToolHandler) reorder(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.ReorderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.toolSvc.ReorderTools(r.Context(), user.ID, req.Order); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
