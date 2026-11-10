package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func executeBotSetState(ctx context.Context, store botSessionStore, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	tpl := templateSystemForRun(runID)
	stateKey := strings.TrimSpace(renderTemplate(node.Config.StateKey, runInput, steps, tpl))
	contextText := strings.TrimSpace(renderTemplate(node.Config.ContextJSON, runInput, steps, tpl))
	if contextText == "" {
		contextText = "{}"
	}
	contextPatch := map[string]any{}
	if err := json.Unmarshal([]byte(contextText), &contextPatch); err != nil {
		err := fmt.Errorf("bot.set_state context_json must be a JSON object: %w", err)
		return failedAction(actionInput(node, runInput), "{}", map[string]any{"provider": "internal", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	if stateKey == "" {
		err := errors.New("bot.set_state state_key is required")
		return failedAction(actionInput(node, runInput), "{}", map[string]any{"provider": "internal", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	var root map[string]any
	_ = json.Unmarshal(runInput, &root)
	sessionID := stringAtRunPath(root, "bot_session.bot_session_id")
	depID := stringAtRunPath(root, "_pulse.action_dependency_id")
	relayKey := firstNonEmpty(stringAtRunPath(root, "bot_session.relay_key"), stringAtRunPath(root, "_pulse.telegram_relay_key"))
	chatID := firstNonEmpty(stringAtRunPath(root, "bot_session.chat_id"), stringAtRunPath(root, "telegram_relay.chat_id"), stringAtRunPath(root, "telegram_relay.chat.id"))
	userID := firstNonEmpty(stringAtRunPath(root, "bot_session.user_id"), stringAtRunPath(root, "telegram_relay.user_id"), stringAtRunPath(root, "telegram_relay.from.id"))
	if sessionID == "" && (depID == "" || relayKey == "" || chatID == "") {
		err := errors.New("bot.set_state requires a telegram relay bot_session")
		return failedAction(actionInput(node, runInput), "{}", map[string]any{"provider": "internal", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	contextJSON, _ := json.Marshal(contextPatch)
	requestInput, _ := json.Marshal(map[string]any{
		"state_key":      stateKey,
		"context_json":   contextPatch,
		"merge_context":  node.Config.MergeContext,
		"bot_session_id": sessionID,
		"relay_key":      relayKey,
		"chat_id":        chatID,
		"user_id":        userID,
	})
	var updatedContext []byte
	var updatedSessionID string
	var err error
	if sessionID != "" {
		err = store.QueryRow(ctx, `
UPDATE bot_sessions
   SET state_key = $2,
       context_json = CASE WHEN $3 THEN context_json || $4::jsonb ELSE $4::jsonb END,
       last_workflow_run_id = $5,
       updated_at = now()
 WHERE bot_session_id = $1
RETURNING bot_session_id::text, context_json`,
			sessionID, stateKey, node.Config.MergeContext, string(contextJSON), nullableUUID(runID)).
			Scan(&updatedSessionID, &updatedContext)
	} else {
		err = store.QueryRow(ctx, `
UPDATE bot_sessions
   SET state_key = $5,
       context_json = CASE WHEN $6 THEN context_json || $7::jsonb ELSE $7::jsonb END,
       last_workflow_run_id = $8,
       updated_at = now()
 WHERE action_dependency_id = $1
   AND relay_key = $2
   AND chat_id = $3
   AND user_id = $4
RETURNING bot_session_id::text, context_json`,
			depID, relayKey, chatID, userID, stateKey, node.Config.MergeContext, string(contextJSON), nullableUUID(runID)).
			Scan(&updatedSessionID, &updatedContext)
	}
	if err != nil {
		err := fmt.Errorf("bot.set_state session update failed: %w", err)
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "internal", "ok": false, "error": "session update failed"}, false, "storage", err)
	}
	updated := map[string]any{}
	_ = json.Unmarshal(updatedContext, &updated)
	output := map[string]any{
		"action_key":      node.ActionKey,
		"action_type_key": node.ActionTypeKey,
		"provider":        "internal",
		"bot_session_id":  updatedSessionID,
		"state_key":       stateKey,
		"context_json":    updated,
		"expose": map[string]any{
			"bot_session_id": updatedSessionID,
			"state_key":      stateKey,
			"context_json":   updated,
		},
	}
	outputJSON, _ := json.Marshal(output)
	providerJSON, _ := json.Marshal(map[string]any{
		"provider":                   "internal",
		"ok":                         true,
		"provider_response_redacted": true,
	})
	return actionExecution{
		InputJSON:            string(requestInput),
		OutputJSON:           string(outputJSON),
		ProviderResponseJSON: string(providerJSON),
	}
}
