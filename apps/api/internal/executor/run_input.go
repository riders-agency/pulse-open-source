package executor

import (
	"encoding/json"
	"strings"
)

// EnrichRunInputProjectID adds project_id (and batch payment markers) when missing from run input.
func EnrichRunInputProjectID(runInput []byte, projectID string) []byte {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" || projectIDFromRunInput(runInput) != "" {
		return runInput
	}
	root := map[string]any{}
	if len(bytesTrim(runInput)) > 0 {
		_ = json.Unmarshal(runInput, &root)
	}
	if root == nil {
		root = map[string]any{}
	}
	pulse, _ := root["_pulse"].(map[string]any)
	if pulse == nil {
		pulse = map[string]any{}
		root["_pulse"] = pulse
	}
	pulse["project_id"] = projectID
	if _, ok := pulse["batch"]; !ok {
		pulse["batch"] = true
	}
	proj, _ := root["project"].(map[string]any)
	if proj == nil {
		proj = map[string]any{}
		root["project"] = proj
	}
	proj["project_id"] = projectID
	pc, _ := root["payment_cycle"].(map[string]any)
	if pc == nil {
		pc = map[string]any{}
		root["payment_cycle"] = pc
	}
	if _, ok := pc["batch"]; !ok {
		pc["batch"] = true
	}
	out, err := json.Marshal(root)
	if err != nil {
		return runInput
	}
	return out
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

// DefinitionHasPaymentDueTrigger reports whether workflow definition JSON includes clients.payment_due.trigger.
func DefinitionHasPaymentDueTrigger(def []byte) bool {
	var parsed struct {
		Nodes []struct {
			ActionTypeKey string `json:"action_type_key"`
		} `json:"nodes"`
	}
	if json.Unmarshal(def, &parsed) != nil {
		return false
	}
	for _, n := range parsed.Nodes {
		if n.ActionTypeKey == "clients.payment_due.trigger" {
			return true
		}
	}
	return false
}
