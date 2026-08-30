package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Up(migrationsPath, sqlitePath string) error {
	m, err := newMigrate(migrationsPath, sqlitePath)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func Down(migrationsPath, sqlitePath string) error {
	m, err := newMigrate(migrationsPath, sqlitePath)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func newMigrate(migrationsPath, sqlitePath string) (*migrate.Migrate, error) {
	return migrate.New(fmt.Sprintf("file://%s", migrationsPath), fmt.Sprintf("sqlite3://%s", sqlitePath))
}
