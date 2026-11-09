package repository

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/errs"
)

func mapFromAny(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if out, ok := value.(map[string]any); ok {
		return out
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func clean(value string) string {
	return strings.TrimSpace(value)
}

func uuidOrNil(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return parsed
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nonEmptyJSON(primary, fallback []byte) []byte {
	trimmed := strings.TrimSpace(string(primary))
	if trimmed != "" && trimmed != "{}" && trimmed != "null" {
		return primary
	}
	return fallback
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func jsonOrNil(value map[string]any) any {
	if value == nil {
		return nil
	}
	raw, _ := json.Marshal(value)
	return raw
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, pgx.ErrNoRows)
}

func IsForbidden(err error) bool {
	return errors.Is(err, errs.ErrForbidden)
}
