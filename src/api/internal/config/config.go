package config

import (
	"os"
	"time"
)

const defaultSessionTTL = 30 * 24 * time.Hour

type Config struct {
	SQLitePath     string
	RecoverySecret string
	CookieSecure   bool
	SessionTTL     time.Duration
}

func Load() Config {
	return Config{
		SQLitePath:     os.Getenv("SQLITE_PATH"),
		RecoverySecret: os.Getenv("RECOVERY_SECRET"),
		CookieSecure:   os.Getenv("COOKIE_SECURE") == "true",
		SessionTTL:     defaultSessionTTL,
	}
}
