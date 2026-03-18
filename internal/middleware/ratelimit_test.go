package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterSkipsSessionCookie(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	called := 0
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "token"})
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "token"})
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec1.Code != http.StatusOK || rec2.Code != http.StatusOK {
		t.Fatalf("expected authenticated requests to bypass rate limiting, got %d and %d", rec1.Code, rec2.Code)
	}
	if called != 2 {
		t.Fatalf("expected handler to be called twice, got %d", called)
	}
}

func TestRateLimiterLimitsUnauthenticatedRequests(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first request to pass, got %d", rec1.Code)
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be rate limited, got %d", rec2.Code)
	}
}
