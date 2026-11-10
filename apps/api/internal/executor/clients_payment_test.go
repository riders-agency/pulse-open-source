package executor

import "testing"

func TestPaymentActionIDFromRunInput(t *testing.T) {
	t.Run("top level", func(t *testing.T) {
		got := paymentActionIDFromRunInput([]byte(`{"payment_action_id":"4c4b3c1d-9611-4970-95ea-166b8d8ad3ad"}`))
		if got != "4c4b3c1d-9611-4970-95ea-166b8d8ad3ad" {
			t.Fatalf("payment action id = %q", got)
		}
	})

	t.Run("pulse metadata", func(t *testing.T) {
		got := paymentActionIDFromRunInput([]byte(`{"_pulse":{"payment_action_id":"8c6ee89d-4f1a-4ec9-aa86-099309a66a20"}}`))
		if got != "8c6ee89d-4f1a-4ec9-aa86-099309a66a20" {
			t.Fatalf("payment action id = %q", got)
		}
	})
}
