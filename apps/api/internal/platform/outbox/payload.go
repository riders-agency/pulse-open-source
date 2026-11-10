package outbox

import (
	"encoding/json"
	"fmt"
	"strings"
)

const EventVersionField = "event_version"

func normalizePayload(typeKey string, payload json.RawMessage, schemaVersion string) (json.RawMessage, error) {
	version := strings.TrimSpace(schemaVersion)
	if version == "" {
		version = strings.TrimSpace(typeKey) + ".v1"
	}
	if version == ".v1" {
		return nil, fmt.Errorf("outbox schema version is required")
	}
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	obj := map[string]json.RawMessage{}
	if err := json.Unmarshal(payload, &obj); err != nil {
		return nil, fmt.Errorf("outbox payload must be a JSON object: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("outbox payload must be a JSON object")
	}
	if rawVersion, ok := obj[EventVersionField]; ok {
		var existing string
		if err := json.Unmarshal(rawVersion, &existing); err != nil {
			return nil, fmt.Errorf("outbox payload %s must be a string", EventVersionField)
		}
		existing = strings.TrimSpace(existing)
		if existing != "" && existing != version {
			return nil, fmt.Errorf("outbox payload version %q does not match %q", existing, version)
		}
	}
	encodedVersion, _ := json.Marshal(version)
	obj[EventVersionField] = encodedVersion
	normalized, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(normalized), nil
}
