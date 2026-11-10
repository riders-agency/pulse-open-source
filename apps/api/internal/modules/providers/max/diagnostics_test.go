package max

import (
	"testing"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func TestAnalyzeClassifiesAuth(t *testing.T) {
	report := Analyze(map[string]any{
		"provider":    "max",
		"status_code": 401,
		"ok":          false,
		"description": "verify.token failed",
	}, false, providerdiagnostics.BaseFailure(map[string]any{"provider": "max"}, false, nil))
	if report.ReasonCode != "max_auth_failed" || report.Retryable {
		t.Fatalf("expected max auth failed non-retryable, got %#v", report)
	}
}

func TestEvaluateMaxCriteria(t *testing.T) {
	results, err := Evaluate(map[string]any{
		"provider":    "max",
		"status_code": 200,
		"message_id":  "m1",
		"raw_response_json": map[string]any{
			"message": map[string]any{"id": "m1"},
		},
	})
	if err != nil {
		t.Fatalf("expected max criteria success, got %v", err)
	}
	if len(results) != len(Criteria()) {
		t.Fatalf("expected %d results, got %d", len(Criteria()), len(results))
	}

	results, err = Evaluate(map[string]any{
		"provider":          "max",
		"status_code":       200,
		"raw_response_json": map[string]any{"ok": true},
	})
	if err == nil {
		t.Fatalf("expected missing message failure")
	}
	if results[2].Passed {
		t.Fatalf("expected message criterion to fail, got %#v", results[2])
	}
}
