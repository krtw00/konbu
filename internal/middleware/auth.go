package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type contextKey string

const userContextKey contextKey = "user"

func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *model.User
			var err error

			// Try Bearer token first (CLI/bot)
			if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				token := strings.TrimPrefix(auth, "Bearer ")
				user, err = authSvc.AuthenticateByAPIKey(r.Context(), token)
				if err != nil {
					http.Error(w, `{"error":{"code":"unauthorized","message":"invalid api key"}}`, http.StatusUnauthorized)
					return
				}
			} else if email := r.Header.Get("X-Forwarded-User"); email != "" {
				// ForwardAuth (Web UI)
				user, err = authSvc.GetOrCreateUser(r.Context(), email)
				if err != nil {
					log.Printf("auth error: %v", err)
					http.Error(w, `{"error":{"code":"unauthorized","message":"authentication failed"}}`, http.StatusUnauthorized)
					return
				}
			} else {
				http.Error(w, `{"error":{"code":"unauthorized","message":"no credentials provided"}}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *model.User {
	u, _ := ctx.Value(userContextKey).(*model.User)
	return u
}
