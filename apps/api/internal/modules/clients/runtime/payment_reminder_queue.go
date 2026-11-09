package runtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/runtime/runqueue"
)

func (h *PaymentDueTrigger) enqueueExistingPaymentActionReminder(ctx context.Context) bool {
	tx, err := h.P.Begin(ctx)
	if err != nil {
		return false
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
SELECT s.workflow_id::text, s.workflow_version_id::text, s.action_key::text, s.remind_every_days,
       pa.payment_action_id::text, pa.project_id::text
  FROM workflow_client_payment_due_state s
  JOIN pulse_read.workflows_v1 w ON w.workflow_id = s.workflow_id
  JOIN payment_actions pa ON pa.project_id = s.project_id
 WHERE s.enabled = true
   AND w.status = 'active'
   AND pa.status IN ('pending','sent','snoozed')
   AND pa.due_date <= (CURRENT_DATE + (s.days_before_due * interval '1 day'))::date
   AND (pa.status <> 'snoozed' OR pa.snoozed_until IS NULL OR pa.snoozed_until <= clock_timestamp())
   AND (pa.remind_after_at IS NULL OR pa.remind_after_at <= clock_timestamp())
 ORDER BY pa.due_date, pa.created_at
 FOR UPDATE OF pa SKIP LOCKED
 LIMIT 1`)
	var workflowID, versionID, actionKey, paymentActionID, projectID string
	var remindEveryDays int
	if err := row.Scan(&workflowID, &versionID, &actionKey, &remindEveryDays, &paymentActionID, &projectID); err != nil {
		return false
	}
	inputJSON, err := paymentActionRunInput(ctx, tx, paymentActionID, actionKey, versionID)
	if err != nil {
		slog.Warn("client payment trigger input failed", slog.Any("err", err))
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
		EventMessage:      "client payment reminder queued",
	}); err != nil {
		slog.Warn("client payment trigger run insert failed", slog.String("workflow_id", workflowID), slog.Any("err", err))
		return false
	}
	if _, err := tx.Exec(ctx, `
UPDATE payment_actions
   SET status = CASE WHEN status = 'pending' THEN 'sent' ELSE status END,
       workflow_run_id = $2,
       last_reminded_at = clock_timestamp(),
       remind_after_at = clock_timestamp() + ($3 * interval '1 day'),
       updated_at = clock_timestamp()
 WHERE payment_action_id = $1`,
		paymentActionID, runID.String(), clampReminderDays(remindEveryDays, 7)); err != nil {
		return false
	}
	if err := tx.Commit(ctx); err != nil {
		return false
	}
	h.notify()
	return true
}

func (h *PaymentDueTrigger) enqueueNewPaymentActionReminder(ctx context.Context) bool {
	tx, err := h.P.Begin(ctx)
	if err != nil {
		return false
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
SELECT s.workflow_id::text, s.workflow_version_id::text, s.action_key::text, s.remind_every_days,
       s.days_before_due, s.project_id::text, cs.company_id::text, cs.client_id::text,
       MIN(cs.next_due_date)::text AS due_date, COALESCE(SUM(cs.amount),0)::float8,
       CASE
         WHEN BOOL_OR(COALESCE(cs.metadata_json #>> '{invoice_scenario,key}', cs.metadata_json ->> 'invoice_scenario_key', '') = 'kontur_elba_invoice') THEN 'kontur_elba'
         WHEN BOOL_OR(COALESCE(cs.metadata_json #>> '{invoice_scenario,key}', cs.metadata_json ->> 'invoice_scenario_key', '') = 'email_telegram_invoice') THEN 'email'
         ELSE 'bank'
       END AS action_source
  FROM workflow_client_payment_due_state s
  JOIN pulse_read.workflows_v1 w ON w.workflow_id = s.workflow_id
  JOIN client_services cs ON cs.project_id = s.project_id
 WHERE s.enabled = true
   AND w.status = 'active'
   AND cs.status = 'active'
   AND cs.next_due_date IS NOT NULL
   AND cs.next_due_date <= (CURRENT_DATE + (s.days_before_due * interval '1 day'))::date
   AND NOT EXISTS (
     SELECT 1 FROM payment_actions pa
      WHERE pa.client_id = cs.client_id
        AND pa.project_id = cs.project_id
        AND pa.status IN ('pending','sent','snoozed')
   )
 GROUP BY s.workflow_id, s.workflow_version_id, s.action_key, s.remind_every_days,
          s.days_before_due, s.project_id, cs.company_id, cs.client_id
 ORDER BY MIN(cs.next_due_date)
 LIMIT 1`)
	var workflowID, versionID, actionKey, projectID, companyID, clientID, dueDate string
	var actionSource string
	var remindEveryDays, daysBeforeDue int
	var total float64
	if err := row.Scan(&workflowID, &versionID, &actionKey, &remindEveryDays, &daysBeforeDue, &projectID, &companyID, &clientID, &dueDate, &total, &actionSource); err != nil {
		return false
	}
	paymentActionID := uuid.New()
	tag, err := tx.Exec(ctx, `
INSERT INTO payment_actions (
  payment_action_id, company_id, client_id, project_id, status, source,
  period_start, period_end, due_date, total_amount, currency,
  last_reminded_at, remind_after_at)
VALUES ($1,$2,$3,$4,'sent',$8,$5::date,$5::date,$5::date,$6,'RUB',
        clock_timestamp(), clock_timestamp() + ($7 * interval '1 day'))
ON CONFLICT DO NOTHING`,
		paymentActionID.String(), companyID, clientID, projectID, dueDate, total, clampReminderDays(remindEveryDays, 7), firstNonEmpty(actionSource, "bank"))
	if err != nil {
		slog.Warn("client payment action insert failed", slog.Any("err", err))
		return false
	}
	if tag.RowsAffected() == 0 {
		return false
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO payment_action_items (
  payment_action_id, client_service_id, client_service_type_id, title, due_date, amount, currency)
SELECT $1, cs.client_service_id, cs.client_service_type_id, cs.name, cs.next_due_date,
       COALESCE(cs.amount,0), cs.currency
  FROM client_services cs
 WHERE cs.client_id = $2
   AND cs.project_id = $3
   AND cs.status = 'active'
   AND cs.next_due_date IS NOT NULL
   AND cs.next_due_date <= (CURRENT_DATE + ($4 * interval '1 day'))::date
ON CONFLICT DO NOTHING`,
		paymentActionID.String(), clientID, projectID, clampDaysBeforeDue(daysBeforeDue, 7)); err != nil {
		slog.Warn("client payment action item insert failed", slog.Any("err", err))
		return false
	}
	if _, err := tx.Exec(ctx, `
UPDATE payment_actions pa
   SET total_amount = sub.total
  FROM (
    SELECT COALESCE(SUM(amount),0)::numeric AS total
      FROM payment_action_items
     WHERE payment_action_id = $1
  ) sub
 WHERE pa.payment_action_id = $1`, paymentActionID.String()); err != nil {
		return false
	}
	inputJSON, err := paymentActionRunInput(ctx, tx, paymentActionID.String(), actionKey, versionID)
	if err != nil {
		slog.Warn("client payment trigger input failed", slog.Any("err", err))
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
		EventMessage:      "client payment reminder queued",
	}); err != nil {
		slog.Warn("client payment trigger run insert failed", slog.String("workflow_id", workflowID), slog.Any("err", err))
		return false
	}
	if _, err := tx.Exec(ctx, `
UPDATE payment_actions
   SET workflow_run_id = $2,
       updated_at = clock_timestamp()
 WHERE payment_action_id = $1`, paymentActionID.String(), runID.String()); err != nil {
		return false
	}
	if err := tx.Commit(ctx); err != nil {
		return false
	}
	h.notify()
	return true
}

func paymentActionRunInput(ctx context.Context, tx pgx.Tx, paymentActionID, actionKey, versionID string) (string, error) {
	var row struct {
		CompanyID    string
		ClientID     string
		ClientName   string
		ClientType   string
		INN          string
		KPP          string
		ProjectID    string
		ProjectName  string
		ProjectSlug  string
		Status       string
		DueDate      string
		PeriodStart  string
		PeriodEnd    string
		Total        float64
		Currency     string
		ContactName  string
		ContactEmail string
	}
	err := tx.QueryRow(ctx, `
SELECT pa.company_id::text, pa.client_id::text, cl.name, cl.client_type,
       COALESCE(cl.inn,''), COALESCE(cl.kpp,''), pa.project_id::text, p.name, p.slug,
       pa.status, pa.due_date::text, pa.period_start::text, pa.period_end::text,
       pa.total_amount::float8, pa.currency, COALESCE(pc.contact_name,''), COALESCE(pc.email,'')
  FROM payment_actions pa
  JOIN clients cl ON cl.client_id = pa.client_id
  JOIN pulse_read.projects_v1 p ON p.project_id = pa.project_id
  LEFT JOIN LATERAL (
    SELECT NULLIF(TRIM(CONCAT_WS(' ', NULLIF(cc.first_name,''), NULLIF(cc.last_name,''))), '') AS contact_name,
           cc.email
      FROM client_contacts cc
     WHERE cc.client_id = pa.client_id
       AND cc.status = 'active'
       AND COALESCE(cc.email, '') <> ''
     ORDER BY cc.is_primary DESC, cc.updated_at DESC
     LIMIT 1
  ) pc ON true
 WHERE pa.payment_action_id = $1`, paymentActionID).Scan(
		&row.CompanyID, &row.ClientID, &row.ClientName, &row.ClientType, &row.INN, &row.KPP,
		&row.ProjectID, &row.ProjectName, &row.ProjectSlug, &row.Status, &row.DueDate,
		&row.PeriodStart, &row.PeriodEnd, &row.Total, &row.Currency, &row.ContactName, &row.ContactEmail,
	)
	if err != nil {
		return "", err
	}
	items, err := tx.Query(ctx, `
SELECT i.payment_action_item_id::text, i.client_service_id::text, i.title, i.due_date::text,
       i.amount::float8, i.currency, COALESCE(st.service_type_key,''),
       COALESCE(cs.metadata_json #>> '{invoice_scenario,key}', cs.metadata_json ->> 'invoice_scenario_key', '')
  FROM payment_action_items i
  LEFT JOIN client_services cs ON cs.client_service_id = i.client_service_id
  LEFT JOIN client_service_types st ON st.client_service_type_id = i.client_service_type_id
 WHERE i.payment_action_id = $1
 ORDER BY i.due_date, i.title`, paymentActionID)
	if err != nil {
		return "", err
	}
	defer items.Close()
	services := make([]map[string]any, 0)
	invoiceScenarioKey := ""
	for items.Next() {
		var paymentActionItemID, serviceID, title, dueDate, currency, serviceType, itemScenarioKey string
		var amount float64
		if err := items.Scan(&paymentActionItemID, &serviceID, &title, &dueDate, &amount, &currency, &serviceType, &itemScenarioKey); err != nil {
			return "", err
		}
		itemScenarioKey = strings.TrimSpace(itemScenarioKey)
		if invoiceScenarioKey == "" {
			invoiceScenarioKey = itemScenarioKey
		}
		services = append(services, map[string]any{
			"payment_action_item_id": paymentActionItemID,
			"client_service_id":      serviceID,
			"service_type_key":       serviceType,
			"invoice_scenario_key":   itemScenarioKey,
			"name":                   title,
			"due_date":               dueDate,
			"amount":                 amount,
			"currency":               currency,
		})
	}
	if err := items.Err(); err != nil {
		return "", err
	}
	paymentURL := paymentActionURL(paymentActionID)
	input := map[string]any{
		"_pulse": map[string]any{
			"source_type":         "client_payment_due",
			"trigger_action_key":  actionKey,
			"workflow_version_id": versionID,
			"payment_action_id":   paymentActionID,
			"payment_action_url":  paymentURL,
			"project_id":          row.ProjectID,
			"client_id":           row.ClientID,
		},
		"client": map[string]any{
			"client_id":   row.ClientID,
			"company_id":  row.CompanyID,
			"name":        row.ClientName,
			"client_type": row.ClientType,
			"inn":         row.INN,
			"kpp":         row.KPP,
			"primary_contact": map[string]any{
				"name":  firstNonEmpty(row.ContactName, row.ClientName),
				"email": row.ContactEmail,
			},
		},
		"project": map[string]any{
			"project_id": row.ProjectID,
			"name":       row.ProjectName,
			"slug":       row.ProjectSlug,
		},
		"services": services,
		"totals": map[string]any{
			"amount":       row.Total,
			"currency":     row.Currency,
			"due_date":     row.DueDate,
			"period_start": row.PeriodStart,
			"period_end":   row.PeriodEnd,
			"status":       row.Status,
			"status_label": paymentStatusLabel(row.Status),
		},
		"payment_action": map[string]any{
			"payment_action_id":    paymentActionID,
			"status":               row.Status,
			"status_label":         paymentStatusLabel(row.Status),
			"url":                  paymentURL,
			"invoice_scenario_key": invoiceScenarioKey,
			"due_date":             row.DueDate,
			"period_start":         row.PeriodStart,
			"period_end":           row.PeriodEnd,
			"amount":               row.Total,
			"currency":             row.Currency,
		},
		"payment_action_id": paymentActionID,
	}
	out, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
