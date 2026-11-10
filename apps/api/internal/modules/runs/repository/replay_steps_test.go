package repository

import (
	"encoding/json"
	"testing"
)

func TestInputWithReplayStepsKeepsWebhookAndNormalizedContact(t *testing.T) {
	input := []byte(`{"webhook":{"selected_date":"8 — 19 декабря 2026 г.","phone":"+79000000000"},"_pulse":{"webhook_slug":"va-booking","steps":{"manual":{"value":"keep"}}}}`)
	steps := map[string]any{
		"phone_normalize": map[string]any{"phone_e164": "+79000000000"},
		"fio_normalize":   map[string]any{"u_name": "Иван"},
	}
	merged, err := inputWithReplaySteps(input, steps)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(merged, &root); err != nil {
		t.Fatal(err)
	}
	webhook := root["webhook"].(map[string]any)
	if webhook["selected_date"] != "8 — 19 декабря 2026 г." {
		t.Fatalf("webhook payload changed: %#v", webhook)
	}
	pulse := root["_pulse"].(map[string]any)
	gotSteps := pulse["steps"].(map[string]any)
	if gotSteps["phone_normalize"].(map[string]any)["phone_e164"] != "+79000000000" ||
		gotSteps["fio_normalize"].(map[string]any)["u_name"] != "Иван" ||
		gotSteps["manual"].(map[string]any)["value"] != "keep" {
		t.Fatalf("replay lost existing or normalized steps: %#v", gotSteps)
	}
}
