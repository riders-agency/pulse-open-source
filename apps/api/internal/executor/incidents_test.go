package executor

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIncidentFromExecutionInvalidProviderToken(t *testing.T) {
	final := actionExecution{
		OutputJSON: `{
			"diagnostics": {
				"summary": "Telegram did not accept the message",
				"provider": "telegram",
				"reason_code": "telegram_invalid_token",
				"human_reason": "HTTP 401 Unauthorized",
				"next_action": "Rotate the bot token",
				"retryable": false
			}
		}`,
		ProviderResponseJSON: `{"provider":"telegram","status_code":401,"provider_response_redacted":true}`,
		ErrorType:            "provider_http",
	}

	incident, ok := incidentFromExecution(final)
	if !ok {
		t.Fatal("expected invalid token to create incident")
	}
	if incident.Message != "Telegram did not accept the message" {
		t.Fatalf("unexpected message %q", incident.Message)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(incident.PayloadJSON), &payload); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if payload["reason_code"] != "telegram_invalid_token" {
		t.Fatalf("unexpected reason_code %#v", payload["reason_code"])
	}
	if _, ok := payload["provider_response"].(map[string]any); !ok {
		t.Fatalf("provider_response missing in payload: %#v", payload)
	}
}

func TestIncidentFromExecutionIgnoresNonIncidentDiagnostics(t *testing.T) {
	final := actionExecution{
		OutputJSON:           `{"diagnostics":{"reason_code":"telegram_chat_not_found"}}`,
		ProviderResponseJSON: `{}`,
	}
	incident, ok := incidentFromExecution(final)
	if ok || strings.TrimSpace(incident.PayloadJSON) != "" {
		t.Fatalf("expected no incident, got %#v", incident)
	}
}
