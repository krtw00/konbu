package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type contextKey string

const userContextKey contextKey = "user"

func Auth(authSvc *service.AuthService, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *model.User
			var err error

			if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				token := strings.TrimPrefix(auth, "Bearer ")
				user, err = authSvc.AuthenticateByAPIKey(r.Context(), token)
				if err != nil {
					http.Error(w, `{"error":{"code":"unauthorized","message":"invalid api key"}}`, http.StatusUnauthorized)
					return
				}
			} else if userID, ok := GetSessionUserID(r, cfg.SessionSecret); ok {
				id, parseErr := uuid.Parse(userID)
				if parseErr != nil {
					http.Error(w, `{"error":{"code":"unauthorized","message":"invalid session"}}`, http.StatusUnauthorized)
					return
				}
				user, err = authSvc.GetUserByID(r.Context(), id)
				if err != nil {
					http.Error(w, `{"error":{"code":"unauthorized","message":"session user not found"}}`, http.StatusUnauthorized)
					return
				}
			} else if cfg.DevUser != "" {
				user, err = authSvc.GetOrCreateUser(r.Context(), cfg.DevUser)
				if err != nil {
					log.Printf("dev auth error: %v", err)
					http.Error(w, `{"error":{"code":"unauthorized","message":"dev user auth failed"}}`, http.StatusUnauthorized)
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
