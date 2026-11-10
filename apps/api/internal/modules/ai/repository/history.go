package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

const (
	langGraphHistoryMaxMessages = 12
	langGraphHistoryMaxChars    = 12000
	langGraphHistoryBodyChars   = 1400
)

var (
	historyBearerPattern  = regexp.MustCompile(`(?i)bearer\s+[a-z0-9._~+/=-]+`)
	historySecretPattern  = regexp.MustCompile(`(?i)\b(password|secret|token|access_token|refresh_token|api[_-]?key|authorization)\s*[:=]\s*["']?[^\s"',}]+`)
	historyLiveKeyPattern = regexp.MustCompile(`\brp_(?:live|test)_[A-Za-z0-9._~+/=-]+`)
)

type LangGraphHistoryMessage struct {
	MessageID string
	Role      string
	Body      string
	Payload   map[string]any
	CreatedAt time.Time
}

type langGraphHistoryRow struct {
	MessageID string
	Role      string
	Body      string
	Payload   []byte
	CreatedAt time.Time
}

func (a *AI) LangGraphTurnHistory(ctx context.Context, projectID, sessionID uuid.UUID) ([]LangGraphHistoryMessage, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	var total int
	if err := a.P.QueryRow(ctx, `
SELECT count(*)
  FROM ai_chat_messages m
  JOIN ai_chat_sessions s ON s.ai_chat_session_id = m.ai_chat_session_id
 WHERE m.ai_chat_session_id = $1
   AND s.project_id = $2`, sessionID, projectID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := a.P.Query(ctx, `
SELECT ai_chat_message_id::text, role::text, body::text, payload_json, created_at
  FROM ai_chat_messages
 WHERE ai_chat_session_id = $1
 ORDER BY created_at DESC
 LIMIT $2`, sessionID, langGraphHistoryMaxMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rawRows := make([]langGraphHistoryRow, 0, langGraphHistoryMaxMessages)
	for rows.Next() {
		var row langGraphHistoryRow
		if err := rows.Scan(&row.MessageID, &row.Role, &row.Body, &row.Payload, &row.CreatedAt); err != nil {
			return nil, err
		}
		rawRows = append(rawRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.Reverse(rawRows)
	omitted := total - len(rawRows)
	if omitted < 0 {
		omitted = 0
	}
	return sanitizeLangGraphHistory(rawRows, omitted), nil
}

func sanitizeLangGraphHistory(rows []langGraphHistoryRow, omitted int) []LangGraphHistoryMessage {
	out := make([]LangGraphHistoryMessage, 0, len(rows)+1)
	usedChars := 0
	for _, row := range rows {
		message := LangGraphHistoryMessage{
			MessageID: strings.TrimSpace(row.MessageID),
			Role:      sanitizeHistoryRole(row.Role),
			Body:      truncateHistoryBody(redactHistoryText(row.Body)),
			Payload:   sanitizeHistoryPayload(row.Payload),
			CreatedAt: row.CreatedAt.UTC(),
		}
		if message.Body == "" && len(message.Payload) == 0 {
			continue
		}
		usedChars += len([]rune(message.Body))
		out = append(out, message)
	}
	for usedChars > langGraphHistoryMaxChars && len(out) > 0 {
		usedChars -= len([]rune(out[0].Body))
		out = out[1:]
		omitted++
	}
	if omitted > 0 {
		summary := LangGraphHistoryMessage{
			MessageID: "summary:langgraph-history",
			Role:      "system",
			Body:      fmt.Sprintf("Earlier conversation summary: %d older messages were omitted by bounded history policy.", omitted),
			Payload: map[string]any{
				"kind":          "conversation_summary",
				"omitted_count": omitted,
				"redacted":      true,
			},
			CreatedAt: time.Time{},
		}
		out = append([]LangGraphHistoryMessage{summary}, out...)
	}
	return out
}

func sanitizeHistoryRole(role string) string {
	switch strings.TrimSpace(role) {
	case "system", "user", "assistant", "tool", "approval", "event":
		return strings.TrimSpace(role)
	default:
		return "event"
	}
}

func truncateHistoryBody(body string) string {
	runes := []rune(strings.TrimSpace(body))
	if len(runes) <= langGraphHistoryBodyChars {
		return string(runes)
	}
	return string(runes[:langGraphHistoryBodyChars]) + "..."
}

func redactHistoryText(value string) string {
	value = historyBearerPattern.ReplaceAllString(value, "Bearer [redacted]")
	value = historyLiveKeyPattern.ReplaceAllString(value, "[redacted-token]")
	return historySecretPattern.ReplaceAllString(value, "$1=[redacted]")
}

func sanitizeHistoryPayload(raw []byte) map[string]any {
	var payload map[string]any
	_ = json.Unmarshal(raw, &payload)
	if len(payload) == 0 {
		return map[string]any{}
	}
	allowed := map[string]bool{
		"kind":                  true,
		"runtime":               true,
		"intent":                true,
		"tool_key":              true,
		"risk_level":            true,
		"ai_tool_call_id":       true,
		"client_command":        true,
		"local_date":            true,
		"timezone":              true,
		"source_surface":        true,
		"proposal_updated":      true,
		"context_message_count": true,
	}
	out := map[string]any{}
	for key, value := range payload {
		if allowed[key] {
			out[key] = value
		}
	}
	if len(out) > 0 {
		out["redacted"] = true
	}
	return out
}
