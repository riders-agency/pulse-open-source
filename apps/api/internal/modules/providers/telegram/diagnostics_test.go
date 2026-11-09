package telegram

import (
	"testing"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func TestAnalyzeClassifiesAuthAndRateLimit(t *testing.T) {
	auth := Analyze(map[string]any{
		"provider":    "telegram",
		"status_code": 401,
		"ok":          false,
		"description": "Unauthorized",
	}, false, providerdiagnostics.BaseFailure(map[string]any{"provider": "telegram"}, false, nil))
	if auth.ReasonCode != "telegram_invalid_token" || auth.Retryable {
		t.Fatalf("expected invalid token non-retryable, got %#v", auth)
	}

	limited := Analyze(map[string]any{
		"provider":    "telegram",
		"status_code": 429,
		"ok":          false,
		"parameters":  map[string]any{"retry_after": 12},
	}, true, providerdiagnostics.BaseFailure(map[string]any{"provider": "telegram"}, true, nil))
	if limited.ReasonCode != "telegram_rate_limited" || !limited.Retryable {
		t.Fatalf("expected rate limit retryable, got %#v", limited)
	}
}

func TestEvaluateTelegramCriteria(t *testing.T) {
	results, err := Evaluate(map[string]any{
		"provider":          "telegram",
		"status_code":       200,
		"ok":                true,
		"message_id":        float64(123),
		"raw_response_json": map[string]any{"ok": true},
	})
	if err != nil {
		t.Fatalf("expected criteria success, got %v", err)
	}
	if len(results) != len(Criteria()) {
		t.Fatalf("expected %d results, got %d", len(Criteria()), len(results))
	}

	results, err = Evaluate(map[string]any{
		"provider":          "telegram",
		"status_code":       200,
		"ok":                false,
		"raw_response_json": map[string]any{"ok": false},
	})
	if err == nil {
		t.Fatalf("expected criteria failure")
	}
	if results[2].Passed {
		t.Fatalf("expected ok criterion to fail, got %#v", results[2])
	}
}
