package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	"riders.agency/pulse/api/internal/modules/ai/execution"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

func streamLangGraphTurn(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, ai *aisvc.Commands, projectID, sessionID, userID uuid.UUID, in airepo.MessageInput) (map[string]any, error) {
	in.ReceivedAt = time.Now().UTC()
	runKey := uuid.NewString()
	turnContext, err := loadLangGraphTurnContext(ctx, ai, projectID, sessionID, userID)
	if err != nil {
		return nil, err
	}
	capability := evaluateAICapability(ctx, capabilityOptions{CheckPythonGRPC: true})
	rollout := evaluateAIRolloutPolicy(turnContext.CompanyID, projectID, in, capability)
	if err := rollout.ValidateExecution(in); err != nil {
		return nil, err
	}

	addr := aiAgentGRPCTarget()
	dialOptions, _, err := aiAgentGRPCDialOptions(addr)
	if err != nil {
		return nil, err
	}
	dialCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	conn, err := connectAgent(dialCtx, addr, dialOptions)
	if err != nil {
		return nil, fmt.Errorf("python LangGraph unavailable: %w", err)
	}
	defer conn.Close()

	client := aiagentv1.NewAgentRuntimeServiceClient(conn)
	stream, err := client.StartTurn(aiAgentOutgoingContext(ctx), &aiagentv1.TurnRequest{
		RequestId:         runKey,
		IdempotencyKey:    projectID.String() + ":" + sessionID.String() + ":" + runKey,
		CompanyId:         turnContext.CompanyID,
		ProjectId:         projectID.String(),
		UserId:            userID.String(),
		SessionId:         sessionID.String(),
		ThreadId:          turnContext.ThreadID,
		Body:              strings.TrimSpace(in.Body),
		LocalDate:         strings.TrimSpace(in.LocalDate),
		Timezone:          strings.TrimSpace(in.Timezone),
		Locale:            "ru-RU",
		PageContextJson:   rawStruct(in.PageContextJSON),
		ClientContextJson: rawStruct(in.ClientContextJSON),
		UiStateJson:       rawStruct(in.UIStateJSON),
		ActiveToolCallId:  strings.TrimSpace(in.ActiveToolCallID),
		ActiveProposalId:  strings.TrimSpace(in.ActivePatchProposalID),
		History:           turnContext.History,
		AuthContext: &aiagentv1.AuthContext{
			CompanyId: turnContext.CompanyID,
			ProjectId: projectID.String(),
			UserId:    userID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("langgraph start turn: %w", err)
	}

	finalBody := ""
	payload := map[string]any{
		"kind":       "answer",
		"risk_level": "read",
		"runtime":    "langgraph",
	}
	payload = attachAIRolloutPayload(payload, rollout)
	readResults := make([]map[string]any, 0)
	assistantTextStreamed := false
	runtimeErr := ""
	var trace *execution.Trace
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			if runtimeErr != "" {
				return nil, fmt.Errorf("%s", runtimeErr)
			}
			if finalBody == "" {
				finalBody = "Готово."
			}
			payload = attachAIRolloutPayload(aggregateReadResultsPayload(payload, readResults), rollout)
			if trace != nil {
				payload["execution"] = trace
				payload["token_usage_json"] = trace.Usage()
			}
			out, persistErr := ai.AddLangGraphTurn(ctx, projectID, sessionID, userID, in, finalBody, payload)
			if persistErr != nil {
				return nil, persistErr
			}
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("langgraph stream: %w", err)
		}
		if evidence := execution.Parse(event.GetToolCall().AsMap()["execution"]); evidence != nil {
			trace = evidence
			writeSSE(w, flusher, "trace", map[string]any{"type": "EXECUTION_TRACE", "execution": trace})
		}
		switch event.GetEventType() {
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_RUN_STARTED, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_STATUS:
			writeSSE(w, flusher, "status", map[string]any{
				"type":    "STATUS",
				"status":  strings.TrimSpace(event.GetMessage()),
				"runtime": "langgraph",
				"seq":     event.GetSeq(),
			})
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_CARD_CREATED, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_CARD_UPDATED, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_CARD_PATCHED:
			card := event.GetCard().AsMap()
			writeSSE(w, flusher, "card", map[string]any{
				"type":    event.GetEventType().String(),
				"message": strings.TrimSpace(event.GetMessage()),
				"runtime": "langgraph",
				"seq":     event.GetSeq(),
				"card":    card,
			})
			if cardPayload, ok := proposalPayloadFromCard(card); ok {
				payload = cardPayload
				finalBody = firstNonEmpty(strings.TrimSpace(event.GetMessage()), "Подготовил карточку. Проверьте поля и подтвердите действие.")
				writeSSE(w, flusher, "approval", map[string]any{
					"type":      "APPROVAL_REQUIRED",
					"message":   finalBody,
					"runtime":   "langgraph",
					"seq":       event.GetSeq(),
					"tool_key":  cardPayload["tool_key"],
					"card_type": cardPayload["card_type"],
					"card":      card,
				})
			} else if stringFromAny(card["card_type"]) == "conflict_choice" {
				payload = map[string]any{
					"kind":           "card",
					"card_type":      "conflict_choice",
					"risk_level":     string(aitools.RiskHigh),
					"card_json":      card,
					"developer_json": map[string]any{"langgraph_card": card, "hidden_by_default": true},
				}
				finalBody = firstNonEmpty(strings.TrimSpace(event.GetMessage()), "Сначала завершите или отмените текущую карточку.")
			}
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_PROPOSED, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_STARTED, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_RESULT:
			toolCall := event.GetToolCall().AsMap()
			writeSSE(w, flusher, "tool", map[string]any{
				"type":      event.GetEventType().String(),
				"message":   strings.TrimSpace(event.GetMessage()),
				"runtime":   "langgraph",
				"seq":       event.GetSeq(),
				"tool_call": toolCall,
			})
			if event.GetEventType() == aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_RESULT {
				mapped := mapLangGraphToolResult(toolCall, event.GetMessage(), finalBody)
				finalBody = mapped.FinalBody
				if mapped.Payload != nil {
					payload = mapped.Payload
				}
				if mapped.ReadResult != nil {
					readResults = append(readResults, mapped.ReadResult)
				}
				if mapped.StreamText && !assistantTextStreamed {
					streamAssistantText(w, flusher, finalBody)
					assistantTextStreamed = true
				}
			}
			if event.GetEventType() == aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_STARTED && stringFromAny(toolCall["tool_key"]) == "proposal.edit" {
				finalBody = firstNonEmpty(strings.TrimSpace(event.GetMessage()), "Обновляю текущую карточку.")
				payload = map[string]any{
					"kind":       "proposal_edit",
					"tool_key":   "proposal.edit",
					"risk_level": string(aitools.RiskMedium),
					"runtime":    "langgraph",
					"tool_call":  toolCall,
				}
				writeSSE(w, flusher, "approval", map[string]any{
					"type":      "APPROVAL_REQUIRED",
					"message":   finalBody,
					"runtime":   "langgraph",
					"seq":       event.GetSeq(),
					"tool_key":  "proposal.edit",
					"tool_call": toolCall,
				})
			}
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_APPROVAL_REQUIRED, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_INTERRUPT:
			writeSSE(w, flusher, "status", map[string]any{
				"type":      event.GetEventType().String(),
				"status":    strings.TrimSpace(event.GetMessage()),
				"runtime":   "langgraph",
				"seq":       event.GetSeq(),
				"interrupt": event.GetInterrupt().AsMap(),
			})
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_MESSAGE_DELTA:
			writeSSE(w, flusher, "delta", map[string]any{
				"type":    "MESSAGE_DELTA",
				"text":    firstNonEmpty(event.GetBodyDelta(), event.GetMessage()),
				"runtime": "langgraph",
				"seq":     event.GetSeq(),
			})
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_MESSAGE_FINAL:
			finalBody = firstNonEmpty(event.GetMessage(), event.GetBodyDelta(), finalBody)
			if finalBody != "" && !assistantTextStreamed {
				streamAssistantText(w, flusher, finalBody)
				assistantTextStreamed = true
			}
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_ERROR:
			errPayload := event.GetError().AsMap()
			errCode := strings.TrimSpace(stringFromAny(errPayload["code"]))
			errMessage := firstNonEmpty(strings.TrimSpace(stringFromAny(errPayload["message"])), strings.TrimSpace(event.GetMessage()), "LangGraph runtime error")
			runtimeErr = errMessage
			if errCode != "" {
				runtimeErr = errCode + ": " + errMessage
			}
			writeSSE(w, flusher, "error", map[string]any{
				"type":    "ERROR",
				"message": errMessage,
				"code":    errCode,
				"runtime": "langgraph",
				"seq":     event.GetSeq(),
			})
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_MEMORY_UPDATED:
			writeSSE(w, flusher, "memory", map[string]any{
				"type":    "MEMORY_UPDATED",
				"message": strings.TrimSpace(event.GetMessage()),
				"runtime": "langgraph",
				"seq":     event.GetSeq(),
				"debug":   event.GetDebug().AsMap(),
			})
		case aiagentv1.AgentEventType_AGENT_EVENT_TYPE_DONE:
			continue
		default:
			writeSSE(w, flusher, "status", map[string]any{
				"type":       "UNKNOWN_AGENT_EVENT",
				"message":    strings.TrimSpace(event.GetMessage()),
				"runtime":    "langgraph",
				"seq":        event.GetSeq(),
				"event_type": int32(event.GetEventType()),
			})
		}
	}
}
