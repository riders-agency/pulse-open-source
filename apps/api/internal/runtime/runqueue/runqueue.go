package runqueue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func RequestRun(ctx context.Context, pool dbpkg.DB, req Request) (uuid.UUID, bool, error) {
	if pool == nil {
		return uuid.Nil, false, fmt.Errorf("run queue pool is nil")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	defer tx.Rollback(ctx)
	runID, inserted, err := RequestRunTx(ctx, tx, req)
	if err != nil {
		return uuid.Nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, false, err
	}
	return runID, inserted, nil
}

func RequestRunTx(ctx context.Context, tx pgx.Tx, req Request) (uuid.UUID, bool, error) {
	if tx == nil {
		return uuid.Nil, false, fmt.Errorf("run queue transaction is nil")
	}
	req, err := normalizeRequest(req)
	if err != nil {
		return uuid.Nil, false, err
	}
	var runID uuid.UUID
	var inserted bool
	err = tx.QueryRow(ctx, `
WITH inserted AS (
  INSERT INTO workflow_runs (
    workflow_run_id,
    project_id,
    workflow_id,
    workflow_version_id,
    source_type,
    source_webhook_event_id,
    replayed_from_workflow_run_id,
    run_request_version,
    run_request_idempotency_key,
    run_request_source_json,
    status,
    input_json
  )
  VALUES ($1,$2,$3,$4,$5,NULLIF($6, '')::uuid,NULLIF($7, '')::uuid,$13,$14,$15::jsonb,'pending',$8::jsonb)
  ON CONFLICT DO NOTHING
  RETURNING workflow_run_id, true AS inserted
),
selected AS (
  SELECT workflow_run_id, inserted FROM inserted
  UNION ALL
  SELECT workflow_run_id, false AS inserted
    FROM workflow_runs
   WHERE run_request_idempotency_key = NULLIF($14, '')
     AND NOT EXISTS (SELECT 1 FROM inserted)
  UNION ALL
  SELECT workflow_run_id, false AS inserted
    FROM workflow_runs
   WHERE workflow_run_id = $1
     AND NOT EXISTS (SELECT 1 FROM inserted)
  LIMIT 1
),
event_inserted AS (
  INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
  SELECT workflow_run_id, NULLIF($9, ''), $10, $11, $12::jsonb, clock_timestamp()
    FROM selected
   WHERE inserted
  RETURNING workflow_run_id
)
SELECT workflow_run_id, inserted FROM selected
`, req.WorkflowRunID.String(), req.ProjectID, req.WorkflowID, req.WorkflowVersionID, req.SourceType,
		req.SourceWebhookEventID, req.ReplayedFromWorkflowRunID, string(req.InputJSON), req.EventActionKey,
		req.EventType, req.EventMessage, string(req.EventPayloadJSON), req.ContractVersion, req.IdempotencyKey,
		string(req.SourceMetadataJSON)).Scan(&runID, &inserted)
	if err != nil {
		return uuid.Nil, false, err
	}
	return runID, inserted, nil
}
