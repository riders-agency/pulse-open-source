package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/domainevent"
)

func enqueueFirstSeen(ctx context.Context, tx pgx.Tx, widget widgetScope, visitorID, sessionID string, snapshot map[string]any) error {
	rows, err := tx.Query(ctx, `
SELECT mt.workflow_id::text, COALESCE(mt.workflow_version_id, w.current_workflow_version_id)::text
  FROM pulse_read.module_workflow_triggers_v1 mt
  JOIN pulse_read.workflows_v1 w ON w.workflow_id = mt.workflow_id
 WHERE mt.company_id = $1
   AND (mt.project_id IS NULL OR mt.project_id = $2)
   AND mt.module_key = 'site_chat'
   AND mt.event_key = 'visitor.first_seen'
   AND mt.enabled = true
   AND w.status = 'active'
   AND COALESCE(mt.workflow_version_id, w.current_workflow_version_id) IS NOT NULL`, widget.CompanyID, widget.ProjectID)
	if err != nil {
		return err
	}
	type triggerTarget struct {
		workflowID string
		versionID  string
	}
	targets := []triggerTarget{}
	for rows.Next() {
		var workflowID, versionID string
		if err := rows.Scan(&workflowID, &versionID); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, triggerTarget{workflowID: workflowID, versionID: versionID})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, target := range targets {
		runID := uuid.NewString()
		input := map[string]any{
			"event_type":             "visitor.first_seen",
			"event_source":           "site_chat",
			"visitor_id":             visitorID,
			"visit_session_id":       sessionID,
			"chat_widget_id":         widget.ChatWidgetID,
			"project_id":             widget.ProjectID,
			"visitor_state_snapshot": snapshot,
			"conversation_id":        stringFromMap(snapshot, "conversation_id"),
			"conversation_url":       stringFromMap(snapshot, "conversation_url"),
			"panel_url":              stringFromMap(snapshot, "panel_url"),
			"call_url":               stringFromMap(snapshot, "call_url"),
			"current_url":            stringFromMap(snapshot, "current_url"),
			"current_title":          stringFromMap(snapshot, "current_title"),
			"_pulse": map[string]any{
				"source_type": "module_trigger",
				"module_key":  "site_chat",
				"event_key":   "visitor.first_seen",
			},
		}
		payload := map[string]any{
			"workflow_run_id":     runID,
			"project_id":          widget.ProjectID,
			"workflow_id":         target.workflowID,
			"workflow_version_id": target.versionID,
			"source_type":         "module_trigger",
			"input_json":          input,
			"event_action_key":    "visitor_first_seen",
			"event_message":       "site chat visitor.first_seen queued",
			"event_payload_json":  input,
		}
		if err := domainevent.Record(ctx, domainevent.Event{
			TypeKey:        "workflow.run.requested",
			AggregateType:  "workflow_run",
			AggregateID:    runID,
			IdempotencyKey: "workflow.run.requested:" + runID,
			PayloadJSON:    jsonBytes(payload),
		}); err != nil {
			return err
		}
	}
	return nil
}
