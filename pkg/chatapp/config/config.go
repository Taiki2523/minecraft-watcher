package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port               string
	DatabaseDSN        string
	SessionKey         string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendOrigin     string
	StaticDir          string
}

func Load() (Config, error) {
	cfg := Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseDSN:        getEnv("DB_DSN", "root:password@tcp(127.0.0.1:3306)/chatapp?parseTime=true"),
		SessionKey:         getEnv("SESSION_KEY", "change-me"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		FrontendOrigin:     getEnv("FRONTEND_ORIGIN", ""),
		StaticDir:          getEnv("STATIC_DIR", "web/dist"),
	}

	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return cfg, fmt.Errorf("missing required Google OAuth credentials")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
