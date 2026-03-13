package config

import (
	"os"
)

type Config struct {
	DatabaseURL   string
	Port          string
	DevUser       string
	LoginUser     string
	LoginPass     string
	SessionSecret string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "konbu-dev-secret-change-me"
	}

	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          port,
		DevUser:       os.Getenv("DEV_USER"),
		LoginUser:     os.Getenv("KONBU_USER"),
		LoginPass:     os.Getenv("KONBU_PASS"),
		SessionSecret: sessionSecret,
	}
}
