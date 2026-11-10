package repository

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionState struct {
	SessionID             uuid.UUID
	ProjectID             uuid.UUID
	ActiveToolCallID      uuid.UUID
	ActivePatchProposalID uuid.UUID
	ActiveToolCall        *ToolCallState
	WorkingMemory         map[string]any
	LastPageContext       json.RawMessage
}

type ToolCallState struct {
	ID        uuid.UUID
	ToolKey   string
	RiskLevel string
	Status    string
	Input     map[string]any
}

type ProposalEdit struct {
	ToolCallID   uuid.UUID
	ToolKey      string
	Patch        map[string]any
	HumanSummary string
}

func (a *AI) loadSessionState(ctx context.Context, tx pgx.Tx, projectID, sessionID uuid.UUID) (SessionState, error) {
	state := SessionState{SessionID: sessionID, ProjectID: projectID, WorkingMemory: map[string]any{}}
	var activeToolText, activePatchText string
	var workingMemoryRaw, lastPageRaw []byte
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(active_tool_call_id::text, ''),
       COALESCE(active_patch_proposal_id::text, ''),
       COALESCE(working_memory_json, '{}'::jsonb),
       COALESCE(last_page_context_json, page_context_json, '{}'::jsonb)
  FROM ai_chat_sessions
 WHERE ai_chat_session_id = $1 AND project_id = $2`, sessionID, projectID).Scan(&activeToolText, &activePatchText, &workingMemoryRaw, &lastPageRaw); err != nil {
		return state, err
	}
	state.ActiveToolCallID = uuidOrNilValue(activeToolText)
	state.ActivePatchProposalID = uuidOrNilValue(activePatchText)
	state.LastPageContext = json.RawMessage(lastPageRaw)
	_ = json.Unmarshal(workingMemoryRaw, &state.WorkingMemory)
	if state.WorkingMemory == nil {
		state.WorkingMemory = map[string]any{}
	}
	if call, ok := a.loadActiveToolCallState(ctx, tx, projectID, sessionID, state.ActiveToolCallID); ok {
		state.ActiveToolCall = &call
		state.ActiveToolCallID = call.ID
		return state, nil
	}
	if call, ok := a.loadLatestPendingToolCallState(ctx, tx, projectID, sessionID); ok {
		state.ActiveToolCall = &call
		state.ActiveToolCallID = call.ID
	}
	return state, nil
}

func (a *AI) loadActiveToolCallState(ctx context.Context, tx pgx.Tx, projectID, sessionID, toolCallID uuid.UUID) (ToolCallState, bool) {
	if toolCallID == uuid.Nil {
		return ToolCallState{}, false
	}
	return a.loadToolCallState(ctx, tx, projectID, sessionID, `ai_tool_call_id = $1`, []any{toolCallID})
}

func (a *AI) loadLatestPendingToolCallState(ctx context.Context, tx pgx.Tx, projectID, sessionID uuid.UUID) (ToolCallState, bool) {
	return a.loadToolCallState(ctx, tx, projectID, sessionID, `status = 'pending_approval'`, nil)
}

func (a *AI) loadToolCallState(ctx context.Context, tx pgx.Tx, projectID, sessionID uuid.UUID, extraWhere string, args []any) (ToolCallState, bool) {
	queryArgs := append(args, sessionID, projectID)
	rows, err := tx.Query(ctx, `
SELECT ai_tool_call_id, tool_key, risk_level, status, input_json
  FROM ai_tool_calls
 WHERE `+extraWhere+`
   AND ai_chat_session_id = $`+itoa(len(args)+1)+`
   AND project_id = $`+itoa(len(args)+2)+`
 ORDER BY created_at DESC
 LIMIT 1`, queryArgs...)
	if err != nil {
		return ToolCallState{}, false
	}
	defer rows.Close()
	if !rows.Next() {
		return ToolCallState{}, false
	}
	var call ToolCallState
	var raw []byte
	if err := rows.Scan(&call.ID, &call.ToolKey, &call.RiskLevel, &call.Status, &raw); err != nil {
		return ToolCallState{}, false
	}
	_ = json.Unmarshal(raw, &call.Input)
	if call.Input == nil {
		call.Input = map[string]any{}
	}
	return call, true
}

func (a *AI) ResolveProposalEdit(ctx context.Context, projectID, userID uuid.UUID, in MessageInput, state SessionState) (ProposalEdit, bool) {
	if state.ActiveToolCall == nil || state.ActiveToolCall.Status != "pending_approval" || state.ActiveToolCall.ToolKey != "task.create" {
		return ProposalEdit{}, false
	}
	body := strings.TrimSpace(in.Body)
	lower := strings.ToLower(body)
	if body == "" || isExplicitNewCommand(lower) {
		return ProposalEdit{}, false
	}
	patch := map[string]any{}
	summary := []string{}
	if due, label := proposalDueDate(body, in.LocalDate, in.Timezone); due != "" {
		patch["due_date"] = due
		patch["due_date_label"] = label
		summary = append(summary, "срок — "+label)
	}
	if strings.Contains(lower, "срочно") || strings.Contains(lower, "urgent") {
		patch["priority"] = "urgent"
		summary = append(summary, "приоритет — срочный")
	} else if strings.Contains(lower, "высок") || strings.Contains(lower, "важн") || strings.Contains(lower, "high") {
		patch["priority"] = "high"
		summary = append(summary, "приоритет — высокий")
	} else if strings.Contains(lower, "низк") || strings.Contains(lower, "low") {
		patch["priority"] = "low"
		summary = append(summary, "приоритет — низкий")
	} else if strings.Contains(lower, "средн") || strings.Contains(lower, "medium") {
		patch["priority"] = "medium"
		summary = append(summary, "приоритет — средний")
	}
	if strings.Contains(lower, "без клиента") || strings.Contains(lower, "клиента не надо") {
		patch["client_id"] = ""
		patch["client_name"] = ""
		patch["client_query"] = ""
		patch["client_resolved"] = false
		summary = append(summary, "клиент убран")
	} else if query := clientQueryFromText(body); query != "" && !isDateOnlyFollowUp(body) {
		matches := a.findClients(ctx, projectID, userID, query, 5)
		clientID, clientName := firstClientMatch(matches)
		patch["client_query"] = query
		patch["client_id"] = clientID
		patch["client_name"] = clientName
		patch["client_resolved"] = clientID != ""
		snippet := clientSearchSnippet(query, matches)
		patch["client_search_snippet"] = snippet
		patch["entity_snippets"] = []map[string]any{snippet}
		if clientName != "" {
			summary = append(summary, "клиент — "+clientName)
		} else {
			summary = append(summary, "клиент требует уточнения")
		}
	}
	if title := titleEditFromText(body); title != "" {
		patch["title"] = title
		summary = append(summary, "название обновлено")
	}
	if len(patch) == 0 {
		return ProposalEdit{}, false
	}
	return ProposalEdit{
		ToolCallID:   state.ActiveToolCall.ID,
		ToolKey:      state.ActiveToolCall.ToolKey,
		Patch:        patch,
		HumanSummary: strings.Join(summary, ", "),
	}, true
}

func (a *AI) ApplyProposalEdit(ctx context.Context, tx pgx.Tx, projectID, sessionID, userID uuid.UUID, edit ProposalEdit) (string, map[string]any, error) {
	var beforeRaw []byte
	var toolKey string
	if err := tx.QueryRow(ctx, `
SELECT tool_key, input_json
  FROM ai_tool_calls
 WHERE ai_tool_call_id = $1
   AND project_id = $2
   AND status = 'pending_approval'`, edit.ToolCallID, projectID).Scan(&toolKey, &beforeRaw); err != nil {
		return "", nil, err
	}
	before := map[string]any{}
	_ = json.Unmarshal(beforeRaw, &before)
	after := copyMap(before)
	for key, value := range edit.Patch {
		after[key] = value
	}
	if err := a.updateToolCallInputTx(ctx, tx, projectID, sessionID, edit.ToolCallID, userID, before, after, "proposal follow-up edit"); err != nil {
		return "", nil, err
	}
	_, _ = tx.Exec(ctx, `
UPDATE ai_chat_sessions
   SET active_tool_call_id = $2,
       status = 'waiting_tool_approval',
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, edit.ToolCallID)
	payload := proposalPayload(edit.ToolCallID.String(), "task.create.update", toolKey, "write", after)
	payload["proposal_updated"] = true
	body := "Обновил черновик задачи"
	if edit.HumanSummary != "" {
		body += ": " + edit.HumanSummary
	}
	body += ". Проверьте карточку и нажмите «Применить», когда все верно."
	return body, payload, nil
}

func proposalDueDate(text, localDate, timezone string) (string, string) {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "послезавтра"):
		return dueDateFromTextWithBase("послезавтра", localDate, timezone), "послезавтра"
	case strings.Contains(lower, "завтра"):
		return dueDateFromTextWithBase("завтра", localDate, timezone), "завтра"
	case strings.Contains(lower, "сегодня"):
		return dueDateFromTextWithBase("сегодня", localDate, timezone), "сегодня"
	default:
		return "", ""
	}
}

func titleEditFromText(text string) string {
	patterns := []string{
		`(?i)назови\s+(.{3,120})`,
		`(?i)название\s+(.{3,120})`,
	}
	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(text); len(match) > 1 {
			return strings.Trim(match[1], " \t\r\n.,!?\"'«»")
		}
	}
	return ""
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
