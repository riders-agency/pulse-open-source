package http

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
	clientsmodule "riders.agency/pulse/api/internal/modules/clients"
	clientsvc "riders.agency/pulse/api/internal/modules/clients/service"
)

type langGraphToolGateway struct {
	ai          *aisvc.Commands
	clientTools *clientsvc.AIReadTools
	policy      aitools.ToolGateway
}

func NewLangGraphToolGateway(ai *aisvc.Commands) aitools.ToolGateway {
	var clientTools *clientsvc.AIReadTools
	if ai != nil && ai.Clients != nil {
		clientTools = clientsvc.NewAIReadTools(ai.P, ai.Clients.Clients)
	}
	return &langGraphToolGateway{
		ai:          ai,
		clientTools: clientTools,
		policy:      aitools.NewPolicyGateway(defaultToolCapabilities()),
	}
}

func defaultToolCapabilities() []aitools.ModuleCapability {
	capabilities := append([]aitools.ModuleCapability{}, aitools.DefaultModuleCapabilities()...)
	capabilities = append(capabilities, clientsmodule.AIToolCapabilities()...)
	return capabilities
}

func (g *langGraphToolGateway) ListTools(ctx context.Context, req aitools.ToolCatalogRequest) (aitools.ToolCatalog, error) {
	catalog, err := g.policy.ListTools(ctx, req)
	if err != nil {
		return catalog, err
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return aitools.ToolCatalog{}, err
	}
	policy, err := g.ai.EmployeePolicy(ctx, projectID)
	if err != nil {
		return aitools.ToolCatalog{}, err
	}
	filtered := make([]aitools.ToolDescriptor, 0, len(catalog.Tools))
	for _, tool := range catalog.Tools {
		if policy.Allows(tool.ToolKey) {
			filtered = append(filtered, tool)
		}
	}
	catalog.Tools = filtered
	return catalog, nil
}

func (g *langGraphToolGateway) ValidateToolCall(ctx context.Context, req aitools.ToolCallSpec) (aitools.ValidationResult, error) {
	validation, err := g.policy.ValidateToolCall(ctx, req)
	if err != nil || !validation.OK {
		return validation, err
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return aitools.ValidationResult{}, err
	}
	if err := g.ai.RequireEmployeeTool(ctx, projectID, req.ToolKey); err != nil {
		return aitools.ValidationResult{OK: false, ErrorCode: "tool_forbidden", UserMessage: "Инструмент недоступен этому сотруднику."}, nil
	}
	return validation, nil
}

func (g *langGraphToolGateway) ExecuteReadTool(ctx context.Context, req aitools.ToolCallSpec) (aitools.ToolObservation, error) {
	if !isLangGraphReadTool(req.ToolKey) {
		return aitools.ToolObservation{ToolCallID: req.ToolCallID, ToolKey: req.ToolKey, Status: "failed", ErrorCode: "read_only_endpoint", UserMessage: "Изменение данных требует подтверждения."}, nil
	}
	validation, err := g.ValidateToolCall(ctx, req)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	if !validation.OK {
		return aitools.ToolObservation{
			ToolCallID:  req.ToolCallID,
			ToolKey:     req.ToolKey,
			Status:      "failed",
			ErrorCode:   validation.ErrorCode,
			UserMessage: validation.UserMessage,
		}, nil
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	input := map[string]any{}
	if len(validation.NormalizedInputJSON) > 0 {
		if err := json.Unmarshal(validation.NormalizedInputJSON, &input); err != nil {
			return aitools.ToolObservation{}, err
		}
	}
	out, err := g.executeRead(ctx, projectID, sessionID, userID, req.ToolKey, input)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return aitools.ToolObservation{}, err
	}
	return aitools.ToolObservation{
		ToolCallID: req.ToolCallID,
		ToolKey:    req.ToolKey,
		Status:     "succeeded",
		OutputJSON: raw,
	}, nil
}
