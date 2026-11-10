package runtime

import (
	"context"
	"encoding/json"
)

func (r *Runner) executeManualPaymentTrigger(_ context.Context, node actionPlan, runInput []byte) actionExecution {
	input := actionInput(node, runInput)
	event := paymentDigestEventFromRunInput(runInput)
	currentPayment := currentPaymentContextFromEvent(event)
	paymentEvent := paymentEventContextFromEvent(event)
	output := map[string]any{
		"provider":        "clients",
		"ok":              true,
		"action_key":      node.ActionKey,
		"action_type_key": node.ActionTypeKey,
		"manual_payment":  true,
		"current_payment": currentPayment,
		"payment_event":   paymentEvent,
		"expose": map[string]any{
			"current_payment": currentPayment,
			"payment_event":   paymentEvent,
		},
	}
	provider := map[string]any{
		"provider":                   "clients",
		"ok":                         true,
		"provider_response_redacted": true,
		"payment_action_id":          event.PaymentActionID,
	}
	outputJSON, _ := json.Marshal(output)
	providerJSON, _ := json.Marshal(provider)
	return actionExecution{InputJSON: input, OutputJSON: string(outputJSON), ProviderResponseJSON: string(providerJSON)}
}
