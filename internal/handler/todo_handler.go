package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type TodoHandler struct {
	todoSvc *service.TodoService
}

func NewTodoHandler(todoSvc *service.TodoService) *TodoHandler {
	return &TodoHandler{todoSvc: todoSvc}
}

func (h *TodoHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Patch("/{id}/done", h.markDone)
	r.Patch("/{id}/reopen", h.reopen)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	params := parseListParams(r)

	result, err := h.todoSvc.ListTodos(r.Context(), user.ID, params)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	todo, err := h.todoSvc.GetTodo(r.Context(), id, user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, todo)
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateTodoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	todo, err := h.todoSvc.CreateTodo(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, todo)
}

func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.UpdateTodoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	todo, err := h.todoSvc.UpdateTodo(r.Context(), id, user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, todo)
}

func (h *TodoHandler) markDone(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.todoSvc.MarkDone(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *TodoHandler) reopen(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.todoSvc.Reopen(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.todoSvc.DeleteTodo(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
