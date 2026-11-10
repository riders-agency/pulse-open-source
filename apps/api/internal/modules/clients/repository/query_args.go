package repository

import (
	"strings"

	"github.com/google/uuid"
)

func nullableUUID(id *uuid.UUID) any {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id.String()
}

func nullStr(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func optStr(value *string) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(*value)
}

func optTrim(value *string) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(*value)
}

func optServiceTypeKey(value *string) any {
	if value == nil {
		return nil
	}
	key := normalizeServiceTypeKey(*value)
	if key == "" {
		return nil
	}
	return key
}

func optServiceTypeStatus(value *string) any {
	if value == nil {
		return nil
	}
	return normalizeServiceTypeStatus(*value)
}

func optDigits(value *string) any {
	if value == nil {
		return nil
	}
	return normalizeDigits(*value)
}

func optDate(value *string) any {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	if normalized, err := parsePaymentDate(v); err == nil {
		return normalized
	}
	return v
}

func optNormClientType(value *string) any {
	if value == nil {
		return nil
	}
	return normalizeClientType(*value)
}

func optNormLegalEntityType(value *string) any {
	if value == nil {
		return nil
	}
	return normalizeLegalEntityType(*value)
}

func normalizeDigits(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
