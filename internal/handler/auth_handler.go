package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Get("/", h.getMe)
		r.Put("/", h.updateMe)
	})

	return r
}

func (h *AuthHandler) getMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	writeData(w, user)
}

func (h *AuthHandler) updateMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.UpdateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	updated, err := h.authSvc.UpdateUser(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, updated)
}

// --- API Keys ---

type APIKeyHandler struct {
	authSvc *service.AuthService
}

func NewAPIKeyHandler(authSvc *service.AuthService) *APIKeyHandler {
	return &APIKeyHandler{authSvc: authSvc}
}

func (h *APIKeyHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *APIKeyHandler) list(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	keys, err := h.authSvc.ListAPIKeys(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, keys)
}

func (h *APIKeyHandler) create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.CreateAPIKeyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	key, err := h.authSvc.CreateAPIKey(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, key)
}

func (h *APIKeyHandler) delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.authSvc.DeleteAPIKey(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}
