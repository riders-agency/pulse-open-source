package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/ai/execution"
)

func (a *AI) SaveExecutionTrace(ctx context.Context, projectID, sessionID, userID, callID uuid.UUID, trace *execution.Trace) error {
	if trace == nil {
		return nil
	}
	if _, err := a.GetSession(ctx, projectID, sessionID, userID); err != nil {
		return err
	}
	raw, err := json.Marshal(trace)
	if err != nil {
		return err
	}
	usage, err := json.Marshal(trace.Usage())
	if err != nil {
		return err
	}
	_, err = a.P.Exec(ctx, `
UPDATE ai_chat_messages SET payload_json = jsonb_set(payload_json, '{execution}', $3::jsonb), token_usage_json=$4::jsonb
WHERE ai_chat_session_id=$1 AND role='assistant' AND payload_json->>'ai_tool_call_id'=$2`, sessionID, callID.String(), raw, usage)
	return err
}
