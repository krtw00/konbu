package config

import "testing"

func TestLoadRequiresSessionSecretOutsideDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "")
	t.Setenv("DEV_USER", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("OPEN_REGISTRATION", "")
	t.Setenv("R2_ENDPOINT", "")
	t.Setenv("R2_BUCKET", "")
	t.Setenv("DEFAULT_AI_PROVIDER", "")
	t.Setenv("OPENAI_ENDPOINT", "")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("ANTHROPIC_ENDPOINT", "")
	t.Setenv("ANTHROPIC_MODEL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing SESSION_SECRET to fail outside development")
	}
}

func TestLoadDefaultsInDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "")
	t.Setenv("DEV_USER", "dev@example.com")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("OPEN_REGISTRATION", "")
	t.Setenv("R2_ENDPOINT", "")
	t.Setenv("R2_BUCKET", "")
	t.Setenv("DEFAULT_AI_PROVIDER", "")
	t.Setenv("OPENAI_ENDPOINT", "")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("ANTHROPIC_ENDPOINT", "")
	t.Setenv("ANTHROPIC_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.SessionSecret != "konbu-dev-secret-change-me" {
		t.Fatalf("unexpected default session secret: %s", cfg.SessionSecret)
	}
	if cfg.R2Endpoint != "https://7a4bdd42ddce5841858a7d6fe6119430.r2.cloudflarestorage.com" {
		t.Fatalf("unexpected default R2 endpoint: %s", cfg.R2Endpoint)
	}
	if cfg.R2Bucket != "konbu-attachments" {
		t.Fatalf("unexpected default R2 bucket: %s", cfg.R2Bucket)
	}
	if cfg.DefaultAIProvider != "openai" || cfg.OpenAIModel != "gpt-4o" {
		t.Fatalf("unexpected AI defaults: provider=%s model=%s", cfg.DefaultAIProvider, cfg.OpenAIModel)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("PORT", "9090")
	t.Setenv("DEV_USER", "dev@example.com")
	t.Setenv("SESSION_SECRET", "secret")
	t.Setenv("OPEN_REGISTRATION", "true")
	t.Setenv("BASE_URL", "https://konbu.example.com")
	t.Setenv("GOOGLE_CLIENT_ID", "gid")
	t.Setenv("GOOGLE_CLIENT_SECRET", "gsecret")
	t.Setenv("WEBHOOK_SECRET", "whsec")
	t.Setenv("KOFI_TOKEN", "kofi")
	t.Setenv("AI_ENCRYPTION_KEY", "enc")
	t.Setenv("R2_ACCESS_KEY_ID", "r2id")
	t.Setenv("R2_SECRET_ACCESS_KEY", "r2secret")
	t.Setenv("R2_ENDPOINT", "https://r2.example.com")
	t.Setenv("R2_BUCKET", "bucket")
	t.Setenv("R2_PUBLIC_URL", "https://cdn.example.com")
	t.Setenv("DEFAULT_AI_PROVIDER", "anthropic")
	t.Setenv("DEFAULT_AI_API_KEY", "api-key")
	t.Setenv("DEFAULT_AI_ENDPOINT", "https://llm.example.com")
	t.Setenv("DEFAULT_AI_MODEL", "model-x")
	t.Setenv("OPENAI_ENDPOINT", "https://openai.example.com")
	t.Setenv("OPENAI_MODEL", "gpt-x")
	t.Setenv("ANTHROPIC_ENDPOINT", "https://anthropic.example.com")
	t.Setenv("ANTHROPIC_MODEL", "claude-x")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://example" || cfg.Port != "9090" || !cfg.OpenRegistration {
		t.Fatalf("unexpected config core fields: %#v", cfg)
	}
	if cfg.GoogleClientID != "gid" || cfg.GoogleSecret != "gsecret" || cfg.BaseURL != "https://konbu.example.com" {
		t.Fatalf("unexpected oauth config: %#v", cfg)
	}
	if cfg.R2Endpoint != "https://r2.example.com" || cfg.R2Bucket != "bucket" || cfg.R2PublicURL != "https://cdn.example.com" {
		t.Fatalf("unexpected attachment config: %#v", cfg)
	}
	if cfg.DefaultAIProvider != "anthropic" || cfg.DefaultAIAPIKey != "api-key" || cfg.DefaultAIModel != "model-x" {
		t.Fatalf("unexpected AI override config: %#v", cfg)
	}
}
