package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

func pageContextAnswer(pageContext json.RawMessage, body string) string {
	var ctx map[string]any
	_ = json.Unmarshal(pageContext, &ctx)
	page := stringValue(ctx["pageKind"])
	if page == "" {
		page = stringValue(ctx["screen_scope"])
	}
	path := stringValue(ctx["pathname"])
	projectID := stringValue(ctx["projectId"])
	if projectID == "" {
		projectID = stringValue(ctx["project_id"])
	}
	if strings.TrimSpace(body) == "" {
		body = "Где я сейчас?"
	}
	return fmt.Sprintf("Вы сейчас в контексте проекта %s, экран: %s, путь: %s. Я буду читать детали через разрешенные project-scoped tools, не из памяти.", projectID, pageContextLabel(page), path)
}

func pageContextLabel(page string) string {
	switch page {
	case "site_chat":
		return "Чат сайта / Live Assist"
	case "project_dashboard", "dashboard":
		return "Обзор проекта"
	case "task_board", "tasks":
		return "Задачи"
	case "task_detail", "task":
		return "Карточка задачи"
	case "client_detail", "client":
		return "Карточка клиента"
	case "clients":
		return "Клиенты"
	case "workflow_editor":
		return "Редактор workflow"
	case "workflow_run":
		return "Run workflow"
	case "integrations":
		return "Интеграции"
	case "payments":
		return "Оплаты"
	default:
		return defaultString(page, "неизвестный экран")
	}
}

type aiClientMatch struct {
	ID   string `json:"client_id"`
	Name string `json:"name"`
}

func (a *AI) findClients(ctx context.Context, projectID, userID uuid.UUID, query string, limit int) []aiClientMatch {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	if a == nil || a.P == nil {
		return nil
	}
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil
	}
	if limit <= 0 {
		limit = 3
	}
	tokens := clientSearchTokens(query)
	rows, err := a.P.Query(ctx, `
SELECT c.client_id::text, c.name
  FROM pulse_read.clients_v1 c
  JOIN pulse_read.client_projects_v1 cp ON cp.client_id = c.client_id
 WHERE cp.project_id = $1
   AND c.status <> 'deleted'
   AND (
     lower(c.name) = lower($2)
     OR lower(c.name) LIKE '%' || lower($2) || '%'
     OR lower(COALESCE(c.full_name,'')) LIKE '%' || lower($2) || '%'
     OR EXISTS (
       SELECT 1
         FROM unnest($3::text[]) token
        WHERE length(token) >= 3
          AND (
            lower(c.name) LIKE '%' || lower(token) || '%'
            OR lower(COALESCE(c.full_name,'')) LIKE '%' || lower(token) || '%'
          )
     )
   )
 ORDER BY CASE WHEN lower(c.name) = lower($2) THEN 0 ELSE 1 END, c.updated_at DESC
 LIMIT $4`, projectID, query, tokens, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]aiClientMatch, 0, limit)
	for rows.Next() {
		var item aiClientMatch
		if err := rows.Scan(&item.ID, &item.Name); err == nil {
			out = append(out, item)
		}
	}
	return out
}

func clientSearchTokens(query string) []string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	tokens := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.Trim(part, " \t\r\n.,!?\"'«»()[]{}")
		if len([]rune(part)) < 3 || seen[part] {
			continue
		}
		seen[part] = true
		tokens = append(tokens, part)
	}
	if len(tokens) == 0 && strings.TrimSpace(query) != "" {
		tokens = append(tokens, strings.ToLower(strings.TrimSpace(query)))
	}
	return tokens
}

func firstClientMatch(matches []aiClientMatch) (string, string) {
	if len(matches) == 0 {
		return "", ""
	}
	return matches[0].ID, matches[0].Name
}

func clientSearchSnippet(query string, matches []aiClientMatch) map[string]any {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	items := make([]map[string]any, 0, len(matches))
	for i, match := range matches {
		items = append(items, map[string]any{
			"client_id": match.ID,
			"name":      match.Name,
			"rank":      i + 1,
			"url":       "/clients/" + match.ID,
		})
	}
	selectedID := ""
	selectedName := ""
	selectedURL := ""
	status := "not_found"
	if len(matches) > 0 {
		selectedID = matches[0].ID
		selectedName = matches[0].Name
		selectedURL = "/clients/" + matches[0].ID
		status = "matched"
		if len(matches) > 1 {
			status = "multiple_matches"
		}
	}
	return map[string]any{
		"entity_type":   "client",
		"tool_key":      "client.search",
		"query":         query,
		"status":        status,
		"selected_id":   selectedID,
		"selected_name": selectedName,
		"selected_url":  selectedURL,
		"matches":       items,
		"source":        "clients",
	}
}

func entitySnippetsFromFields(fields map[string]any) []map[string]any {
	raw := fields["entity_snippets"]
	out := normalizeEntitySnippets(raw)
	if len(out) > 0 {
		return out
	}
	if snippet := normalizeEntitySnippet(fields["client_search_snippet"]); len(snippet) > 0 {
		return []map[string]any{snippet}
	}
	return nil
}

func normalizeEntitySnippets(raw any) []map[string]any {
	switch value := raw.(type) {
	case []map[string]any:
		return value
	case []any:
		out := make([]map[string]any, 0, len(value))
		for _, item := range value {
			if snippet := normalizeEntitySnippet(item); len(snippet) > 0 {
				out = append(out, snippet)
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeEntitySnippet(raw any) map[string]any {
	value, ok := raw.(map[string]any)
	if !ok || len(value) == 0 {
		return nil
	}
	return value
}
