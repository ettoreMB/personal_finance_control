package config

import (
	"os"
	"time"
)

const defaultSessionTTL = 30 * 24 * time.Hour

const defaultMigrationsPath = "migrations"

type Config struct {
	SQLitePath            string
	MigrationsPath        string
	RecoverySecret        string
	CookieSecure          bool
	SessionTTL            time.Duration
	WhatsAppInstanceToken string
	WhatsAppOwnerPhone    string
	EvolutionBaseURL      string
	EvolutionAPIKey       string
	EvolutionInstanceID   string
	JevAPIKey             string
	JevModel              string
}

func Load() Config {
	return Config{
		SQLitePath:            os.Getenv("SQLITE_PATH"),
		MigrationsPath:        envOrDefault("MIGRATIONS_PATH", defaultMigrationsPath),
		RecoverySecret:        os.Getenv("RECOVERY_SECRET"),
		CookieSecure:          os.Getenv("COOKIE_SECURE") == "true",
		SessionTTL:            defaultSessionTTL,
		WhatsAppInstanceToken: os.Getenv("WHATSAPP_INSTANCE_TOKEN"),
		WhatsAppOwnerPhone:    os.Getenv("WHATSAPP_OWNER_PHONE"),
		EvolutionBaseURL:      os.Getenv("EVOLUTION_BASE_URL"),
		EvolutionAPIKey:       os.Getenv("EVOLUTION_API_KEY"),
		EvolutionInstanceID:   os.Getenv("EVOLUTION_INSTANCE_ID"),
		JevAPIKey:             os.Getenv("JEV_API_KEY"),
		JevModel:              envOrDefault("JEV_MODEL", "jev-latest"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
