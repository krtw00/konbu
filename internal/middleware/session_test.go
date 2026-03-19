package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krtw00/konbu/internal/config"
)

func TestSetSessionCookieUsesFirebaseSessionName(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "https://app.example.com/api/v1/auth/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	SetSessionCookie(rec, req, "user-123", "secret", &config.Config{BaseURL: "https://app.example.com"})

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	if cookies[0].Name != sessionCookieName {
		t.Fatalf("expected cookie name %q, got %q", sessionCookieName, cookies[0].Name)
	}
	if !cookies[0].HttpOnly {
		t.Fatal("expected session cookie to be HttpOnly")
	}
	if !cookies[0].Secure {
		t.Fatal("expected session cookie to be Secure")
	}
	if cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookies[0].SameSite)
	}
}

func TestSetSessionCookieUsesSameSiteNoneForCrossOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "https://api.example.com/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	SetSessionCookie(rec, req, "user-123", "secret", &config.Config{BaseURL: "https://app.example.com"})

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	if cookies[0].SameSite != http.SameSiteNoneMode {
		t.Fatalf("expected SameSite=None, got %v", cookies[0].SameSite)
	}
}

func TestClearSessionCookieExpiresFirebaseSessionCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "https://app.example.com/api/v1/auth/logout", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	ClearSessionCookie(rec, req, &config.Config{BaseURL: "https://app.example.com"})

	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "__session=") {
		t.Fatalf("expected __session cookie to be cleared, got %q", setCookie)
	}
	if !strings.Contains(setCookie, "Max-Age=0") && !strings.Contains(setCookie, "Max-Age=-1") {
		t.Fatalf("expected Max-Age to expire cookie, got %q", setCookie)
	}
}
