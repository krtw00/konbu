package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
	cfg     *config.Config
}

func NewAuthHandler(authSvc *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, cfg: cfg}
}

func (h *AuthHandler) HandleSetupStatus(w http.ResponseWriter, r *http.Request) {
	needsSetup, userCount, err := h.authSvc.NeedsSetup(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, map[string]interface{}{
		"needs_setup":       needsSetup,
		"user_count":        userCount,
		"open_registration": h.cfg.OpenRegistration,
	})
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	needsSetup, _, err := h.authSvc.NeedsSetup(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	if !needsSetup && !h.cfg.OpenRegistration {
		user := middleware.UserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"error": map[string]string{
					"code":    "forbidden",
					"message": "registration is not open",
				},
			})
			return
		}
	}

	created, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}

	writeCreated(w, created)
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	user, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	middleware.SetSessionCookie(w, r, user.ID.String(), h.cfg.SessionSecret, h.cfg)
	writeData(w, user)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w, r, h.cfg)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "logged out"},
	})
}

func (h *AuthHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "not logged in",
			},
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.authSvc.ChangePassword(r.Context(), user.ID, req.OldPassword, req.NewPassword); err != nil {
		writeError(w, err)
		return
	}

	writeData(w, map[string]string{"message": "password changed"})
}

func (h *AuthHandler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{"code": "unauthorized", "message": "not logged in"},
		})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.authSvc.DeleteAccount(r.Context(), user.ID, req.Password); err != nil {
		writeError(w, err)
		return
	}

	middleware.ClearSessionCookie(w, r, h.cfg)
	writeData(w, map[string]string{"message": "account deleted"})
}

func (h *AuthHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	writeData(w, user)
}

func (h *AuthHandler) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
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

func (h *AuthHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	settings, err := h.authSvc.GetSettings(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, json.RawMessage(settings))
}

func (h *AuthHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var settings json.RawMessage
	if err := decodeJSON(r, &settings); err != nil {
		writeError(w, err)
		return
	}
	if err := h.authSvc.UpdateSettings(r.Context(), user.ID, settings); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, json.RawMessage(settings))
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
	r.Get("/calendar-feed", h.calendarFeedStatus)
	r.Post("/calendar-feed", h.rotateCalendarFeed)
	r.Delete("/calendar-feed", h.deleteCalendarFeed)
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

func (h *APIKeyHandler) calendarFeedStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	status, err := h.authSvc.GetCalendarFeedTokenStatus(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, status)
}

func (h *APIKeyHandler) rotateCalendarFeed(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	baseURL := h.baseURL(r)
	token, err := h.authSvc.RotateCalendarFeedToken(r.Context(), user.ID, baseURL)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, token)
}

func (h *APIKeyHandler) deleteCalendarFeed(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.authSvc.DeleteCalendarFeedToken(r.Context(), user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *APIKeyHandler) baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
