package kernel

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUpsertActionAttemptSQLPostgres(t *testing.T) {
	databaseURL := os.Getenv("PULSE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PULSE_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin test transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })

	if _, err := tx.Exec(ctx, `
CREATE TEMP TABLE action_attempts (
  action_run_id uuid NOT NULL,
  attempt_no integer NOT NULL,
  status text NOT NULL,
  input_json jsonb NOT NULL,
  output_json jsonb NOT NULL,
  provider_response_json jsonb NOT NULL,
  error text NOT NULL,
  error_type text NOT NULL,
  started_at timestamptz NOT NULL,
  finished_at timestamptz NOT NULL,
  duration_ms integer NOT NULL,
  PRIMARY KEY (action_run_id, attempt_no)
) ON COMMIT DROP`); err != nil {
		t.Fatalf("create temporary action attempts table: %v", err)
	}

	actionRunID := uuid.New()
	args := []any{actionRunID, 1, "failed", `{}`, `{}`, `{}`, "temporary", "provider"}
	if _, err := tx.Exec(ctx, upsertActionAttemptSQL, args...); err != nil {
		t.Fatalf("insert action attempt: %v", err)
	}
	args[2] = "succeeded"
	args[6] = ""
	args[7] = ""
	if _, err := tx.Exec(ctx, upsertActionAttemptSQL, args...); err != nil {
		t.Fatalf("upsert action attempt: %v", err)
	}

	var status string
	var durationMS int
	if err := tx.QueryRow(ctx, `
SELECT status, duration_ms
FROM action_attempts
WHERE action_run_id = $1 AND attempt_no = 1`, actionRunID).Scan(&status, &durationMS); err != nil {
		t.Fatalf("read action attempt: %v", err)
	}
	if status != "succeeded" {
		t.Fatalf("status = %q, want succeeded", status)
	}
	if durationMS < 1 {
		t.Fatalf("duration_ms = %d, want at least 1", durationMS)
	}
}
