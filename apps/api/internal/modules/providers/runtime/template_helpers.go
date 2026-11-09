package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	return strings.TrimRight(strings.TrimSpace(firstNonEmpty(os.Getenv("RIDERS_PULSE_PUBLIC_URL"), os.Getenv("NEXT_PUBLIC_APP_URL"), os.Getenv("PUBLIC_APP_URL"), "https://pulse.riders.agency")), "/")
}

func (s templateSystem) Map() map[string]any {
	return map[string]any{"id": s.RunID, "link": s.RunURL, "run_id": s.RunID, "run_url": s.RunURL}
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

func stringMetadata(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	if value, ok := metadata[key].(string); ok {
		return value
	}
	return ""
}

func maxBoolMetadata(metadata map[string]any, key string) (bool, bool) {
	if metadata == nil {
		return false, false
	}
	value, exists := metadata[key]
	if !exists || value == nil {
		return false, false
	}
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return false, false
	}
}

func durationOrDefault(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func shouldSkipTelegramBotReplyToChannel(runInput []byte) bool {
	var root map[string]any
	if len(runInput) == 0 || json.Unmarshal(runInput, &root) != nil {
		return false
	}
	if _, ok := root["telegram_relay"]; !ok {
		return false
	}
	chatType := stringAtRunPath(root, "telegram_relay.chat.type")
	updateType := stringAtRunPath(root, "telegram_relay.update_type")
	return chatType == "channel" || updateType == "channel_post" || updateType == "edited_channel_post"
}

func splitTelegramHTML(text string, maxRunes int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxRunes {
		if text == "" {
			return nil
		}
		return []string{text}
	}
	blocks := strings.Split(text, "\n\n")
	if len(blocks) <= 1 {
		return splitTelegramHTMLHard(runes, maxRunes)
	}
	chunks := make([]string, 0, len(blocks))
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(current.String()))
		current.Reset()
	}
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		extra := block
		if current.Len() > 0 {
			extra = "\n\n" + block
		}
		if len([]rune(current.String()+extra)) > maxRunes {
			flush()
			if len([]rune(block)) > maxRunes {
				chunks = append(chunks, splitTelegramHTMLHard([]rune(block), maxRunes)...)
				continue
			}
			current.WriteString(block)
			continue
		}
		current.WriteString(extra)
	}
	flush()
	if len(chunks) == 0 {
		return splitTelegramHTMLHard(runes, maxRunes)
	}
	return chunks
}

func splitTelegramHTMLHard(runes []rune, maxRunes int) []string {
	if maxRunes <= 0 {
		maxRunes = 3900
	}
	out := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
	}
	return out
}
