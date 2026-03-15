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
		R2PublicURL:      os.Getenv("R2_PUBLIC_URL"),
	}
}
