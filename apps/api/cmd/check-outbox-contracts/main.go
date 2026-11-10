package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/modules/workflows"
	"riders.agency/pulse/api/internal/platform/outbox"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

type fixtureIDs struct {
	UserID     uuid.UUID
	CompanyID  uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	VersionID  uuid.UUID
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
	defer cleanupFixture(context.Background(), pool, fixture)

	if err := checkPayloadContract(ctx, pool); err != nil {
		fail("payload contract: %v", err)
	}
	if err := checkRollbackAtomicity(ctx, pool); err != nil {
		fail("rollback atomicity: %v", err)
	}
	if err := checkWorkflowRunRequestedIdempotency(ctx, pool, fixture); err != nil {
		fail("workflow run requested idempotency: %v", err)
	}
}

func checkPayloadContract(ctx context.Context, pool *pgxpool.Pool) error {
	repo := &outbox.Repository{P: pool}
	if _, err := repo.Enqueue(ctx, outbox.EnqueueRequest{
		TypeKey:        "outbox.contract.bad-array",
		AggregateType:  "contract",
		AggregateID:    uuid.NewString(),
		IdempotencyKey: "outbox-contract:bad-array:" + uuid.NewString(),
		PayloadJSON:    []byte(`[]`),
	}); err == nil {
		return fmt.Errorf("non-object payload was accepted")
	}
	if _, err := repo.Enqueue(ctx, outbox.EnqueueRequest{
		TypeKey:        "outbox.contract.version",
		SchemaVersion:  "outbox.contract.version.v1",
		AggregateType:  "contract",
		AggregateID:    uuid.NewString(),
		IdempotencyKey: "outbox-contract:bad-version:" + uuid.NewString(),
		PayloadJSON:    []byte(`{"event_version":"outbox.contract.version.v0"}`),
	}); err == nil {
		return fmt.Errorf("payload version drift was accepted")
	}
	return nil
}

func checkRollbackAtomicity(ctx context.Context, pool *pgxpool.Pool) error {
	key := "outbox-contract:rollback:" + uuid.NewString()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := outbox.EnqueueTx(ctx, tx, outbox.EnqueueRequest{
		TypeKey:        "outbox.contract.rollback",
		AggregateType:  "contract",
		AggregateID:    uuid.NewString(),
		IdempotencyKey: key,
		PayloadJSON:    []byte(`{"kind":"rollback"}`),
	}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	var inTxCount int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*)::int FROM outbox_events WHERE idempotency_key = $1`, key).Scan(&inTxCount); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if inTxCount != 1 {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("expected event inside transaction, got %d", inTxCount)
	}
	if err := tx.Rollback(ctx); err != nil {
		return err
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM outbox_events WHERE idempotency_key = $1`, key).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("rolled back outbox event remained: %d", count)
	}
	return nil
}

func checkWorkflowRunRequestedIdempotency(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) error {
	repo := &outbox.Repository{P: pool}
	runID := uuid.New()
	key := "outbox-contract:workflow-run-requested:" + runID.String()
	payloadJSON, err := json.Marshal(map[string]any{
		"workflow_run_id":     runID.String(),
		"project_id":          f.ProjectID.String(),
		"workflow_id":         f.WorkflowID.String(),
		"workflow_version_id": f.VersionID.String(),
		"source_type":         runqueue.SourceTaskEvent,
		"input_json": map[string]any{
			"contract": "outbox",
		},
		"event_action_key":   "outbox_contract",
		"event_message":      "outbox contract smoke queued",
		"event_payload_json": map[string]any{"contract": "outbox"},
	})
	if err != nil {
		return err
	}
	req := outbox.EnqueueRequest{
		TypeKey:        workflows.WorkflowRunRequestedEvent,
		SchemaVersion:  workflows.WorkflowRunRequestedEventVersion,
		AggregateType:  "workflow_run",
		AggregateID:    runID.String(),
		IdempotencyKey: key,
		PayloadJSON:    payloadJSON,
	}
	ev, err := repo.Enqueue(ctx, req)
	if err != nil {
		return err
	}
	duplicate, err := repo.Enqueue(ctx, req)
	if err != nil {
		return err
	}
	if duplicate.ID != ev.ID {
		return fmt.Errorf("duplicate enqueue returned a different event id: %s != %s", duplicate.ID, ev.ID)
	}
	driftReq := req
	driftReq.PayloadJSON = []byte(`{"event_version":"workflow.run.requested.v1","source_type":"task_event"}`)
	if _, err := repo.Enqueue(ctx, driftReq); err == nil {
		return fmt.Errorf("idempotency key payload drift was accepted")
	}
	var version string
	if err := pool.QueryRow(ctx, `
SELECT payload_json ->> 'event_version'
  FROM outbox_events
 WHERE outbox_event_id = $1`, ev.ID).Scan(&version); err != nil {
		return err
	}
	if version != workflows.WorkflowRunRequestedEventVersion {
		return fmt.Errorf("unexpected event version %q", version)
	}
	reg := outbox.NewRegistry()
	if err := workflows.RegisterOutboxHandlers(reg, pool, nil); err != nil {
		return err
	}
	handler, ok := reg.Handler(workflows.WorkflowRunRequestedEvent)
	if !ok {
		return fmt.Errorf("workflow run requested handler not registered")
	}
	if err := handler.Handle(ctx, ev); err != nil {
		return err
	}
	if err := handler.Handle(ctx, ev); err != nil {
		return err
	}
	var runCount, eventCount int
	if err := pool.QueryRow(ctx, `
SELECT
  (SELECT COUNT(*)::int FROM workflow_runs WHERE workflow_run_id = $1 AND run_request_idempotency_key = $2),
  (SELECT COUNT(*)::int FROM run_events WHERE workflow_run_id = $1)`, runID, key).Scan(&runCount, &eventCount); err != nil {
		return err
	}
	if runCount != 1 || eventCount != 1 {
		return fmt.Errorf("unexpected idempotent consumer result runs=%d run_events=%d", runCount, eventCount)
	}
	return nil
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
VALUES ($1, 'Outbox Contract Smoke', 'Outbox Contract Smoke', true)`, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO companies (company_id, owner_user_id, slug, name, created_by_user_id)
VALUES ($1, $2, $3, 'Outbox Contract Smoke', $2)`, f.CompanyID, f.UserID, "outbox-contract-"+f.CompanyID.String()); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO projects (project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id)
VALUES ($1, $2, 'Outbox Contract Smoke', $3, $4, 'hash', $5)`,
		f.ProjectID, f.CompanyID, "outbox-contract-"+f.ProjectID.String(), "hook-"+f.ProjectID.String(), f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	def := `{"nodes":[{"action_key":"outbox_contract","action_type_key":"stop","config":{}}]}`
	if _, err := pool.Exec(ctx, `
INSERT INTO workflows (workflow_id, project_id, name, slug, status, draft_definition_json, created_by_user_id)
VALUES ($1, $2, 'Outbox Contract Smoke', $3, 'active', $4::jsonb, $5)`,
		f.WorkflowID, f.ProjectID, "outbox-contract-"+f.WorkflowID.String(), def, f.UserID); err != nil {
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

func cleanupFixture(ctx context.Context, pool *pgxpool.Pool, f fixtureIDs) {
	_, _ = pool.Exec(ctx, `DELETE FROM companies WHERE company_id = $1`, f.CompanyID)
	_, _ = pool.Exec(ctx, `DELETE FROM outbox_events WHERE idempotency_key LIKE 'outbox-contract:%'`)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
