package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
)

func (a *AI) CreateSession(ctx context.Context, projectID, userID uuid.UUID, in SessionInput) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	employeeID, companyID, err := a.ensureEmployeeRow(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	sessionID := uuid.New()
	_, err = a.P.Exec(ctx, `
INSERT INTO ai_chat_sessions (
  ai_chat_session_id, project_ai_employee_id, company_id, project_id, screen_scope,
  screen_entity_type, screen_entity_id, workflow_id, workflow_run_id, task_id, client_id,
  action_dependency_id, model_id, created_by_user_id, page_context_json
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb)`,
		sessionID, employeeID, companyID, projectID, clean(in.ScreenScope), clean(in.ScreenEntityType), clean(in.ScreenEntityID),
		uuidOrNil(in.WorkflowID), uuidOrNil(in.WorkflowRunID), uuidOrNil(in.TaskID), uuidOrNil(in.ClientID),
		uuidOrNil(in.ActionDependency), clean(in.ModelID), userID, jsonDefault(in.PageContextJSON))
	if err != nil {
		return nil, err
	}
	return a.GetSession(ctx, projectID, sessionID, userID)
}

func (a *AI) GetSession(ctx context.Context, projectID, sessionID, userID uuid.UUID) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	var sid, companyID, status, scope, entityType, entityID, modelID, langgraphThreadID, agentRuntime string
	var sessionRevision int
	var activeToolCallID, activePatchID *string
	var createdAt, lastActivity string
	var pageContext, lastPageContext, uiState, pendingContext, workingMemory []byte
	err := a.P.QueryRow(ctx, `
SELECT ai_chat_session_id::text, company_id::text, status::text, screen_scope::text, screen_entity_type::text,
       screen_entity_id::text, model_id::text, created_at::text, last_activity_at::text,
       COALESCE(session_revision, 0),
       COALESCE(langgraph_thread_id, ''),
       COALESCE(agent_runtime, ''),
       COALESCE(page_context_json, '{}'::jsonb),
       COALESCE(last_page_context_json, '{}'::jsonb),
       COALESCE(ui_state_json, '{}'::jsonb),
       COALESCE(pending_context_json, '{}'::jsonb),
       COALESCE(working_memory_json, '{}'::jsonb),
       active_tool_call_id::text,
       active_patch_proposal_id::text
  FROM ai_chat_sessions
 WHERE ai_chat_session_id = $1
   AND project_id = $2`,
		sessionID, projectID).Scan(&sid, &companyID, &status, &scope, &entityType, &entityID, &modelID, &createdAt, &lastActivity, &sessionRevision, &langgraphThreadID, &agentRuntime, &pageContext, &lastPageContext, &uiState, &pendingContext, &workingMemory, &activeToolCallID, &activePatchID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	messages, err := a.messages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	toolCalls, err := a.toolCalls(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ai_chat_session_id":       sid,
		"company_id":               companyID,
		"project_id":               projectID.String(),
		"status":                   status,
		"agent_runtime":            agentRuntime,
		"langgraph_thread_id":      langgraphThreadID,
		"screen_scope":             scope,
		"screen_entity_type":       entityType,
		"screen_entity_id":         entityID,
		"model_id":                 modelID,
		"session_revision":         sessionRevision,
		"page_context_json":        json.RawMessage(pageContext),
		"last_page_context_json":   json.RawMessage(nonEmptyJSON(lastPageContext, pageContext)),
		"ui_state_json":            json.RawMessage(uiState),
		"pending_context_json":     json.RawMessage(pendingContext),
		"working_memory_json":      json.RawMessage(workingMemory),
		"active_tool_call_id":      stringPtrValue(activeToolCallID),
		"active_patch_proposal_id": stringPtrValue(activePatchID),
		"created_at":               createdAt,
		"last_activity_at":         lastActivity,
		"messages":                 messages,
		"tool_calls":               toolCalls,
	}, nil
}

func (a *AI) CancelSession(ctx context.Context, projectID, sessionID, userID uuid.UUID) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	tag, err := a.P.Exec(ctx, `
UPDATE ai_chat_sessions
   SET status = 'canceled',
       active_tool_call_id = NULL,
       active_patch_proposal_id = NULL,
       pending_context_json = '{}'::jsonb,
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1
   AND project_id = $2`, sessionID, projectID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return a.GetSession(ctx, projectID, sessionID, userID)
}

func (a *AI) messages(ctx context.Context, sessionID uuid.UUID) ([]map[string]any, error) {
	rows, err := a.P.Query(ctx, `
SELECT ai_chat_message_id::text, role::text, body::text, payload_json, redaction_json,
       token_usage_json, created_at::text
  FROM ai_chat_messages
 WHERE ai_chat_session_id = $1
 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, role, body, createdAt string
		var payload, redaction, usage []byte
		if err := rows.Scan(&id, &role, &body, &payload, &redaction, &usage, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"ai_chat_message_id": id,
			"role":               role,
			"body":               body,
			"payload_json":       json.RawMessage(payload),
			"redaction_json":     json.RawMessage(redaction),
			"token_usage_json":   json.RawMessage(usage),
			"created_at":         createdAt,
		})
	}
	return out, rows.Err()
}
