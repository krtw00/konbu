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
	searchSvc := service.NewSearchService(db)

	exportSvc := service.NewExportService(db, memoSvc, todoSvc, eventSvc, toolSvc)

	// Handlers
	authH := handler.NewAuthHandler(authSvc, cfg)
	apiKeyH := handler.NewAPIKeyHandler(authSvc)
	tagH := handler.NewTagHandler(tagSvc)
	toolH := handler.NewToolHandler(toolSvc)
	memoH := handler.NewMemoHandler(memoSvc)
	todoH := handler.NewTodoHandler(todoSvc)
	eventH := handler.NewEventHandler(eventSvc)
	searchH := handler.NewSearchHandler(searchSvc)
	exportH := handler.NewExportHandler(exportSvc)
	importH := handler.NewImportHandler(eventSvc)

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

	// Static files (unauthenticated)
	staticDir := http.Dir("web/static")
	r.Handle("/assets/*", http.FileServer(staticDir))
	r.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/favicon.svg")
	})
	r.Get("/hero.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		http.ServeFile(w, r, "web/static/hero.png")
	})

	// Auth public endpoints (no session required)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authH.HandleRegister)
		r.Post("/login", authH.HandleLogin)
		r.Post("/logout", authH.HandleLogout)
		r.Get("/setup-status", authH.HandleSetupStatus)

		// Authenticated endpoints under /auth
		r.Group(func(r chi.Router) {
			r.Use(middleware.SessionAuth(cfg))
			r.Use(middleware.Auth(authSvc, cfg))
			r.Get("/me", authH.HandleGetMe)
			r.Put("/me", authH.HandleUpdateMe)
			r.Post("/change-password", authH.HandleChangePassword)
			r.Get("/settings", authH.HandleGetSettings)
			r.Put("/settings", authH.HandleUpdateSettings)
			r.Post("/delete-account", authH.HandleDeleteAccount)
		})
	})

	// Session-protected API routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.SessionAuth(cfg))

		r.Route("/api/v1", func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, cfg))

			r.Mount("/api-keys", apiKeyH.Routes())
			r.Mount("/tags", tagH.Routes())
			r.Mount("/tools", toolH.Routes())
			r.Mount("/memos", memoH.Routes())
			r.Mount("/todos", todoH.Routes())
			r.Mount("/events", eventH.Routes())
			r.Get("/search", searchH.HandleSearch)
			r.Mount("/export", exportH.Routes())
			r.Mount("/import", importH.Routes())
		})
	})

	// SPA fallback (unauthenticated — login page etc. must be accessible)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/index.html")
	})

	addr := ":" + cfg.Port
	log.Printf("konbu server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
