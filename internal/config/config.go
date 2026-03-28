package config

import (
	"os"
	"strconv"
)

// Config holds runtime settings loaded from the environment.
type Config struct {
	Port         int
	DatabaseURL  string
	CORSOrigin   string
	Environment  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			port = n
		}
	}
	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CORSOrigin:  os.Getenv("CORS_ORIGIN"),
		Environment: getenvDefault("APP_ENV", "development"),
	}
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
