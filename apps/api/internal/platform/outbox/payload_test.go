package outbox

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizePayloadAddsDefaultEventVersion(t *testing.T) {
	payload, err := normalizePayload("workflow.run.requested", json.RawMessage(`{"project_id":"p1"}`), "")
	if err != nil {
		t.Fatalf("normalize payload failed: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if got[EventVersionField] != "workflow.run.requested.v1" {
		t.Fatalf("unexpected event version: %#v", got[EventVersionField])
	}
	if got["project_id"] != "p1" {
		t.Fatalf("payload field was not preserved: %#v", got)
	}
}

func TestNormalizePayloadRejectsNonObjectPayload(t *testing.T) {
	_, err := normalizePayload("demo.event", json.RawMessage(`[]`), "")
	if err == nil || !strings.Contains(err.Error(), "JSON object") {
		t.Fatalf("expected object error, got %v", err)
	}
}

func TestNormalizePayloadRejectsVersionDrift(t *testing.T) {
	_, err := normalizePayload("demo.event", json.RawMessage(`{"event_version":"demo.event.v0"}`), "demo.event.v1")
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected version drift error, got %v", err)
	}
}
