package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/krtw00/konbu/internal/config"
)

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	allowedOrigin := ""
	if cfg.BaseURL != "" {
		if u, err := url.Parse(cfg.BaseURL); err == nil && u.Scheme != "" && u.Host != "" {
			allowedOrigin = u.Scheme + "://" + u.Host
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isAllowedOrigin(origin, allowedOrigin, cfg.DevUser != "") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin, configuredOrigin string, allowLocalhost bool) bool {
	if configuredOrigin != "" && origin == configuredOrigin {
		return true
	}
	if !allowLocalhost {
		return false
	}
	return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
}
