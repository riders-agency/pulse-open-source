package http

import (
	"testing"

	sitechat "riders.agency/pulse/api/internal/modules/sitechat/repository"
)

func TestLowPriorityEnvelopeTypes(t *testing.T) {
	for _, kind := range []string{"cursor.move", "touch.move", "scroll", "viewport", "monitoring.incremental", "event.batch"} {
		if !isLowPriority(kind) {
			t.Fatalf("%s should be low priority", kind)
		}
	}
	for _, kind := range []string{"chat.message", "operator.invite", "call.offer", "call.recording", "call.audio_chunk", "webrtc.offer"} {
		if isLowPriority(kind) {
			t.Fatalf("%s should be critical priority", kind)
		}
	}
}

func TestWithEnvelopeScopeFillsRoutingFields(t *testing.T) {
	scope := sitechat.LiveAssistScope{
		ProjectID:      "project-1",
		ChatWidgetID:   "widget-1",
		VisitorID:      "visitor-1",
		VisitSessionID: "session-1",
		ConversationID: "conversation-1",
	}

	out := withEnvelopeScope(WSEnvelope{Type: "cursor.move"}, scope)
	if out.ID == "" || out.SentAt.IsZero() {
		t.Fatal("expected generated envelope id and sent_at")
	}
	if out.ProjectID != scope.ProjectID || out.ConversationID != scope.ConversationID {
		t.Fatalf("scope was not applied: %+v", out)
	}
}

func TestVisitorLeftEnvelopeCarriesOfflineState(t *testing.T) {
	scope := sitechat.LiveAssistScope{
		ProjectID:      "project-1",
		ChatWidgetID:   "widget-1",
		VisitorID:      "visitor-1",
		VisitSessionID: "session-1",
		ConversationID: "conversation-1",
	}

	out := visitorLeftEnvelope(scope, "pagehide")
	if out.Type != "visitor.left" || out.VisitSessionID != scope.VisitSessionID || out.ConversationID != scope.ConversationID {
		t.Fatalf("unexpected envelope routing: %+v", out)
	}
	if out.Payload["connection_status"] != "offline" || out.Payload["reason"] != "pagehide" {
		t.Fatalf("unexpected payload: %+v", out.Payload)
	}
	if out.Payload["last_seen_at"] == "" {
		t.Fatal("last_seen_at should be set")
	}
}
