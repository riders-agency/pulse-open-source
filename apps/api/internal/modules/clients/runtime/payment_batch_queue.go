package runtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/runtime/runqueue"
)

func (h *PaymentDueTrigger) enqueueBatchPaymentDueRun(ctx context.Context) bool {
	tx, err := h.P.Begin(ctx)
	if err != nil {
		return false
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
SELECT s.workflow_id::text, s.workflow_version_id::text, s.action_key::text, s.remind_every_days,
       s.days_before_due, s.project_id::text, p.name, COALESCE(p.slug, '')
  FROM workflow_client_payment_due_state s
  JOIN pulse_read.workflows_v1 w ON w.workflow_id = s.workflow_id
  JOIN pulse_read.projects_v1 p ON p.project_id = s.project_id
 WHERE s.enabled = true
   AND w.status = 'active'
 ORDER BY s.updated_at
 FOR UPDATE OF s SKIP LOCKED
 LIMIT 1`)
	var workflowID, versionID, actionKey, projectID, projectName, projectSlug string
	var remindEveryDays, daysBeforeDue int
	if err := row.Scan(&workflowID, &versionID, &actionKey, &remindEveryDays, &daysBeforeDue, &projectID, &projectName, &projectSlug); err != nil {
		return false
	}
	remindEveryDays = clampReminderDays(remindEveryDays, 7)
	daysBeforeDue = clampDaysBeforeDue(daysBeforeDue, 7)

	var pendingRuns int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)::int
  FROM pulse_read.workflow_runs_v1
 WHERE workflow_id = $1
   AND source_type = 'client_payment_due'
   AND status IN ('pending', 'running')`, workflowID).Scan(&pendingRuns); err != nil || pendingRuns > 0 {
		return false
	}
	var recentBatch int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)::int
  FROM pulse_read.workflow_runs_v1
 WHERE workflow_id = $1
   AND source_type = 'client_payment_due'
   AND started_at > clock_timestamp() - ($2 * interval '1 day')`,
		workflowID, remindEveryDays).Scan(&recentBatch); err != nil || recentBatch > 0 {
		return false
	}
	var notifyPending bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM payment_actions pa
   WHERE pa.project_id = $1::uuid
     AND pa.status IN ('pending','sent','snoozed')
     AND pa.due_date <= (CURRENT_DATE + ($2 * interval '1 day'))::date
     AND (pa.status <> 'snoozed' OR pa.snoozed_until IS NULL OR pa.snoozed_until <= clock_timestamp())
     AND (pa.remind_after_at IS NULL OR pa.remind_after_at <= clock_timestamp())
  UNION ALL
  SELECT 1
    FROM client_services cs
   WHERE cs.project_id = $1::uuid
     AND cs.status = 'active'
     AND cs.next_due_date IS NOT NULL
     AND cs.next_due_date <= (CURRENT_DATE + ($2 * interval '1 day'))::date
     AND NOT EXISTS (
       SELECT 1 FROM payment_actions pa
        WHERE pa.client_id = cs.client_id
          AND pa.project_id = cs.project_id
          AND pa.status IN ('pending','sent','snoozed')
     )
)`, projectID, daysBeforeDue).Scan(&notifyPending); err != nil || !notifyPending {
		return false
	}

	if err := h.ensureDuePaymentActions(ctx, tx, projectID, daysBeforeDue); err != nil {
		slog.Warn("ensure due payment actions failed", slog.String("project_id", projectID), slog.Any("err", err))
		return false
	}

	rows, err := h.collectPaymentCycleRowsTx(ctx, tx, projectID, daysBeforeDue)
	if err != nil || len(rows) == 0 {
		return false
	}

	digest := buildPaymentDigest(rows, projectName, projectID, paymentDigestEvent{})
	inputJSON, err := batchPaymentRunInput(projectID, projectName, projectSlug, actionKey, versionID, len(rows), digest.Subject)
	if err != nil {
		return false
	}
	runID := uuid.New()
	if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{
		WorkflowRunID:     runID,
		ProjectID:         projectID,
		WorkflowID:        workflowID,
		WorkflowVersionID: versionID,
		SourceType:        "client_payment_due",
		InputJSON:         json.RawMessage(inputJSON),
		EventActionKey:    actionKey,
		EventMessage:      "batch client payment digest queued",
	}); err != nil {
		slog.Warn("batch client payment run insert failed", slog.String("workflow_id", workflowID), slog.Any("err", err))
		return false
	}
	if _, err := tx.Exec(ctx, `
UPDATE payment_actions pa
   SET status = CASE WHEN pa.status = 'pending' THEN 'sent' ELSE pa.status END,
       workflow_run_id = $2,
       last_reminded_at = clock_timestamp(),
       remind_after_at = clock_timestamp() + ($3 * interval '1 day'),
       updated_at = clock_timestamp()
 WHERE pa.project_id = $1
   AND pa.status IN ('pending','sent','snoozed')
   AND pa.due_date <= (CURRENT_DATE + ($4 * interval '1 day'))::date
   AND (pa.status <> 'snoozed' OR pa.snoozed_until IS NULL OR pa.snoozed_until <= clock_timestamp())
   AND (pa.remind_after_at IS NULL OR pa.remind_after_at <= clock_timestamp())`,
		projectID, runID.String(), remindEveryDays, daysBeforeDue); err != nil {
		return false
	}
	if err := tx.Commit(ctx); err != nil {
		return false
	}
	h.notify()
	return true
}

func batchPaymentRunInput(projectID, projectName, projectSlug, actionKey, versionID string, dueCount int, subject string) (string, error) {
	input := map[string]any{
		"_pulse": map[string]any{
			"source_type":         "client_payment_due",
			"batch":               true,
			"project_id":          projectID,
			"trigger_action_key":  actionKey,
			"workflow_version_id": versionID,
			"digest_subject":      subject,
		},
		"project": map[string]any{
			"project_id": projectID,
			"name":       projectName,
			"slug":       projectSlug,
		},
		"payment_cycle": map[string]any{
			"batch":     true,
			"due_count": dueCount,
		},
	}
	out, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (h *PaymentDueTrigger) collectPaymentCycleRowsTx(ctx context.Context, tx pgx.Tx, projectID string, daysBeforeDue int) ([]paymentCycleRow, error) {
	rows, err := tx.Query(ctx, paymentCycleCollectSQL, projectID, daysBeforeDue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPaymentCycleRows(rows)
}
