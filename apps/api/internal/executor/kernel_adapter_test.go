package executor

import (
	"testing"
	"time"
)

func TestKernelNodePreservesWorkflowRetryTiming(t *testing.T) {
	node := kernelNode(actionPlan{
		ActionKey:     "notify",
		ActionTypeKey: "email.send",
		Config: actionConfig{RetryPolicy: retryPolicy{
			MaxAttempts:     10,
			InitialInterval: "10s",
			MaxInterval:     "90s",
			MaxElapsed:      "12m",
		}},
	})
	if node.MaxAttempts != 10 {
		t.Fatalf("max attempts = %d, want 10", node.MaxAttempts)
	}
	if node.RetryInitialDelay != 10*time.Second || node.RetryMaxDelay != 90*time.Second || node.RetryMaxElapsed != 12*time.Minute {
		t.Fatalf("unexpected retry timing: initial=%s max=%s elapsed=%s", node.RetryInitialDelay, node.RetryMaxDelay, node.RetryMaxElapsed)
	}
}

func TestParsePositiveDurationRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"", "invalid", "0s", "-1s"} {
		if got := parsePositiveDuration(raw); got != 0 {
			t.Fatalf("parsePositiveDuration(%q) = %s, want zero", raw, got)
		}
	}
}
