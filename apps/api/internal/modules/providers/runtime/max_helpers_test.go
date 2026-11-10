package runtime

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitMaxMessageTextKeepsEveryPartWithinProviderLimit(t *testing.T) {
	text := strings.Repeat("Развернутая техническая строка сертификата. ", 220)
	parts := splitMaxMessageText(text)
	if len(parts) < 2 {
		t.Fatalf("parts = %d, want multiple parts", len(parts))
	}
	for index, part := range parts {
		if length := len([]byte(part)); length > maxMessageTextBytes {
			t.Fatalf("part %d bytes = %d, limit = %d", index, length, maxMessageTextBytes)
		}
	}
	if strings.Join(parts, " ") != strings.TrimSpace(text) {
		t.Fatal("split message does not preserve text")
	}
}

func TestSplitMaxMessageTextPrefersParagraphBoundary(t *testing.T) {
	first := strings.Repeat("a", 3500)
	second := strings.Repeat("b", 900)
	parts := splitMaxMessageText(first + "\n" + second)
	if len(parts) != 2 {
		t.Fatalf("parts = %d, want 2", len(parts))
	}
	if parts[0] != first || parts[1] != second {
		t.Fatal("paragraph boundary was not preserved")
	}
}

func TestBuildMaxMessageBodiesDistributesTextAndAttachments(t *testing.T) {
	text := strings.Repeat("data ", 900)
	attachments := []map[string]any{
		{"type": "file", "payload": map[string]any{"token": "one"}},
		{"type": "file", "payload": map[string]any{"token": "two"}},
		{"type": "file", "payload": map[string]any{"token": "three"}},
	}
	bodies := buildMaxMessageBodies(map[string]any{"text": text, "format": "markdown"}, attachments)
	if len(bodies) != 3 {
		t.Fatalf("message bodies = %d, want 3", len(bodies))
	}
	for index, body := range bodies {
		if len([]byte(body["text"].(string))) > maxMessageTextBytes {
			t.Fatalf("body %d exceeds MAX text limit", index)
		}
		attached, ok := body["attachments"].([]map[string]any)
		if !ok || len(attached) != 1 {
			t.Fatalf("body %d attachments = %#v", index, body["attachments"])
		}
	}
	if bodies[2]["text"] != "📎 Вложение 3 из 3" {
		t.Fatalf("third message text = %q", bodies[2]["text"])
	}
}

func TestBuildMaxMessageBodiesWithoutAttachmentsUsesAllTextParts(t *testing.T) {
	text := strings.Repeat("x", maxMessageTextBytes+1)
	bodies := buildMaxMessageBodies(map[string]any{"text": text}, nil)
	if len(bodies) != 2 {
		t.Fatalf("message bodies = %d, want 2", len(bodies))
	}
	if _, exists := bodies[0]["attachments"]; exists {
		t.Fatal("unexpected attachments in text-only message")
	}
}

func TestSplitMaxMessageTextCountsUTF8Bytes(t *testing.T) {
	text := strings.Repeat("сертификат 🎁 ", 400)
	parts := splitMaxMessageText(text)
	if len(parts) < 2 {
		t.Fatalf("parts = %d, want multiple UTF-8-safe parts", len(parts))
	}
	for index, part := range parts {
		if len([]byte(part)) > maxMessageTextBytes {
			t.Fatalf("part %d bytes = %d, limit = %d", index, len([]byte(part)), maxMessageTextBytes)
		}
		if !utf8.ValidString(part) {
			t.Fatalf("part %d is not valid UTF-8", index)
		}
	}
}
