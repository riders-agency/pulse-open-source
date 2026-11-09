package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	wfrepo "riders.agency/pulse/api/internal/modules/workflows/contracts"
)

func (a *AI) executeWorkflowPatch(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if a.Workflows == nil {
		return nil, fmt.Errorf("workflow service not configured")
	}
	workflowID, err := uuid.Parse(stringValue(input["workflow_id"]))
	if err != nil {
		return nil, fmt.Errorf("workflow_id required")
	}
	rawPatch, _ := json.Marshal(input["patch_json"])
	nextDef, nextGraph, testInput, err := a.applyWorkflowPatchJSON(ctx, workflowID, userID, rawPatch)
	if err != nil {
		return nil, err
	}
	patch := wfrepo.DraftPatch{Definition: nextDef, Graph: nextGraph}
	if len(testInput) > 0 {
		patch.TestInput = testInput
	}
	if err := a.Workflows.PutDraft(ctx, workflowID, userID, patch); err != nil {
		return nil, err
	}
	_, _ = a.P.Exec(ctx, `
UPDATE ai_patch_proposals
   SET status = 'applied', applied_by_user_id = $4, applied_at = now()
 WHERE project_id = $1
   AND target_entity_id = $2
   AND patch_target_type = 'workflow_draft'
   AND patch_json = $3::jsonb
   AND status IN ('proposed','validated','pending_approval')`, projectID, workflowID.String(), string(rawPatch), userID)
	return map[string]any{
		"ok":          true,
		"workflow_id": workflowID.String(),
		"url":         "/workflows/" + workflowID.String() + "/edit",
		"message":     "Patch применен в черновик workflow. Публикация не выполнялась.",
	}, nil
}

func (a *AI) executeTaskProcessPatch(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	workflowID, err := uuid.Parse(stringValue(input["task_workflow_id"]))
	if err != nil {
		return nil, fmt.Errorf("task_workflow_id required")
	}
	if err := a.WorkTracker.AddReviewStage(ctx, projectID, userID, workflowID); err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":               true,
		"task_workflow_id": workflowID.String(),
		"url":              "/tasks/processes",
		"message":          "В процесс задач добавлен статус Review и переход из In progress.",
	}, nil
}

func (a *AI) applyWorkflowPatchJSON(ctx context.Context, workflowID, userID uuid.UUID, rawPatch []byte) (json.RawMessage, json.RawMessage, json.RawMessage, error) {
	wf, err := a.Workflows.Get(ctx, workflowID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	var patch struct {
		Format         string           `json:"format"`
		BaseWorkflowID string           `json:"base_workflow_id"`
		Operations     []map[string]any `json:"operations"`
	}
	if err := json.Unmarshal(rawPatch, &patch); err != nil {
		return nil, nil, nil, err
	}
	if patch.Format != "riders-pulse-workflow-patch-v1" {
		return nil, nil, nil, fmt.Errorf("invalid workflow patch format")
	}
	if patch.BaseWorkflowID != "" && patch.BaseWorkflowID != workflowID.String() {
		return nil, nil, nil, fmt.Errorf("workflow patch target mismatch")
	}
	def := workflowDefinitionFromRaw(wf.DraftDefinition)
	graph := workflowGraphFromRaw(wf.DraftGraph, def)
	var testInput json.RawMessage
	for _, op := range patch.Operations {
		switch stringValue(op["op"]) {
		case "add_action", "upsert_action":
			action, ok := op["action"].(map[string]any)
			if !ok {
				return nil, nil, nil, fmt.Errorf("add_action requires action")
			}
			addOrUpdateWorkflowAction(def, graph, action)
		case "add_connection":
			from, to := stringValue(op["from"]), stringValue(op["to"])
			if from == "" || to == "" {
				return nil, nil, nil, fmt.Errorf("add_connection requires from/to")
			}
			branch := stringValue(op["branch"])
			addWorkflowConnection(def, graph, from, to, branch)
		case "set_test_input":
			testInput, _ = json.Marshal(op["test_input_json"])
		default:
			return nil, nil, nil, fmt.Errorf("unsupported workflow patch op %s", stringValue(op["op"]))
		}
	}
	defRaw, _ := json.Marshal(def)
	graphRaw, _ := json.Marshal(graph)
	return defRaw, graphRaw, testInput, nil
}
