package config_test

import (
	"testing"
	"time"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
)

func TestLoadReadsEnvVars(t *testing.T) {
	t.Setenv("SQLITE_PATH", "/tmp/test.sqlite")
	t.Setenv("MIGRATIONS_PATH", "/app/migrations")
	t.Setenv("RECOVERY_SECRET", "s3cr3t")
	t.Setenv("COOKIE_SECURE", "true")

	cfg := config.Load()

	if cfg.SQLitePath != "/tmp/test.sqlite" {
		t.Errorf("expected SQLitePath %q, got %q", "/tmp/test.sqlite", cfg.SQLitePath)
	}
	if cfg.MigrationsPath != "/app/migrations" {
		t.Errorf("expected MigrationsPath %q, got %q", "/app/migrations", cfg.MigrationsPath)
	}
	if cfg.RecoverySecret != "s3cr3t" {
		t.Errorf("expected RecoverySecret %q, got %q", "s3cr3t", cfg.RecoverySecret)
	}
	if !cfg.CookieSecure {
		t.Error("expected CookieSecure to be true")
	}
	if cfg.SessionTTL != 30*24*time.Hour {
		t.Errorf("expected SessionTTL %v, got %v", 30*24*time.Hour, cfg.SessionTTL)
	}
}

func TestLoadDefaultsCookieSecureToFalse(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "")

	cfg := config.Load()

	if cfg.CookieSecure {
		t.Error("expected CookieSecure to default to false")
	}
}

func TestLoadDefaultsMigrationsPath(t *testing.T) {
	t.Setenv("MIGRATIONS_PATH", "")

	cfg := config.Load()

	if cfg.MigrationsPath != "migrations" {
		t.Errorf("expected MigrationsPath %q, got %q", "migrations", cfg.MigrationsPath)
	}
}
