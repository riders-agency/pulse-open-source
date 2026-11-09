package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var templateExpression = regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)

type templateSystem struct {
	RunID  string
	RunURL string
}

func templateSystemForRun(runID string) templateSystem {
	runID = strings.TrimSpace(runID)
	base := publicAppBaseURL()
	runURL := ""
	if runID != "" {
		runURL = base + "/runs/" + runID
	}
	return templateSystem{RunID: runID, RunURL: runURL}
}

func publicAppBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(firstNonEmpty(
		os.Getenv("RIDERS_PULSE_PUBLIC_URL"),
		os.Getenv("NEXT_PUBLIC_APP_URL"),
		os.Getenv("PUBLIC_APP_URL"),
		"https://pulse.riders.agency",
	)), "/")
}

func (s templateSystem) Map() map[string]any {
	return map[string]any{
		"id":      s.RunID,
		"link":    s.RunURL,
		"run_id":  s.RunID,
		"run_url": s.RunURL,
	}
}

// runInputTemplateRoot returns the object used for {{ input.* }} expansion.
// Live webhooks store the body under {"webhook": payload}; replays use {"payload": ...}.
func runInputTemplateRoot(root any) any {
	m, ok := root.(map[string]any)
	if !ok || m == nil {
		return root
	}
	if w, ok := m["webhook"]; ok && w != nil {
		return w
	}
	if relay, ok := m["telegram_relay"]; ok && relay != nil {
		return relay
	}
	if p, ok := m["payload"]; ok && p != nil {
		return p
	}
	return root
}

func mergeActionExpose(steps map[string]any, actionKey, outputJSON string) {
	if steps == nil || strings.TrimSpace(actionKey) == "" || strings.TrimSpace(outputJSON) == "" {
		return
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(outputJSON), &output); err != nil {
		return
	}
	expose, ok := output["expose"].(map[string]any)
	if !ok || len(expose) == 0 {
		return
	}
	steps[actionKey] = expose
}

func renderTemplate(template string, input []byte, steps map[string]any, system templateSystem) string {
	if template == "" {
		return ""
	}
	var root any
	if len(input) > 0 {
		_ = json.Unmarshal(input, &root)
	}
	tplRoot := runInputTemplateRoot(root)
	return templateExpression.ReplaceAllStringFunc(template, func(match string) string {
		parts := templateExpression.FindStringSubmatch(match)
		if len(parts) != 2 {
			return ""
		}
		path := strings.TrimSpace(parts[1])
		if strings.HasPrefix(path, "json ") {
			rawPath := strings.TrimSpace(strings.TrimPrefix(path, "json "))
			value, ok := templateRawValue(root, tplRoot, steps, system, rawPath)
			if !ok {
				return "null"
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				return "null"
			}
			return string(encoded)
		}
		if strings.HasPrefix(path, "input.") {
			return fmt.Sprint(valueAtPath(tplRoot, strings.TrimPrefix(path, "input.")))
		}
		if strings.HasPrefix(path, "steps.") {
			return fmt.Sprint(valueAtPath(steps, strings.TrimPrefix(path, "steps.")))
		}
		if strings.HasPrefix(path, "system.") {
			return fmt.Sprint(valueAtPath(system.Map(), strings.TrimPrefix(path, "system.")))
		}
		if strings.HasPrefix(path, "run.") {
			return fmt.Sprint(valueAtPath(map[string]any{"id": system.RunID, "url": system.RunURL}, strings.TrimPrefix(path, "run.")))
		}
		if value, ok := webhookTemplateValue(root, tplRoot, path); ok {
			return fmt.Sprint(value)
		}
		return ""
	})
}

func templateRawValue(root any, inputRoot any, steps map[string]any, system templateSystem, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false
	}
	if strings.HasPrefix(path, "input.") {
		return valueAtPathRaw(inputRoot, strings.TrimPrefix(path, "input."))
	}
	if strings.HasPrefix(path, "steps.") {
		return valueAtPathRaw(steps, strings.TrimPrefix(path, "steps."))
	}
	if strings.HasPrefix(path, "system.") {
		return valueAtPathRaw(system.Map(), strings.TrimPrefix(path, "system."))
	}
	if strings.HasPrefix(path, "run.") {
		return valueAtPathRaw(map[string]any{"id": system.RunID, "url": system.RunURL}, strings.TrimPrefix(path, "run."))
	}
	return webhookRawValue(root, inputRoot, path)
}

func webhookRawValue(root any, inputRoot any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false
	}
	head, tail := splitTemplatePath(path)
	if head == "webhook" || head == "payload" {
		if tail == "" {
			return inputRoot, true
		}
		return valueAtPathRaw(inputRoot, tail)
	}
	slug := webhookSlugFromInput(root)
	if slug != "" && head == slug {
		if tail == "" {
			return inputRoot, true
		}
		return valueAtPathRaw(inputRoot, tail)
	}
	return valueAtPathRaw(root, path)
}

func webhookTemplateValue(root any, inputRoot any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false
	}
	head, tail := splitTemplatePath(path)
	if head == "webhook" || head == "payload" {
		if tail == "" {
			return templateValue(inputRoot), true
		}
		value, ok := valueAtPathRaw(inputRoot, tail)
		return templateValue(value), ok
	}
	slug := webhookSlugFromInput(root)
	if slug != "" && head == slug {
		if tail == "" {
			return templateValue(inputRoot), true
		}
		value, ok := valueAtPathRaw(inputRoot, tail)
		return templateValue(value), ok
	}
	value, ok := valueAtPathRaw(root, path)
	return templateValue(value), ok
}

func splitTemplatePath(path string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(path), ".", 2)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func webhookSlugFromInput(root any) string {
	m, ok := root.(map[string]any)
	if !ok || m == nil {
		return ""
	}
	if pulse, ok := m["_pulse"].(map[string]any); ok {
		for _, key := range []string{"webhook_slug", "webhook_endpoint_key", "endpoint_key", "telegram_relay_key"} {
			if value, ok := pulse[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	for _, key := range []string{"webhook_slug", "webhook_endpoint_key", "endpoint_key"} {
		if value, ok := m[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func valueAtPath(root any, path string) any {
	value, ok := valueAtPathRaw(root, path)
	if !ok {
		return ""
	}
	return templateValue(value)
}

func stringAtRunPath(root any, path string) string {
	value, ok := valueAtPathRaw(root, path)
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(templateValue(value)))
}

func paymentActionIDFromRunInput(runInput []byte) string {
	if len(bytes.TrimSpace(runInput)) == 0 {
		return ""
	}
	var root any
	if err := json.Unmarshal(runInput, &root); err != nil {
		return ""
	}
	for _, path := range []string{"payment_action_id", "_pulse.payment_action_id"} {
		if value := stringAtRunPath(root, path); value != "" {
			return value
		}
	}
	return ""
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func clampAttempts(value int) int {
	if value < 1 {
		return 1
	}
	if value > 10 {
		return 10
	}
	return value
}

func nullableUUID(value string) *string {
	if value == "" {
		return nil
	}
	if _, err := uuid.Parse(value); err != nil {
		return nil
	}
	return &value
}
