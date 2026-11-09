package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/platform/authz"
)

func (a *AI) AddLangGraphTurnTx(ctx context.Context, projectID, sessionID, userID uuid.UUID, in MessageInput, reply string, payload map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return nil, fmt.Errorf("body required")
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		reply = "Готово."
	}
	if payload == nil {
		payload = map[string]any{"kind": "answer", "risk_level": "read", "runtime": "langgraph"}
	}
	var exists bool
	if err := a.P.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM ai_chat_sessions
   WHERE ai_chat_session_id = $1
     AND project_id = $2
)`, sessionID, projectID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	tx, err := a.transaction()
	if err != nil {
		return nil, err
	}

	pageContextJSON := jsonDefault(in.PageContextJSON)
	clientContextJSON := jsonDefault(in.ClientContextJSON)
	uiStateJSON := jsonDefault(in.UIStateJSON)
	userPayload := map[string]any{
		"kind":                     "user_message",
		"page_context":             json.RawMessage(pageContextJSON),
		"client_context":           json.RawMessage(clientContextJSON),
		"ui_state":                 json.RawMessage(uiStateJSON),
		"active_tool_call_id":      clean(in.ActiveToolCallID),
		"active_patch_proposal_id": clean(in.ActivePatchProposalID),
		"reply_to_message_id":      clean(in.ReplyToMessageID),
		"client_command":           clean(in.ClientCommand),
		"local_date":               clean(in.LocalDate),
		"timezone":                 clean(in.Timezone),
		"source_surface":           "widget",
		"runtime":                  "langgraph",
	}
	userPayloadJSON, _ := json.Marshal(userPayload)
	receivedAt := in.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO ai_chat_messages (ai_chat_session_id, role, body, payload_json, created_at)
VALUES ($1, 'user', $2, $3::jsonb, $4)`, sessionID, body, userPayloadJSON, receivedAt); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO ai_session_context_frames (
  ai_chat_session_id, project_id, page_context_json, client_context_json, ui_state_json,
  active_tool_call_id, active_patch_proposal_id, local_date, timezone
) VALUES ($1,$2,$3::jsonb,$4::jsonb,$5::jsonb,$6::uuid,$7::uuid,$8,$9)`,
		sessionID, projectID, pageContextJSON, clientContextJSON, uiStateJSON, uuidOrNil(in.ActiveToolCallID), uuidOrNil(in.ActivePatchProposalID), clean(in.LocalDate), clean(in.Timezone)); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET page_context_json = CASE WHEN $2::jsonb = '{}'::jsonb THEN page_context_json ELSE $2::jsonb END,
       last_page_context_json = CASE WHEN $2::jsonb = '{}'::jsonb THEN last_page_context_json ELSE $2::jsonb END,
       ui_state_json = CASE WHEN $3::jsonb = '{}'::jsonb THEN ui_state_json ELSE $3::jsonb END,
       agent_runtime = 'langgraph',
       langgraph_thread_id = CASE
         WHEN COALESCE(langgraph_thread_id, '') = '' THEN 'pulse:' || project_id::text || ':' || ai_chat_session_id::text
         ELSE langgraph_thread_id
       END,
       session_revision = session_revision + 1,
       pending_context_json = jsonb_strip_nulls(jsonb_build_object(
         'active_tool_call_id', NULLIF($4, ''),
         'active_patch_proposal_id', NULLIF($5, ''),
         'reply_to_message_id', NULLIF($6, ''),
         'client_command', NULLIF($7, ''),
         'local_date', NULLIF($8, ''),
         'timezone', NULLIF($9, ''),
         'client_context', $10::jsonb,
         'runtime', 'langgraph'
       )),
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, pageContextJSON, uiStateJSON, clean(in.ActiveToolCallID), clean(in.ActivePatchProposalID), clean(in.ReplyToMessageID), clean(in.ClientCommand), clean(in.LocalDate), clean(in.Timezone), clientContextJSON); err != nil {
		return nil, err
	}

	status := "completed"
	if stringValue(payload["kind"]) == "proposal_edit" {
		state, err := a.loadSessionState(ctx, tx, projectID, sessionID)
		if err != nil {
			return nil, err
		}
		edit, ok := a.ResolveProposalEdit(ctx, projectID, userID, in, state)
		if !ok {
			payload = map[string]any{
				"kind":       "answer",
				"risk_level": "read",
				"runtime":    "langgraph",
				"plan_json": map[string]any{
					"steps": []map[string]any{
						{"label": "Проверить активную карточку", "tool": "proposal_editor"},
						{"label": "Не найдено уверенное изменение", "tool": "proposal_editor"},
					},
				},
			}
			reply = "Не увидел точное изменение для текущей карточки. Напишите, какое поле обновить: срок, клиент, приоритет или название."
		} else {
			updatedReply, updatedPayload, err := a.ApplyProposalEdit(ctx, tx, projectID, sessionID, userID, edit)
			if err != nil {
				return nil, err
			}
			reply = updatedReply
			payload = updatedPayload
			payload["runtime"] = "langgraph"
			payload["plan_json"] = map[string]any{
				"steps": []map[string]any{
					{"label": "Найти активную pending карточку", "tool": "proposal_editor"},
					{"label": "Обновить тот же task.create proposal", "tool": "proposal_editor"},
					{"label": "Оставить применение за пользователем", "tool": "human_approval"},
				},
			}
			status = "waiting_tool_approval"
		}
	}
	if stringValue(payload["kind"]) == "proposal" && stringValue(payload["ai_tool_call_id"]) == "" {
		toolKey := stringValue(payload["tool_key"])
		if toolKey == "" {
			toolKey = "task.create"
		}
		riskLevel := normalizeAIRiskLevel(firstNonEmptyString(stringValue(payload["risk_level"]), "high"))
		payload["risk_level"] = riskLevel
		toolInput := mapFromAny(payload["input_json"])
		if len(toolInput) == 0 {
			toolInput = mapFromAny(payload["fields_json"])
		}
		if toolKey == "task.create" {
			if query := stringValue(toolInput["client_query"]); query != "" {
				matches := a.findClients(ctx, projectID, userID, query, 5)
				clientID, clientName := firstClientMatch(matches)
				if clientID != "" {
					toolInput["client_id"] = clientID
					toolInput["client_name"] = clientName
					toolInput["client_resolved"] = true
				}
				snippet := clientSearchSnippet(query, matches)
				toolInput["client_search_snippet"] = snippet
				toolInput["entity_snippets"] = []map[string]any{snippet}
				payload["input_json"] = toolInput
				payload["fields_json"] = toolInput
				display := mapFromAny(payload["display_json"])
				if clientName != "" {
					display["client_name"] = clientName
				}
				display["entity_snippets"] = []map[string]any{snippet}
				payload["display_json"] = display
				payload["entity_snippets"] = []map[string]any{snippet}
			}
		}
		callID, err := a.createToolCall(ctx, tx, sessionID, projectID, toolKey, riskLevel, toolInput, true)
		if err != nil {
			return nil, err
		}
		payload["ai_tool_call_id"] = callID
		payload["next_tool"] = toolKey
		payload["requires_approval"] = true
		status = "waiting_tool_approval"
	}

	payload["runtime"] = "langgraph"
	payloadJSON, _ := json.Marshal(payload)
	if _, err := tx.Exec(ctx, `
INSERT INTO ai_chat_messages (ai_chat_session_id, role, body, payload_json, token_usage_json, created_at)
VALUES ($1, 'assistant', $2, $3::jsonb, $4::jsonb, clock_timestamp())`, sessionID, reply, payloadJSON, mustJSON(mapFromAny(payload["token_usage_json"]))); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET status = $2,
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, status); err != nil {
		return nil, err
	}
	return a.GetSession(ctx, projectID, sessionID, userID)
}
