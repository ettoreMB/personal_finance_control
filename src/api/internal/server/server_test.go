package server_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
	"github.com/ettoreMB/personal_finance_control/api/internal/db"
	"github.com/ettoreMB/personal_finance_control/api/internal/migrate"
	"github.com/ettoreMB/personal_finance_control/api/internal/server"
)

func migrationsDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine caller")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

func newTestApp(t *testing.T, opts ...func(*config.Config)) *fiber.App {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")

	if err := migrate.Up(migrationsDir(t), dbPath); err != nil {
		t.Fatalf("unexpected error applying migrations: %v", err)
	}

	conn, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to db: %v", err)
	}

	cfg := config.Config{SQLitePath: dbPath}
	for _, opt := range opts {
		opt(&cfg)
	}

	return server.New(conn, cfg)
}

func TestHealthzReturnsOK(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
