package repository

import (
	"encoding/json"
	"testing"
)

func TestTelegramRelayMetadataDefaultsToAllUpdates(t *testing.T) {
	meta := json.RawMessage(`{"relay_key":"support_bot","webhook_secret":"secret_token"}`)
	_, _, _, allowed, _, _, _, polling, err := telegramRelayMetadataForStorage(meta, true)
	if err != nil {
		t.Fatalf("telegramRelayMetadataForStorage: %v", err)
	}
	if len(allowed) != len(allTelegramUpdateTypes()) {
		t.Fatalf("allowed update count = %d, want %d", len(allowed), len(allTelegramUpdateTypes()))
	}
	if allowed[0] != "message" || allowed[len(allowed)-1] != "managed_bot" {
		t.Fatalf("unexpected allowed updates: %#v", allowed)
	}
	if polling.Enabled || polling.TimeoutSeconds != 50 || polling.Limit != 100 || !polling.DeleteWebhook || polling.DropPendingUpdates {
		t.Fatalf("unexpected polling defaults: %#v", polling)
	}
}

func TestTelegramRelayMetadataParsesPolling(t *testing.T) {
	meta := json.RawMessage(`{
		"relay_key":"support_bot",
		"webhook_secret":"secret_token",
		"polling_enabled":true,
		"polling_timeout_seconds":999,
		"polling_limit":0,
		"polling_delete_webhook":false,
		"polling_drop_pending_updates":true
	}`)
	_, _, _, _, _, _, _, polling, err := telegramRelayMetadataForStorage(meta, true)
	if err != nil {
		t.Fatalf("telegramRelayMetadataForStorage: %v", err)
	}
	if !polling.Enabled {
		t.Fatalf("polling should be enabled")
	}
	if polling.TimeoutSeconds != 60 {
		t.Fatalf("timeout = %d, want clamped 60", polling.TimeoutSeconds)
	}
	if polling.Limit != 100 {
		t.Fatalf("limit = %d, want fallback 100", polling.Limit)
	}
	if polling.DeleteWebhook {
		t.Fatalf("delete webhook should respect explicit false")
	}
	if !polling.DropPendingUpdates {
		t.Fatalf("drop pending should be true")
	}
}

func TestGenerateTelegramSecretTokenIsTelegramCompatible(t *testing.T) {
	token, err := generateTelegramSecretToken(48)
	if err != nil {
		t.Fatalf("generateTelegramSecretToken: %v", err)
	}
	if len(token) != 48 {
		t.Fatalf("token length = %d", len(token))
	}
	if !validateTelegramSecretToken(token) {
		t.Fatalf("token is not Telegram-compatible: %q", token)
	}
}
