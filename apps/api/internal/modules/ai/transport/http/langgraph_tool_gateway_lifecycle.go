package http

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

func (g *langGraphToolGateway) PatchProposal(ctx context.Context, req aitools.PatchProposalRequest) (aitools.ToolProposal, error) {
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return aitools.ToolProposal{}, err
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return aitools.ToolProposal{}, err
	}
	toolCallID, err := uuid.Parse(req.ToolCallID)
	if err != nil {
		return aitools.ToolProposal{}, err
	}
	patch := map[string]any{}
	if len(req.PatchJSON) > 0 {
		if err := json.Unmarshal(req.PatchJSON, &patch); err != nil {
			return aitools.ToolProposal{}, err
		}
	}
	if _, err := g.ai.UpdateToolCall(ctx, projectID, toolCallID, userID, patch, "langgraph proposal patch"); err != nil {
		return aitools.ToolProposal{}, err
	}
	return aitools.ToolProposal{
		ToolCallID: req.ToolCallID,
		CardID:     req.CardID,
		Revision:   req.BaseRevision + 1,
		Status:     "needs_approval",
		InputJSON:  req.PatchJSON,
		CardJSON:   json.RawMessage(`{}`),
	}, nil
}

func (g *langGraphToolGateway) ApproveProposal(ctx context.Context, req aitools.ApprovalRequest) (aitools.ToolObservation, error) {
	projectID, toolCallID, userID, err := parseApprovalIDs(req)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	session, err := g.ai.ApproveToolCall(ctx, projectID, toolCallID, userID)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	output := approvalCompletionObservation(session)
	return lifecycleObservation(req, output)
}

func approvalCompletionObservation(session map[string]any) map[string]any {
	// The previous assistant turn describes the proposal, not the operation just applied.
	return approvalObservation(session, "approved", firstNonEmpty(lastToolBody(session), "Действие обработано."))
}

func (g *langGraphToolGateway) RejectProposal(ctx context.Context, req aitools.ApprovalRequest) (aitools.ToolObservation, error) {
	projectID, toolCallID, userID, err := parseApprovalIDs(req)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	session, err := g.ai.RejectToolCall(ctx, projectID, toolCallID, userID, "natural language rejection")
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	output := approvalObservation(session, "rejected", firstNonEmpty(lastToolBody(session), "Действие отменено. Данные не менялись."))
	return lifecycleObservation(req, output)
}

func lifecycleObservation(req aitools.ApprovalRequest, output map[string]any) (aitools.ToolObservation, error) {
	raw, err := json.Marshal(output)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	return aitools.ToolObservation{
		ToolCallID:   req.ToolCallID,
		ToolKey:      stringFromAny(output["tool_key"]),
		Status:       "succeeded",
		OutputJSON:   raw,
		AuditEventID: req.IdempotencyKey,
	}, nil
}

func parseApprovalIDs(req aitools.ApprovalRequest) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	toolCallID, err := uuid.Parse(req.ToolCallID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectID, toolCallID, userID, nil
}
