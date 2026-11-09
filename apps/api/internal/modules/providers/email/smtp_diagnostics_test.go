package email

import (
	"errors"
	"testing"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func TestAnalyzeSMTPClassifiesAuthAndRecipient(t *testing.T) {
	auth := AnalyzeSMTP(map[string]any{
		"provider": "smtp",
		"smtp": map[string]any{
			"accepted":          false,
			"failed_stage":      "auth",
			"smtp_code":         "535",
			"last_server_reply": "535 5.7.8 Authentication credentials invalid",
		},
	}, false, providerdiagnostics.BaseFailure(map[string]any{"provider": "smtp"}, false, nil))
	if auth.ReasonCode != "smtp_auth_failed" || auth.Retryable {
		t.Fatalf("expected smtp auth failed non-retryable, got %#v", auth)
	}

	rcpt := AnalyzeSMTP(map[string]any{
		"provider": "smtp",
		"smtp": map[string]any{
			"accepted":          false,
			"failed_stage":      "rcpt_to",
			"smtp_code":         "550",
			"last_server_reply": "550 5.1.1 User unknown",
		},
	}, false, providerdiagnostics.BaseFailure(map[string]any{"provider": "smtp"}, false, nil))
	if rcpt.ReasonCode != "smtp_recipient_rejected" || rcpt.Retryable {
		t.Fatalf("expected recipient rejected non-retryable, got %#v", rcpt)
	}
}

func TestIsRetryableSMTPFailure(t *testing.T) {
	if !IsRetryableSMTPFailure(map[string]any{
		"smtp": map[string]any{"failed_stage": "tcp_dial"},
	}, errors.New("connection reset")) {
		t.Fatalf("expected tcp_dial to be retryable")
	}
	if IsRetryableSMTPFailure(map[string]any{
		"smtp": map[string]any{"failed_stage": "auth", "smtp_code": "535"},
	}, errors.New("535 auth failed")) {
		t.Fatalf("expected auth failure to be final")
	}
}

func TestEvaluateSMTPCriteria(t *testing.T) {
	results, err := Evaluate(map[string]any{
		"provider": "smtp",
		"smtp": map[string]any{
			"accepted": true,
			"stages": []any{
				"tcp_connected",
				"smtp_banner",
				"mail_from",
				"rcpt",
				"data_sent",
				"quit",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected smtp criteria success, got %v", err)
	}
	if len(results) != len(Criteria()) {
		t.Fatalf("expected %d results, got %d", len(Criteria()), len(results))
	}

	results, err = Evaluate(map[string]any{
		"provider": "smtp",
		"smtp": map[string]any{
			"accepted": false,
			"stages":   []any{"tcp_connected", "smtp_banner", "mail_from"},
		},
	})
	if err == nil {
		t.Fatalf("expected smtp criteria failure")
	}
	if results[2].Passed {
		t.Fatalf("expected rcpt criterion to fail, got %#v", results[2])
	}
}
