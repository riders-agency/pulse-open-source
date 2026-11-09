package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/domainevent"
)

type paymentSavedModuleEvent struct {
	CompanyID       string
	ClientID        string
	ClientName      string
	ProjectID       string
	ProjectName     string
	ProjectSlug     string
	PaymentActionID string
	PaymentFactID   string
	Total           float64
	Currency        string
	PaidAt          time.Time
	Source          string
	PaymentMethod   string
}

func (c *Clients) enqueuePaymentSavedModuleTriggers(ctx context.Context, tx pgx.Tx, ev paymentSavedModuleEvent) error {
	dbTriggerInstalled, err := paymentSavedDBTriggerInstalled(ctx, tx)
	if err != nil {
		return err
	}
	if dbTriggerInstalled {
		return nil
	}
	rows, err := tx.Query(ctx, `
SELECT mt.module_workflow_trigger_id::text,
       mt.workflow_id::text,
       COALESCE(mt.workflow_version_id::text, w.current_workflow_version_id::text, '') AS workflow_version_id,
       mt.config_json
  FROM pulse_read.module_workflow_triggers_v1 mt
  JOIN pulse_read.workflows_v1 w ON w.workflow_id = mt.workflow_id
 WHERE mt.company_id = $1
   AND (mt.project_id = $2 OR mt.project_id IS NULL)
   AND mt.module_key = 'clients.payments'
   AND mt.event_key = 'payment.saved'
   AND mt.enabled = true
   AND w.status = 'active'
   AND COALESCE(mt.workflow_version_id, w.current_workflow_version_id) IS NOT NULL
 ORDER BY mt.project_id NULLS LAST, mt.created_at`, ev.CompanyID, ev.ProjectID)
	if err != nil {
		return err
	}
	type moduleTriggerRun struct {
		TriggerID         string
		WorkflowID        string
		WorkflowVersionID string
		ConfigJSON        json.RawMessage
	}
	pendingRuns := make([]moduleTriggerRun, 0)
	for rows.Next() {
		var triggerID, workflowID, workflowVersionID string
		var configRaw []byte
		if err := rows.Scan(&triggerID, &workflowID, &workflowVersionID, &configRaw); err != nil {
			rows.Close()
			return err
		}
		if workflowVersionID == "" {
			continue
		}
		configJSON := json.RawMessage(bytes.TrimSpace(configRaw))
		if len(configJSON) == 0 {
			configJSON = json.RawMessage(`{}`)
		}
		pendingRuns = append(pendingRuns, moduleTriggerRun{
			TriggerID:         triggerID,
			WorkflowID:        workflowID,
			WorkflowVersionID: workflowVersionID,
			ConfigJSON:        configJSON,
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	paidAt := ev.PaidAt.UTC().Format(time.RFC3339)
	var periodStart, periodEnd, nextPeriodStart, serviceSummary, clientServiceID, serviceName string
	_ = tx.QueryRow(ctx, `
SELECT COALESCE(pa.period_start::text, ''),
       COALESCE(pa.period_end::text, ''),
       COALESCE(MIN(cs.next_due_date)::text, ''),
       COALESCE(string_agg(DISTINCT NULLIF(i.title, ''), ', ' ORDER BY NULLIF(i.title, '')), ''),
       COALESCE((array_agg(i.client_service_id::text ORDER BY i.due_date, i.title) FILTER (WHERE i.client_service_id IS NOT NULL))[1], ''),
       COALESCE((array_agg(NULLIF(i.title, '') ORDER BY i.due_date, i.title) FILTER (WHERE NULLIF(i.title, '') IS NOT NULL))[1], '')
  FROM payment_actions pa
  LEFT JOIN payment_action_items i ON i.payment_action_id = pa.payment_action_id
  LEFT JOIN client_services cs ON cs.client_service_id = i.client_service_id
 WHERE pa.payment_action_id = $1
 GROUP BY pa.payment_action_id`, ev.PaymentActionID).
		Scan(&periodStart, &periodEnd, &nextPeriodStart, &serviceSummary, &clientServiceID, &serviceName)
	for _, trigger := range pendingRuns {
		runID := uuid.New()
		paymentID := firstNonEmpty(ev.PaymentFactID, ev.PaymentActionID)
		links := paymentLinks(ev.ClientID, ev.ClientName, clientServiceID, firstNonEmpty(serviceName, serviceSummary), ev.PaymentActionID, paymentID, ev.Total, ev.Currency, periodStart, periodEnd)
		currentPayment := map[string]any{
			"payment_id":            paymentID,
			"payment_action_id":     ev.PaymentActionID,
			"payment_fact_id":       ev.PaymentFactID,
			"client_id":             ev.ClientID,
			"client_service_id":     clientServiceID,
			"client_name":           ev.ClientName,
			"service_name":          firstNonEmpty(serviceName, serviceSummary),
			"service_summary":       serviceSummary,
			"amount":                ev.Total,
			"currency":              ev.Currency,
			"period_label":          paymentPeriodLabel(periodStart, periodEnd),
			"period_start":          periodStart,
			"period_end":            periodEnd,
			"next_period_start":     nextPeriodStart,
			"next_payment_due_date": nextPeriodStart,
			"paid_at":               paidAt,
			"source":                ev.Source,
			"source_label":          paymentSourcePublicLabel(ev.Source),
			"payment_method":        ev.PaymentMethod,
			"payment_method_label":  paymentMethodPublicLabel(ev.PaymentMethod),
			"links":                 links,
		}
		paymentEvent := map[string]any{
			"present":               true,
			"payment_action_id":     ev.PaymentActionID,
			"payment_id":            paymentID,
			"payment_fact_id":       ev.PaymentFactID,
			"source":                ev.Source,
			"source_label":          paymentSourcePublicLabel(ev.Source),
			"payment_method":        ev.PaymentMethod,
			"payment_method_label":  paymentMethodPublicLabel(ev.PaymentMethod),
			"amount":                ev.Total,
			"currency":              ev.Currency,
			"paid_at":               paidAt,
			"period_label":          paymentPeriodLabel(periodStart, periodEnd),
			"period_start":          periodStart,
			"period_end":            periodEnd,
			"next_period_start":     nextPeriodStart,
			"next_payment_due_date": nextPeriodStart,
			"service_summary":       serviceSummary,
			"service_name":          firstNonEmpty(serviceName, serviceSummary),
			"url":                   paymentActionPublicURL(ev.PaymentActionID),
			"links":                 links,
			"client": map[string]any{
				"client_id":  ev.ClientID,
				"company_id": ev.CompanyID,
				"name":       ev.ClientName,
				"url":        pulsePublicURL("/clients/" + ev.ClientID),
				"href":       clientWorkspaceHref(ev.ClientID),
			},
			"client_service": map[string]any{
				"client_id":         ev.ClientID,
				"client_service_id": clientServiceID,
				"name":              firstNonEmpty(serviceName, serviceSummary),
				"url":               pulsePublicURL(clientServiceWorkspaceHref(ev.ClientID, clientServiceID)),
				"href":              clientServiceWorkspaceHref(ev.ClientID, clientServiceID),
			},
			"project": map[string]any{
				"project_id": ev.ProjectID,
				"name":       ev.ProjectName,
				"slug":       ev.ProjectSlug,
				"url":        pulsePublicURL("/projects/" + ev.ProjectID),
			},
		}
		input := map[string]any{
			"_pulse": map[string]any{
				"source_type":                "module_trigger",
				"module_workflow_trigger_id": trigger.TriggerID,
				"module_key":                 "clients.payments",
				"event_key":                  "payment.saved",
				"workflow_version_id":        trigger.WorkflowVersionID,
				"company_id":                 ev.CompanyID,
				"project_id":                 ev.ProjectID,
				"client_id":                  ev.ClientID,
				"payment_action_id":          ev.PaymentActionID,
			},
			"current_payment": currentPayment,
			"payment_event":   paymentEvent,
			"module_trigger": map[string]any{
				"module_workflow_trigger_id": trigger.TriggerID,
				"module_key":                 "clients.payments",
				"event_key":                  "payment.saved",
				"config":                     trigger.ConfigJSON,
			},
			"payment_action": map[string]any{
				"payment_action_id": ev.PaymentActionID,
				"status":            "paid",
				"status_label":      "Оплачен",
				"amount":            ev.Total,
				"currency":          ev.Currency,
				"paid_at":           paidAt,
				"source":            ev.Source,
				"payment_method":    ev.PaymentMethod,
				"url":               paymentActionPublicURL(ev.PaymentActionID),
				"period_start":      periodStart,
				"period_end":        periodEnd,
				"next_period_start": nextPeriodStart,
				"service_summary":   serviceSummary,
				"links":             links,
			},
			"client": map[string]any{
				"client_id":  ev.ClientID,
				"company_id": ev.CompanyID,
				"name":       ev.ClientName,
				"url":        pulsePublicURL("/clients/" + ev.ClientID),
			},
			"project": map[string]any{
				"project_id": ev.ProjectID,
				"name":       ev.ProjectName,
				"slug":       ev.ProjectSlug,
				"url":        pulsePublicURL("/projects/" + ev.ProjectID),
			},
			"totals": map[string]any{
				"amount":       ev.Total,
				"currency":     ev.Currency,
				"status":       "paid",
				"status_label": "Оплачен",
				"paid_at":      paidAt,
				"period_start": periodStart,
				"period_end":   periodEnd,
			},
		}
		outboxPayload := map[string]any{
			"workflow_run_id":     runID.String(),
			"project_id":          ev.ProjectID,
			"workflow_id":         trigger.WorkflowID,
			"workflow_version_id": trigger.WorkflowVersionID,
			"source_type":         "module_trigger",
			"input_json":          input,
			"event_action_key":    "manual_payment_trigger",
			"event_message":       "module trigger payment.saved queued",
			"event_payload_json":  input,
		}
		outboxJSON, err := json.Marshal(outboxPayload)
		if err != nil {
			return err
		}
		if err := domainevent.Record(ctx, domainevent.Event{
			TypeKey:        "workflow.run.requested",
			AggregateType:  "workflow_run",
			AggregateID:    runID.String(),
			IdempotencyKey: "workflow.run.requested:" + runID.String(),
			PayloadJSON:    outboxJSON,
		}); err != nil {
			return err
		}
	}
	return nil
}

func paymentSavedDBTriggerInstalled(ctx context.Context, tx pgx.Tx) (bool, error) {
	var installed bool
	err := tx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM pg_trigger
   WHERE tgname = 'trg_payment_facts_enqueue_module_triggers'
     AND NOT tgisinternal
)`).Scan(&installed)
	return installed, err
}
