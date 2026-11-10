package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/executor"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

// ErrNeedsPublish is returned when a manual run is requested without a workflow version.
var ErrNeedsPublish = errors.New("publish workflow before manual run")

// ErrReplayActionKeyRequired is returned by ReplayAction without action_key.
var ErrReplayActionKeyRequired = errors.New("action_key is required")

type Runs struct {
	P dbpkg.DB
	E *executor.Engine
}

func (r *Runs) assertWorkflow(ctx context.Context, wfID, uid uuid.UUID) error {
	if err := authz.RequireWorkflow(ctx, r.P, wfID, authz.PermRead); err != nil {
		return err
	}
	row := r.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.workflows_v1 w
 JOIN pulse_read.projects_v1 pr ON pr.project_id = w.project_id
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE w.workflow_id = $1`, wfID.String(), uid.String())
	var one int
	if err := row.Scan(&one); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (r *Runs) assertRun(ctx context.Context, runID, uid uuid.UUID) error {
	if err := authz.RequireRun(ctx, r.P, runID, authz.PermRead); err != nil {
		return err
	}
	row := r.P.QueryRow(ctx, `
SELECT 1 FROM workflow_runs wr
 JOIN pulse_read.projects_v1 pr ON pr.project_id = wr.project_id
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE wr.workflow_run_id = $1`, runID.String(), uid.String())
	var one int
	if err := row.Scan(&one); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

// ManualTx creates a run using the workflow's current published version.
func (r *Runs) ManualTx(ctx context.Context, wfID, uid uuid.UUID, input map[string]any) (string, error) {
	if err := r.assertWorkflow(ctx, wfID, uid); err != nil {
		return "", err
	}
	var pid, wvid string
	if err := r.P.QueryRow(ctx, `
SELECT w.project_id::text, w.current_workflow_version_id::text
  FROM pulse_read.workflows_v1 w WHERE w.workflow_id = $1`, wfID.String()).
		Scan(&pid, &wvid); err != nil {
		return "", err
	}
	if wvid == "" {
		return "", ErrNeedsPublish
	}
	rid := uuid.New()
	inb, _ := json.Marshal(input)
	if input == nil {
		inb = []byte("{}")
	}
	tx, err := r.transaction()
	if err != nil {
		return "", err
	}
	if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{
		WorkflowRunID:     rid,
		ProjectID:         pid,
		WorkflowID:        wfID.String(),
		WorkflowVersionID: wvid,
		SourceType:        "manual",
		InputJSON:         json.RawMessage(inb),
		EventMessage:      "manual run queued",
	}); err != nil {
		return "", err
	}
	return rid.String(), nil
}

type RunRow struct {
	RunID                   string          `json:"workflow_run_id"`
	ProjectID               string          `json:"project_id,omitempty"`
	WorkflowID              string          `json:"workflow_id,omitempty"`
	CompanyID               string          `json:"company_id,omitempty"`
	Status                  string          `json:"status"`
	SourceType              string          `json:"source_type,omitempty"`
	StartedAt               any             `json:"started_at,omitempty"`
	FinishedAt              any             `json:"finished_at,omitempty"`
	Input                   json.RawMessage `json:"input_json,omitempty"`
	Output                  json.RawMessage `json:"output_json,omitempty"`
	Error                   *string         `json:"error,omitempty"`
	StateMachineVersion     string          `json:"state_machine_version,omitempty"`
	TerminalReason          string          `json:"terminal_reason,omitempty"`
	CancelReason            string          `json:"cancel_reason,omitempty"`
	CancelRequestedByUserID *string         `json:"cancel_requested_by_user_id,omitempty"`
	CancelledAt             any             `json:"cancelled_at,omitempty"`
	WaitingDeadlineAt       any             `json:"waiting_deadline_at,omitempty"`
	WaitingReason           string          `json:"waiting_reason,omitempty"`
	WakeReason              string          `json:"wake_reason,omitempty"`
	LastStateTransitionAt   any             `json:"last_state_transition_at,omitempty"`
}

func (r *Runs) ListByWorkflow(ctx context.Context, wfID, uid uuid.UUID) ([]RunRow, error) {
	if err := r.assertWorkflow(ctx, wfID, uid); err != nil {
		return nil, err
	}
	rows, err := r.P.Query(ctx, `
SELECT wr.workflow_run_id::text, wr.project_id::text, wr.workflow_id::text, p.company_id::text,
       wr.status::text, wr.source_type::text, wr.started_at, wr.finished_at, wr.input_json, wr.output_json, wr.error,
       wr.state_machine_version::text, wr.terminal_reason::text, wr.cancel_reason::text,
       wr.cancel_requested_by_user_id::text, wr.cancelled_at, wr.waiting_deadline_at,
       wr.waiting_reason::text, wr.wake_reason::text, wr.last_state_transition_at
  FROM workflow_runs wr
  JOIN pulse_read.projects_v1 p ON p.project_id = wr.project_id
 WHERE wr.workflow_id = $1
 ORDER BY started_at DESC NULLS LAST, workflow_run_id DESC
 LIMIT 100`, wfID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RunRow, 0)
	for rows.Next() {
		var x RunRow
		var in, outJSON []byte
		if err := rows.Scan(
			&x.RunID, &x.ProjectID, &x.WorkflowID, &x.CompanyID,
			&x.Status, &x.SourceType, &x.StartedAt, &x.FinishedAt, &in, &outJSON, &x.Error,
			&x.StateMachineVersion, &x.TerminalReason, &x.CancelReason, &x.CancelRequestedByUserID,
			&x.CancelledAt, &x.WaitingDeadlineAt, &x.WaitingReason, &x.WakeReason, &x.LastStateTransitionAt,
		); err != nil {
			return nil, err
		}
		x.Input = in
		x.Output = outJSON
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *Runs) Get(ctx context.Context, runID, uid uuid.UUID) (*RunRow, error) {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return nil, err
	}
	row := r.P.QueryRow(ctx, `
SELECT wr.workflow_run_id::text, wr.project_id::text, wr.workflow_id::text, p.company_id::text,
       wr.status::text, wr.source_type::text, wr.started_at, wr.finished_at, wr.input_json, wr.output_json, wr.error,
       wr.state_machine_version::text, wr.terminal_reason::text, wr.cancel_reason::text,
       wr.cancel_requested_by_user_id::text, wr.cancelled_at, wr.waiting_deadline_at,
       wr.waiting_reason::text, wr.wake_reason::text, wr.last_state_transition_at
  FROM workflow_runs wr
  JOIN pulse_read.projects_v1 p ON p.project_id = wr.project_id
 WHERE wr.workflow_run_id = $1`, runID.String())
	var x RunRow
	var in, outJSON []byte
	if err := row.Scan(
		&x.RunID, &x.ProjectID, &x.WorkflowID, &x.CompanyID,
		&x.Status, &x.SourceType, &x.StartedAt, &x.FinishedAt, &in, &outJSON, &x.Error,
		&x.StateMachineVersion, &x.TerminalReason, &x.CancelReason, &x.CancelRequestedByUserID,
		&x.CancelledAt, &x.WaitingDeadlineAt, &x.WaitingReason, &x.WakeReason, &x.LastStateTransitionAt,
	); err != nil {
		return nil, err
	}
	x.Input = in
	x.Output = outJSON
	return &x, nil
}

func (r *Runs) ActionRuns(ctx context.Context, runID, uid uuid.UUID) ([]map[string]any, error) {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return nil, err
	}
	attemptsByActionRun := map[string][]map[string]any{}
	attemptRows, err := r.P.Query(ctx, `
SELECT ar.action_run_id::text, ar.action_type_key::text, aa.action_attempt_id::text, aa.attempt_no, aa.status::text,
       aa.input_json, aa.output_json, aa.provider_response_json, aa.error, aa.error_type,
       aa.started_at, aa.finished_at, aa.duration_ms
  FROM action_runs ar
  JOIN action_attempts aa ON aa.action_run_id = ar.action_run_id
 WHERE ar.workflow_run_id = $1
 ORDER BY ar.action_run_id, aa.attempt_no`, runID.String())
	if err != nil {
		return nil, err
	}
	for attemptRows.Next() {
		var actionRunID, actionTypeKey, attemptID, status string
		var attemptNo int
		var input, output, provider []byte
		var errMsg, errType *string
		var startedAt, finishedAt any
		var durationMS *int
		if err := attemptRows.Scan(
			&actionRunID, &actionTypeKey, &attemptID, &attemptNo, &status,
			&input, &output, &provider, &errMsg, &errType,
			&startedAt, &finishedAt, &durationMS,
		); err != nil {
			attemptRows.Close()
			return nil, err
		}
		attemptsByActionRun[actionRunID] = append(attemptsByActionRun[actionRunID], map[string]any{
			"action_attempt_id":      attemptID,
			"attempt_no":             attemptNo,
			"status":                 status,
			"input_json":             json.RawMessage(input),
			"output_json":            json.RawMessage(output),
			"provider_response_json": json.RawMessage(provider),
			"diagnostics":            executor.DiagnosticsFromStored(actionTypeKey, output, provider, errMsg, status),
			"error":                  errMsg,
			"error_type":             errType,
			"started_at":             startedAt,
			"finished_at":            finishedAt,
			"duration_ms":            durationMS,
		})
	}
	if err := attemptRows.Err(); err != nil {
		attemptRows.Close()
		return nil, err
	}
	attemptRows.Close()

	rows, err := r.P.Query(ctx, `
	SELECT ar.action_run_id::text, ar.action_key::text, ar.action_type_key::text,
	       dt.dependency_type_key::text AS dep_type_key,
	       ar.status::text,
	       ar.attempt_count, COALESCE(ar.max_attempts, 0), ar.input_json, ar.output_json, ar.provider_response_json, ar.error,
	       ar.started_at, ar.finished_at,
	       ar.retry_class, ar.retry_policy_json, ar.next_retry_at, ar.retry_exhausted_at,
	       ar.dead_lettered_at, ar.dead_letter_reason, ar.dead_letter_owner, ar.replayable,
	       ar.state_machine_version::text, ar.terminal_reason::text, ar.cancelled_at,
	       ar.waiting_deadline_at, ar.waiting_reason::text, ar.wake_reason::text, ar.last_state_transition_at
	  FROM action_runs ar
  LEFT JOIN pulse_read.action_dependencies_v1 ad ON ad.action_dependency_id = ar.action_dependency_id
  LEFT JOIN pulse_read.action_dependency_types_v1 dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ar.workflow_run_id = $1 ORDER BY ar.action_run_id`, runID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var aid, ak, atk, st, stateMachineVersion, terminalReason, waitingReason, wakeReason string
		var depType *string
		var ac, ma int
		var in, o, pr []byte
		var errMsg, retryClass, deadLetterReason, deadLetterOwner *string
		var retryPolicy []byte
		var sa, fa, nextRetryAt, retryExhaustedAt, deadLetteredAt, cancelledAt, waitingDeadlineAt, lastStateTransitionAt any
		var replayable bool
		if err := rows.Scan(
			&aid, &ak, &atk, &depType, &st, &ac, &ma, &in, &o, &pr, &errMsg,
			&sa, &fa, &retryClass, &retryPolicy, &nextRetryAt, &retryExhaustedAt,
			&deadLetteredAt, &deadLetterReason, &deadLetterOwner, &replayable,
			&stateMachineVersion, &terminalReason, &cancelledAt, &waitingDeadlineAt,
			&waitingReason, &wakeReason, &lastStateTransitionAt,
		); err != nil {
			return nil, err
		}
		m := map[string]any{
			"action_run_id":            aid,
			"action_key":               ak,
			"action_type_key":          atk,
			"status":                   st,
			"attempt_count":            ac,
			"max_attempts":             ma,
			"input_json":               json.RawMessage(in),
			"output_json":              json.RawMessage(o),
			"provider_response_json":   json.RawMessage(pr),
			"diagnostics":              executor.DiagnosticsFromStored(atk, o, pr, errMsg, st),
			"error":                    errMsg,
			"started_at":               sa,
			"finished_at":              fa,
			"retry_class":              retryClass,
			"retry_policy":             json.RawMessage(retryPolicy),
			"next_retry_at":            nextRetryAt,
			"retry_exhausted_at":       retryExhaustedAt,
			"dead_lettered_at":         deadLetteredAt,
			"dead_letter_reason":       deadLetterReason,
			"dead_letter_owner":        deadLetterOwner,
			"replayable":               replayable,
			"state_machine_version":    stateMachineVersion,
			"terminal_reason":          terminalReason,
			"cancelled_at":             cancelledAt,
			"waiting_deadline_at":      waitingDeadlineAt,
			"waiting_reason":           waitingReason,
			"wake_reason":              wakeReason,
			"last_state_transition_at": lastStateTransitionAt,
			"attempts":                 attemptsByActionRun[aid],
		}
		if depType != nil && strings.TrimSpace(*depType) != "" {
			m["dependency_type_key"] = strings.TrimSpace(*depType)
			m["integration_profile"] = strings.TrimSpace(*depType)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Runs) Events(ctx context.Context, runID, uid uuid.UUID) ([]map[string]any, error) {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return nil, err
	}
	rows, err := r.P.Query(ctx, `
SELECT run_event_id::text, event_type::text, COALESCE(message,'')::text, COALESCE(action_key,'')::text, payload_json, created_at
  FROM run_events WHERE workflow_run_id = $1 ORDER BY created_at, run_event_id`, runID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, et, msg, ak string
		var pj []byte
		var ca any
		if err := rows.Scan(&id, &et, &msg, &ak, &pj, &ca); err != nil {
			return nil, err
		}
		m := map[string]any{
			"run_event_id": id,
			"event_type":   et,
			"message":      msg,
			"created_at":   ca,
		}
		if ak != "" {
			m["action_key"] = ak
		}
		if len(pj) > 0 {
			m["payload_json"] = json.RawMessage(pj)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Runs) CancelTx(ctx context.Context, runID, uid uuid.UUID) error {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return err
	}
	tx, err := r.transaction()
	if err != nil {
		return err
	}

	var previousStatus string
	var changed bool
	err = tx.QueryRow(ctx, `
WITH target AS (
  SELECT workflow_run_id, status
    FROM workflow_runs
   WHERE workflow_run_id = $1
   FOR UPDATE
),
updated AS (
  UPDATE workflow_runs wr
     SET status = 'cancelled',
         finished_at = now(),
         error = 'cancelled:user_cancel',
         terminal_reason = 'user_cancel',
         cancel_reason = 'user_cancel',
         cancel_requested_by_user_id = $2,
         cancelled_at = COALESCE(wr.cancelled_at, now()),
         run_lease_expires_at = NULL,
         waiting_deadline_at = NULL,
         wake_reason = CASE WHEN target.status = 'waiting' THEN 'cancelled' ELSE wr.wake_reason END,
         last_state_transition_at = now()
    FROM target
   WHERE wr.workflow_run_id = target.workflow_run_id
     AND target.status IN ('pending','running','waiting')
   RETURNING target.status::text AS previous_status, true AS changed
)
SELECT previous_status, changed FROM updated
UNION ALL
SELECT target.status::text, false FROM target
 WHERE NOT EXISTS (SELECT 1 FROM updated)`, runID.String(), uid.String()).Scan(&previousStatus, &changed)
	if err != nil {
		return err
	}
	if !changed {
		if previousStatus == "cancelled" {
			return nil
		}
		return fmt.Errorf("not cancellable")
	}
	payload, err := json.Marshal(map[string]string{
		"reason":          "user_cancel",
		"previous_status": previousStatus,
	})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json)
VALUES ($1,'workflow.cancelled','cancel requested: user_cancel',$2::jsonb)`, runID.String(), string(payload)); err != nil {
		return err
	}
	return nil
}

func (r *Runs) Graph(ctx context.Context, runID, uid uuid.UUID) (map[string]any, error) {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return nil, err
	}
	var graph []byte
	if err := r.P.QueryRow(ctx, `
SELECT COALESCE(wv.graph_json, '{}'::jsonb)
  FROM workflow_runs wr
  JOIN pulse_read.workflow_versions_v1 wv ON wv.workflow_version_id = wr.workflow_version_id
 WHERE wr.workflow_run_id = $1`, runID.String()).Scan(&graph); err != nil {
		return nil, err
	}
	rows, err := r.P.Query(ctx, `
SELECT action_key::text, status::text, output_json, error
  FROM action_runs WHERE workflow_run_id = $1`, runID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	overlay := map[string]map[string]any{}
	for rows.Next() {
		var ak, st string
		var out []byte
		var errStr *string
		if err := rows.Scan(&ak, &st, &out, &errStr); err != nil {
			return nil, err
		}
		entry := map[string]any{"status": st}
		if len(out) > 0 {
			entry["output_json"] = json.RawMessage(out)
		}
		if errStr != nil {
			entry["error"] = *errStr
		}
		overlay[ak] = entry
	}
	var base any
	if err := json.Unmarshal(graph, &base); err != nil {
		base = map[string]any{}
	}
	return map[string]any{
		"graph":          base,
		"action_overlay": overlay,
	}, rows.Err()
}

func (r *Runs) Replay(ctx context.Context, runID, uid uuid.UUID) (string, error) {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return "", err
	}
	var pid, wf, wv string
	var input []byte
	err := r.P.QueryRow(ctx, `
SELECT project_id::text, workflow_id::text, workflow_version_id::text, input_json
  FROM workflow_runs WHERE workflow_run_id = $1`, runID.String()).
		Scan(&pid, &wf, &wv, &input)
	if err != nil {
		return "", err
	}
	replayInput, err := inputWithReplayAnchor(input, runID.String(), "")
	if err != nil {
		return "", err
	}
	nid := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, r.P, runqueue.Request{
		WorkflowRunID:             nid,
		ProjectID:                 pid,
		WorkflowID:                wf,
		WorkflowVersionID:         wv,
		SourceType:                "replay",
		ReplayedFromWorkflowRunID: runID.String(),
		InputJSON:                 json.RawMessage(replayInput),
		EventMessage:              "replay from previous run",
	}); err != nil {
		return "", err
	}
	if r.E != nil {
		r.E.Notify()
	}
	return nid.String(), nil
}

// ReplayAction enqueues a new run that executes only the given action_key (same workflow_version and merged input_json).
func (r *Runs) ReplayAction(ctx context.Context, runID, uid uuid.UUID, actionKey string) (string, error) {
	if err := r.assertRun(ctx, runID, uid); err != nil {
		return "", err
	}
	actionKey = strings.TrimSpace(actionKey)
	if actionKey == "" {
		return "", ErrReplayActionKeyRequired
	}
	var pid, wf, wv string
	var input []byte
	err := r.P.QueryRow(ctx, `
SELECT project_id::text, workflow_id::text, workflow_version_id::text, input_json
  FROM workflow_runs WHERE workflow_run_id = $1`, runID.String()).
		Scan(&pid, &wf, &wv, &input)
	if err != nil {
		return "", err
	}
	mergedInput, err := inputWithReplayAnchor(input, runID.String(), actionKey)
	if err != nil {
		return "", err
	}
	steps, err := r.replayActionSteps(ctx, runID)
	if err != nil {
		return "", err
	}
	mergedInput, err = inputWithReplaySteps(mergedInput, steps)
	if err != nil {
		return "", err
	}

	nid := uuid.New()
	if _, _, err := runqueue.RequestRun(ctx, r.P, runqueue.Request{
		WorkflowRunID:             nid,
		ProjectID:                 pid,
		WorkflowID:                wf,
		WorkflowVersionID:         wv,
		SourceType:                "replay",
		ReplayedFromWorkflowRunID: runID.String(),
		InputJSON:                 json.RawMessage(mergedInput),
		EventActionKey:            actionKey,
		EventMessage:              "replay single action from previous run",
	}); err != nil {
		return "", err
	}
	if r.E != nil {
		r.E.Notify()
	}
	return nid.String(), nil
}

func (r *Runs) replayActionSteps(ctx context.Context, runID uuid.UUID) (map[string]any, error) {
	rows, err := r.P.Query(ctx, `
SELECT action_key, output_json->'expose'
  FROM action_runs
 WHERE workflow_run_id = $1
   AND status = 'succeeded'
   AND jsonb_typeof(output_json->'expose') = 'object'`, runID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	steps := map[string]any{}
	for rows.Next() {
		var actionKey string
		var raw []byte
		if err := rows.Scan(&actionKey, &raw); err != nil {
			return nil, err
		}
		var expose map[string]any
		if err := json.Unmarshal(raw, &expose); err != nil {
			return nil, err
		}
		steps[actionKey] = expose
	}
	return steps, rows.Err()
}

func inputWithReplaySteps(input []byte, steps map[string]any) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(input, &root); err != nil {
		return nil, err
	}
	pulse, ok := root["_pulse"].(map[string]any)
	if !ok {
		pulse = map[string]any{}
		root["_pulse"] = pulse
	}
	existing, ok := pulse["steps"].(map[string]any)
	if !ok {
		existing = map[string]any{}
	}
	for key, expose := range steps {
		if _, present := existing[key]; !present {
			existing[key] = expose
		}
	}
	pulse["steps"] = existing
	return json.Marshal(root)
}

func inputWithReplayAnchor(input []byte, baseRunID string, actionKey string) ([]byte, error) {
	inObj := map[string]any{}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &inObj)
	}
	var pulse map[string]any
	if p, ok := inObj["_pulse"].(map[string]any); ok && p != nil {
		pulse = p
	} else {
		pulse = map[string]any{}
	}
	pulse["base_workflow_run_id"] = baseRunID
	pulse["replay_idempotency_anchor_workflow_run_id"] = baseRunID
	if actionKey != "" {
		pulse["retry_action_key"] = actionKey
	}
	inObj["_pulse"] = pulse
	return json.Marshal(inObj)
}

// StreamEvents writes SSE updates by polling newly inserted rows.
