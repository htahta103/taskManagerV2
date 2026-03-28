package config

import (
	"testing"
)

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("CORS_ORIGIN", "")
	t.Setenv("APP_ENV", "")
	cfg := Load()
	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Fatalf("expected development, got %q", cfg.Environment)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("APP_ENV", "test")
	cfg := Load()
	if cfg.Port != 3000 {
		t.Fatalf("expected port 3000, got %d", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Fatalf("unexpected DATABASE_URL: %q", cfg.DatabaseURL)
	}
}

func TestLoad_InvalidPortFallsBack(t *testing.T) {
	t.Setenv("PORT", "abc")
	cfg := Load()
	if cfg.Port != 8080 {
		t.Fatalf("expected fallback 8080 for invalid PORT, got %d", cfg.Port)
	}
}
