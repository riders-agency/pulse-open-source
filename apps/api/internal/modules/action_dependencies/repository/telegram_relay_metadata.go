package repository

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrTelegramRelayMetadata = errors.New("telegram relay metadata invalid")

type telegramRelayMetadata struct {
	RelayKey       string          `json:"relay_key"`
	WebhookSecret  string          `json:"webhook_secret"`
	RelaySecret    string          `json:"relay_secret"`
	SecretToken    string          `json:"secret_token"`
	RelayURL       string          `json:"relay_url"`
	AllowedUpdates []string        `json:"allowed_updates"`
	DefaultChatID  string          `json:"default_chat_id"`
	CountryRelays  json.RawMessage `json:"country_relays"`
}

type telegramRelayPolling struct {
	Enabled            bool
	TimeoutSeconds     int
	Limit              int
	DeleteWebhook      bool
	DropPendingUpdates bool
}

func allTelegramUpdateTypes() []string {
	return []string{
		"message",
		"edited_message",
		"channel_post",
		"edited_channel_post",
		"business_connection",
		"business_message",
		"edited_business_message",
		"deleted_business_messages",
		"guest_message",
		"message_reaction",
		"message_reaction_count",
		"inline_query",
		"chosen_inline_result",
		"callback_query",
		"shipping_query",
		"pre_checkout_query",
		"purchased_paid_media",
		"poll",
		"poll_answer",
		"my_chat_member",
		"chat_member",
		"chat_join_request",
		"chat_boost",
		"removed_chat_boost",
		"managed_bot",
	}
}

func sanitizeRelayKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		case r == ' ' || r == '.':
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), "_-")
}

func validateTelegramSecretToken(value string) bool {
	if len(value) < 1 || len(value) > 256 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func generateTelegramSecretToken(length int) (string, error) {
	if length <= 0 {
		length = 48
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}

func relayCountryRelays(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return json.RawMessage("[]")
	}
	var relays []map[string]any
	if err := json.Unmarshal(raw, &relays); err != nil {
		return json.RawMessage("[]")
	}
	cleaned := make([]map[string]any, 0, len(relays))
	for _, relay := range relays {
		if relay == nil {
			continue
		}
		cleaned = append(cleaned, relay)
	}
	if len(cleaned) == 0 {
		return json.RawMessage("[]")
	}
	out, _ := json.Marshal(cleaned)
	return out
}

func visibleCountryRelays(raw json.RawMessage) []map[string]any {
	var relays []map[string]any
	_ = json.Unmarshal(relayCountryRelays(raw), &relays)
	out := make([]map[string]any, 0, len(relays))
	for _, relay := range relays {
		item := map[string]any{}
		for _, key := range []string{"name", "region", "url", "enabled"} {
			if value, ok := relay[key]; ok {
				item[key] = value
			}
		}
		if _, ok := relay["authorization_header"]; ok {
			item["authorization_header"] = "***"
		}
		if _, ok := relay["secret_token"]; ok {
			item["secret_token"] = "***"
		}
		out = append(out, item)
	}
	return out
}

func telegramRelayMetadataForStorage(meta json.RawMessage, requireSecret bool) (string, string, string, []string, string, json.RawMessage, string, telegramRelayPolling, error) {
	if len(meta) == 0 || strings.TrimSpace(string(meta)) == "" {
		return "", "", "", nil, "", nil, "", telegramRelayPolling{}, fmt.Errorf("%w: задайте relay_key и webhook_secret", ErrTelegramRelayMetadata)
	}
	var raw map[string]any
	if err := json.Unmarshal(meta, &raw); err != nil {
		return "", "", "", nil, "", nil, "", telegramRelayPolling{}, fmt.Errorf("%w: %v", ErrTelegramRelayMetadata, err)
	}
	var parsed telegramRelayMetadata
	_ = json.Unmarshal(meta, &parsed)
	relayKey := sanitizeRelayKey(parsed.RelayKey)
	if len(relayKey) < 3 || len(relayKey) > 80 {
		return "", "", "", nil, "", nil, "", telegramRelayPolling{}, fmt.Errorf("%w: relay_key должен быть 3-80 символов A-Z, a-z, 0-9, _ или -", ErrTelegramRelayMetadata)
	}
	secret := strings.TrimSpace(firstNonEmpty(parsed.WebhookSecret, parsed.SecretToken, parsed.RelaySecret))
	if requireSecret && secret == "" {
		return "", "", "", nil, "", nil, "", telegramRelayPolling{}, fmt.Errorf("%w: webhook_secret обязателен", ErrTelegramRelayMetadata)
	}
	if secret != "" && !validateTelegramSecretToken(secret) {
		return "", "", "", nil, "", nil, "", telegramRelayPolling{}, fmt.Errorf("%w: webhook_secret должен совпадать с Telegram secret_token: 1-256 символов A-Z, a-z, 0-9, _ или -", ErrTelegramRelayMetadata)
	}
	allowed := make([]string, 0, len(parsed.AllowedUpdates))
	for _, item := range parsed.AllowedUpdates {
		item = strings.TrimSpace(item)
		if item != "" {
			allowed = append(allowed, item)
		}
	}
	if len(allowed) == 0 {
		allowed = allTelegramUpdateTypes()
	}
	countryRelays := relayCountryRelays(parsed.CountryRelays)
	publicURL := strings.TrimSpace(parsed.RelayURL)
	if publicURL == "" {
		publicURL = "https://pulse.riders.agency/api/v1/telegram/relay/" + relayKey
	}

	delete(raw, "webhook_secret")
	delete(raw, "relay_secret")
	delete(raw, "secret_token")
	raw["relay_key"] = relayKey
	raw["relay_url"] = publicURL
	raw["allowed_updates"] = allowed
	raw["country_relays"] = visibleCountryRelays(countryRelays)
	if strings.TrimSpace(parsed.DefaultChatID) != "" {
		raw["default_chat_id"] = strings.TrimSpace(parsed.DefaultChatID)
	}
	polling := telegramRelayPolling{
		Enabled:            boolFromRaw(raw, "polling_enabled", false),
		TimeoutSeconds:     intFromRaw(raw, "polling_timeout_seconds", 50, 1, 60),
		Limit:              intFromRaw(raw, "polling_limit", 100, 1, 100),
		DeleteWebhook:      boolFromRaw(raw, "polling_delete_webhook", true),
		DropPendingUpdates: boolFromRaw(raw, "polling_drop_pending_updates", false),
	}
	raw["polling_enabled"] = polling.Enabled
	raw["polling_timeout_seconds"] = polling.TimeoutSeconds
	raw["polling_limit"] = polling.Limit
	raw["polling_delete_webhook"] = polling.DeleteWebhook
	raw["polling_drop_pending_updates"] = polling.DropPendingUpdates
	storage, _ := json.Marshal(raw)
	return string(storage), relayKey, publicURL, allowed, strings.TrimSpace(parsed.DefaultChatID), countryRelays, secret, polling, nil
}

func boolFromRaw(raw map[string]any, key string, fallback bool) bool {
	if raw == nil {
		return fallback
	}
	switch value := raw[key].(type) {
	case bool:
		return value
	case string:
		if strings.EqualFold(strings.TrimSpace(value), "true") {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(value), "false") {
			return false
		}
	}
	return fallback
}

func intFromRaw(raw map[string]any, key string, fallback, min, max int) int {
	if raw == nil {
		return fallback
	}
	out := fallback
	switch value := raw[key].(type) {
	case float64:
		out = int(value)
	case int:
		out = value
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			out = parsed
		}
	}
	if out == 0 {
		return fallback
	}
	if out < min {
		return min
	}
	if out > max {
		return max
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func clampInt(value, fallback, min, max int) int {
	if value == 0 {
		value = fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
