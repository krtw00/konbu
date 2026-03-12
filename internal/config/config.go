package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL   string
	Port          string
	AdminEmail    string
	AllowedEmails []string
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

	allowedRaw := os.Getenv("ALLOWED_EMAILS")
	if allowedRaw == "" {
		allowedRaw = "*"
	}

	var allowed []string
	if allowedRaw != "*" {
		for _, e := range strings.Split(allowedRaw, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				allowed = append(allowed, e)
			}
		}
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "konbu-dev-secret-change-me"
	}

	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          port,
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AllowedEmails: allowed,
		DevUser:       os.Getenv("DEV_USER"),
		LoginUser:     os.Getenv("KONBU_USER"),
		LoginPass:     os.Getenv("KONBU_PASS"),
		SessionSecret: sessionSecret,
	}
}

func (c *Config) IsEmailAllowed(email string) bool {
	if len(c.AllowedEmails) == 0 {
		return true
	}
	for _, e := range c.AllowedEmails {
		if strings.EqualFold(e, email) {
			return true
		}
	}
	return false
}
