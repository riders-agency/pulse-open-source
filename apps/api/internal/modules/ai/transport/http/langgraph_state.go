package http

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
)

func attachLangGraphThreadState(ctx context.Context, session map[string]any, projectID, sessionID, userID uuid.UUID) map[string]any {
	if session == nil || aiEmployeeRuntime() != "langgraph" {
		return session
	}
	state, err := loadLangGraphThreadState(ctx, session, projectID, sessionID, userID)
	if err != nil {
		session["runtime_state_json"] = map[string]any{
			"status":          "unavailable",
			"runtime":         "langgraph",
			"error_code":      "runtime_state_unavailable",
			"error_text":      err.Error(),
			"thread_revision": intFromAny(session["session_revision"]),
		}
		return session
	}
	session["runtime_state_json"] = state
	return session
}

func loadLangGraphThreadState(ctx context.Context, session map[string]any, projectID, sessionID, userID uuid.UUID) (map[string]any, error) {
	addr := aiAgentGRPCTarget()
	dialOptions, _, err := aiAgentGRPCDialOptions(addr)
	if err != nil {
		return nil, err
	}
	rpcCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	conn, err := connectAgent(rpcCtx, addr, dialOptions)
	if err != nil {
		return nil, fmt.Errorf("python LangGraph unavailable: %w", err)
	}
	defer conn.Close()

	threadID := strings.TrimSpace(stringFromAny(session["langgraph_thread_id"]))
	if threadID == "" {
		threadID = "pulse:" + projectID.String() + ":" + sessionID.String()
	}
	client := aiagentv1.NewAgentRuntimeServiceClient(conn)
	resp, err := client.GetThreadState(aiAgentOutgoingContext(rpcCtx), &aiagentv1.GetThreadStateRequest{
		RequestId: uuid.NewString(),
		ProjectId: projectID.String(),
		SessionId: sessionID.String(),
		ThreadId:  threadID,
		AuthContext: &aiagentv1.AuthContext{
			CompanyId: stringFromAny(session["company_id"]),
			ProjectId: projectID.String(),
			UserId:    userID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("langgraph get thread state: %w", err)
	}
	out := sanitizeLangGraphThreadState(resp.GetStateJson().AsMap())
	out["status"] = "ok"
	out["thread_id"] = resp.GetThreadId()
	out["run_id"] = resp.GetRunId()
	out["thread_revision"] = intFromAny(session["session_revision"])
	out["session_status"] = stringFromAny(session["status"])
	out["active_tool_call_id"] = firstNonEmpty(stringFromAny(out["active_tool_call_id"]), stringFromAny(session["active_tool_call_id"]))
	out["active_patch_proposal_id"] = stringFromAny(session["active_patch_proposal_id"])
	out["cards"] = sanitizeLangGraphCards(resp.GetCards())
	return out, nil
}

func sanitizeLangGraphThreadState(raw map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{
		"ready",
		"runtime",
		"service_version",
		"python_version",
		"langgraph_version",
		"grpcio_version",
		"project_id",
		"session_id",
		"checkpoint_enabled",
		"checkpoint_namespace",
		"checkpoint_thread_id",
		"checkpoint_found",
		"state_schema_version",
		"current_step",
		"active_mode",
		"dialogue_state",
		"active_topic",
		"suspended_topics",
		"resolved_entities",
		"unresolved_slots",
		"active_card_id",
		"active_tool_call_id",
		"active_proposal_id",
		"pending_interrupt",
		"cancel_requested",
		"canceled_at",
		"cancel_reason",
		"next_nodes",
		"last_error",
	} {
		if value, ok := raw[key]; ok {
			out[key] = value
		}
	}
	return out
}

func sanitizeLangGraphCards(cards []*structpb.Struct) []map[string]any {
	out := make([]map[string]any, 0, len(cards))
	for _, card := range cards {
		if card == nil {
			continue
		}
		out = append(out, sanitizeLangGraphCard(card.AsMap()))
	}
	return out
}

func sanitizeLangGraphCard(raw map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{
		"card_id",
		"card_type",
		"revision",
		"status",
		"risk_level",
		"requires_approval",
		"title",
		"subtitle",
		"fields",
		"actions",
		"updated_at",
	} {
		if value, ok := raw[key]; ok {
			out[key] = value
		}
	}
	return out
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		var out int
		if _, err := fmt.Sscanf(typed, "%d", &out); err == nil {
			return out
		}
	}
	return 0
}
