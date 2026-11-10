package clients

import "testing"

func TestAIToolCapabilitiesOwnClientAndPaymentTools(t *testing.T) {
	caps := AIToolCapabilities()
	tools := map[string]string{}
	intents := map[string]string{}
	for _, cap := range caps {
		for _, tool := range cap.Tools {
			if previous := tools[tool.ToolKey]; previous != "" {
				t.Fatalf("duplicate tool key %s in %s and %s", tool.ToolKey, previous, cap.ModuleKey)
			}
			tools[tool.ToolKey] = cap.ModuleKey
		}
		for _, intent := range cap.Intents {
			if previous := intents[intent.Intent]; previous != "" {
				t.Fatalf("duplicate intent %s in %s and %s", intent.Intent, previous, cap.ModuleKey)
			}
			intents[intent.Intent] = cap.ModuleKey
		}
	}
	for _, key := range []string{"client.search", "client.create", "client.read", "payment.lookup_due", "payment.action.confirm"} {
		if tools[key] == "" {
			t.Fatalf("missing tool %s in clients-owned AI capabilities", key)
		}
	}
	for _, key := range []string{
		"client.search",
		"client.read",
		"client.count",
		"client.create",
		"client.update",
		"payment.lookup_due",
		"payment.schedule.read",
		"payment.action.create",
		"payment.confirm",
	} {
		if intents[key] == "" {
			t.Fatalf("missing intent %s in clients-owned AI capabilities", key)
		}
	}
}
