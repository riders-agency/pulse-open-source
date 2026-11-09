package kernel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type runClaim struct {
	RunID                     string
	WorkflowVersionID         string
	ProjectID                 string
	ReplayedFromWorkflowRunID string
	DefinitionJSON            []byte
	InputJSON                 []byte
	Owner                     string
	FencingToken              int64
	ClaimCount                int
	LeaseDuration             time.Duration
	HeartbeatInterval         time.Duration
}

type claimExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func defaultRunClaimOwner() string {
	host, _ := os.Hostname()
	if host == "" {
		host = "unknown-host"
	}
	return fmt.Sprintf("%s:%d:%s", host, os.Getpid(), uuid.NewString())
}

func (k *Kernel) claimNextRun(ctx context.Context) (runClaim, bool) {
	leaseSeconds := int(k.runLeaseDuration.Seconds())
	if leaseSeconds <= 0 {
		leaseSeconds = int(defaultRunLeaseDuration.Seconds())
	}
	claim := runClaim{
		Owner:             k.runClaimOwner,
		LeaseDuration:     k.runLeaseDuration,
		HeartbeatInterval: k.runHeartbeatInterval,
	}
	err := k.pool.QueryRow(ctx, `
WITH candidate AS (
  SELECT workflow_run_id
    FROM workflow_runs
   WHERE (
         status = 'pending'
         OR (status = 'running' AND (run_lease_expires_at IS NULL OR run_lease_expires_at < now()))
       )
     AND (NOT $3::boolean OR workflow_run_id = ANY($4::uuid[]))
   ORDER BY CASE WHEN status = 'pending' THEN 0 ELSE 1 END, workflow_run_id
   FOR UPDATE SKIP LOCKED
   LIMIT 1
),
claimed AS (
  UPDATE workflow_runs wr
     SET status = 'running',
         started_at = COALESCE(started_at, now()),
         run_claim_owner = $1,
         run_lease_expires_at = now() + ($2 * interval '1 second'),
         run_heartbeat_at = now(),
         run_fencing_token = wr.run_fencing_token + 1,
         run_claim_count = wr.run_claim_count + 1
    FROM candidate
   WHERE wr.workflow_run_id = candidate.workflow_run_id
	  RETURNING wr.workflow_run_id, wr.workflow_version_id, wr.project_id::text,
	            COALESCE(wr.replayed_from_workflow_run_id::text, '') AS replayed_from_workflow_run_id,
	            wr.input_json, wr.run_fencing_token, wr.run_claim_count
)
SELECT c.workflow_run_id::text, c.workflow_version_id::text, c.project_id,
       c.replayed_from_workflow_run_id, c.input_json,
       c.run_fencing_token, c.run_claim_count, wv.definition_json
  FROM claimed c
  JOIN workflow_versions wv ON wv.workflow_version_id = c.workflow_version_id
`, claim.Owner, leaseSeconds, k.runClaimFilterActive, k.runClaimAllowlist).Scan(
		&claim.RunID,
		&claim.WorkflowVersionID,
		&claim.ProjectID,
		&claim.ReplayedFromWorkflowRunID,
		&claim.InputJSON,
		&claim.FencingToken,
		&claim.ClaimCount,
		&claim.DefinitionJSON,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return runClaim{}, false
	}
	if err != nil {
		slog.Error("workflow run claim failed", slog.Any("err", err))
		return runClaim{}, false
	}
	return claim, true
}

func (c runClaim) heartbeat(ctx context.Context, exec claimExecutor) bool {
	leaseSeconds := int(c.LeaseDuration.Seconds())
	if leaseSeconds <= 0 {
		leaseSeconds = int(defaultRunLeaseDuration.Seconds())
	}
	tag, err := exec.Exec(ctx, `
UPDATE workflow_runs
   SET run_heartbeat_at = now(),
       run_lease_expires_at = now() + ($4 * interval '1 second')
 WHERE workflow_run_id = $1
   AND run_claim_owner = $2
   AND run_fencing_token = $3
   AND status = 'running'`, c.RunID, c.Owner, c.FencingToken, leaseSeconds)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		slog.Warn("workflow run heartbeat failed", slog.String("workflow_run_id", c.RunID), slog.Any("err", err))
		return false
	}
	return tag.RowsAffected() == 1
}

func (c runClaim) startHeartbeat(ctx context.Context, exec claimExecutor) (context.Context, func()) {
	interval := c.HeartbeatInterval
	if interval <= 0 {
		interval = defaultRunHeartbeatInterval
	}
	actionCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-actionCtx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				if !c.heartbeat(actionCtx, exec) {
					cancel()
					return
				}
			}
		}
	}()
	return actionCtx, func() {
		close(done)
		cancel()
	}
}

func (c runClaim) waitForRetry(ctx context.Context, exec claimExecutor, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	interval := c.HeartbeatInterval
	if interval <= 0 {
		interval = defaultRunHeartbeatInterval
	}
	if interval > delay {
		interval = delay
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case <-timer.C:
			return c.heartbeat(ctx, exec)
		case <-ticker.C:
			if !c.heartbeat(ctx, exec) {
				return false
			}
		}
	}
}

func (c runClaim) markFailed(ctx context.Context, exec claimExecutor, message string) bool {
	reason := "action_failed"
	tag, err := exec.Exec(ctx, `
UPDATE workflow_runs
   SET status = 'failed',
       finished_at = now(),
       error = $2,
       terminal_reason = $5,
       last_state_transition_at = now(),
       run_lease_expires_at = NULL,
       run_heartbeat_at = now()
 WHERE workflow_run_id = $1
   AND run_claim_owner = $3
   AND run_fencing_token = $4
   AND status = 'running'`, c.RunID, message, c.Owner, c.FencingToken, reason)
	if err != nil {
		slog.Warn("workflow run fenced failure update failed", slog.String("workflow_run_id", c.RunID), slog.Any("err", err))
		return false
	}
	return tag.RowsAffected() == 1
}

func (c runClaim) markFailedWithOutput(ctx context.Context, exec claimExecutor, message, outputJSON, reason string) bool {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "action_failed"
	}
	tag, err := exec.Exec(ctx, `
UPDATE workflow_runs
   SET status = 'failed',
       finished_at = now(),
       output_json = $2::jsonb,
       error = $3,
       terminal_reason = $6,
       last_state_transition_at = now(),
       run_lease_expires_at = NULL,
       run_heartbeat_at = now()
 WHERE workflow_run_id = $1
   AND run_claim_owner = $4
   AND run_fencing_token = $5
   AND status = 'running'`, c.RunID, outputJSON, message, c.Owner, c.FencingToken, reason)
	if err != nil {
		slog.Warn("workflow run fenced failure output update failed", slog.String("workflow_run_id", c.RunID), slog.Any("err", err))
		return false
	}
	return tag.RowsAffected() == 1
}

func (c runClaim) markWaiting(ctx context.Context, exec claimExecutor, reason string, deadline *time.Time) bool {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "manual_approval_required"
	}
	tag, err := exec.Exec(ctx, `
UPDATE workflow_runs
   SET status = 'waiting',
       finished_at = NULL,
       error = NULL,
       waiting_reason = $4,
       waiting_deadline_at = $5,
       wake_reason = '',
       last_state_transition_at = now(),
       run_lease_expires_at = NULL,
       run_heartbeat_at = now()
 WHERE workflow_run_id = $1
   AND run_claim_owner = $2
   AND run_fencing_token = $3
   AND status = 'running'`, c.RunID, c.Owner, c.FencingToken, reason, deadline)
	if err != nil {
		slog.Warn("workflow run fenced waiting update failed", slog.String("workflow_run_id", c.RunID), slog.Any("err", err))
		return false
	}
	return tag.RowsAffected() == 1
}

func (c runClaim) markSucceeded(ctx context.Context, exec claimExecutor, outputJSON string, clearError bool) bool {
	query := `
UPDATE workflow_runs
   SET status = 'succeeded',
       finished_at = now(),
       output_json = $2::jsonb,
       error = CASE WHEN $5 THEN NULL ELSE error END,
       terminal_reason = 'completed',
       last_state_transition_at = now(),
       run_lease_expires_at = NULL,
       run_heartbeat_at = now()
 WHERE workflow_run_id = $1
   AND run_claim_owner = $3
   AND run_fencing_token = $4
   AND status = 'running'`
	tag, err := exec.Exec(ctx, query, c.RunID, outputJSON, c.Owner, c.FencingToken, clearError)
	if err != nil {
		slog.Warn("workflow run fenced success update failed", slog.String("workflow_run_id", c.RunID), slog.Any("err", err))
		return false
	}
	return tag.RowsAffected() == 1
}

func (c runClaim) insertRunEvent(ctx context.Context, exec claimExecutor, actionKey, eventType, message, payloadJSON string) {
	if payloadJSON == "" {
		payloadJSON = `{}`
	}
	_, _ = exec.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
SELECT workflow_run_id, NULLIF($2, ''), $3, $4, $5::jsonb, clock_timestamp()
  FROM workflow_runs
 WHERE workflow_run_id = $1
   AND run_claim_owner = $6
   AND run_fencing_token = $7`,
		c.RunID, actionKey, eventType, message, payloadJSON, c.Owner, c.FencingToken)
}
