package http

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

func (g *langGraphToolGateway) CreateProposal(ctx context.Context, req aitools.ToolCallSpec) (aitools.ToolProposal, error) {
	validation, err := g.ValidateToolCall(ctx, req)
	if err != nil {
		return aitools.ToolProposal{}, err
	}
	if !validation.OK {
		return aitools.ToolProposal{}, fmt.Errorf("%s", validation.ErrorCode)
	}
	return g.policy.CreateProposal(ctx, req)
}

func (g *langGraphToolGateway) executeRead(ctx context.Context, projectID, sessionID, userID uuid.UUID, toolKey string, input map[string]any) (map[string]any, error) {
	switch toolKey {
	case "ui_context.read":
		return g.ai.ReadLangGraphPageContext(ctx, projectID, sessionID, userID, input)
	case "context.read", "integration.catalog.list", "integration.dependency.read", "task_process.read":
		return g.ai.ReadLangGraphContext(ctx, projectID, sessionID, userID, input)
	case "task.search", "task.read":
		return g.ai.ReadLangGraphTaskSearch(ctx, projectID, sessionID, userID, input)
	case "client.search", "client.read":
		return g.clientTools.ReadClientSearch(ctx, projectID, userID, input)
	case "client.count":
		return g.clientTools.ReadClientCount(ctx, projectID, userID, input)
	case "payment.lookup_due", "payment.schedule.read":
		return g.clientTools.ReadPaymentDue(ctx, projectID, userID, input)
	case "workflow.read_draft", "workflow.read_context", "workflow.compile_prompt":
		return g.ai.ReadLangGraphWorkflowDraft(ctx, projectID, sessionID, userID, input)
	default:
		return nil, fmt.Errorf("unknown read tool")
	}
}
