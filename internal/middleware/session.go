package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/krtw00/konbu/internal/config"
)

const (
	sessionCookieName = "konbu_session"
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

func SetSessionCookie(w http.ResponseWriter, r *http.Request, userID, secret string) {
	token := MakeSessionToken(userID, secret)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookieName,
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}

func GetSessionUserID(r *http.Request, secret string) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	return VerifySessionToken(cookie.Value, secret)
}

func LoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			msg := ""
			if r.URL.Query().Get("error") == "1" {
				msg = "Invalid username or password"
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, loginHTML, msg)
			return
		}

		r.ParseForm()
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if user == cfg.LoginUser && pass == cfg.LoginPass {
			token := MakeSessionToken(user, cfg.SessionSecret)
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    token,
				Path:     "/",
				MaxAge:   sessionMaxAge,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
			})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/login?error=1", http.StatusFound)
	}
}

func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ClearSessionCookie(w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func SessionAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.LoginUser != "" {
				cookie, err := r.Cookie(sessionCookieName)
				if err == nil {
					if _, ok := VerifySessionToken(cookie.Value, cfg.SessionSecret); ok {
						next.ServeHTTP(w, r)
						return
					}
				}

				if strings.HasPrefix(r.URL.Path, "/api/") {
					http.Error(w, `{"error":{"code":"unauthorized","message":"not logged in"}}`, http.StatusUnauthorized)
					return
				}

				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

const loginHTML = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>konbu — Login</title>
<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{min-height:100vh;display:flex;align-items:center;justify-content:center;background:#0f1a14;font-family:system-ui,sans-serif;color:#c8d6d0}
.login-box{width:320px;padding:40px 32px;background:#1a2e23;border-radius:12px;border:1px solid #2a4a38}
.logo{text-align:center;margin-bottom:24px}
.logo img{width:48px;height:48px}
.logo h1{font-size:1.2rem;margin-top:8px;color:#4ade80}
.field{margin-bottom:16px}
.field label{display:block;font-size:.8rem;color:#8aa898;margin-bottom:4px}
.field input{width:100%%;padding:10px 12px;background:#0f1a14;border:1px solid #2a4a38;border-radius:6px;color:#c8d6d0;font-size:.9rem;outline:none}
.field input:focus{border-color:#4ade80}
button{width:100%%;padding:10px;background:#4ade80;color:#0f1a14;border:none;border-radius:6px;font-size:.9rem;font-weight:600;cursor:pointer}
button:hover{background:#3db86a}
.error{color:#ef4444;font-size:.8rem;text-align:center;margin-bottom:12px}
</style>
</head>
<body>
<div class="login-box">
  <div class="logo"><img src="/static/favicon.svg" alt="konbu"><h1>konbu</h1></div>
  <div class="error">%s</div>
  <form method="POST" action="/login">
    <div class="field"><label>Username</label><input type="text" name="username" autofocus></div>
    <div class="field"><label>Password</label><input type="password" name="password"></div>
    <button type="submit">Log in</button>
  </form>
</div>
</body>
</html>`
