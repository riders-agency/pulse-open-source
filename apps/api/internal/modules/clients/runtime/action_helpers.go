package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func actionInput(node actionPlan, runInput []byte) string {
	if len(bytes.TrimSpace(runInput)) > 0 {
		return string(runInput)
	}
	return "{}"
}

func bytesTrim(b []byte) []byte {
	return bytes.TrimSpace(b)
}

func failedAction(inputJSON, outputJSON string, provider map[string]any, retryable bool, errorType string, err error) actionExecution {
	if strings.TrimSpace(outputJSON) == "" {
		outputJSON = "{}"
	}
	providerJSON, _ := json.Marshal(provider)
	return actionExecution{
		InputJSON:            inputJSON,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: string(providerJSON),
		Retryable:            retryable,
		ErrorType:            errorType,
		Err:                  err,
	}
}

func projectIDFromRunInput(runInput []byte) string {
	if len(bytes.TrimSpace(runInput)) == 0 {
		return ""
	}
	var root any
	if err := json.Unmarshal(runInput, &root); err != nil {
		return ""
	}
	for _, path := range []string{"project.project_id", "_pulse.project_id", "module_trigger.project_id"} {
		if value := stringAtRunPath(root, path); value != "" {
			return value
		}
	}
	return ""
}

func stringAtRunPath(root any, path string) string {
	value, ok := valueAtPathRaw(root, path)
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(templateValue(value)))
}

func valueAtPathRaw(root any, path string) (any, bool) {
	cur := root
	for _, part := range strings.Split(path, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		object, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		next, ok := object[part]
		if !ok || next == nil {
			return "", false
		}
		cur = next
	}
	return cur, true
}

func templateValue(value any) any {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%.0f", typed)
		}
		return fmt.Sprint(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(typed)
		return string(b)
	}
}

func publicAppBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(firstNonEmpty(
		os.Getenv("RIDERS_PULSE_PUBLIC_URL"),
		os.Getenv("NEXT_PUBLIC_APP_URL"),
		os.Getenv("PUBLIC_APP_URL"),
		"https://pulse.riders.agency",
	)), "/")
}

func paymentActionURL(paymentActionID string) string {
	return publicAppBaseURL() + "/payment-actions/" + strings.TrimSpace(paymentActionID)
}

func paymentStatusLabel(status string) string {
	switch strings.TrimSpace(status) {
	case "pending", "scheduled":
		return "Запланирован"
	case "sent", "needs_decision":
		return "Отправлен"
	case "snoozed":
		return "Отложен"
	case "paid":
		return "Оплачен"
	case "cancelled":
		return "Отменен"
	default:
		return strings.TrimSpace(status)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func clampDaysBeforeDue(value, fallback int) int {
	if value <= 0 {
		value = fallback
	}
	if value < 1 {
		return 1
	}
	if value > 90 {
		return 90
	}
	return value
}

func clampReminderDays(value, fallback int) int {
	if value == 0 {
		value = fallback
	}
	if value < 1 {
		return 1
	}
	if value > 365 {
		return 365
	}
	return value
}

func scheduleEnabled(cfg actionConfig, now time.Time) bool {
	if cfg.Enabled != nil && !*cfg.Enabled {
		return false
	}
	return activePeriodIncludes(cfg.ActiveFrom, cfg.ActiveUntil, now)
}

func activePeriodIncludes(fromRaw, untilRaw string, now time.Time) bool {
	now = now.UTC()
	if from, ok := parseActionBoundary(fromRaw, false); ok && now.Before(from) {
		return false
	}
	if until, ok := parseActionBoundary(untilRaw, true); ok && now.After(until) {
		return false
	}
	return true
}

func parseActionBoundary(raw string, endOfDay bool) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.UTC(), true
	}
	if parsed, err := time.Parse("2006-01-02", raw); err == nil {
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		if endOfDay {
			parsed = parsed.Add(24*time.Hour - time.Nanosecond)
		}
		return parsed, true
	}
	return time.Time{}, false
}

func flattenNodes(def []byte) []actionPlan {
	var d workflowDefinition
	_ = json.Unmarshal(def, &d)
	out := make([]actionPlan, 0, len(d.Nodes))
	for i := range d.Nodes {
		node := &d.Nodes[i]
		node.ActionKey = strings.TrimSpace(node.ActionKey)
		node.ActionTypeKey = strings.TrimSpace(node.ActionTypeKey)
		if node.ActionKey == "" {
			node.ActionKey = node.ActionTypeKey
		}
		if node.ActionTypeKey == "" {
			node.ActionTypeKey = "unknown"
		}
		if node.ActionKey != "" {
			out = append(out, *node)
		}
	}
	return out
}
