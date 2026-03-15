package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/handler"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/migrate"
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

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Auto-migrate
	migrationsDir := "sql/migrations"
	if _, err := os.Stat("/migrations"); err == nil {
		migrationsDir = "/migrations"
	}
	if err := migrate.Run(db, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// Services
	r2Svc := service.NewR2Service(cfg)
	authSvc := service.NewAuthService(db, cfg)
	tagSvc := service.NewTagService(db)
	toolSvc := service.NewToolService(db)
	memoSvc := service.NewMemoService(db, tagSvc)
	todoSvc := service.NewTodoService(db, tagSvc)
	eventSvc := service.NewEventService(db, tagSvc)
	searchSvc := service.NewSearchService(db)

	exportSvc := service.NewExportService(db, memoSvc, todoSvc, eventSvc, toolSvc)
	chatSvc := service.NewChatService(db, cfg, memoSvc, todoSvc, eventSvc, searchSvc)

	// Background tasks
	toolSvc.StartIconRefreshLoop(6 * time.Hour)

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
	chatH := handler.NewChatHandler(chatSvc)
	attachH := handler.NewAttachmentHandler(r2Svc)

	// Rate limiters
	apiLimiter := middleware.NewRateLimiter(100, time.Minute)
	authLimiter := middleware.NewRateLimiter(10, time.Minute)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logging)

	// Webhooks (unauthenticated, signature-verified)
	webhookH := handler.NewWebhookHandler(authSvc, cfg.WebhookSecret)
	r.Post("/webhooks/github-sponsors", webhookH.HandleGitHubSponsors)

	// Health check (unauthenticated)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Attachment serving (unauthenticated, for Markdown image display)
	r.Get("/api/v1/attachments/*", attachH.Serve)

	// Static files (unauthenticated, immutable hashed assets)
	staticDir := http.Dir("web/static")
	r.Handle("/assets/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.FileServer(staticDir).ServeHTTP(w, r)
	}))
	r.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/favicon.svg")
	})
	r.Get("/hero.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		http.ServeFile(w, r, "web/static/hero.png")
	})

	// OAuth
	oauthH := handler.NewOAuthHandler(authSvc, cfg)

	// Auth public endpoints (no session required)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(authLimiter.Middleware)
		r.Post("/register", authH.HandleRegister)
		r.Post("/login", authH.HandleLogin)
		r.Post("/logout", authH.HandleLogout)
		r.Get("/setup-status", authH.HandleSetupStatus)
		r.Get("/providers", oauthH.HandleProviders)

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

	// OAuth routes (outside /api/v1, no session)
	r.Get("/auth/google/login", oauthH.HandleGoogleLogin)
	r.Get("/auth/google/callback", oauthH.HandleGoogleCallback)

	// Session-protected API routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.SessionAuth(cfg))

		r.Route("/api/v1", func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, cfg))
			r.Use(apiLimiter.Middleware)

			r.Mount("/api-keys", apiKeyH.Routes())
			r.Mount("/tags", tagH.Routes())
			r.Mount("/tools", toolH.Routes())
			r.Mount("/memos", memoH.Routes())
			r.Mount("/todos", todoH.Routes())
			r.Mount("/events", eventH.Routes())
			r.Get("/search", searchH.HandleSearch)
			r.Mount("/export", exportH.Routes())
			r.Mount("/import", importH.Routes())
			r.Mount("/chat", chatH.Routes())
			r.Mount("/attachments", attachH.UploadRoutes())
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
