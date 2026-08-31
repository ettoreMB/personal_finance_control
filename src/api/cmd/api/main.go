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

	conn, err := db.Connect(cfg.SQLitePath)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	app := server.New(conn, cfg)

	if err := app.Listen(":3000"); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
