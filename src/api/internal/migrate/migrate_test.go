package migrate_test

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ettoreMB/personal_finance_control/api/internal/migrate"
)

func migrationsDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine caller")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

func TestUpCreatesSchemaAndSeedsCategories(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")

	if err := migrate.Up(migrationsDir(t), dbPath); err != nil {
		t.Fatalf("unexpected error applying migrations: %v", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("unexpected error opening db: %v", err)
	}
	defer conn.Close()

	for _, table := range []string{"users", "categories", "entries"} {
		var name string
		if err := conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table,
		).Scan(&name); err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}

	rows, err := conn.Query("SELECT name FROM categories ORDER BY name")
	if err != nil {
		t.Fatalf("unexpected error querying categories: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("unexpected error scanning category: %v", err)
		}
		got = append(got, name)
	}

	want := []string{"carro", "casa", "comida", "lazer"}
	if !slices.Equal(got, want) {
		t.Fatalf("expected seeded categories %v, got %v", want, got)
	}
}

func TestDownReversesSchemaAndSeed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")

	if err := migrate.Up(migrationsDir(t), dbPath); err != nil {
		t.Fatalf("unexpected error applying migrations: %v", err)
	}

	if err := migrate.Down(migrationsDir(t), dbPath); err != nil {
		t.Fatalf("unexpected error reverting migrations: %v", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("unexpected error opening db: %v", err)
	}
	defer conn.Close()

	for _, table := range []string{"users", "categories", "entries"} {
		var name string
		err := conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table,
		).Scan(&name)
		if err != sql.ErrNoRows {
			t.Fatalf("expected table %q to no longer exist, got err=%v", table, err)
		}
	}
}
