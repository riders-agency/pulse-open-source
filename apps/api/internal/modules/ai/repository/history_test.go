package repository

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSanitizeLangGraphHistoryBoundsAndRedacts(t *testing.T) {
	rows := []langGraphHistoryRow{
		historyRow("m-1", "user", "мой token=abc123 и Authorization: Bearer secret-token", map[string]any{
			"kind":            "user_message",
			"source_surface":  "widget",
			"page_context":    map[string]any{"route": "/secret"},
			"provider_secret": "must not leak",
		}),
		historyRow("m-2", "assistant", "готово", map[string]any{
			"kind":          "tool_result",
			"tool_key":      "task.search",
			"output_json":   map[string]any{"raw": "must not leak"},
			"runtime":       "langgraph",
			"risk_level":    "read",
			"debug_payload": "must not leak",
		}),
	}

	out := sanitizeLangGraphHistory(rows, 3)

	if len(out) != 3 {
		t.Fatalf("history len = %d, want summary + 2 messages", len(out))
	}
	if out[0].Role != "system" || out[0].Payload["kind"] != "conversation_summary" {
		t.Fatalf("summary missing: %#v", out[0])
	}
	if !strings.Contains(out[1].Body, "token=[redacted]") ||
		strings.Contains(out[1].Body, "abc123") ||
		strings.Contains(out[1].Body, "secret-token") {
		t.Fatalf("secret text was not redacted: %q", out[1].Body)
	}
	if _, ok := out[1].Payload["page_context"]; ok {
		t.Fatalf("page_context leaked into history payload: %#v", out[1].Payload)
	}
	if _, ok := out[2].Payload["output_json"]; ok {
		t.Fatalf("tool output leaked into history payload: %#v", out[2].Payload)
	}
	if out[1].CreatedAt.Location() != time.UTC {
		t.Fatalf("created_at must be UTC: %s", out[1].CreatedAt.Location())
	}
}

func TestSanitizeLangGraphHistoryDropsOldestWhenCharBudgetExceeded(t *testing.T) {
	long := strings.Repeat("x", langGraphHistoryBodyChars)
	rows := make([]langGraphHistoryRow, 0, langGraphHistoryMaxMessages)
	for i := 0; i < langGraphHistoryMaxMessages; i++ {
		rows = append(rows, historyRow("m", "user", long, map[string]any{"kind": "user_message"}))
	}

	out := sanitizeLangGraphHistory(rows, 0)

	if out[0].Payload["kind"] != "conversation_summary" {
		t.Fatalf("expected synthetic summary after char-budget truncation: %#v", out[0])
	}
	if len(out) >= len(rows)+1 {
		t.Fatalf("history was not bounded: got %d rows for %d source rows", len(out), len(rows))
	}
}

func historyRow(id string, role string, body string, payload map[string]any) langGraphHistoryRow {
	raw, _ := json.Marshal(payload)
	return langGraphHistoryRow{
		MessageID: id,
		Role:      role,
		Body:      body,
		Payload:   raw,
		CreatedAt: time.Date(2026, 6, 25, 8, 0, 0, 0, time.FixedZone("MSK", 3*60*60)),
	}
}
