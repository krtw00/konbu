package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL      string
	Port             string
	DevUser          string
	SessionSecret    string
	OpenRegistration bool
	WebhookSecret    string
	GoogleClientID   string
	GoogleSecret     string
	BaseURL          string
	AIEncryptionKey  string
	R2AccessKeyID    string
	R2SecretAccessKey string
	R2Endpoint       string
	R2Bucket         string
	R2PublicURL      string
	KofiToken            string
	OpenAIEndpoint       string
	OpenAIModel          string
	AnthropicEndpoint    string
	AnthropicModel       string
	DefaultAIProvider    string
	DefaultAIAPIKey      string
	DefaultAIEndpoint    string
	DefaultAIModel       string
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "konbu-dev-secret-change-me"
		if os.Getenv("DEV_USER") == "" {
			log.Println("WARNING: SESSION_SECRET is not set. Using insecure default. Set SESSION_SECRET in production.")
		}
	}

	r2Endpoint := os.Getenv("R2_ENDPOINT")
	if r2Endpoint == "" {
		r2Endpoint = "https://7a4bdd42ddce5841858a7d6fe6119430.r2.cloudflarestorage.com"
	}
	r2Bucket := os.Getenv("R2_BUCKET")
	if r2Bucket == "" {
		r2Bucket = "konbu-attachments"
	}

	return &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		Port:             port,
		DevUser:          os.Getenv("DEV_USER"),
		SessionSecret:    sessionSecret,
		OpenRegistration: os.Getenv("OPEN_REGISTRATION") == "true",
		WebhookSecret:    os.Getenv("WEBHOOK_SECRET"),
		GoogleClientID:   os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:     os.Getenv("GOOGLE_CLIENT_SECRET"),
		BaseURL:          os.Getenv("BASE_URL"),
		AIEncryptionKey:  os.Getenv("AI_ENCRYPTION_KEY"),
		R2AccessKeyID:    os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Endpoint:       r2Endpoint,
		R2Bucket:         r2Bucket,
		R2PublicURL:        os.Getenv("R2_PUBLIC_URL"),
		KofiToken:            os.Getenv("KOFI_TOKEN"),
		OpenAIEndpoint:       getEnvDefault("OPENAI_ENDPOINT", "https://api.openai.com/v1/chat/completions"),
		OpenAIModel:          getEnvDefault("OPENAI_MODEL", "gpt-4o"),
		AnthropicEndpoint:    getEnvDefault("ANTHROPIC_ENDPOINT", "https://api.anthropic.com/v1/messages"),
		AnthropicModel:       getEnvDefault("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
		DefaultAIProvider:    getEnvDefault("DEFAULT_AI_PROVIDER", "openai"),
		DefaultAIAPIKey:      os.Getenv("DEFAULT_AI_API_KEY"),
		DefaultAIEndpoint:    os.Getenv("DEFAULT_AI_ENDPOINT"),
		DefaultAIModel:       os.Getenv("DEFAULT_AI_MODEL"),
	}
}
