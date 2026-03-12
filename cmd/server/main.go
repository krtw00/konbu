package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/handler"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Services
	authSvc := service.NewAuthService(db, cfg)
	tagSvc := service.NewTagService(db)
	toolSvc := service.NewToolService(db)
	memoSvc := service.NewMemoService(db, tagSvc)
	todoSvc := service.NewTodoService(db, tagSvc)
	eventSvc := service.NewEventService(db, tagSvc)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	apiKeyH := handler.NewAPIKeyHandler(authSvc)
	tagH := handler.NewTagHandler(tagSvc)
	toolH := handler.NewToolHandler(toolSvc)
	memoH := handler.NewMemoHandler(memoSvc)
	todoH := handler.NewTodoHandler(todoSvc)
	eventH := handler.NewEventHandler(eventSvc)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logging)

	// Health check (unauthenticated)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Login/logout (unauthenticated)
	r.HandleFunc("/login", middleware.LoginHandler(cfg))
	r.Get("/logout", middleware.LogoutHandler())

	// Static files (unauthenticated — needed for login page favicon/css)
	staticDir := http.Dir("/web/static")
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(staticDir)))

	// Session-protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.SessionAuth(cfg))

		// API v1 routes
		r.Route("/api/v1", func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, cfg.DevUser))

			r.Mount("/auth", authH.Routes())
			r.Mount("/api-keys", apiKeyH.Routes())
			r.Mount("/tags", tagH.Routes())
			r.Mount("/tools", toolH.Routes())
			r.Mount("/memos", memoH.Routes())
			r.Mount("/todos", todoH.Routes())
			r.Mount("/events", eventH.Routes())
		})

		// Web UI
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/web/static/index.html")
		})
	})

	addr := ":" + cfg.Port
	log.Printf("konbu server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
