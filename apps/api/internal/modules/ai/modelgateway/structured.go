package modelgateway

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseStructuredResponse(content string, responseFormat map[string]any) (map[string]any, error) {
	if len(responseFormat) == 0 {
		return map[string]any{}, nil
	}
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, schemaError(fmt.Errorf("empty structured response"))
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
		return nil, schemaError(err)
	}
	return redactMap(out), nil
}

func redactMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		if isSensitiveKey(key) {
			out[key] = "[redacted]"
			continue
		}
		out[key] = redactValue(value)
	}
	return out
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactMap(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, redactValue(item))
		}
		return out
	default:
		return typed
	}
}

func isSensitiveKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization", "cookie", "password", "provider_payload", "raw", "secret", "token":
		return true
	default:
		return false
	}
}
