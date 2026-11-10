package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (r *Runner) executePaymentDueTrigger(ctx context.Context, node actionPlan, runInput []byte) actionExecution {
	return r.buildPaymentCycleCollectExecution(ctx, node, runInput, clampDaysBeforeDue(node.Config.DaysBeforeDue, 7))
}

func (r *Runner) executePaymentCycleCollect(ctx context.Context, node actionPlan, runInput []byte) actionExecution {
	return r.buildPaymentCycleCollectExecution(ctx, node, runInput, clampDaysBeforeDue(node.Config.DaysBeforeDue, 30))
}

func (r *Runner) buildPaymentCycleCollectExecution(ctx context.Context, node actionPlan, runInput []byte, daysBeforeDue int) actionExecution {
	input := actionInput(node, runInput)
	projectID := projectIDFromRunInput(runInput)
	if projectID == "" {
		err := errors.New("project_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "clients", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	if _, err := uuid.Parse(projectID); err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "clients", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	rows, err := r.collectPaymentCycleRows(ctx, projectID, daysBeforeDue)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "clients", "ok": false, "error": err.Error()}, true, "database", err)
	}
	payments := make([]map[string]any, 0, len(rows))
	scheduled := make([]map[string]any, 0)
	sent := make([]map[string]any, 0)
	counts := map[string]int{"scheduled": 0, "sent": 0, "total": 0}
	var totalAmount float64
	for _, row := range rows {
		item := row.Map()
		payments = append(payments, item)
		totalAmount += row.TotalAmount
		counts["total"]++
		if row.CycleState == "sent" {
			sent = append(sent, item)
			counts["sent"]++
		} else {
			scheduled = append(scheduled, item)
			counts["scheduled"]++
		}
	}
	output := map[string]any{
		"provider":              "clients",
		"ok":                    true,
		"action_key":            node.ActionKey,
		"action_type_key":       node.ActionTypeKey,
		"project_id":            projectID,
		"days_before_due":       daysBeforeDue,
		"payments":              payments,
		"scheduled":             scheduled,
		"sent":                  sent,
		"counts":                counts,
		"payment_due":           node.ActionTypeKey == "clients.payment_due.trigger",
		"payment_cycle_collect": node.ActionTypeKey == "clients.payment_cycle.collect",
		"totals": map[string]any{
			"amount":   totalAmount,
			"currency": "RUB",
		},
	}
	projectName := projectNameFromPaymentRows(rows)
	event := paymentDigestEventFromRunInput(runInput)
	currentPayment := currentPaymentContextFromEvent(event)
	paymentSnapshot := map[string]any{
		"summary": map[string]any{
			"project_id":        projectID,
			"project_name":      projectName,
			"days_before_due":   daysBeforeDue,
			"clients_count":     counts["total"],
			"scheduled_count":   counts["scheduled"],
			"sent_count":        counts["sent"],
			"total_amount":      totalAmount,
			"currency":          "RUB",
			"triggered_payment": currentPayment,
		},
		"clients": payments,
	}
	digest := buildPaymentDigest(rows, projectName, projectID, event)
	output["project"] = map[string]any{
		"project_id": projectID,
		"name":       projectName,
	}
	output["current_payment"] = currentPayment
	output["payment_snapshot"] = paymentSnapshot
	output["expose"] = map[string]any{
		"digest_subject":   digest.Subject,
		"digest_telegram":  digest.TelegramHTML,
		"digest_max":       digest.MaxText,
		"digest_html":      digest.EmailHTML,
		"digest_plain":     digest.PlainText,
		"current_payment":  currentPayment,
		"payment_snapshot": paymentSnapshot,
		"counts":           counts,
		"totals":           output["totals"],
	}
	provider := map[string]any{
		"provider":                   "clients",
		"ok":                         true,
		"provider_response_redacted": true,
		"project_id":                 projectID,
		"rows":                       len(rows),
	}
	outputJSON, _ := json.Marshal(output)
	providerJSON, _ := json.Marshal(provider)
	return actionExecution{InputJSON: input, OutputJSON: string(outputJSON), ProviderResponseJSON: string(providerJSON)}
}

func projectNameFromPaymentRows(rows []paymentCycleRow) string {
	for _, row := range rows {
		if name := strings.TrimSpace(row.ProjectName); name != "" {
			return name
		}
	}
	return ""
}
