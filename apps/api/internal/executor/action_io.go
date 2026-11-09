package executor

import (
	"encoding/json"
	"strings"
)

func actionInput(node actionPlan, runInput []byte) string {
	if node.ActionTypeKey == "webhook.trigger" ||
		node.ActionTypeKey == "manual.trigger" ||
		node.ActionTypeKey == "telegram.relay.trigger" ||
		node.ActionTypeKey == "clients.manual_payment.trigger" ||
		node.ActionTypeKey == "clients.payment_due.trigger" ||
		node.ActionTypeKey == "clients.payment_cycle.collect" {
		if len(runInput) > 0 {
			return string(runInput)
		}
		return "{}"
	}
	b, _ := json.Marshal(map[string]any{
		"config":           node.Config,
		"success_criteria": node.Config.SuccessCriteria,
	})
	return string(b)
}

func skippedOutput(node actionPlan) string {
	return skippedOutputWithReason(node, "test_run_downstream")
}

func skippedOutputWithReason(node actionPlan, reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "skipped"
	}
	output := map[string]any{
		"skipped":         true,
		"reason":          reason,
		"action_key":      node.ActionKey,
		"action_type_key": node.ActionTypeKey,
	}
	b, _ := json.Marshal(output)
	return string(b)
}
