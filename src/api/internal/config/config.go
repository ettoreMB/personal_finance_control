package config

import (
	"os"
	"time"
)

const defaultSessionTTL = 30 * 24 * time.Hour

const defaultMigrationsPath = "migrations"

type Config struct {
	SQLitePath     string
	MigrationsPath string
	RecoverySecret string
	CookieSecure   bool
	SessionTTL     time.Duration
}

func Load() Config {
	return Config{
		SQLitePath:     os.Getenv("SQLITE_PATH"),
		MigrationsPath: envOrDefault("MIGRATIONS_PATH", defaultMigrationsPath),
		RecoverySecret: os.Getenv("RECOVERY_SECRET"),
		CookieSecure:   os.Getenv("COOKIE_SECURE") == "true",
		SessionTTL:     defaultSessionTTL,
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
