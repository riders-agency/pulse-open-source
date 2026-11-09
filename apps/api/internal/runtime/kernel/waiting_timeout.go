package kernel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (k *Kernel) timeoutWaitingRun(ctx context.Context) bool {
	tx, err := k.pool.Begin(ctx)
	if err != nil {
		slog.Warn("workflow waiting timeout begin failed", slog.Any("err", err))
		return false
	}
	defer tx.Rollback(ctx)

	var runID string
	var waitingReason string
	var deadline time.Time
	err = tx.QueryRow(ctx, `
WITH target AS (
  SELECT workflow_run_id
    FROM workflow_runs
   WHERE status = 'waiting'
     AND waiting_deadline_at IS NOT NULL
     AND waiting_deadline_at <= now()
   ORDER BY waiting_deadline_at, workflow_run_id
   FOR UPDATE SKIP LOCKED
   LIMIT 1
),
updated AS (
  UPDATE workflow_runs wr
     SET status = 'failed',
         finished_at = now(),
         error = 'waiting timeout',
         terminal_reason = 'waiting_timeout',
         wake_reason = 'timeout',
         run_lease_expires_at = NULL,
         last_state_transition_at = now()
    FROM target
   WHERE wr.workflow_run_id = target.workflow_run_id
   RETURNING wr.workflow_run_id::text, wr.waiting_reason::text, wr.waiting_deadline_at
)
SELECT workflow_run_id, waiting_reason, waiting_deadline_at FROM updated`).Scan(&runID, &waitingReason, &deadline)
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	if err != nil {
		slog.Warn("workflow waiting timeout claim failed", slog.Any("err", err))
		return false
	}

	if _, err := tx.Exec(ctx, `
UPDATE action_runs
   SET status = 'failed',
       error = 'waiting timeout',
       terminal_reason = 'waiting_timeout',
       wake_reason = 'timeout',
       finished_at = COALESCE(finished_at, now()),
       last_state_transition_at = now()
 WHERE workflow_run_id = $1
   AND status = 'waiting'`, runID); err != nil {
		slog.Warn("workflow waiting timeout action update failed", slog.String("workflow_run_id", runID), slog.Any("err", err))
		return false
	}

	waitingReason = strings.TrimSpace(waitingReason)
	payload, err := json.Marshal(map[string]string{
		"reason":           "waiting_timeout",
		"waiting_reason":   waitingReason,
		"deadline":         deadline.UTC().Format(time.RFC3339Nano),
		"terminal_reason":  "waiting_timeout",
		"state_transition": "waiting_to_failed",
	})
	if err != nil {
		slog.Warn("workflow waiting timeout payload failed", slog.String("workflow_run_id", runID), slog.Any("err", err))
		return false
	}
	message := "waiting timeout"
	if waitingReason != "" {
		message = fmt.Sprintf("waiting timeout: %s", waitingReason)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json, created_at)
VALUES ($1,'workflow.waiting_timed_out',$2,$3::jsonb, clock_timestamp())`, runID, message, string(payload)); err != nil {
		slog.Warn("workflow waiting timeout event failed", slog.String("workflow_run_id", runID), slog.Any("err", err))
		return false
	}
	if err := tx.Commit(ctx); err != nil {
		slog.Warn("workflow waiting timeout commit failed", slog.String("workflow_run_id", runID), slog.Any("err", err))
		return false
	}
	slog.Warn("workflow waiting timed out", slog.String("workflow_run_id", runID), slog.String("waiting_reason", waitingReason))
	return true
}
