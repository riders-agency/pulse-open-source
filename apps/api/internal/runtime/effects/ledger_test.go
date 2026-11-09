package effects

import "testing"

func TestStableEffectIDIncludesAttempt(t *testing.T) {
	first := StableEffectID("run", "action", 1)
	second := StableEffectID("run", "action", 2)
	if first == second {
		t.Fatal("expected different effect ids per attempt")
	}
	if first != StableEffectID("run", "action", 1) {
		t.Fatal("expected stable effect id")
	}
}

func TestRequiresLedgerForExternalEffects(t *testing.T) {
	for _, key := range []string{"telegram.send_message", "max.send_message", "email.send", "http.request", "uon.create_lead", "clients.manual_payment.trigger"} {
		if !RequiresLedger(key) {
			t.Fatalf("expected %s to require ledger", key)
		}
	}
	if RequiresLedger("condition.branch") {
		t.Fatal("condition.branch should not require ledger")
	}
}
