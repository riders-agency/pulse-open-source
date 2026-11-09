package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	runssvc "riders.agency/pulse/api/internal/modules/runs/service"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	runsrepo "riders.agency/pulse/api/internal/modules/runs/repository"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/runtime/effects"
	"riders.agency/pulse/api/internal/runtime/kernel"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

type smokeRuntime struct {
	delay                 time.Duration
	actionTypeKey         string
	maxAttempts           int
	failuresBeforeSuccess int
	failureRetryable      bool
	failureErrorType      string
	failureProviderJSON   string
	failureErr            error
	calls                 *int
	waiting               bool
	waitingReason         string
	waitingDeadline       time.Time
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		fail("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fail("connect: %v", err)
	}
	defer pool.Close()

	fixture, err := createFixture(ctx, pool)
	if err != nil {
		fail("fixture: %v", err)
	}
	defer cleanupFixture(context.Background(), pool, fixture.CompanyID)

	if err := checkHeartbeat(ctx, pool, fixture); err != nil {
		fail("heartbeat: %v", err)
	}
	if err := checkReclaimAndFence(ctx, pool, fixture); err != nil {
		fail("reclaim/fence: %v", err)
	}
	if err := checkCancellationSemantics(ctx, pool, fixture); err != nil {
		fail("cancellation: %v", err)
	}
	if err := checkEffectLedger(ctx, pool, fixture); err != nil {
		fail("effect ledger: %v", err)
	}
	if err := checkRetryDeadLetterAndReplay(ctx, pool, fixture); err != nil {
		fail("retry/dead-letter/replay: %v", err)
	}
}

type fixtureIDs struct {
	UserID     uuid.UUID
	CompanyID  uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	VersionID  uuid.UUID
}

func createFixture(ctx context.Context, pool *pgxpool.Pool) (fixtureIDs, error) {
	f := fixtureIDs{
		UserID:     uuid.New(),
		CompanyID:  uuid.New(),
		ProjectID:  uuid.New(),
		WorkflowID: uuid.New(),
		VersionID:  uuid.New(),
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO users (user_id, first_name, display_name, is_activated)
VALUES ($1, 'Runtime Claim Smoke', 'Runtime Claim Smoke', true)`, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO companies (company_id, owner_user_id, slug, name, created_by_user_id)
VALUES ($1, $2, $3, 'Runtime Claim Smoke', $2)`, f.CompanyID, f.UserID, "runtime-claim-"+f.CompanyID.String()); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO projects (project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id)
VALUES ($1, $2, 'Runtime Claim Smoke', $3, $4, 'hash', $5)`,
		f.ProjectID, f.CompanyID, "runtime-claim-"+f.ProjectID.String(), "hook-"+f.ProjectID.String(), f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO company_memberships (company_id, user_id, role)
VALUES ($1, $2, 'owner')`, f.CompanyID, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, role, status)
VALUES ($1, $2, 'owner', 'active')`, f.ProjectID, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	def := `{"nodes":[{"action_key":"claim_smoke","action_type_key":"smoke.action","config":{}}]}`
	if _, err := pool.Exec(ctx, `
INSERT INTO workflows (workflow_id, project_id, name, slug, status, draft_definition_json, created_by_user_id)
VALUES ($1, $2, 'Runtime Claim Smoke', $3, 'active', $4::jsonb, $5)`,
		f.WorkflowID, f.ProjectID, "runtime-claim-"+f.WorkflowID.String(), def, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO workflow_versions (workflow_version_id, project_id, workflow_id, version_no, status, definition_json, created_by_user_id)
VALUES ($1, $2, $3, 1, 'published', $4::jsonb, $5)`,
		f.VersionID, f.ProjectID, f.WorkflowID, def, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	_, err := pool.Exec(ctx, `UPDATE workflows SET current_workflow_version_id = $2 WHERE workflow_id = $1`, f.WorkflowID, f.VersionID)
	return f, err
}

func cleanupFixture(ctx context.Context, pool *pgxpool.Pool, companyID uuid.UUID) {
	_, _ = pool.Exec(ctx, `DELETE FROM companies WHERE company_id = $1`, companyID)
}

func checkHeartbeat(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) error {
	runID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, runID, "heartbeat")); err != nil {
		return err
	}
	runtime := smokeRuntime{delay: 1500 * time.Millisecond}
	k := kernel.New(pool, runtime, kernel.WithRunClaimLease("claim-smoke-heartbeat", time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected heartbeat run to process")
	}
	var status string
	var claimCount int
	var heartbeatSet bool
	if err := pool.QueryRow(ctx, `
SELECT status, run_claim_count, run_heartbeat_at IS NOT NULL
  FROM workflow_runs
 WHERE workflow_run_id = $1`, runID).Scan(&status, &claimCount, &heartbeatSet); err != nil {
		return err
	}
	if status != "succeeded" || claimCount != 1 || !heartbeatSet {
		return fmt.Errorf("unexpected heartbeat run state status=%s claim_count=%d heartbeat=%v", status, claimCount, heartbeatSet)
	}
	return nil
}

func checkReclaimAndFence(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) error {
	runID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, runID, "reclaim")); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
UPDATE workflow_runs
   SET status = 'running',
       run_claim_owner = 'stale-worker',
       run_fencing_token = 1,
       run_claim_count = 1,
       run_lease_expires_at = now() - interval '5 seconds',
       run_heartbeat_at = now() - interval '10 seconds'
 WHERE workflow_run_id = $1`, runID); err != nil {
		return err
	}
	k := kernel.New(pool, smokeRuntime{}, kernel.WithRunClaimLease("claim-smoke-reclaimer", time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected expired lease to be reclaimed")
	}
	tag, err := pool.Exec(ctx, `
UPDATE workflow_runs
   SET status = 'failed', error = 'stale write should not land'
 WHERE workflow_run_id = $1
   AND run_claim_owner = 'stale-worker'
   AND run_fencing_token = 1`, runID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 0 {
		return errors.New("stale owner/token updated reclaimed run")
	}
	var status, owner string
	var token int64
	var claimCount int
	var reclaimedEvents int
	if err := pool.QueryRow(ctx, `
SELECT wr.status, wr.run_claim_owner, wr.run_fencing_token, wr.run_claim_count,
       (SELECT COUNT(*)::int FROM run_events re WHERE re.workflow_run_id = wr.workflow_run_id AND re.event_type = 'workflow.reclaimed')
  FROM workflow_runs wr
 WHERE wr.workflow_run_id = $1`, runID).Scan(&status, &owner, &token, &claimCount, &reclaimedEvents); err != nil {
		return err
	}
	if status != "succeeded" || owner != "claim-smoke-reclaimer" || token != 2 || claimCount != 2 || reclaimedEvents != 1 {
		return fmt.Errorf("unexpected reclaim state status=%s owner=%s token=%d claim_count=%d reclaimed_events=%d",
			status, owner, token, claimCount, reclaimedEvents)
	}
	return nil
}

func checkCancellationSemantics(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) error {
	runs := runssvc.NewCommands(&runsrepo.Runs{P: pool})
	actorCtx := authz.WithSessionClaims(ctx, authz.SessionClaims{
		UserID:    f.UserID,
		ActorType: authz.ActorSession,
	})

	pendingRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, pendingRunID, "cancel-pending")); err != nil {
		return err
	}
	if err := runs.Cancel(actorCtx, pendingRunID, f.UserID); err != nil {
		return fmt.Errorf("cancel pending: %w", err)
	}
	if err := runs.Cancel(actorCtx, pendingRunID, f.UserID); err != nil {
		return fmt.Errorf("cancel pending idempotency: %w", err)
	}
	var pendingStatus, pendingTerminalReason, pendingCancelReason, pendingCancelUser string
	var pendingCancelledAt bool
	var pendingEvents int
	if err := pool.QueryRow(ctx, `
SELECT wr.status, wr.terminal_reason, wr.cancel_reason, COALESCE(wr.cancel_requested_by_user_id::text, ''),
       wr.cancelled_at IS NOT NULL,
       (SELECT COUNT(*)::int FROM run_events re WHERE re.workflow_run_id = wr.workflow_run_id AND re.event_type = 'workflow.cancelled')
  FROM workflow_runs wr
 WHERE wr.workflow_run_id = $1`, pendingRunID).Scan(
		&pendingStatus, &pendingTerminalReason, &pendingCancelReason, &pendingCancelUser,
		&pendingCancelledAt, &pendingEvents,
	); err != nil {
		return err
	}
	if pendingStatus != "cancelled" || pendingTerminalReason != "user_cancel" || pendingCancelReason != "user_cancel" ||
		pendingCancelUser != f.UserID.String() || !pendingCancelledAt || pendingEvents != 1 {
		return fmt.Errorf("unexpected pending cancel state status=%s terminal=%s cancel=%s user=%s cancelled_at=%v events=%d",
			pendingStatus, pendingTerminalReason, pendingCancelReason, pendingCancelUser, pendingCancelledAt, pendingEvents)
	}
	k := kernel.New(pool, smokeRuntime{}, kernel.WithRunClaimLease("cancel-smoke-pending", time.Second, 100*time.Millisecond))
	if k.ProcessOne(ctx) {
		return errors.New("cancelled pending run was claimed")
	}

	runningRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, runningRunID, "cancel-running")); err != nil {
		return err
	}
	k = kernel.New(pool, smokeRuntime{delay: 1500 * time.Millisecond}, kernel.WithRunClaimLease("cancel-smoke-running", 5*time.Second, 100*time.Millisecond))
	done := make(chan bool, 1)
	go func() {
		done <- k.ProcessOne(ctx)
	}()
	if err := waitForRunStatus(ctx, pool, runningRunID, "running", 2*time.Second); err != nil {
		return err
	}
	if err := runs.Cancel(actorCtx, runningRunID, f.UserID); err != nil {
		return fmt.Errorf("cancel running: %w", err)
	}
	select {
	case <-time.After(3 * time.Second):
		return errors.New("timed out waiting for cancelled running action to stop")
	case processed := <-done:
		if processed {
			return errors.New("cancelled running action reported successful processing")
		}
	}
	var runningStatus, runningTerminalReason, runningError string
	var runningEvents int
	if err := pool.QueryRow(ctx, `
SELECT wr.status, wr.terminal_reason, COALESCE(wr.error, ''),
       (SELECT COUNT(*)::int FROM run_events re WHERE re.workflow_run_id = wr.workflow_run_id AND re.event_type = 'workflow.cancelled')
  FROM workflow_runs wr
 WHERE wr.workflow_run_id = $1`, runningRunID).Scan(&runningStatus, &runningTerminalReason, &runningError, &runningEvents); err != nil {
		return err
	}
	if runningStatus != "cancelled" || runningTerminalReason != "user_cancel" || runningError != "cancelled:user_cancel" || runningEvents != 1 {
		return fmt.Errorf("unexpected running cancel state status=%s terminal=%s error=%s events=%d",
			runningStatus, runningTerminalReason, runningError, runningEvents)
	}

	completedRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, completedRunID, "cancel-completed")); err != nil {
		return err
	}
	k = kernel.New(pool, smokeRuntime{}, kernel.WithRunClaimLease("cancel-smoke-completed", time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected completed run to process")
	}
	if err := runs.Cancel(actorCtx, completedRunID, f.UserID); err == nil {
		return errors.New("completed run cancel unexpectedly succeeded")
	}
	var completedStatus, completedTerminalReason string
	if err := pool.QueryRow(ctx, `
SELECT status, terminal_reason FROM workflow_runs WHERE workflow_run_id = $1`, completedRunID).
		Scan(&completedStatus, &completedTerminalReason); err != nil {
		return err
	}
	if completedStatus != "succeeded" || completedTerminalReason != "completed" {
		return fmt.Errorf("unexpected completed cancel state status=%s terminal=%s", completedStatus, completedTerminalReason)
	}

	waitingRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, waitingRunID, "waiting-timeout")); err != nil {
		return err
	}
	waitingRuntime := smokeRuntime{
		waiting:         true,
		waitingReason:   "approval_timeout",
		waitingDeadline: time.Now().Add(-time.Minute),
	}
	k = kernel.New(pool, waitingRuntime, kernel.WithRunClaimLease("cancel-smoke-waiting-timeout", time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected waiting run to process")
	}
	var waitingStatus, waitingReason string
	if err := pool.QueryRow(ctx, `
SELECT status, waiting_reason FROM workflow_runs WHERE workflow_run_id = $1`, waitingRunID).
		Scan(&waitingStatus, &waitingReason); err != nil {
		return err
	}
	if waitingStatus != "waiting" || waitingReason != "approval_timeout" {
		return fmt.Errorf("unexpected waiting state status=%s reason=%s", waitingStatus, waitingReason)
	}
	if !k.ProcessOne(ctx) {
		return errors.New("expected waiting timeout to process")
	}
	var timeoutWorkflowStatus, timeoutWorkflowTerminal, timeoutWorkflowWake, timeoutActionStatus, timeoutActionTerminal, timeoutActionWake string
	var timeoutEvents int
	if err := pool.QueryRow(ctx, `
SELECT wr.status, wr.terminal_reason, wr.wake_reason, ar.status, ar.terminal_reason, ar.wake_reason,
       (SELECT COUNT(*)::int FROM run_events re WHERE re.workflow_run_id = wr.workflow_run_id AND re.event_type = 'workflow.waiting_timed_out')
  FROM workflow_runs wr
  JOIN action_runs ar ON ar.workflow_run_id = wr.workflow_run_id
 WHERE wr.workflow_run_id = $1`, waitingRunID).Scan(
		&timeoutWorkflowStatus, &timeoutWorkflowTerminal, &timeoutWorkflowWake,
		&timeoutActionStatus, &timeoutActionTerminal, &timeoutActionWake, &timeoutEvents,
	); err != nil {
		return err
	}
	if timeoutWorkflowStatus != "failed" || timeoutWorkflowTerminal != "waiting_timeout" || timeoutWorkflowWake != "timeout" ||
		timeoutActionStatus != "failed" || timeoutActionTerminal != "waiting_timeout" || timeoutActionWake != "timeout" || timeoutEvents != 1 {
		return fmt.Errorf("unexpected waiting timeout state workflow=%s/%s/%s action=%s/%s/%s events=%d",
			timeoutWorkflowStatus, timeoutWorkflowTerminal, timeoutWorkflowWake,
			timeoutActionStatus, timeoutActionTerminal, timeoutActionWake, timeoutEvents)
	}
	return nil
}

func waitForRunStatus(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID, want string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		var status string
		err := pool.QueryRow(ctx, `SELECT status FROM workflow_runs WHERE workflow_run_id = $1`, runID).Scan(&status)
		if err != nil {
			return err
		}
		if status == want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for run %s status %q, last status %q", runID, want, status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func checkEffectLedger(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) error {
	runID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, runID, "effect-ledger")); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
UPDATE workflow_runs
   SET status = 'succeeded',
       started_at = COALESCE(started_at, now()),
       finished_at = now()
 WHERE workflow_run_id = $1`, runID); err != nil {
		return err
	}
	actionRunID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO action_runs (
  action_run_id, workflow_run_id, action_key, action_type_key, status,
  attempt_count, max_attempts, input_json, started_at)
VALUES ($1,$2,'effect_smoke','telegram.send_message','running',0,1,'{}'::jsonb,now())`,
		actionRunID, runID); err != nil {
		return err
	}
	ledger := &effects.Ledger{P: pool}
	effectID := effects.StableEffectID(runID.String(), "effect_smoke", 1)
	started, err := ledger.Start(ctx, effects.StartRequest{
		EffectID:       effectID,
		WorkflowRunID:  runID.String(),
		ActionRunID:    actionRunID.String(),
		ActionKey:      "effect_smoke",
		ActionTypeKey:  "telegram.send_message",
		AttemptNo:      1,
		IdempotencyKey: effectID,
		InputJSON:      `{"message":"hello"}`,
		NonIdempotent:  true,
	})
	if err != nil {
		return err
	}
	if started.Existing || started.Status != effects.StatusStarted {
		return fmt.Errorf("unexpected first effect start: %#v", started)
	}
	duplicateStarted, err := ledger.Start(ctx, effects.StartRequest{
		EffectID:       effectID,
		WorkflowRunID:  runID.String(),
		ActionRunID:    actionRunID.String(),
		ActionKey:      "effect_smoke",
		ActionTypeKey:  "telegram.send_message",
		AttemptNo:      1,
		IdempotencyKey: effectID,
		InputJSON:      `{"message":"hello"}`,
		NonIdempotent:  true,
	})
	if err != nil {
		return err
	}
	if !duplicateStarted.Existing || duplicateStarted.Status != effects.StatusStarted {
		return fmt.Errorf("expected duplicate started effect, got %#v", duplicateStarted)
	}
	if err := ledger.Unknown(ctx, effectID, "manual reconciliation required"); err != nil {
		return err
	}
	var unknownStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM effect_ledger WHERE effect_id = $1`, effectID).Scan(&unknownStatus); err != nil {
		return err
	}
	if unknownStatus != effects.StatusUnknown {
		return fmt.Errorf("expected unknown effect, got %s", unknownStatus)
	}

	secondEffectID := effects.StableEffectID(runID.String(), "effect_smoke", 2)
	if _, err := ledger.Start(ctx, effects.StartRequest{
		EffectID:       secondEffectID,
		WorkflowRunID:  runID.String(),
		ActionRunID:    actionRunID.String(),
		ActionKey:      "effect_smoke",
		ActionTypeKey:  "telegram.send_message",
		AttemptNo:      2,
		IdempotencyKey: secondEffectID,
		InputJSON:      `{"message":"hello again"}`,
		NonIdempotent:  true,
	}); err != nil {
		return err
	}
	if err := ledger.Succeeded(ctx, secondEffectID, `{"sent":true}`, `{"provider":"telegram","ok":true}`); err != nil {
		return err
	}
	duplicateSucceeded, err := ledger.Start(ctx, effects.StartRequest{
		EffectID:       secondEffectID,
		WorkflowRunID:  runID.String(),
		ActionRunID:    actionRunID.String(),
		ActionKey:      "effect_smoke",
		ActionTypeKey:  "telegram.send_message",
		AttemptNo:      2,
		IdempotencyKey: secondEffectID,
		InputJSON:      `{"message":"hello again"}`,
		NonIdempotent:  true,
	})
	if err != nil {
		return err
	}
	replayedOutput := map[string]bool{}
	_ = json.Unmarshal([]byte(duplicateSucceeded.OutputJSON), &replayedOutput)
	if !duplicateSucceeded.Existing || duplicateSucceeded.Status != effects.StatusSucceeded || !replayedOutput["sent"] {
		return fmt.Errorf("expected duplicate succeeded replay, got %#v", duplicateSucceeded)
	}
	return nil
}

func checkRetryDeadLetterAndReplay(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) error {
	retryRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, retryRunID, "retry-success")); err != nil {
		return err
	}
	retryRuntime := smokeRuntime{
		actionTypeKey:         "http.request",
		maxAttempts:           3,
		failuresBeforeSuccess: 2,
		failureRetryable:      true,
		failureErrorType:      "provider_http",
		failureProviderJSON:   `{"status":503}`,
		failureErr:            errors.New("provider 503"),
	}
	k := kernel.New(pool, retryRuntime, kernel.WithRunClaimLease("retry-smoke-success", 5*time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected retry success run to process")
	}
	var retryStatus, retryActionStatus string
	var attempts, retryEvents int
	if err := pool.QueryRow(ctx, `
	SELECT wr.status, ar.status, ar.attempt_count,
	       (SELECT COUNT(*)::int FROM run_events re WHERE re.workflow_run_id = wr.workflow_run_id AND re.event_type = 'action.retry_scheduled')
	  FROM workflow_runs wr
	  JOIN action_runs ar ON ar.workflow_run_id = wr.workflow_run_id
	 WHERE wr.workflow_run_id = $1`, retryRunID).Scan(&retryStatus, &retryActionStatus, &attempts, &retryEvents); err != nil {
		return err
	}
	if retryStatus != "succeeded" || retryActionStatus != "succeeded" || attempts != 3 || retryEvents != 2 {
		return fmt.Errorf("unexpected retry success state workflow=%s action=%s attempts=%d retry_events=%d",
			retryStatus, retryActionStatus, attempts, retryEvents)
	}

	deadRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, deadRunID, "dead-letter")); err != nil {
		return err
	}
	deadRuntime := smokeRuntime{
		actionTypeKey:       "max.send_message",
		maxAttempts:         3,
		failureRetryable:    false,
		failureErrorType:    "auth",
		failureProviderJSON: `{"status":401}`,
		failureErr:          errors.New("authorization: bearer secret token=abc user@example.com"),
	}
	k = kernel.New(pool, deadRuntime, kernel.WithRunClaimLease("retry-smoke-dead", 5*time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected dead-letter run to process")
	}
	var deadWorkflowStatus, deadActionStatus, retryClass, deadReason string
	var deadAttempts, deadLetters int
	if err := pool.QueryRow(ctx, `
	SELECT wr.status, ar.status, ar.attempt_count, ar.retry_class, ar.dead_letter_reason,
	       (SELECT COUNT(*)::int FROM workflow_action_dead_letters dl WHERE dl.action_run_id = ar.action_run_id)
	  FROM workflow_runs wr
	  JOIN action_runs ar ON ar.workflow_run_id = wr.workflow_run_id
	 WHERE wr.workflow_run_id = $1`, deadRunID).Scan(&deadWorkflowStatus, &deadActionStatus, &deadAttempts, &retryClass, &deadReason, &deadLetters); err != nil {
		return err
	}
	if deadWorkflowStatus != "failed" || deadActionStatus != "dead_lettered" || deadAttempts != 1 || retryClass != "auth" || deadLetters != 1 {
		return fmt.Errorf("unexpected dead-letter state workflow=%s action=%s attempts=%d class=%s letters=%d",
			deadWorkflowStatus, deadActionStatus, deadAttempts, retryClass, deadLetters)
	}
	if strings.Contains(deadReason, "secret") || strings.Contains(deadReason, "user@example.com") {
		return fmt.Errorf("dead-letter reason was not redacted: %s", deadReason)
	}

	effectRunID := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, pool, requestFor(f, effectRunID, "effect-replay-source")); err != nil {
		return err
	}
	calls := 0
	effectRuntime := smokeRuntime{actionTypeKey: "telegram.send_message", maxAttempts: 1, calls: &calls}
	k = kernel.New(pool, effectRuntime, kernel.WithRunClaimLease("retry-smoke-effect-source", 5*time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected source effect run to process")
	}
	replayRunID := uuid.New()
	replayInput, _ := json.Marshal(map[string]any{
		"smoke": "effect-replay",
		"_pulse": map[string]any{
			"base_workflow_run_id":                      effectRunID.String(),
			"replay_idempotency_anchor_workflow_run_id": effectRunID.String(),
		},
	})
	if _, _, err := runqueue.RequestRun(ctx, pool, runqueue.Request{
		WorkflowRunID:             replayRunID,
		ProjectID:                 f.ProjectID.String(),
		WorkflowID:                f.WorkflowID.String(),
		WorkflowVersionID:         f.VersionID.String(),
		SourceType:                runqueue.SourceReplay,
		ReplayedFromWorkflowRunID: effectRunID.String(),
		IdempotencyKey:            "runtime-claim-smoke:effect-replay:" + replayRunID.String(),
		InputJSON:                 replayInput,
		EventMessage:              "runtime replay smoke queued",
	}); err != nil {
		return err
	}
	k = kernel.New(pool, effectRuntime, kernel.WithRunClaimLease("retry-smoke-effect-replay", 5*time.Second, 100*time.Millisecond))
	if !k.ProcessOne(ctx) {
		return errors.New("expected replay effect run to process")
	}
	if calls != 1 {
		return fmt.Errorf("expected replay to reuse completed effect without provider execution, calls=%d", calls)
	}
	var replayActionStatus string
	if err := pool.QueryRow(ctx, `
	SELECT status FROM action_runs WHERE workflow_run_id = $1`, replayRunID).Scan(&replayActionStatus); err != nil {
		return err
	}
	if replayActionStatus != "succeeded" {
		return fmt.Errorf("expected replay action succeeded from ledger, got %s", replayActionStatus)
	}
	return nil
}

func requestFor(f fixtureIDs, runID uuid.UUID, suffix string) runqueue.Request {
	input, _ := json.Marshal(map[string]any{"smoke": suffix})
	return runqueue.Request{
		WorkflowRunID:     runID,
		ProjectID:         f.ProjectID.String(),
		WorkflowID:        f.WorkflowID.String(),
		WorkflowVersionID: f.VersionID.String(),
		SourceType:        runqueue.SourceTest,
		IdempotencyKey:    "runtime-claim-smoke:" + suffix + ":" + runID.String(),
		InputJSON:         input,
		EventMessage:      "runtime claim smoke queued",
	}
}

func (r smokeRuntime) BuildRunPlan(context.Context, []byte, []byte, string) (kernel.RunPlan, error) {
	actionTypeKey := strings.TrimSpace(r.actionTypeKey)
	if actionTypeKey == "" {
		actionTypeKey = "smoke.action"
	}
	maxAttempts := r.maxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return kernel.RunPlan{
		Execute: []kernel.Node{{
			ActionKey:     "claim_smoke",
			ActionTypeKey: actionTypeKey,
			MaxAttempts:   maxAttempts,
		}},
		InitialSteps: map[string]any{},
	}, nil
}

func (smokeRuntime) DefinitionNeedsProjectID([]byte) bool                  { return false }
func (smokeRuntime) EnrichRunInputProjectID(input []byte, _ string) []byte { return input }
func (smokeRuntime) ActionInput(kernel.Node, []byte) string                { return `{}` }
func (smokeRuntime) SkippedOutput(kernel.Node) string                      { return `{"skipped":true}` }
func (smokeRuntime) SkippedOutputWithReason(kernel.Node, string) string    { return `{"skipped":true}` }
func (smokeRuntime) ShouldRunGraphNode(kernel.Node, any, map[string]bool, map[string]string) bool {
	return true
}
func (smokeRuntime) BranchFromOutput(string) string                   { return "" }
func (smokeRuntime) MergeActionExpose(map[string]any, string, string) {}

func (r smokeRuntime) ExecuteAction(ctx context.Context, _ kernel.Node, _ []byte, _ map[string]any, _ string, _ string, attempt int) kernel.ActionExecution {
	if r.calls != nil {
		*r.calls++
	}
	if r.delay > 0 {
		select {
		case <-ctx.Done():
			return kernel.ActionExecution{InputJSON: `{}`, OutputJSON: `{}`, ProviderResponseJSON: `{}`, Err: ctx.Err(), ErrorType: "canceled"}
		case <-time.After(r.delay):
		}
	}
	if attempt < 1 {
		attempt = 1
	}
	if r.waiting {
		reason := strings.TrimSpace(r.waitingReason)
		if reason == "" {
			reason = "manual_approval_required"
		}
		return kernel.ActionExecution{
			InputJSON:            `{}`,
			OutputJSON:           `{"waiting":true}`,
			ProviderResponseJSON: `{}`,
			Waiting:              true,
			WaitingReason:        reason,
			WaitingDeadline:      r.waitingDeadline,
		}
	}
	if r.failuresBeforeSuccess == 0 && r.failureErr != nil {
		return kernel.ActionExecution{
			InputJSON:            `{}`,
			OutputJSON:           `{}`,
			ProviderResponseJSON: providerOrDefault(r.failureProviderJSON),
			Retryable:            r.failureRetryable,
			ErrorType:            r.failureErrorType,
			Err:                  r.failureErr,
		}
	}
	if attempt <= r.failuresBeforeSuccess {
		err := r.failureErr
		if err == nil {
			err = errors.New("smoke failure")
		}
		return kernel.ActionExecution{
			InputJSON:            `{}`,
			OutputJSON:           `{}`,
			ProviderResponseJSON: providerOrDefault(r.failureProviderJSON),
			Retryable:            r.failureRetryable,
			ErrorType:            r.failureErrorType,
			Err:                  err,
		}
	}
	return kernel.ActionExecution{InputJSON: `{}`, OutputJSON: `{"ok":true}`, ProviderResponseJSON: `{}`}
}

func providerOrDefault(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return `{}`
	}
	return raw
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
