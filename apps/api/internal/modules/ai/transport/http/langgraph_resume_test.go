package http

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
)

type resumeErrorRuntime struct {
	aiagentv1.UnimplementedAgentRuntimeServiceServer
}

func (resumeErrorRuntime) ResumeTurn(_ *aiagentv1.ResumeRequest, stream grpc.ServerStreamingServer[aiagentv1.AgentEvent]) error {
	event := contractEventWithError(1, "run", "thread", "session", "denied", map[string]any{"code": "policy_changed"})
	event.ToolCall = mustStruct(map[string]any{"execution": map[string]any{
		"schema": "pulse-execution-v1", "status": "completed",
	}})
	return stream.Send(event)
}

func TestResumeTraceDoesNotHideRuntimeError(t *testing.T) {
	addr, stop := serveContractRuntime(t, resumeErrorRuntime{})
	defer stop()
	t.Setenv("AI_AGENT_GRPC_ADDR", addr)
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	t.Setenv("PULSE_AI_EMPLOYEE_RUNTIME", "langgraph")
	t.Setenv("PULSE_AI_RESUME_READY", "true")
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
	_, handled, err := resumeLangGraphToolDecision(context.Background(), nil, airepo.ToolCallRuntimeContext{
		ProjectID: uuid.New(), SessionID: uuid.New(), ToolCallID: uuid.New(),
		AgentRuntime: "langgraph", Status: "pending_approval",
	}, uuid.New(), "approve", "")
	if !handled || err == nil || err.Error() != "policy_changed" {
		t.Fatalf("trace hid the terminal error: handled=%v err=%v", handled, err)
	}
}
