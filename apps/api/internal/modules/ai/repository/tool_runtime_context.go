package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

type ToolCallRuntimeContext struct {
	CompanyID         string
	ProjectID         uuid.UUID
	SessionID         uuid.UUID
	ToolCallID        uuid.UUID
	AgentRuntime      string
	LangGraphThreadID string
	Status            string
}

type SessionRuntimeContext struct {
	CompanyID         string
	ProjectID         uuid.UUID
	SessionID         uuid.UUID
	AgentRuntime      string
	LangGraphThreadID string
	Status            string
}

func (a *AI) SessionRuntimeContext(ctx context.Context, projectID, sessionID uuid.UUID) (SessionRuntimeContext, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return SessionRuntimeContext{}, err
	}
	var out SessionRuntimeContext
	out.ProjectID = projectID
	out.SessionID = sessionID
	err := a.P.QueryRow(ctx, `
SELECT company_id::text,
       COALESCE(agent_runtime, ''),
       COALESCE(langgraph_thread_id, ''),
       status
  FROM ai_chat_sessions
 WHERE ai_chat_session_id = $1
   AND project_id = $2`, sessionID, projectID).Scan(
		&out.CompanyID,
		&out.AgentRuntime,
		&out.LangGraphThreadID,
		&out.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SessionRuntimeContext{}, ErrNotFound
		}
		return SessionRuntimeContext{}, err
	}
	return out, nil
}

func (a *AI) ToolCallRuntimeContext(ctx context.Context, projectID, toolCallID uuid.UUID) (ToolCallRuntimeContext, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return ToolCallRuntimeContext{}, err
	}
	var out ToolCallRuntimeContext
	out.ProjectID = projectID
	out.ToolCallID = toolCallID
	err := a.P.QueryRow(ctx, `
SELECT s.company_id::text,
       t.ai_chat_session_id,
       COALESCE(s.agent_runtime, ''),
       COALESCE(s.langgraph_thread_id, ''),
       t.status
  FROM ai_tool_calls t
  JOIN ai_chat_sessions s ON s.ai_chat_session_id = t.ai_chat_session_id
 WHERE t.ai_tool_call_id = $1
   AND t.project_id = $2
   AND s.project_id = $2`, toolCallID, projectID).Scan(
		&out.CompanyID,
		&out.SessionID,
		&out.AgentRuntime,
		&out.LangGraphThreadID,
		&out.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ToolCallRuntimeContext{}, ErrNotFound
		}
		return ToolCallRuntimeContext{}, err
	}
	return out, nil
}
