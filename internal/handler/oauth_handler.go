package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type OAuthHandler struct {
	authSvc *service.AuthService
	cfg     *config.Config
	google  *oauth2.Config
}

func NewOAuthHandler(authSvc *service.AuthService, cfg *config.Config) *OAuthHandler {
	h := &OAuthHandler{authSvc: authSvc, cfg: cfg}
	if cfg.GoogleClientID != "" && cfg.GoogleSecret != "" {
		h.google = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleSecret,
			RedirectURL:  cfg.BaseURL + "/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}
	return h
}

func (h *OAuthHandler) Enabled() bool {
	return h.google != nil
}

func (h *OAuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotFound)
		return
	}
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
	url := h.google.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotFound)
		return
	}

	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	token, err := h.google.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusBadRequest)
		return
	}

	client := h.google.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		http.Error(w, "failed to parse user info", http.StatusInternalServerError)
		return
	}

	user, err := h.authSvc.GetOrCreateUser(r.Context(), strings.ToLower(userInfo.Email))
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	if user.Name == "" && userInfo.Name != "" {
		h.authSvc.UpdateUser(r.Context(), user.ID, model.UpdateUserRequest{Name: userInfo.Name})
	}

	middleware.SetSessionCookie(w, r, user.ID.String(), h.cfg.SessionSecret)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) HandleProviders(w http.ResponseWriter, r *http.Request) {
	providers := map[string]bool{
		"google": h.google != nil,
	}
	writeData(w, providers)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
