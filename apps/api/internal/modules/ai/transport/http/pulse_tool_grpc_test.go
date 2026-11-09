package http

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
	"riders.agency/pulse/api/internal/platform/authz"
)

func TestPulseToolServiceServerExecuteToolAddsSessionClaims(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()
	projectID := uuid.New()
	gateway := &recordingPulseToolGateway{}
	server := NewPulseToolServiceServer(gateway)

	resp, err := server.ExecuteTool(context.Background(), &aiagentv1.ExecuteToolRequest{
		ToolCallId: "tool-call-1",
		CompanyId:  uuid.NewString(),
		ProjectId:  projectID.String(),
		UserId:     userID.String(),
		SessionId:  sessionID.String(),
		ToolKey:    "ui_context.read",
	})
	if err != nil {
		t.Fatalf("ExecuteTool returned error: %v", err)
	}
	if !resp.GetOk() {
		t.Fatalf("ExecuteTool ok = false")
	}
	if gateway.lastExecute.ToolKey != "ui_context.read" {
		t.Fatalf("tool key = %q, want ui_context.read", gateway.lastExecute.ToolKey)
	}
	claims, ok := authz.SessionClaimsFrom(gateway.lastExecuteCtx)
	if !ok {
		t.Fatalf("missing authz session claims")
	}
	if claims.ActorType != authz.ActorSession {
		t.Fatalf("actor type = %q, want %q", claims.ActorType, authz.ActorSession)
	}
	if claims.UserID != userID {
		t.Fatalf("user id = %s, want %s", claims.UserID, userID)
	}
	if claims.SessionID != sessionID {
		t.Fatalf("session id = %s, want %s", claims.SessionID, sessionID)
	}
}

func TestPulseToolServiceServerExecuteToolRejectsMissingUserID(t *testing.T) {
	t.Parallel()

	server := NewPulseToolServiceServer(&recordingPulseToolGateway{})
	_, err := server.ExecuteTool(context.Background(), &aiagentv1.ExecuteToolRequest{
		ToolCallId: "tool-call-1",
		ProjectId:  uuid.NewString(),
		SessionId:  uuid.NewString(),
		ToolKey:    "ui_context.read",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s (err=%v)", status.Code(err), codes.InvalidArgument, err)
	}
}

type recordingPulseToolGateway struct {
	lastExecuteCtx context.Context
	lastExecute    aitools.ToolCallSpec
}

func (g *recordingPulseToolGateway) ListTools(ctx context.Context, req aitools.ToolCatalogRequest) (aitools.ToolCatalog, error) {
	return aitools.ToolCatalog{}, nil
}

func (g *recordingPulseToolGateway) ValidateToolCall(ctx context.Context, req aitools.ToolCallSpec) (aitools.ValidationResult, error) {
	return aitools.ValidationResult{OK: true, NormalizedInputJSON: req.InputJSON}, nil
}

func (g *recordingPulseToolGateway) ExecuteReadTool(ctx context.Context, req aitools.ToolCallSpec) (aitools.ToolObservation, error) {
	g.lastExecuteCtx = ctx
	g.lastExecute = req
	return aitools.ToolObservation{
		ToolCallID: req.ToolCallID,
		ToolKey:    req.ToolKey,
		Status:     "succeeded",
		OutputJSON: json.RawMessage(`{"answer":"ok"}`),
	}, nil
}

func (g *recordingPulseToolGateway) CreateProposal(ctx context.Context, req aitools.ToolCallSpec) (aitools.ToolProposal, error) {
	return aitools.ToolProposal{}, nil
}

func (g *recordingPulseToolGateway) PatchProposal(ctx context.Context, req aitools.PatchProposalRequest) (aitools.ToolProposal, error) {
	return aitools.ToolProposal{}, nil
}

func (g *recordingPulseToolGateway) ApproveProposal(ctx context.Context, req aitools.ApprovalRequest) (aitools.ToolObservation, error) {
	return aitools.ToolObservation{}, nil
}

func (g *recordingPulseToolGateway) RejectProposal(ctx context.Context, req aitools.ApprovalRequest) (aitools.ToolObservation, error) {
	return aitools.ToolObservation{}, nil
}
