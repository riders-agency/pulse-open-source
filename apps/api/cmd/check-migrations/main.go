package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/platform/migrate"
)

func main() {
	var migrationsPath string
	var databaseURL string
	var forbidDemoTenant bool
	var verifyIdempotent bool
	var verifyDown bool
	var verifyIntegrity bool
	flag.StringVar(&migrationsPath, "migrations", "../../db/migrations", "path to migration files")
	flag.StringVar(&databaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	flag.BoolVar(&forbidDemoTenant, "forbid-demo-tenant", false, "fail if the active demo tenant exists after migrations")
	flag.BoolVar(&verifyIdempotent, "verify-idempotent", false, "run migrations up twice against the same database")
	flag.BoolVar(&verifyDown, "verify-down", false, "run all down migrations and then migrate back up")
	flag.BoolVar(&verifyIntegrity, "verify-integrity", false, "check invalid indexes and unvalidated constraints after migrating up")
	flag.Parse()

	if strings.TrimSpace(databaseURL) == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}
	if strings.TrimSpace(migrationsPath) == "" {
		fmt.Fprintln(os.Stderr, "migrations path is required")
		os.Exit(2)
	}
	if err := migrateUpAndValidate(databaseURL, migrationsPath, forbidDemoTenant, verifyIntegrity); err != nil {
		fmt.Fprintf(os.Stderr, "migration check failed: %v\n", err)
		os.Exit(1)
	}
	if verifyIdempotent {
		if err := migrateUpAndValidate(databaseURL, migrationsPath, forbidDemoTenant, verifyIntegrity); err != nil {
			fmt.Fprintf(os.Stderr, "migration idempotency check failed: %v\n", err)
			os.Exit(1)
		}
	}
	if verifyDown {
		if err := migrate.Down(databaseURL, migrationsPath); err != nil {
			fmt.Fprintf(os.Stderr, "migration down check failed: %v\n", err)
			os.Exit(1)
		}
		if err := migrateUpAndValidate(databaseURL, migrationsPath, forbidDemoTenant, verifyIntegrity); err != nil {
			fmt.Fprintf(os.Stderr, "migration down/up replay failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func migrateUpAndValidate(databaseURL, migrationsPath string, forbidDemoTenant bool, verifyIntegrity bool) error {
	if err := migrate.Up(databaseURL, migrationsPath); err != nil {
		return err
	}
	if forbidDemoTenant {
		if err := checkNoDemoTenant(databaseURL); err != nil {
			return fmt.Errorf("demo tenant check: %w", err)
		}
	}
	if verifyIntegrity {
		if err := checkSchemaIntegrity(databaseURL); err != nil {
			return fmt.Errorf("schema integrity: %w", err)
		}
	}
	return nil
}

func checkNoDemoTenant(databaseURL string) error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM users
   WHERE email = 'demo@riders-pulse.local'::citext
     AND deleted_at IS NULL
)`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("active demo@riders-pulse.local exists")
	}
	return nil
}

func checkSchemaIntegrity(databaseURL string) error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	var invalidIndexes int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)::int
  FROM pg_index i
  JOIN pg_class c ON c.oid = i.indexrelid
  JOIN pg_namespace n ON n.oid = c.relnamespace
 WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
   AND NOT i.indisvalid`).Scan(&invalidIndexes); err != nil {
		return err
	}
	if invalidIndexes != 0 {
		return fmt.Errorf("invalid indexes: %d", invalidIndexes)
	}

	var unvalidatedConstraints int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)::int
  FROM pg_constraint c
  JOIN pg_namespace n ON n.oid = c.connamespace
 WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
   AND NOT c.convalidated`).Scan(&unvalidatedConstraints); err != nil {
		return err
	}
	if unvalidatedConstraints != 0 {
		return fmt.Errorf("unvalidated constraints: %d", unvalidatedConstraints)
	}
	return nil
}
