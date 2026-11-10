package repository

import (
	"encoding/json"
	"testing"
)

func TestCobrowseKind(t *testing.T) {
	cases := map[string]string{
		"monitoring.snapshot":    "snapshot",
		"monitoring.incremental": "mutation",
		"cursor.move":            "cursor",
		"touch.move":             "touch",
		"scroll":                 "scroll",
		"viewport":               "viewport",
		"page.navigate":          "navigation",
		"click":                  "click",
		"diagnostic":             "diagnostic",
		"chat.message":           "",
	}

	for input, want := range cases {
		if got := cobrowseKind(input); got != want {
			t.Fatalf("cobrowseKind(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSequenceFromPayload(t *testing.T) {
	if got := sequenceFromPayload(map[string]any{"sequence": float64(42)}); got != 42 {
		t.Fatalf("float sequence = %d, want 42", got)
	}
	if got := sequenceFromPayload(map[string]any{"sequence": json.Number("77")}); got != 77 {
		t.Fatalf("json.Number sequence = %d, want 77", got)
	}
	if got := sequenceFromPayload(nil); got <= 0 {
		t.Fatalf("fallback sequence should be positive, got %d", got)
	}
}

func TestApplyEventsMarksVisitorLeftOffline(t *testing.T) {
	snapshot := map[string]any{"activity": map[string]any{"connection_status": "online", "is_active": true}}
	applyEvents(snapshot, []EventInput{{EventType: "visitor.left", Payload: map[string]any{"reason": "pagehide"}}})

	activity, ok := snapshot["activity"].(map[string]any)
	if !ok {
		t.Fatal("activity snapshot missing")
	}
	if activity["connection_status"] != "offline" {
		t.Fatalf("connection_status = %v, want offline", activity["connection_status"])
	}
	if activity["is_active"] != false {
		t.Fatalf("is_active = %v, want false", activity["is_active"])
	}
	if activity["last_seen_at"] == "" {
		t.Fatal("last_seen_at should be set")
	}
}

func TestApplyEventsPersistsMirrorSnapshotAndScrollContainers(t *testing.T) {
	snapshot := map[string]any{}
	applyEvents(snapshot, []EventInput{{
		EventType: "monitoring.snapshot",
		Payload: map[string]any{
			"mirror": map[string]any{
				"html":              "<html><body>modal</body></html>",
				"scroll_containers": []any{map[string]any{"path": "html.1.0", "top": float64(120)}},
			},
		},
	}, {
		EventType: "scroll",
		Payload: map[string]any{
			"scroll_containers": []any{map[string]any{"path": "html.1.0", "top": float64(240)}},
		},
	}})

	mirror, ok := snapshot["mirror"].(map[string]any)
	if !ok {
		t.Fatal("mirror snapshot missing")
	}
	if mirror["html"] != "<html><body>modal</body></html>" {
		t.Fatalf("mirror html = %v", mirror["html"])
	}
	containers, ok := mirror["scroll_containers"].([]any)
	if !ok || len(containers) != 1 {
		t.Fatalf("scroll_containers = %#v", mirror["scroll_containers"])
	}
	first, _ := containers[0].(map[string]any)
	if first["top"] != float64(240) {
		t.Fatalf("scroll container top = %v, want 240", first["top"])
	}
}

func TestApplyEventsIgnoresStaleSequence(t *testing.T) {
	snapshot := map[string]any{
		"activity": map[string]any{"last_sequence": int64(20)},
		"scroll":   map[string]any{"x": float64(0), "y": float64(900)},
	}

	applyEvents(snapshot, []EventInput{{
		EventType: "scroll",
		Payload: map[string]any{
			"sequence": float64(18),
			"scroll_x": float64(0),
			"scroll_y": float64(120),
		},
	}, {
		EventType: "scroll",
		Payload: map[string]any{
			"sequence": float64(21),
			"scroll_x": float64(0),
			"scroll_y": float64(960),
		},
	}})

	scroll, ok := snapshot["scroll"].(map[string]any)
	if !ok {
		t.Fatal("scroll snapshot missing")
	}
	if scroll["y"] != float64(960) {
		t.Fatalf("scroll y = %v, want latest sequence y=960", scroll["y"])
	}
	activity, _ := snapshot["activity"].(map[string]any)
	if activity["last_sequence"] != int64(21) {
		t.Fatalf("last_sequence = %v, want 21", activity["last_sequence"])
	}
}

func TestMergePresenceSnapshotPreservesLiveWebvisorState(t *testing.T) {
	existing := map[string]any{
		"current_url": "https://riders.agency/?old=1",
		"mirror":      map[string]any{"html": "<html><body>open modal</body></html>"},
		"cursor":      map[string]any{"x": float64(100), "y": float64(200)},
		"last_click":  map[string]any{"x": float64(90), "y": float64(180)},
		"activity":    map[string]any{"last_sequence": int64(42), "connection_status": "online"},
		"signals":     map[string]any{"monitoring_active": true, "has_dead_click": true},
	}
	fresh := map[string]any{
		"current_url": "https://riders.agency/?rp_e2e=1",
		"viewport":    map[string]any{"width": 1280, "height": 720},
		"scroll":      map[string]any{"x": 0, "y": 600},
		"activity":    map[string]any{"last_seen_at": "now", "is_active": true, "connection_status": "online"},
		"signals":     map[string]any{"monitoring_active": true, "chat_open": false, "operator_invited": false},
	}

	merged := mergePresenceSnapshot(existing, fresh)

	if merged["current_url"] != "https://riders.agency/?rp_e2e=1" {
		t.Fatalf("current_url = %v", merged["current_url"])
	}
	if _, ok := merged["mirror"].(map[string]any); !ok {
		t.Fatal("mirror should be preserved")
	}
	if _, ok := merged["cursor"].(map[string]any); !ok {
		t.Fatal("cursor should be preserved")
	}
	if _, ok := merged["last_click"].(map[string]any); !ok {
		t.Fatal("last_click should be preserved")
	}
	signals, _ := merged["signals"].(map[string]any)
	if signals["has_dead_click"] != true {
		t.Fatal("existing diagnostic signal should be preserved")
	}
}

func TestAnalyticsPayloadKeepsOnlySafePresenceMetadata(t *testing.T) {
	payload := analyticsPayload(PresenceInput{
		Metadata: map[string]any{
			"page_session_id":    "page_123",
			"widget_instance_id": "widget_456",
			"conversation_id":    "secret-ish-id",
			"token":              "should-not-be-stored",
		},
	})

	metadata, ok := payload["metadata"].(map[string]any)
	if !ok {
		t.Fatal("metadata missing")
	}
	if metadata["page_session_id"] != "page_123" {
		t.Fatalf("page_session_id = %v", metadata["page_session_id"])
	}
	if metadata["widget_instance_id"] != "widget_456" {
		t.Fatalf("widget_instance_id = %v", metadata["widget_instance_id"])
	}
	if _, ok := metadata["conversation_id"]; ok {
		t.Fatal("conversation_id should not be copied into analytics metadata")
	}
	if _, ok := metadata["token"]; ok {
		t.Fatal("token should not be copied into analytics metadata")
	}
}
