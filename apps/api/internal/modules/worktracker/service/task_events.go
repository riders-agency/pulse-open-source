package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/outbox"
)

type TaskEvents struct{}

func (TaskEvents) EnqueueTaskModuleTriggers(ctx context.Context, tx pgx.Tx, companyID, projectID, workflowID, taskID uuid.UUID, eventKey string, payload map[string]any) error {
	raw, _ := json.Marshal(payload)
	rows, err := tx.Query(ctx, `
  SELECT gen_random_uuid()::text AS workflow_run_id,
         ($2::uuid)::text AS project_id,
         mt.workflow_id::text AS workflow_id,
         COALESCE(mt.workflow_version_id, w.current_workflow_version_id)::text AS workflow_version_id,
         jsonb_build_object(
           '_pulse', jsonb_build_object(
             'source_type', 'task_event',
             'module_key', 'work.tasks',
             'event_key', $5::text,
             'module_workflow_trigger_id', mt.module_workflow_trigger_id::text,
             'project_id', $2::text,
             'task_workflow_id', $3::text,
             'task_id', $4::text
           ),
           'event', jsonb_build_object(
             'module_key', 'work.tasks',
             'event_key', $5::text,
             'payload', $6::jsonb,
             'config', mt.config_json
           ),
           'task_id', $4::text,
           'task', jsonb_build_object('task_id', $4::text, 'url', 'https://pulse.riders.agency/tasks/boards')
         )::jsonb AS input_json
    FROM pulse_read.module_workflow_triggers_v1 mt
    JOIN pulse_read.workflows_v1 w ON w.workflow_id = mt.workflow_id
   WHERE mt.company_id = $1::uuid
     AND (mt.project_id = $2::uuid OR mt.project_id IS NULL)
     AND mt.module_key = 'work.tasks'
     AND mt.event_key = $5
     AND mt.enabled = true
     AND w.status = 'active'
     AND COALESCE(mt.workflow_version_id, w.current_workflow_version_id) IS NOT NULL
     AND COALESCE(NULLIF(mt.config_json ->> 'task_workflow_id', ''), $3::text) = $3::text`,
		companyID.String(), projectID.String(), workflowID.String(), taskID.String(), eventKey, string(raw))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var runID string
		var queuedProjectID, queuedWorkflowID, versionID string
		var inputJSON json.RawMessage
		if err := rows.Scan(&runID, &queuedProjectID, &queuedWorkflowID, &versionID, &inputJSON); err != nil {
			return err
		}
		payloadJSON, err := json.Marshal(map[string]any{
			"workflow_run_id":     runID,
			"project_id":          queuedProjectID,
			"workflow_id":         queuedWorkflowID,
			"workflow_version_id": versionID,
			"source_type":         "task_event",
			"input_json":          inputJSON,
			"event_action_key":    "task_event",
			"event_message":       "task event " + eventKey + " queued",
			"event_payload_json":  inputJSON,
		})
		if err != nil {
			return err
		}
		if _, err := outbox.EnqueueTx(ctx, tx, outbox.EnqueueRequest{
			TypeKey:        "workflow.run.requested",
			AggregateType:  "workflow_run",
			AggregateID:    runID,
			IdempotencyKey: "workflow.run.requested:" + runID,
			PayloadJSON:    payloadJSON,
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}
