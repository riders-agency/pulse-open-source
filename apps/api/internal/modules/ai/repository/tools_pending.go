package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (a *AI) applyTaskCreateFollowUp(ctx context.Context, projectID, userID uuid.UUID, input map[string]any, body string, in MessageInput) (map[string]any, bool, string) {
	updated := copyMap(input)
	lower := strings.ToLower(body)
	changed := false
	if due := dueDateFromTextWithBase(body, in.LocalDate, in.Timezone); due != "" {
		updated["due_date"] = due
		changed = true
	}
	if strings.Contains(lower, "срочно") || strings.Contains(lower, "важн") || strings.Contains(lower, "высок") || strings.Contains(lower, "high") {
		updated["priority"] = "high"
		changed = true
	} else if strings.Contains(lower, "низк") || strings.Contains(lower, "low") {
		updated["priority"] = "low"
		changed = true
	} else if strings.Contains(lower, "средн") || strings.Contains(lower, "medium") {
		updated["priority"] = "medium"
		changed = true
	}
	if strings.Contains(lower, "не бил") || strings.Contains(lower, "небилл") {
		updated["billable"] = false
		changed = true
	}
	if strings.Contains(lower, "билл") && !strings.Contains(lower, "не бил") {
		updated["billable"] = true
		changed = true
	}
	if strings.Contains(lower, "мне") || strings.Contains(lower, "на меня") || strings.Contains(lower, "назначь меня") {
		updated["assignee_user_id"] = userID.String()
		updated["assignee_name"] = "мне"
		changed = true
	}
	if query := clientQueryFromText(body); query != "" && !isDateOnlyFollowUp(body) {
		matches := a.findClients(ctx, projectID, userID, query, 5)
		clientID, clientName := firstClientMatch(matches)
		updated["client_query"] = query
		updated["client_id"] = clientID
		updated["client_name"] = clientName
		updated["client_resolved"] = clientID != ""
		updated["client_search_snippet"] = clientSearchSnippet(query, matches)
		updated["entity_snippets"] = []map[string]any{clientSearchSnippet(query, matches)}
		changed = true
	}
	if !changed {
		return input, false, ""
	}
	replyParts := []string{"Обновил черновик задачи."}
	if due := stringValue(updated["due_date"]); due != "" {
		replyParts = append(replyParts, "Срок: "+dateLabel(due)+".")
	}
	if clientName := stringValue(updated["client_name"]); clientName != "" {
		replyParts = append(replyParts, "Клиент: "+clientName+".")
	} else if query := stringValue(updated["client_query"]); query != "" {
		replyParts = append(replyParts, "Клиент «"+query+"» пока не найден точно.")
	}
	if priority := stringValue(updated["priority"]); priority != "" {
		replyParts = append(replyParts, "Приоритет: "+priorityLabel(priority)+".")
	}
	replyParts = append(replyParts, "Проверьте карточку и нажмите «Применить».")
	return updated, true, strings.Join(replyParts, " ")
}

func priorityLabel(priority string) string {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "urgent":
		return "срочный"
	case "high":
		return "высокий"
	case "low":
		return "низкий"
	case "medium":
		return "средний"
	default:
		return priority
	}
}

func (a *AI) updateToolCallInputTx(ctx context.Context, tx pgx.Tx, projectID, sessionID, toolCallID, userID uuid.UUID, before, after map[string]any, reason string) error {
	var toolKey string
	_ = tx.QueryRow(ctx, `SELECT tool_key FROM ai_tool_calls WHERE ai_tool_call_id = $1 AND project_id = $2`, toolCallID, projectID).Scan(&toolKey)
	if _, err := tx.Exec(ctx, `
UPDATE ai_tool_calls
   SET input_json = $3::jsonb
 WHERE ai_tool_call_id = $1
   AND project_id = $2
   AND status = 'pending_approval'`, toolCallID, projectID, mustJSON(after)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO ai_tool_call_edits (
  ai_tool_call_id, ai_chat_session_id, project_id, edited_by_user_id,
  edit_reason, before_json, after_json
) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb)`,
		toolCallID, sessionID, projectID, userID, reason, mustJSON(before), mustJSON(after)); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET active_tool_call_id = $2::uuid,
       pending_context_json = jsonb_build_object('active_tool_call_id', ($2::uuid)::text, 'tool_key', $3::text, 'input_json', $4::jsonb),
       status = 'waiting_tool_approval',
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, toolCallID, toolKey, mustJSON(after))
	return err
}

func (a *AI) toolCalls(ctx context.Context, sessionID uuid.UUID) ([]map[string]any, error) {
	rows, err := a.P.Query(ctx, `
SELECT ai_tool_call_id::text, tool_key, risk_level, status, input_json, output_json,
       error_text, requires_approval, created_at::text, COALESCE(finished_at::text, '')
  FROM ai_tool_calls
 WHERE ai_chat_session_id = $1
 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, toolKey, risk, status, errorText, createdAt, finishedAt string
		var input, output []byte
		var requires bool
		if err := rows.Scan(&id, &toolKey, &risk, &status, &input, &output, &errorText, &requires, &createdAt, &finishedAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"ai_tool_call_id":   id,
			"tool_key":          toolKey,
			"risk_level":        risk,
			"status":            status,
			"input_json":        json.RawMessage(input),
			"output_json":       json.RawMessage(output),
			"error_text":        errorText,
			"requires_approval": requires,
			"created_at":        createdAt,
			"finished_at":       finishedAt,
		})
	}
	return out, rows.Err()
}

func (a *AI) createToolCall(ctx context.Context, tx pgx.Tx, sessionID, projectID uuid.UUID, toolKey, riskLevel string, input map[string]any, requiresApproval bool) (string, error) {
	callID := uuid.New()
	riskLevel = normalizeAIRiskLevel(riskLevel)
	_, err := tx.Exec(ctx, `
INSERT INTO ai_tool_calls (
  ai_tool_call_id, ai_chat_session_id, project_id, tool_key, risk_level,
  status, input_json, requires_approval, idempotency_key
) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9)`,
		callID, sessionID, projectID, toolKey, riskLevel, "pending_approval", mustJSON(input), requiresApproval,
		fmt.Sprintf("%s:%s:%d", sessionID.String(), toolKey, time.Now().UnixNano()))
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET active_tool_call_id = $2::uuid,
       status = 'waiting_tool_approval',
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, callID); err != nil {
		return "", err
	}
	return callID.String(), nil
}
