package http

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
	"riders.agency/pulse/api/internal/platform/authz"
)

type pulseToolServiceServer struct {
	aiagentv1.UnimplementedPulseToolServiceServer
	gateway aitools.ToolGateway
}

func NewPulseToolServiceServer(gateway aitools.ToolGateway) aiagentv1.PulseToolServiceServer {
	return &pulseToolServiceServer{gateway: gateway}
}

func (s *pulseToolServiceServer) ListTools(ctx context.Context, req *aiagentv1.ListToolsRequest) (*aiagentv1.ListToolsResponse, error) {
	catalog, err := s.gateway.ListTools(ctx, aitools.ToolCatalogRequest{
		RequestID: req.GetRequestId(),
		CompanyID: req.GetCompanyId(),
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
	})
	if err != nil {
		return nil, err
	}
	out := &aiagentv1.ListToolsResponse{Tools: make([]*aiagentv1.ToolSpec, 0, len(catalog.Tools))}
	for _, tool := range catalog.Tools {
		out.Tools = append(out.Tools, &aiagentv1.ToolSpec{
			ToolKey:             tool.ToolKey,
			DisplayName:         tool.DisplayName,
			Description:         tool.Description,
			RiskLevel:           riskToProto(tool.RiskLevel),
			InputSchemaJson:     rawToStruct(tool.InputSchema),
			OutputSchemaJson:    rawToStruct(tool.OutputSchema),
			RequiredPermissions: []string{strings.TrimSpace(tool.Permission)},
			RequiresApproval:    tool.RequiresApproval,
			IdempotencyRequired: len(tool.IdempotencyKeyFields) > 0,
		})
	}
	return out, nil
}

func (s *pulseToolServiceServer) ReadContext(ctx context.Context, req *aiagentv1.ReadContextRequest) (*aiagentv1.ReadContextResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	observation, err := s.gateway.ExecuteReadTool(ctx, aitools.ToolCallSpec{
		ToolCallID:      req.GetRequestId(),
		CompanyID:       req.GetCompanyId(),
		ProjectID:       req.GetProjectId(),
		UserID:          req.GetUserId(),
		SessionID:       req.GetSessionId(),
		ToolKey:         "context.read",
		RiskLevel:       aitools.RiskRead,
		InputJSON:       structToRaw(req.GetPageContextJson()),
		PageContextJSON: structToRaw(req.GetPageContextJson()),
	})
	if err != nil {
		return nil, err
	}
	return &aiagentv1.ReadContextResponse{ContextJson: rawToStruct(observation.OutputJSON)}, nil
}

func (s *pulseToolServiceServer) ValidateToolCall(ctx context.Context, req *aiagentv1.ValidateToolCallRequest) (*aiagentv1.ValidateToolCallResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	result, err := s.gateway.ValidateToolCall(ctx, aitools.ToolCallSpec{
		ToolCallID: req.GetToolCallId(),
		CompanyID:  req.GetCompanyId(),
		ProjectID:  req.GetProjectId(),
		UserID:     req.GetUserId(),
		SessionID:  req.GetSessionId(),
		ToolKey:    req.GetToolKey(),
		RiskLevel:  aitools.RiskRead,
		InputJSON:  structToRaw(req.GetInputJson()),
	})
	if err != nil {
		return nil, err
	}
	return &aiagentv1.ValidateToolCallResponse{
		Ok:                  result.OK,
		NormalizedInputJson: rawToStruct(result.NormalizedInputJSON),
		ErrorCode:           result.ErrorCode,
		UserMessage:         result.UserMessage,
	}, nil
}

func (s *pulseToolServiceServer) CreateProposal(ctx context.Context, req *aiagentv1.CreateProposalRequest) (*aiagentv1.ToolProposalResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	proposal, err := s.gateway.CreateProposal(ctx, aitools.ToolCallSpec{
		ToolCallID:       req.GetToolCallId(),
		CompanyID:        req.GetCompanyId(),
		ProjectID:        req.GetProjectId(),
		UserID:           req.GetUserId(),
		SessionID:        req.GetSessionId(),
		ToolKey:          req.GetToolKey(),
		IntentKey:        req.GetIntentKey(),
		RiskLevel:        riskFromProto(req.GetRiskLevel()),
		RequiresApproval: req.GetRequiresApproval(),
		InputJSON:        structToRaw(req.GetInputJson()),
		PageContextJSON:  structToRaw(req.GetPageContextJson()),
		IdempotencyKey:   req.GetIdempotencyKey(),
		ExpectedCardType: req.GetExpectedCardType(),
		ParentPlanID:     req.GetParentPlanId(),
		StepID:           req.GetStepId(),
	})
	if err != nil {
		return nil, err
	}
	return proposalToProto(proposal), nil
}

func (s *pulseToolServiceServer) PatchProposal(ctx context.Context, req *aiagentv1.PatchProposalRequest) (*aiagentv1.ToolProposalResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	proposal, err := s.gateway.PatchProposal(ctx, aitools.PatchProposalRequest{
		CompanyID:       req.GetCompanyId(),
		ProjectID:       req.GetProjectId(),
		UserID:          req.GetUserId(),
		SessionID:       req.GetSessionId(),
		ToolCallID:      req.GetToolCallId(),
		CardID:          req.GetCardId(),
		BaseRevision:    int(req.GetBaseRevision()),
		PatchJSON:       structToRaw(req.GetPatchJson()),
		PageContextJSON: structToRaw(req.GetPageContextJson()),
	})
	if err != nil {
		return nil, err
	}
	return proposalToProto(proposal), nil
}

func (s *pulseToolServiceServer) ApproveProposal(ctx context.Context, req *aiagentv1.ApprovalRequest) (*aiagentv1.ExecuteToolResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	observation, err := s.gateway.ApproveProposal(ctx, approvalRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return observationToProto(observation), nil
}

func (s *pulseToolServiceServer) RejectProposal(ctx context.Context, req *aiagentv1.ApprovalRequest) (*aiagentv1.ExecuteToolResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	observation, err := s.gateway.RejectProposal(ctx, approvalRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return observationToProto(observation), nil
}

func (s *pulseToolServiceServer) ExecuteTool(ctx context.Context, req *aiagentv1.ExecuteToolRequest) (*aiagentv1.ExecuteToolResponse, error) {
	ctx, err := contextWithToolSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	observation, err := s.gateway.ExecuteReadTool(ctx, aitools.ToolCallSpec{
		ToolCallID:     req.GetToolCallId(),
		CompanyID:      req.GetCompanyId(),
		ProjectID:      req.GetProjectId(),
		UserID:         req.GetUserId(),
		SessionID:      req.GetSessionId(),
		ToolKey:        req.GetToolKey(),
		RiskLevel:      aitools.RiskRead,
		InputJSON:      structToRaw(req.GetInputJson()),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, err
	}
	return observationToProto(observation), nil
}

func (s *pulseToolServiceServer) ResolveEntity(ctx context.Context, req *aiagentv1.ResolveEntityRequest) (*aiagentv1.ResolveEntityResponse, error) {
	return &aiagentv1.ResolveEntityResponse{Status: "not_implemented"}, nil
}

func contextWithToolSession(ctx context.Context, userIDRaw, sessionIDRaw string) (context.Context, error) {
	userIDText := strings.TrimSpace(userIDRaw)
	if userIDText == "" {
		return ctx, status.Error(codes.InvalidArgument, "user_id is required")
	}
	userID, err := uuid.Parse(userIDText)
	if err != nil || userID == uuid.Nil {
		return ctx, status.Error(codes.InvalidArgument, "user_id must be a UUID")
	}

	sessionID := uuid.Nil
	sessionIDText := strings.TrimSpace(sessionIDRaw)
	if sessionIDText != "" {
		sessionID, err = uuid.Parse(sessionIDText)
		if err != nil {
			return ctx, status.Error(codes.InvalidArgument, "session_id must be a UUID")
		}
	}

	return authz.WithSessionClaims(ctx, authz.SessionClaims{
		SessionID: sessionID,
		UserID:    userID,
		ActorType: authz.ActorSession,
	}), nil
}

func approvalRequestFromProto(req *aiagentv1.ApprovalRequest) aitools.ApprovalRequest {
	return aitools.ApprovalRequest{
		CompanyID:      req.GetCompanyId(),
		ProjectID:      req.GetProjectId(),
		UserID:         req.GetUserId(),
		SessionID:      req.GetSessionId(),
		ToolCallID:     req.GetToolCallId(),
		CardID:         req.GetCardId(),
		CardRevision:   int(req.GetCardRevision()),
		Decision:       req.GetDecision(),
		IdempotencyKey: req.GetIdempotencyKey(),
	}
}

func proposalToProto(proposal aitools.ToolProposal) *aiagentv1.ToolProposalResponse {
	return &aiagentv1.ToolProposalResponse{
		ToolCallId: proposal.ToolCallID,
		ToolKey:    proposal.ToolKey,
		CardId:     proposal.CardID,
		CardType:   proposal.CardType,
		Revision:   int32(proposal.Revision),
		Status:     proposal.Status,
		InputJson:  rawToStruct(proposal.InputJSON),
		CardJson:   rawToStruct(proposal.CardJSON),
	}
}

func observationToProto(observation aitools.ToolObservation) *aiagentv1.ExecuteToolResponse {
	return &aiagentv1.ExecuteToolResponse{
		Ok:           observation.Status == "succeeded",
		Status:       statusToProto(observation.Status),
		OutputJson:   rawToStruct(observation.OutputJSON),
		ErrorCode:    observation.ErrorCode,
		UserMessage:  observation.UserMessage,
		AuditEventId: observation.AuditEventID,
	}
}

func riskFromProto(value aiagentv1.RiskLevel) aitools.RiskLevel {
	switch value {
	case aiagentv1.RiskLevel_RISK_LEVEL_MEDIUM:
		return aitools.RiskMedium
	case aiagentv1.RiskLevel_RISK_LEVEL_HIGH:
		return aitools.RiskHigh
	case aiagentv1.RiskLevel_RISK_LEVEL_CRITICAL:
		return aitools.RiskCritical
	default:
		return aitools.RiskRead
	}
}

func riskToProto(value aitools.RiskLevel) aiagentv1.RiskLevel {
	switch value {
	case aitools.RiskMedium:
		return aiagentv1.RiskLevel_RISK_LEVEL_MEDIUM
	case aitools.RiskHigh:
		return aiagentv1.RiskLevel_RISK_LEVEL_HIGH
	case aitools.RiskCritical:
		return aiagentv1.RiskLevel_RISK_LEVEL_CRITICAL
	default:
		return aiagentv1.RiskLevel_RISK_LEVEL_READ
	}
}

func statusToProto(value string) aiagentv1.ToolExecutionStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "succeeded", "success", "ok":
		return aiagentv1.ToolExecutionStatus_TOOL_EXECUTION_STATUS_SUCCEEDED
	case "denied", "rejected":
		return aiagentv1.ToolExecutionStatus_TOOL_EXECUTION_STATUS_DENIED
	case "needs_approval", "pending_approval":
		return aiagentv1.ToolExecutionStatus_TOOL_EXECUTION_STATUS_NEEDS_APPROVAL
	case "conflict":
		return aiagentv1.ToolExecutionStatus_TOOL_EXECUTION_STATUS_CONFLICT
	case "failed", "error":
		return aiagentv1.ToolExecutionStatus_TOOL_EXECUTION_STATUS_FAILED
	default:
		return aiagentv1.ToolExecutionStatus_TOOL_EXECUTION_STATUS_UNSPECIFIED
	}
}

func structToRaw(value *structpb.Struct) json.RawMessage {
	if value == nil {
		return json.RawMessage(`{}`)
	}
	raw, err := json.Marshal(value.AsMap())
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}

func rawToStruct(raw json.RawMessage) *structpb.Struct {
	return rawStruct(raw)
}
