package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ApprovalProposal struct {
	ToolKey   string
	SessionID uuid.UUID
	Status    string
	Input     map[string]any
}

func (a *AI) LockApproval(ctx context.Context, tx pgx.Tx, projectID, toolCallID, userID uuid.UUID) (ApprovalProposal, error) {
	var proposal ApprovalProposal
	var raw []byte
	err := tx.QueryRow(ctx, `SELECT c.tool_key, c.ai_chat_session_id, c.status, c.input_json
 FROM ai_tool_calls c JOIN ai_chat_sessions s ON s.ai_chat_session_id = c.ai_chat_session_id
 WHERE c.ai_tool_call_id=$1 AND c.project_id=$2 AND s.created_by_user_id=$3
 FOR UPDATE OF c`, toolCallID, projectID, userID).Scan(&proposal.ToolKey, &proposal.SessionID, &proposal.Status, &raw)
	if err == pgx.ErrNoRows {
		return proposal, ErrNotFound
	}
	if err != nil {
		return proposal, err
	}
	err = json.Unmarshal(raw, &proposal.Input)
	return proposal, err
}
