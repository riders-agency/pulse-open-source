package http

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
)

func cancelLangGraphSession(ctx context.Context, sessionCtx airepo.SessionRuntimeContext, userID uuid.UUID) (bool, error) {
	if !shouldCancelLangGraphSession(sessionCtx) {
		return false, nil
	}
	addr := aiAgentGRPCTarget()
	dialOptions, _, err := aiAgentGRPCDialOptions(addr)
	if err != nil {
		return true, err
	}
	dialCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	conn, err := connectAgent(dialCtx, addr, dialOptions)
	if err != nil {
		return true, fmt.Errorf("python LangGraph unavailable: %w", err)
	}
	defer conn.Close()

	threadID := firstNonEmpty(
		strings.TrimSpace(sessionCtx.LangGraphThreadID),
		"pulse:"+sessionCtx.ProjectID.String()+":"+sessionCtx.SessionID.String(),
	)
	requestID := uuid.NewString()
	resp, err := aiagentv1.NewAgentRuntimeServiceClient(conn).CancelRun(
		aiAgentOutgoingContext(dialCtx),
		&aiagentv1.CancelRunRequest{
			RequestId: requestID,
			ProjectId: sessionCtx.ProjectID.String(),
			SessionId: sessionCtx.SessionID.String(),
			ThreadId:  threadID,
			RunId:     requestID,
			AuthContext: &aiagentv1.AuthContext{
				CompanyId: sessionCtx.CompanyID,
				ProjectId: sessionCtx.ProjectID.String(),
				UserId:    userID.String(),
			},
		},
	)
	if err != nil {
		return true, fmt.Errorf("langgraph cancel run: %w", err)
	}
	if resp.GetOk() {
		return true, nil
	}
	if strings.EqualFold(strings.TrimSpace(resp.GetStatus()), "checkpoint_not_found") {
		return true, nil
	}
	return true, fmt.Errorf("langgraph cancel run: %s", firstNonEmpty(resp.GetStatus(), "unknown_status"))
}

func shouldCancelLangGraphSession(sessionCtx airepo.SessionRuntimeContext) bool {
	if !strings.EqualFold(strings.TrimSpace(sessionCtx.AgentRuntime), "langgraph") {
		return false
	}
	if !strings.EqualFold(aiEmployeeRuntime(), "langgraph") {
		return false
	}
	if !boolEnv("PULSE_AI_RESUME_READY") || !boolEnv("PULSE_AI_CHECKPOINTER_READY") {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(sessionCtx.Status)) {
	case "canceled", "completed", "failed":
		return false
	default:
		return true
	}
}
