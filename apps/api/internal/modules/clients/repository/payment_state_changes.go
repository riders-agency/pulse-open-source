package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) ConfirmPaymentActionTx(ctx context.Context, actionID, uid uuid.UUID, in ConfirmPaymentInput) error {
	if err := c.assertPaymentAction(ctx, actionID, uid, authz.PermWrite); err != nil {
		return err
	}
	tx, err := c.transaction()
	if err != nil {
		return err
	}
	var companyID, clientID, projectID, clientName, projectName, projectSlug string
	var status string
	err = tx.QueryRow(ctx, `
SELECT pa.company_id::text, pa.client_id::text, pa.project_id::text, pa.status,
       cl.name, p.name, p.slug
  FROM payment_actions pa
  JOIN clients cl ON cl.client_id = pa.client_id
  JOIN pulse_read.projects_v1 p ON p.project_id = pa.project_id
 WHERE pa.payment_action_id = $1
 FOR UPDATE OF pa`, actionID.String()).Scan(&companyID, &clientID, &projectID, &status, &clientName, &projectName, &projectSlug)
	if err != nil {
		return err
	}
	if status == "paid" || status == "cancelled" {
		return ErrPaymentClosed
	}
	for _, item := range in.Items {
		if item.Amount == nil || strings.TrimSpace(item.PaymentActionItemID) == "" {
			continue
		}
		_, err := tx.Exec(ctx, `
UPDATE payment_action_items
   SET amount = $3
 WHERE payment_action_id = $1 AND payment_action_item_id = $2`, actionID.String(), item.PaymentActionItemID, *item.Amount)
		if err != nil {
			return err
		}
	}
	var total float64
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(SUM(amount),0)::float8 FROM payment_action_items WHERE payment_action_id = $1`,
		actionID.String()).Scan(&total); err != nil {
		return err
	}
	if in.Amount != nil {
		total = *in.Amount
	}
	paidAt := time.Now().UTC()
	if strings.TrimSpace(in.PaidAt) != "" {
		parsed, err := parsePaymentTimestamp(in.PaidAt)
		if err != nil {
			return err
		}
		paidAt = parsed
	}
	source := normalizeDefault(in.Source, "manual")
	paymentMethod, err := normalizePaymentMethod(in.PaymentMethod)
	if err != nil {
		return err
	}
	factID := uuid.New()
	raw, _ := json.Marshal(map[string]any{
		"payment_event": map[string]any{
			"event_type":           "payment.recorded",
			"event_source":         source,
			"source":               source,
			"source_label":         paymentSourcePublicLabel(source),
			"payment_method":       paymentMethod,
			"payment_method_label": paymentMethodPublicLabel(paymentMethod),
			"trigger_type":         "manual_payment_trigger",
		},
		"confirm_payment": in,
	})
	_, err = tx.Exec(ctx, `
INSERT INTO payment_facts (
  payment_fact_id, payment_action_id, company_id, client_id, project_id, source, payment_method, amount, currency,
  paid_at, raw_payload_json, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'RUB',$9,$10::jsonb,$11)`,
		factID.String(), actionID.String(), companyID, clientID, projectID, source, paymentMethod, total, paidAt, string(raw), uid.String())
	if err != nil {
		return err
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
 WHERE payment_action_id = $1`, actionID.String(), source, paymentMethod, total, paidAt, nullStr(in.Comment))
	if err != nil {
		return err
	}
	serviceRows, err := tx.Query(ctx, `
SELECT DISTINCT cs.client_service_id::text, COALESCE(cs.next_due_date::text,''), cs.billing_period
  FROM payment_action_items i
  JOIN client_services cs ON cs.client_service_id = i.client_service_id
 WHERE i.payment_action_id = $1`, actionID.String())
	if err != nil {
		return err
	}
	type serviceDueAdvance struct {
		ServiceID string
		DueDate   string
		Period    string
	}
	services := make([]serviceDueAdvance, 0)
	for serviceRows.Next() {
		var row serviceDueAdvance
		if err := serviceRows.Scan(&row.ServiceID, &row.DueDate, &row.Period); err != nil {
			serviceRows.Close()
			return err
		}
		services = append(services, row)
	}
	if err := serviceRows.Err(); err != nil {
		serviceRows.Close()
		return err
	}
	serviceRows.Close()
	for _, row := range services {
		next, err := AdvanceDueDate(row.DueDate, row.Period)
		if err != nil {
			return err
		}
		if strings.TrimSpace(in.NextDueDate) != "" {
			nextOverride, err := parsePaymentDate(in.NextDueDate)
			if err != nil {
				return fmt.Errorf("next_due_date: %w", err)
			}
			next = &nextOverride
		}
		if next == nil {
			_, err = tx.Exec(ctx, `
UPDATE client_services SET status = 'completed', next_due_date = NULL, updated_at = now()
 WHERE client_service_id = $1`, row.ServiceID)
		} else {
			_, err = tx.Exec(ctx, `
UPDATE client_services SET next_due_date = $2::date, updated_at = now()
 WHERE client_service_id = $1`, row.ServiceID, *next)
		}
		if err != nil {
			return err
		}
	}
	if err := c.enqueuePaymentSavedModuleTriggers(ctx, tx, paymentSavedModuleEvent{
		CompanyID:       companyID,
		ClientID:        clientID,
		ClientName:      clientName,
		ProjectID:       projectID,
		ProjectName:     projectName,
		ProjectSlug:     projectSlug,
		PaymentActionID: actionID.String(),
		PaymentFactID:   factID.String(),
		Total:           total,
		Currency:        "RUB",
		PaidAt:          paidAt,
		Source:          source,
		PaymentMethod:   paymentMethod,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Clients) SnoozePaymentAction(ctx context.Context, actionID, uid uuid.UUID, in SnoozePaymentInput) error {
	if err := c.assertPaymentAction(ctx, actionID, uid, authz.PermWrite); err != nil {
		return err
	}
	until := time.Now().UTC().Add(7 * 24 * time.Hour)
	if strings.TrimSpace(in.SnoozedUntil) != "" {
		parsed, err := parsePaymentTimestamp(in.SnoozedUntil)
		if err != nil {
			return err
		}
		until = parsed
	}
	tag, err := c.P.Exec(ctx, `
UPDATE payment_actions
   SET status = 'snoozed',
       snoozed_until = $2,
       remind_after_at = $2,
       comment = COALESCE($3, comment),
       updated_at = now()
 WHERE payment_action_id = $1
   AND status IN ('pending','sent','snoozed')`, actionID.String(), until, nullStr(in.Comment))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentClosed
	}
	return nil
}
