package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/domainevent"
)

func (a *AI) ExecuteApprovedToolTx(ctx context.Context, projectID, userID uuid.UUID, toolKey string, input map[string]any) (map[string]any, error) {
	var err error
	var output map[string]any
	switch toolKey {
	case "task.create":
		output, err = a.executeTaskCreate(ctx, projectID, userID, input)
	case "task.move":
		output, err = a.executeTaskMove(ctx, projectID, userID, input)
	case "client.create":
		output, err = a.executeClientCreate(ctx, projectID, userID, input)
	case "workflow.patch.apply":
		output, err = a.executeWorkflowPatch(ctx, projectID, userID, input)
	case "task_process.patch.apply":
		output, err = a.executeTaskProcessPatch(ctx, projectID, userID, input)
	default:
		err = fmt.Errorf("unsupported tool %s", toolKey)
	}
	return output, err
}

func (a *AI) RecordApprovalTx(ctx context.Context, tx pgx.Tx, projectID, toolCallID, userID uuid.UUID, proposal ApprovalProposal, output map[string]any, executionErr error) (map[string]any, error) {
	toolKey, sessionID, input := proposal.ToolKey, proposal.SessionID, proposal.Input
	status := "succeeded"
	errorText := ""
	if executionErr != nil {
		status = "failed"
		errorText = executionErr.Error()
		output = map[string]any{"ok": false, "error": "tool_failed"}
	}
	var txErr error
	if _, txErr = tx.Exec(ctx, `
UPDATE ai_tool_calls
   SET status = $2, output_json = $3::jsonb, error_text = $4,
       approved_by_user_id = $5, approved_at = COALESCE(approved_at, now()),
       finished_at = now()
 WHERE ai_tool_call_id = $1`, toolCallID, status, mustJSON(output), errorText, userID); txErr != nil {
		return nil, txErr
	}
	if txErr = a.enqueueAIToolEventTx(ctx, tx, "ai.tool.approved", projectID, sessionID, toolCallID, userID, toolKey, "approved", input, output, ""); txErr != nil {
		return nil, txErr
	}
	followupEvent := "ai.tool.applied"
	if status == "failed" {
		followupEvent = "ai.tool.failed"
	}
	if txErr = a.enqueueAIToolEventTx(ctx, tx, followupEvent, projectID, sessionID, toolCallID, userID, toolKey, status, input, output, errorText); txErr != nil {
		return nil, txErr
	}
	body := "Действие выполнено."
	if executionErr != nil {
		body = "Действие не выполнено: " + errorText
	} else if msg, ok := output["message"].(string); ok && msg != "" {
		body = msg
	}
	if _, txErr = tx.Exec(ctx, `
INSERT INTO ai_chat_messages (ai_chat_session_id, role, body, payload_json)
VALUES ($1, 'tool', $2, $3::jsonb)`, sessionID, body, mustJSON(map[string]any{"kind": "tool_result", "tool_key": toolKey, "tool_call_id": toolCallID.String(), "output": output})); txErr != nil {
		return nil, txErr
	}
	if _, txErr = tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET status = CASE WHEN $2 = 'succeeded' THEN 'completed' ELSE 'failed' END,
       active_tool_call_id = CASE WHEN active_tool_call_id = $3::uuid THEN NULL ELSE active_tool_call_id END,
       pending_context_json = '{}'::jsonb,
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, status, toolCallID); txErr != nil {
		return nil, txErr
	}
	if executionErr == nil {
		a.rememberToolResult(ctx, sessionID, toolKey, output)
	}
	return a.GetSession(ctx, projectID, sessionID, userID)
}

func (a *AI) UpdateToolCallTx(ctx context.Context, projectID, toolCallID, userID uuid.UUID, input map[string]any, reason string) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermWrite); err != nil {
		return nil, err
	}
	tx, err := a.transaction()
	if err != nil {
		return nil, err
	}
	var sessionID uuid.UUID
	var beforeRaw []byte
	var toolKey string
	err = tx.QueryRow(ctx, `
SELECT ai_chat_session_id, tool_key, input_json
  FROM ai_tool_calls
 WHERE ai_tool_call_id = $1
   AND project_id = $2
   AND status = 'pending_approval'
 AND EXISTS(SELECT 1 FROM ai_chat_sessions s WHERE s.ai_chat_session_id=ai_tool_calls.ai_chat_session_id AND s.created_by_user_id=$3)
 FOR UPDATE`, toolCallID, projectID, userID).Scan(&sessionID, &toolKey, &beforeRaw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var before map[string]any
	_ = json.Unmarshal(beforeRaw, &before)
	if before == nil {
		before = map[string]any{}
	}
	after := copyMap(before)
	for key, value := range input {
		after[key] = value
	}
	if toolKey == "task.create" {
		if due := stringValue(after["due_date"]); due != "" {
			after["due_date"] = due
		}
		if query := stringValue(after["client_query"]); query != "" && stringValue(after["client_id"]) == "" {
			matches := a.findClients(ctx, projectID, userID, query, 5)
			clientID, clientName := firstClientMatch(matches)
			after["client_id"] = clientID
			after["client_name"] = clientName
			after["client_resolved"] = clientID != ""
			after["client_search_snippet"] = clientSearchSnippet(query, matches)
			after["entity_snippets"] = []map[string]any{clientSearchSnippet(query, matches)}
		}
	}
	if err := a.updateToolCallInputTx(ctx, tx, projectID, sessionID, toolCallID, userID, before, after, defaultString(reason, "user edited proposal card")); err != nil {
		return nil, err
	}
	return a.GetSession(ctx, projectID, sessionID, userID)
}

func (a *AI) RejectToolCallTx(ctx context.Context, projectID, toolCallID, userID uuid.UUID, reason string) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermWrite); err != nil {
		return nil, err
	}
	var sessionID uuid.UUID
	var toolKey string
	var rawInput []byte
	tx, err := a.transaction()
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx, `
UPDATE ai_tool_calls
   SET status = 'rejected', rejected_by_user_id = $3, rejected_at = now(), approval_reason = $4, finished_at = now()
 WHERE ai_tool_call_id = $1
   AND project_id = $2
   AND status = 'pending_approval'
 RETURNING ai_chat_session_id, tool_key, input_json`, toolCallID, projectID, userID, strings.TrimSpace(reason)).Scan(&sessionID, &toolKey, &rawInput)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var input map[string]any
	_ = json.Unmarshal(rawInput, &input)
	if input == nil {
		input = map[string]any{}
	}
	if err := a.enqueueAIToolEventTx(ctx, tx, "ai.tool.rejected", projectID, sessionID, toolCallID, userID, toolKey, "rejected", input, map[string]any{"ok": false, "rejected": true}, strings.TrimSpace(reason)); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO ai_chat_messages (ai_chat_session_id, role, body, payload_json)
VALUES ($1, 'approval', $2, $3::jsonb)`, sessionID, "Действие отменено. Черновик отклонен, данные не менялись.", mustJSON(map[string]any{"kind": "rejected", "reason": reason})); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET status = 'completed',
       active_tool_call_id = CASE WHEN active_tool_call_id = $2::uuid THEN NULL ELSE active_tool_call_id END,
       pending_context_json = '{}'::jsonb,
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, toolCallID); err != nil {
		return nil, err
	}
	return a.GetSession(ctx, projectID, sessionID, userID)
}

func (a *AI) enqueueAIToolEventTx(
	ctx context.Context,
	tx pgx.Tx,
	typeKey string,
	projectID, sessionID, toolCallID, userID uuid.UUID,
	toolKey, status string,
	input map[string]any,
	output map[string]any,
	errorText string,
) error {
	payload := map[string]any{
		"project_id":         projectID.String(),
		"ai_chat_session_id": sessionID.String(),
		"ai_tool_call_id":    toolCallID.String(),
		"user_id":            userID.String(),
		"tool_key":           strings.TrimSpace(toolKey),
		"status":             strings.TrimSpace(status),
		"input_json":         input,
		"output_json":        output,
		"error":              strings.TrimSpace(errorText),
	}
	err := domainevent.Record(ctx, domainevent.Event{
		TypeKey:        typeKey,
		AggregateType:  "ai_tool_call",
		AggregateID:    toolCallID.String(),
		IdempotencyKey: typeKey + ":" + toolCallID.String() + ":" + strings.TrimSpace(status),
		PayloadJSON:    json.RawMessage(mustJSON(payload)),
	})
	return err
}
