package main

import (
	"log/slog"
	"os"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
	"github.com/ettoreMB/personal_finance_control/api/internal/db"
	"github.com/ettoreMB/personal_finance_control/api/internal/server"
)

func main() {
	cfg := config.Load()

	if _, err := db.Connect(cfg.SQLitePath); err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	app := server.New()

	if err := app.Listen(":3000"); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
