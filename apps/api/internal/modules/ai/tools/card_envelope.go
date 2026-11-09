package tools

import (
	"encoding/json"
	"strings"
)

const CardEnvelopeSchemaV1 = "riders-pulse-ai-card-v1"

type CardEnvelope struct {
	Schema           string          `json:"schema"`
	CardID           string          `json:"card_id"`
	CardType         string          `json:"card_type"`
	Revision         int             `json:"revision"`
	ToolCallID       string          `json:"tool_call_id"`
	IntentKey        string          `json:"intent_key"`
	RiskLevel        RiskLevel       `json:"risk_level"`
	RequiresApproval bool            `json:"requires_approval"`
	Title            string          `json:"title"`
	Subtitle         string          `json:"subtitle"`
	Fields           json.RawMessage `json:"fields,omitempty"`
	Actions          json.RawMessage `json:"actions,omitempty"`
	DisplayJSON      json.RawMessage `json:"display_json,omitempty"`
	InputJSON        json.RawMessage `json:"input_json,omitempty"`
	EntityLinks      json.RawMessage `json:"entity_links,omitempty"`
	DebugJSON        json.RawMessage `json:"debug_json,omitempty"`
}

func NormalizeCardEnvelope(card CardEnvelope) CardEnvelope {
	if strings.TrimSpace(card.Schema) == "" {
		card.Schema = CardEnvelopeSchemaV1
	}
	card.RiskLevel = NormalizeRiskLevel(string(card.RiskLevel))
	if card.Revision <= 0 {
		card.Revision = 1
	}
	return card
}
