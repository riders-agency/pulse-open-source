package engine

import (
	"bytes"
	"encoding/json"
)

func RawJSONOrDefault(raw json.RawMessage, fallback string) string {
	if len(bytes.TrimSpace(raw)) == 0 {
		return fallback
	}
	return string(raw)
}
