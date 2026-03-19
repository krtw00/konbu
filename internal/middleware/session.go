package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/krtw00/konbu/internal/config"
)

const (
	sessionCookieName = "__session"
	sessionMaxAge     = 30 * 24 * 3600 // 30 days
)

func MakeSessionToken(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return payload + ":" + hex.EncodeToString(mac.Sum(nil))
}

func VerifySessionToken(token, secret string) (string, bool) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", false
	}
	expected := MakeSessionToken(parts[0], secret)
	return parts[0], hmac.Equal([]byte(token), []byte(expected))
}

func SetSessionCookie(w http.ResponseWriter, r *http.Request, userID, secret string, cfg *config.Config) {
	token := MakeSessionToken(userID, secret)
	sameSite := sessionSameSiteMode(r, cfg)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
	})
}

func ClearSessionCookie(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	sameSite := sessionSameSiteMode(r, cfg)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
	})
}

func GetSessionUserID(r *http.Request, secret string) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	return VerifySessionToken(cookie.Value, secret)
}

func SessionAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func sessionSameSiteMode(r *http.Request, cfg *config.Config) http.SameSite {
	if !(r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
		return http.SameSiteLaxMode
	}

	origin := r.Header.Get("Origin")
	if origin != "" && origin != requestOrigin(r) {
		return http.SameSiteNoneMode
	}

	if cfg != nil && cfg.BaseURL != "" {
		if u, err := url.Parse(cfg.BaseURL); err == nil && u.Scheme != "" && u.Host != "" {
			if u.Scheme+"://"+u.Host != requestOrigin(r) {
				return http.SameSiteNoneMode
			}
		}
	}

	return http.SameSiteLaxMode
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
