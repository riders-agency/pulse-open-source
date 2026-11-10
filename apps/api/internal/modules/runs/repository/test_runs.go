package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	workflowcommands "riders.agency/pulse/api/internal/modules/workflows/commands"
	"riders.agency/pulse/api/internal/platform/uow"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/executor"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

var ErrInvalidTestMode = errors.New("invalid test run mode")

type TestRunRequest struct {
	InputJSON      map[string]any `json:"input_json"`
	ActionKey      string         `json:"action_key"`
	Mode           string         `json:"mode"`
	StepsOverrides map[string]any `json:"steps_overrides"`
}

func (r *Runs) TestTx(ctx context.Context, wfID, uid uuid.UUID, req TestRunRequest) (string, error) {
	if err := r.assertWorkflow(ctx, wfID, uid); err != nil {
		return "", err
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "full"
	}
	actionKey := strings.TrimSpace(req.ActionKey)
	if actionKey != "" && mode == "full" {
		mode = "through_action"
	}
	if mode != "full" && mode != "through_action" && mode != "selected_only" {
		return "", ErrInvalidTestMode
	}

	tx, err := r.transaction()
	if err != nil {
		return "", err
	}

	snapshot, err := workflowcommands.CreateTestSnapshot(ctx, tx, wfID, uid)
	if err != nil {
		return "", err
	}
	projectID, versionID, def, storedInput := snapshot.ProjectID, snapshot.VersionID, snapshot.Definition, snapshot.TestInput

	inputObject := req.InputJSON
	if inputObject == nil {
		inputObject = map[string]any{}
		_ = json.Unmarshal(storedInput, &inputObject)
	}
	inputBytes, err := testInputWithPulse(inputObject, actionKey, mode, req.StepsOverrides, firstWebhookSlug(def))
	if err != nil {
		return "", err
	}
	if executor.DefinitionHasPaymentDueTrigger(def) {
		inputBytes = executor.EnrichRunInputProjectID(inputBytes, projectID)
	}

	runID := uuid.New()
	if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{
		WorkflowRunID:     runID,
		ProjectID:         projectID,
		WorkflowID:        wfID.String(),
		WorkflowVersionID: versionID.String(),
		SourceType:        "test",
		InputJSON:         json.RawMessage(inputBytes),
		EventActionKey:    actionKey,
		EventMessage:      "editor test run queued",
	}); err != nil {
		return "", err
	}
	if r.E != nil {
		uow.AfterCommit(ctx, r.E.Notify)
	}
	return runID.String(), nil
}

func testInputWithPulse(input map[string]any, actionKey, mode string, steps map[string]any, webhookSlug string) ([]byte, error) {
	if input == nil {
		input = map[string]any{}
	}
	pulse, _ := input["_pulse"].(map[string]any)
	if pulse == nil {
		pulse = map[string]any{}
	}
	pulse["source_type"] = "test"
	pulse["test_mode"] = mode
	if strings.TrimSpace(webhookSlug) != "" {
		pulse["webhook_slug"] = strings.TrimSpace(webhookSlug)
		pulse["webhook_endpoint_key"] = strings.TrimSpace(webhookSlug)
	}
	if actionKey != "" {
		pulse["test_action_key"] = actionKey
	}
	if len(steps) > 0 {
		pulse["steps"] = steps
	}
	input["_pulse"] = pulse
	return json.Marshal(input)
}

func firstWebhookSlug(def []byte) string {
	var parsed struct {
		Nodes []struct {
			ActionTypeKey string         `json:"action_type_key"`
			Config        map[string]any `json:"config"`
		} `json:"nodes"`
	}
	if len(def) == 0 || json.Unmarshal(def, &parsed) != nil {
		return ""
	}
	for _, node := range parsed.Nodes {
		if strings.TrimSpace(node.ActionTypeKey) != "webhook.trigger" {
			continue
		}
		if value, ok := node.Config["endpoint_key"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
