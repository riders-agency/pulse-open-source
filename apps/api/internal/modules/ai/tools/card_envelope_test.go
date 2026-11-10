package tools

import "testing"

func TestNormalizeCardEnvelopeDefaultsSchemaRiskAndRevision(t *testing.T) {
	card := NormalizeCardEnvelope(CardEnvelope{
		CardID:    "card-1",
		CardType:  "task_create_proposal",
		RiskLevel: RiskLevel("write"),
	})
	if card.Schema != CardEnvelopeSchemaV1 {
		t.Fatalf("schema = %q", card.Schema)
	}
	if card.RiskLevel != RiskHigh {
		t.Fatalf("risk = %q", card.RiskLevel)
	}
	if card.Revision != 1 {
		t.Fatalf("revision = %d", card.Revision)
	}
}
