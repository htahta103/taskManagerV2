package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds runtime settings loaded from the environment.
type Config struct {
	Port         int
	DatabaseURL  string
	CORSOrigin   string
	Environment  string
	JWTSecret    string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			port = n
		}
	}
	env := getenvDefault("APP_ENV", "development")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" && env == "development" {
		secret = "dev-insecure-jwt-secret-change-me-32b!!"
	}
	accessMin := 15
	if v := os.Getenv("JWT_ACCESS_TTL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			accessMin = n
		}
	}
	refreshDays := 30
	if v := os.Getenv("JWT_REFRESH_TTL_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			refreshDays = n
		}
	}
	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CORSOrigin:  os.Getenv("CORS_ORIGIN"),
		Environment: env,
		JWTSecret:   secret,
		AccessTTL:   time.Duration(accessMin) * time.Minute,
		RefreshTTL:  time.Duration(refreshDays) * 24 * time.Hour,
	}
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
