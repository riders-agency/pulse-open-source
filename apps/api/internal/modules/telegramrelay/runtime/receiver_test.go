package runtime

import "testing"

func TestTelegramUpdateTypeRecognizesAllConfiguredUpdates(t *testing.T) {
	for _, updateType := range []string{
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
	} {
		got, _ := telegramUpdateType(map[string]any{updateType: map[string]any{}})
		if got != updateType {
			t.Fatalf("telegramUpdateType(%s) = %s", updateType, got)
		}
	}
}

func TestNormalizeTelegramUpdateExtractsCaptionCommand(t *testing.T) {
	normalized := normalizeTelegramUpdate("support_bot", map[string]any{
		"update_id": float64(10),
		"message": map[string]any{
			"message_id": float64(20),
			"caption":    "/photo order_1001",
			"caption_entities": []any{
				map[string]any{"type": "bot_command", "offset": float64(0), "length": float64(6)},
			},
			"chat": map[string]any{"id": float64(123), "type": "private"},
			"from": map[string]any{"id": float64(456)},
		},
	})
	if normalized.Command != "photo" {
		t.Fatalf("command = %q, want photo", normalized.Command)
	}
	if normalized.Text != "/photo order_1001" {
		t.Fatalf("text = %q", normalized.Text)
	}
}

func TestNormalizeTelegramUpdateExtractsCallbackCommand(t *testing.T) {
	normalized := normalizeTelegramUpdate("support_bot", map[string]any{
		"update_id": float64(10),
		"callback_query": map[string]any{
			"id":   "cb-1",
			"data": "/confirm order_1001",
			"from": map[string]any{"id": float64(456)},
			"message": map[string]any{
				"message_id": float64(20),
				"chat":       map[string]any{"id": float64(123), "type": "private"},
			},
		},
	})
	if normalized.Command != "confirm" {
		t.Fatalf("command = %q, want confirm", normalized.Command)
	}
	if normalized.ChatID != "123" || normalized.UserID != "456" {
		t.Fatalf("chat/user = %q/%q", normalized.ChatID, normalized.UserID)
	}
}
