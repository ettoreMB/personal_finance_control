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

func TestLoadReadsWhatsAppEnv(t *testing.T) {
	t.Setenv("WHATSAPP_INSTANCE_TOKEN", "tok")
	t.Setenv("WHATSAPP_OWNER_PHONE", "5511999999999")
	t.Setenv("EVOLUTION_BASE_URL", "http://localhost:8080")
	t.Setenv("EVOLUTION_API_KEY", "evo-key")
	t.Setenv("EVOLUTION_INSTANCE_ID", "inst-1")
	t.Setenv("JEV_API_KEY", "jev-key")
	t.Setenv("JEV_MODEL", "jev-1.13.0")

	cfg := config.Load()

	if cfg.WhatsAppInstanceToken != "tok" || cfg.WhatsAppOwnerPhone != "5511999999999" {
		t.Fatalf("unexpected whatsapp config: %+v", cfg)
	}
	if cfg.EvolutionBaseURL != "http://localhost:8080" || cfg.EvolutionAPIKey != "evo-key" || cfg.EvolutionInstanceID != "inst-1" {
		t.Fatalf("unexpected evolution config: %+v", cfg)
	}
	if cfg.JevAPIKey != "jev-key" || cfg.JevModel != "jev-1.13.0" {
		t.Fatalf("unexpected jev config: %+v", cfg)
	}
}

func TestLoadDefaultsJevModel(t *testing.T) {
	t.Setenv("JEV_MODEL", "")

	cfg := config.Load()

	if cfg.JevModel != "jev-latest" {
		t.Fatalf("expected default model jev-latest, got %q", cfg.JevModel)
	}
}

func TestLoadDefaultsMigrationsPath(t *testing.T) {
	t.Setenv("MIGRATIONS_PATH", "")

	cfg := config.Load()

	if cfg.MigrationsPath != "migrations" {
		t.Errorf("expected MigrationsPath %q, got %q", "migrations", cfg.MigrationsPath)
	}
}
