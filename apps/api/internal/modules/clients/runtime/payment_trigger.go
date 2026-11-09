package runtime

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

func (h *PaymentDueTrigger) TypeKey() string { return "clients.payment_due.trigger" }

func (h *PaymentDueTrigger) Sync(ctx context.Context) error {
	h.syncClientPaymentDueStates(ctx)
	return nil
}

func (h *PaymentDueTrigger) ClaimDue(context.Context, int) ([]contracts.TriggerEvent, error) {
	return nil, nil
}

func (h *PaymentDueTrigger) MarkEnqueued(context.Context, contracts.TriggerEvent, string) error {
	return nil
}

func (h *PaymentDueTrigger) EnqueueDue(ctx context.Context, limit int) (int, error) {
	if h == nil || h.P == nil {
		return 0, nil
	}
	count := 0
	for limit <= 0 || count < limit {
		if !h.enqueueOneDueClientPayment(ctx) {
			break
		}
		count++
	}
	return count, nil
}

func (h *PaymentDueTrigger) notify() {
	if h != nil && h.Notifier != nil {
		h.Notifier.Notify()
	}
}

func (h *PaymentDueTrigger) syncClientPaymentDueStates(ctx context.Context) {
	if h == nil || h.P == nil {
		return
	}
	syncStarted := time.Now().UTC()
	rows, err := h.P.Query(ctx, `
SELECT w.project_id::text, w.workflow_id::text, w.current_workflow_version_id::text, wv.definition_json
  FROM pulse_read.workflows_v1 w
  JOIN pulse_read.workflow_versions_v1 wv ON wv.workflow_version_id = w.current_workflow_version_id
 WHERE w.current_workflow_version_id IS NOT NULL
   AND w.status = 'active'`)
	if err != nil {
		slog.Warn("client payment trigger sync query failed", slog.Any("err", err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var projectID, workflowID, versionID string
		var def []byte
		if err := rows.Scan(&projectID, &workflowID, &versionID, &def); err != nil {
			slog.Warn("client payment trigger sync scan failed", slog.Any("err", err))
			continue
		}
		for _, node := range flattenNodes(def) {
			if node.ActionTypeKey != "clients.payment_due.trigger" {
				continue
			}
			_, err := h.P.Exec(ctx, `
INSERT INTO workflow_client_payment_due_state (
  workflow_id, action_key, workflow_version_id, project_id, enabled,
  remind_every_days, days_before_due, timezone, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,clock_timestamp())
ON CONFLICT (workflow_id, action_key) DO UPDATE SET
  workflow_version_id = EXCLUDED.workflow_version_id,
  project_id = EXCLUDED.project_id,
  enabled = EXCLUDED.enabled,
  remind_every_days = EXCLUDED.remind_every_days,
  days_before_due = EXCLUDED.days_before_due,
  timezone = EXCLUDED.timezone,
  updated_at = clock_timestamp()`,
				workflowID, node.ActionKey, versionID, projectID, scheduleEnabled(node.Config, time.Now().UTC()),
				clampReminderDays(node.Config.RemindEveryDays, 7), clampDaysBeforeDue(node.Config.DaysBeforeDue, 7),
				firstNonEmpty(strings.TrimSpace(node.Config.Timezone), "Europe/Moscow"))
			if err != nil {
				slog.Warn("client payment trigger state upsert failed",
					slog.String("workflow_id", workflowID),
					slog.String("action_key", node.ActionKey),
					slog.Any("err", err),
				)
			}
		}
	}
	if err := rows.Err(); err != nil {
		slog.Warn("client payment trigger sync rows failed", slog.Any("err", err))
	}
	_, _ = h.P.Exec(ctx, `
UPDATE workflow_client_payment_due_state
   SET enabled = false, updated_at = clock_timestamp()
 WHERE updated_at < $1`, syncStarted)
}

func (h *PaymentDueTrigger) enqueueOneDueClientPayment(ctx context.Context) bool {
	if h.enqueueExistingPaymentActionReminder(ctx) {
		return true
	}
	if h.enqueueNewPaymentActionReminder(ctx) {
		return true
	}
	return h.enqueueBatchPaymentDueRun(ctx)
}
