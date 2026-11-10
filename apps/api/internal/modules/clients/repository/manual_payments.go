package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) RecordManualPaymentTx(ctx context.Context, clientID, uid uuid.UUID, in ManualPaymentRecordInput) (*ManualPaymentRecordResult, error) {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return nil, err
	}
	serviceID, err := uuid.Parse(strings.TrimSpace(in.ClientServiceID))
	if err != nil || serviceID == uuid.Nil {
		return nil, ErrPaymentNoService
	}
	idempotencyKey := strings.TrimSpace(in.IdempotencyKey)
	if idempotencyKey == "" {
		return nil, ErrIdempotencyRequired
	}

	type serviceForPayment struct {
		CompanyID      string
		ClientID       string
		ClientName     string
		ProjectID      string
		ProjectName    string
		ProjectSlug    string
		ServiceID      string
		ServiceTypeID  string
		ServiceTypeKey string
		ServiceName    string
		BillingPeriod  string
		NextDueDate    string
		Amount         float64
		Currency       string
	}
	var service serviceForPayment
	err = c.P.QueryRow(ctx, `
SELECT cs.company_id::text, cs.client_id::text, cl.name,
       cs.project_id::text, p.name, p.slug,
       cs.client_service_id::text, cs.client_service_type_id::text,
       COALESCE(st.service_type_key, ''), cs.name, cs.billing_period,
       COALESCE(cs.next_due_date::text, ''), COALESCE(cs.amount,0)::float8, cs.currency
  FROM client_services cs
  JOIN clients cl ON cl.client_id = cs.client_id
  JOIN pulse_read.projects_v1 p ON p.project_id = cs.project_id
  LEFT JOIN client_service_types st ON st.client_service_type_id = cs.client_service_type_id
 WHERE cs.client_service_id = $1
   AND cs.client_id = $2
   AND cs.status = 'active'`, serviceID.String(), clientID.String()).
		Scan(&service.CompanyID, &service.ClientID, &service.ClientName,
			&service.ProjectID, &service.ProjectName, &service.ProjectSlug,
			&service.ServiceID, &service.ServiceTypeID, &service.ServiceTypeKey, &service.ServiceName,
			&service.BillingPeriod, &service.NextDueDate, &service.Amount, &service.Currency)
	if err != nil {
		return nil, err
	}
	projectID, err := uuid.Parse(service.ProjectID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.ProjectID) != "" && strings.TrimSpace(in.ProjectID) != service.ProjectID {
		return nil, fmt.Errorf("client_service_id does not belong to project_id")
	}
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermWrite); err != nil {
		return nil, err
	}
	if err := authz.RequireProjectFinance(ctx, c.P, projectID); err != nil {
		return nil, err
	}

	periodStart, err := parsePaymentDate(firstNonEmpty(in.PeriodStart, service.NextDueDate))
	if err != nil {
		return nil, fmt.Errorf("period_start: %w", err)
	}
	periodEnd := PaymentPeriodEnd(periodStart, service.BillingPeriod)
	if strings.TrimSpace(in.PeriodEnd) != "" {
		periodEnd, err = parsePaymentDate(in.PeriodEnd)
		if err != nil {
			return nil, fmt.Errorf("period_end: %w", err)
		}
	}
	paidAt, err := parsePaymentTimestamp(in.PaidAt)
	if err != nil {
		return nil, fmt.Errorf("paid_at: %w", err)
	}
	amount := service.Amount
	if in.Amount != nil {
		amount = *in.Amount
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	currency := normalizeDefault(in.Currency, service.Currency)
	if currency == "" {
		currency = "RUB"
	}
	if currency != "RUB" {
		return nil, fmt.Errorf("unsupported currency %q", currency)
	}
	paymentMethod, err := normalizePaymentMethod(in.PaymentMethod)
	if err != nil {
		return nil, err
	}
	metadata := json.RawMessage(bytes.TrimSpace(in.MetadataJSON))
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	eventID := uuid.New()
	actionID := uuid.New()
	factID := uuid.New()
	eventSubject := "client/" + clientID.String() + "/service/" + serviceID.String()
	eventSource := "manual"
	triggerType := "manual_payment_trigger"

	tx, err := c.transaction()
	if err != nil {
		return nil, err
	}

	tag, err := tx.Exec(ctx, `
INSERT INTO payment_events (
  payment_event_id, event_type, event_source, event_subject, occurred_at, idempotency_key,
  company_id, project_id, client_id, client_service_id, amount, currency, period_start,
  period_end, paid_at, source, payment_method, trigger_type, status, metadata_json, created_by_user_id)
VALUES ($1,'payment.recorded',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::date,$13::date,$14,$15,$16,$17,'processing',$18::jsonb,$19)
ON CONFLICT (event_source, idempotency_key) DO NOTHING`,
		eventID.String(), eventSource, eventSubject, paidAt, idempotencyKey,
		service.CompanyID, service.ProjectID, clientID.String(), serviceID.String(), amount, currency,
		periodStart, periodEnd, paidAt, eventSource, paymentMethod, triggerType, string(metadata), uid.String())
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		row, err := c.paymentEventByIdempotency(ctx, tx, eventSource, idempotencyKey)
		if err != nil {
			return nil, err
		}
		return &ManualPaymentRecordResult{
			EventID:         row.EventID,
			EventType:       row.EventType,
			EventSource:     row.EventSource,
			EventSubject:    row.EventSubject,
			IdempotencyKey:  row.IdempotencyKey,
			PaymentActionID: row.PaymentActionID,
			PaymentFactID:   row.PaymentFactID,
			Status:          row.Status,
			Duplicate:       true,
		}, nil
	}

	var existingActionID string
	_ = tx.QueryRow(ctx, `
SELECT pa.payment_action_id::text
  FROM payment_actions pa
  JOIN payment_action_items i ON i.payment_action_id = pa.payment_action_id
 WHERE pa.client_id = $1
   AND pa.project_id = $2
   AND pa.period_start = $3::date
   AND pa.status IN ('pending','sent','snoozed')
   AND i.client_service_id = $4
 ORDER BY pa.created_at
 LIMIT 1
 FOR UPDATE OF pa`, clientID.String(), service.ProjectID, periodStart, serviceID.String()).Scan(&existingActionID)
	if existingActionID != "" {
		actionID, err = uuid.Parse(existingActionID)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
UPDATE payment_action_items
   SET amount = $2, currency = $3
 WHERE payment_action_id = $1 AND client_service_id = $4`, actionID.String(), amount, currency, serviceID.String())
		if err != nil {
			return nil, err
		}
	} else {
		_, err = tx.Exec(ctx, `
INSERT INTO payment_actions (
  payment_action_id, company_id, client_id, project_id, status, source,
  period_start, period_end, due_date, total_amount, currency, created_by_user_id)
VALUES ($1,$2,$3,$4,'pending','manual',$5::date,$6::date,$5::date,$7,$8,$9)`,
			actionID.String(), service.CompanyID, clientID.String(), service.ProjectID,
			periodStart, periodEnd, amount, currency, uid.String())
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO payment_action_items (
  payment_action_id, client_service_id, client_service_type_id, title, due_date, amount, currency)
VALUES ($1,$2,$3,$4,$5::date,$6,$7)
ON CONFLICT DO NOTHING`,
			actionID.String(), serviceID.String(), service.ServiceTypeID, service.ServiceName, periodStart, amount, currency)
		if err != nil {
			return nil, err
		}
	}

	next, err := AdvanceDueDate(periodStart, service.BillingPeriod)
	if err != nil {
		return nil, err
	}
	nextPeriodStart := ""
	if next != nil {
		nextPeriodStart = *next
	}
	if strings.TrimSpace(in.NextDueDate) != "" {
		nextOverride, err := parsePaymentDate(in.NextDueDate)
		if err != nil {
			return nil, fmt.Errorf("next_due_date: %w", err)
		}
		nextPeriodStart = nextOverride
		next = &nextPeriodStart
	}
	links := paymentLinks(service.ClientID, service.ClientName, service.ServiceID, service.ServiceName, actionID.String(), factID.String(), amount, currency, periodStart, periodEnd)
	paymentEvent := map[string]any{
		"event_id":              eventID.String(),
		"event_type":            "payment.recorded",
		"event_source":          eventSource,
		"event_subject":         eventSubject,
		"occurred_at":           paidAt.UTC().Format(time.RFC3339),
		"idempotency_key":       idempotencyKey,
		"client_id":             clientID.String(),
		"client_service_id":     serviceID.String(),
		"payment_action_id":     actionID.String(),
		"payment_id":            factID.String(),
		"payment_fact_id":       factID.String(),
		"amount":                amount,
		"currency":              currency,
		"period_start":          periodStart,
		"period_end":            periodEnd,
		"period_label":          paymentPeriodLabel(periodStart, periodEnd),
		"next_period_start":     nextPeriodStart,
		"next_payment_due_date": nextPeriodStart,
		"paid_at":               paidAt.UTC().Format(time.RFC3339),
		"source":                eventSource,
		"source_label":          paymentSourcePublicLabel(eventSource),
		"payment_method":        paymentMethod,
		"payment_method_label":  paymentMethodPublicLabel(paymentMethod),
		"service_name":          service.ServiceName,
		"trigger_type":          triggerType,
		"links":                 links,
		"client": map[string]any{
			"client_id":  service.ClientID,
			"company_id": service.CompanyID,
			"name":       service.ClientName,
			"url":        pulsePublicURL("/clients/" + service.ClientID),
			"href":       clientWorkspaceHref(service.ClientID),
		},
		"client_service": map[string]any{
			"client_id":         service.ClientID,
			"client_service_id": service.ServiceID,
			"name":              service.ServiceName,
			"url":               pulsePublicURL(clientServiceWorkspaceHref(service.ClientID, service.ServiceID)),
			"href":              clientServiceWorkspaceHref(service.ClientID, service.ServiceID),
		},
		"metadata": json.RawMessage(metadata),
	}
	rawPayload := map[string]any{
		"payment_event":  paymentEvent,
		"manual_payment": in,
	}
	rawPayloadJSON, err := json.Marshal(rawPayload)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO payment_facts (
  payment_fact_id, payment_action_id, company_id, client_id, project_id, source, payment_method,
  amount, currency, paid_at, raw_payload_json, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12)`,
		factID.String(), actionID.String(), service.CompanyID, clientID.String(), service.ProjectID,
		eventSource, paymentMethod, amount, currency, paidAt, string(rawPayloadJSON), uid.String())
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
UPDATE payment_actions
   SET status = 'paid',
       source = $2,
       payment_method = $3,
       total_amount = $4,
       confirmed_at = $5,
       comment = COALESCE($6, comment),
       updated_at = now()
 WHERE payment_action_id = $1`, actionID.String(), eventSource, paymentMethod, amount, paidAt, nullStr(in.Comment))
	if err != nil {
		return nil, err
	}
	if next == nil {
		_, err = tx.Exec(ctx, `
UPDATE client_services
   SET status = 'completed', next_due_date = NULL, updated_at = now()
 WHERE client_service_id = $1`, serviceID.String())
	} else {
		_, err = tx.Exec(ctx, `
UPDATE client_services
   SET next_due_date = $2::date, updated_at = now()
 WHERE client_service_id = $1`, serviceID.String(), *next)
	}
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
UPDATE payment_events
   SET payment_action_id = $2,
       payment_fact_id = $3,
       status = 'processed',
       processed_at = now()
 WHERE payment_event_id = $1`, eventID.String(), actionID.String(), factID.String())
	if err != nil {
		return nil, err
	}
	if err := c.enqueuePaymentSavedModuleTriggers(ctx, tx, paymentSavedModuleEvent{
		CompanyID:       service.CompanyID,
		ClientID:        clientID.String(),
		ClientName:      service.ClientName,
		ProjectID:       service.ProjectID,
		ProjectName:     service.ProjectName,
		ProjectSlug:     service.ProjectSlug,
		PaymentActionID: actionID.String(),
		PaymentFactID:   factID.String(),
		Total:           amount,
		Currency:        currency,
		PaidAt:          paidAt,
		Source:          eventSource,
		PaymentMethod:   paymentMethod,
	}); err != nil {
		return nil, err
	}
	return &ManualPaymentRecordResult{
		EventID:         eventID.String(),
		EventType:       "payment.recorded",
		EventSource:     eventSource,
		EventSubject:    eventSubject,
		IdempotencyKey:  idempotencyKey,
		PaymentActionID: actionID.String(),
		PaymentFactID:   factID.String(),
		Status:          "processed",
		Duplicate:       false,
	}, nil
}
