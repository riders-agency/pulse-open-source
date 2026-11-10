package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/modules/ai/execution"

	"github.com/google/uuid"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
)

func resumeLangGraphToolDecision(
	ctx context.Context,
	ai *aisvc.Commands,
	resumeCtx airepo.ToolCallRuntimeContext,
	userID uuid.UUID,
	action string,
	reason string,
) (map[string]any, bool, error) {
	if !shouldResumeLangGraphTool(resumeCtx) {
		return nil, false, nil
	}
	addr := aiAgentGRPCTarget()
	dialOptions, _, err := aiAgentGRPCDialOptions(addr)
	if err != nil {
		return nil, true, err
	}
	dialCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	conn, err := connectAgent(dialCtx, addr, dialOptions)
	if err != nil {
		return nil, true, fmt.Errorf("python LangGraph unavailable: %w", err)
	}
	defer conn.Close()

	action = normalizeResumeAction(action)
	runKey := uuid.NewString()
	client := aiagentv1.NewAgentRuntimeServiceClient(conn)
	stream, err := client.ResumeTurn(aiAgentOutgoingContext(ctx), &aiagentv1.ResumeRequest{
		RequestId:      runKey,
		IdempotencyKey: resumeCtx.ProjectID.String() + ":" + resumeCtx.SessionID.String() + ":" + resumeCtx.ToolCallID.String() + ":" + action,
		CompanyId:      resumeCtx.CompanyID,
		ProjectId:      resumeCtx.ProjectID.String(),
		UserId:         userID.String(),
		SessionId:      resumeCtx.SessionID.String(),
		ThreadId:       firstNonEmpty(strings.TrimSpace(resumeCtx.LangGraphThreadID), "pulse:"+resumeCtx.ProjectID.String()+":"+resumeCtx.SessionID.String()),
		RunId:          runKey,
		InterruptId:    resumeCtx.ToolCallID.String(),
		Decision: &aiagentv1.ApprovalDecision{
			Action: action,
			Reason: strings.TrimSpace(reason),
		},
		AuthContext: &aiagentv1.AuthContext{
			CompanyId: resumeCtx.CompanyID,
			ProjectId: resumeCtx.ProjectID.String(),
			UserId:    userID.String(),
		},
	})
	if err != nil {
		return nil, true, fmt.Errorf("langgraph resume turn: %w", err)
	}

	var runtimeErr string
	var trace *execution.Trace
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			if runtimeErr != "" {
				return nil, true, errors.New(runtimeErr)
			}
			if err := ai.SaveExecutionTrace(ctx, resumeCtx.ProjectID, resumeCtx.SessionID, userID, resumeCtx.ToolCallID, trace); err != nil {
				return nil, true, err
			}
			out, getErr := ai.GetSession(ctx, resumeCtx.ProjectID, resumeCtx.SessionID, userID)
			return out, true, getErr
		}
		if err != nil {
			return nil, true, fmt.Errorf("langgraph resume stream: %w", err)
		}
		if evidence := execution.Parse(event.GetToolCall().AsMap()["execution"]); evidence != nil {
			trace = evidence
		}
		switch event.GetEventType() {
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_ERROR:
			errPayload := event.GetError().AsMap()
			code := strings.TrimSpace(stringFromAny(errPayload["code"]))
			if code == "" {
				code = "langgraph_resume_error"
			}
			runtimeErr = code
		}
	}
}

func shouldResumeLangGraphTool(resumeCtx airepo.ToolCallRuntimeContext) bool {
	if !strings.EqualFold(strings.TrimSpace(resumeCtx.AgentRuntime), "langgraph") {
		return false
	}
	if !strings.EqualFold(aiEmployeeRuntime(), "langgraph") {
		return false
	}
	if !boolEnv("PULSE_AI_RESUME_READY") {
		return false
	}
	if !boolEnv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP") {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(resumeCtx.Status), "pending_approval")
}

func normalizeResumeAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "apply", "approve", "approved":
		return "approve"
	case "reject", "rejected":
		return "reject"
	case "cancel", "canceled", "cancelled":
		return "cancel"
	default:
		return strings.TrimSpace(action)
	}
}
