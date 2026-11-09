package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClassifyProviderHTTP(t *testing.T) {
	if got := Classify("provider_http", errors.New("limited"), true, `{"status_code":429}`); got != ClassRateLimited {
		t.Fatalf("expected rate_limited, got %s", got)
	}
	if got := Classify("provider_http", errors.New("bad token"), true, `{"http":{"status":401}}`); got != ClassAuth {
		t.Fatalf("expected auth, got %s", got)
	}
	if got := Classify("provider_http", errors.New("gateway timeout"), true, `{"status":504}`); got != ClassTimeoutUnknown {
		t.Fatalf("expected timeout_unknown, got %s", got)
	}
	if got := Classify("provider_http", errors.New("bad request"), false, `{"status":400}`); got != ClassPermanent {
		t.Fatalf("expected permanent, got %s", got)
	}
}

func TestClassifyMAXAttachmentProcessingAsTransient(t *testing.T) {
	got := Classify(
		"attachment_processing",
		errors.New("max returned HTTP 400"),
		true,
		`{"status_code":400,"code":"attachment.not.ready"}`,
	)
	if got != ClassTransient {
		t.Fatalf("expected transient attachment processing, got %s", got)
	}
}

func TestEvaluateHonorsRetryabilityAttemptsAndElapsed(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	eval := Evaluate(Failure{
		ActionTypeKey:        "http.request",
		AttemptNo:            1,
		MaxAttempts:          3,
		Retryable:            true,
		ErrorType:            "provider_http",
		Err:                  errors.New("provider failed"),
		ProviderResponseJSON: `{"status":503}`,
		StartedAt:            now,
		Now:                  now,
		Seed:                 "run:action:1",
	})
	if eval.Class != ClassTransient || !eval.ShouldRetry || eval.NextDelay <= 0 {
		t.Fatalf("expected retryable transient evaluation, got %#v", eval)
	}

	exhausted := Evaluate(Failure{
		ActionTypeKey:        "http.request",
		AttemptNo:            3,
		MaxAttempts:          3,
		Retryable:            true,
		ErrorType:            "provider_http",
		Err:                  errors.New("provider failed"),
		ProviderResponseJSON: `{"status":503}`,
		StartedAt:            now,
		Now:                  now,
		Seed:                 "run:action:3",
	})
	if !exhausted.Exhausted || exhausted.ShouldRetry {
		t.Fatalf("expected exhausted terminal evaluation, got %#v", exhausted)
	}

	elapsed := Evaluate(Failure{
		ActionTypeKey: "clients.payment_due.trigger",
		AttemptNo:     1,
		MaxAttempts:   3,
		Retryable:     true,
		ErrorType:     "dependency",
		Err:           errors.New("db down"),
		StartedAt:     now,
		Now:           now.Add(3 * time.Minute),
		Seed:          "run:action:elapsed",
	})
	if !elapsed.Exhausted || elapsed.ShouldRetry {
		t.Fatalf("expected max elapsed exhaustion, got %#v", elapsed)
	}
}

func TestEvaluateHonorsWorkflowRetryTimingOverrides(t *testing.T) {
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	eval := Evaluate(Failure{
		ActionTypeKey: "email.send",
		AttemptNo:     5,
		MaxAttempts:   10,
		Retryable:     true,
		ErrorType:     "smtp",
		Err:           errors.New("temporary rate limit"),
		StartedAt:     now,
		Now:           now,
		InitialDelay:  10 * time.Second,
		MaxDelay:      90 * time.Second,
		MaxElapsed:    12 * time.Minute,
	})
	if !eval.ShouldRetry || eval.NextDelay != 90*time.Second {
		t.Fatalf("workflow retry override was not applied: %#v", eval)
	}
	if eval.MaxElapsedMS != (12 * time.Minute).Milliseconds() {
		t.Fatalf("max elapsed = %d, want %d", eval.MaxElapsedMS, (12 * time.Minute).Milliseconds())
	}
}

func TestTimeoutUnknownAndCanceledAreTerminal(t *testing.T) {
	timeout := Evaluate(Failure{
		ActionTypeKey: "telegram.send_message",
		AttemptNo:     1,
		MaxAttempts:   3,
		Retryable:     true,
		ErrorType:     "network",
		Err:           context.DeadlineExceeded,
		Seed:          "timeout",
	})
	if timeout.Class != ClassTimeoutUnknown || timeout.ShouldRetry {
		t.Fatalf("expected timeout_unknown terminal evaluation, got %#v", timeout)
	}
	canceled := Evaluate(Failure{
		AttemptNo:   1,
		MaxAttempts: 3,
		Retryable:   true,
		ErrorType:   "canceled",
		Err:         context.Canceled,
	})
	if canceled.Class != ClassCanceled || canceled.ShouldRetry {
		t.Fatalf("expected canceled terminal evaluation, got %#v", canceled)
	}
}

func TestRedact(t *testing.T) {
	msg := Redact("token=abc123 email user@example.com Authorization: Bearer secret")
	if msg == "" || msg == "token=abc123 email user@example.com Authorization: Bearer secret" {
		t.Fatalf("expected redacted message, got %q", msg)
	}
}
