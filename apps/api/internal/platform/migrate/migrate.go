package migrate

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Up(databaseURL, migrationsPath string) error {
	m, err := newMigrator(databaseURL, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func Down(databaseURL, migrationsPath string) error {
	m, err := newMigrator(databaseURL, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

func newMigrator(databaseURL, migrationsPath string) (*migrate.Migrate, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required for migrations")
	}
	abs, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("migrations abs: %w", err)
	}
	src := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(abs),
	}).String()
	m, err := migrate.New(
		src,
		databaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("migrate new: %w", err)
	}
	return m, nil
}
