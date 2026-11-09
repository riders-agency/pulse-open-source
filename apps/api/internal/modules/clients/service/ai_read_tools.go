package service

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/authz"
)

type AIReadTools struct {
	P       dbpkg.DB
	Clients *clientsrepo.Clients
}

func NewAIReadTools(pool dbpkg.DB, clients *clientsrepo.Clients) *AIReadTools {
	return &AIReadTools{P: pool, Clients: clients}
}

func (t *AIReadTools) ReadClientSearch(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, t.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	rawQuery := stringValue(input["query"])
	query := firstNonEmptyString(
		stringValue(input["client_query"]),
		stringValue(input["name"]),
		clientQueryFromText(rawQuery),
		rawQuery,
	)
	matches := t.findClients(ctx, projectID, userID, query, 8, true)
	snippet := clientSearchSnippet(query, matches)
	answer := fmt.Sprintf("Клиент по запросу «%s» не найден.", query)
	if query == "" && len(matches) == 0 {
		answer = "В проекте клиенты не найдены."
	} else if query == "" {
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Name)
		}
		answer = fmt.Sprintf("Клиенты проекта: %s.", strings.Join(names, ", "))
	} else if len(matches) == 1 {
		answer = fmt.Sprintf("Найден клиент: %s.", matches[0].Name)
	} else if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Name)
		}
		answer = fmt.Sprintf("Найдено клиентов: %d. Варианты: %s.", len(matches), strings.Join(names, ", "))
	}
	return map[string]any{
		"kind":               "client_search_result",
		"tool_key":           "client.search",
		"query":              query,
		"matches":            matches,
		"client_snippet":     snippet,
		"entity_snippets":    []map[string]any{snippet},
		"answer":             answer,
		"from_memory":        false,
		"runtime":            "langgraph",
		"needs_choice":       len(matches) > 1,
		"selected_client":    firstClientMap(matches),
		"selected_client_id": firstClientID(matches),
	}, nil
}

func (t *AIReadTools) ReadClientCount(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, t.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	var count int
	if err := t.P.QueryRow(ctx, `
SELECT count(*)
  FROM clients c
  JOIN client_projects cp ON cp.client_id = c.client_id
 WHERE cp.project_id = $1
   AND c.status <> 'deleted'`, projectID).Scan(&count); err != nil {
		return nil, err
	}
	return map[string]any{
		"kind":        "client_count",
		"tool_key":    "client.count",
		"count":       count,
		"answer":      fmt.Sprintf("В проекте клиентов: %d.", count),
		"from_memory": false,
		"runtime":     "langgraph",
	}, nil
}

func (t *AIReadTools) ReadPaymentDue(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, t.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	query := stringValue(input["query"])
	if stringValue(input["scope"]) == "project" || isProjectPaymentQuery(query) {
		answer, payload := t.projectPaymentLookup(ctx, projectID, userID, intValue(input["limit"]))
		payload["kind"] = "payment_due_project"
		payload["tool_key"] = "payment.lookup_due"
		payload["answer"] = answer
		payload["runtime"] = "langgraph"
		return payload, nil
	}
	answer, payload := t.paymentLookup(ctx, projectID, userID, query, input)
	payload["kind"] = "payment_due"
	payload["tool_key"] = "payment.lookup_due"
	payload["answer"] = answer
	payload["runtime"] = "langgraph"
	return payload, nil
}

type aiClientMatch struct {
	ID   string `json:"client_id"`
	Name string `json:"name"`
}

func (t *AIReadTools) paymentLookup(ctx context.Context, projectID, userID uuid.UUID, query string, input map[string]any) (string, map[string]any) {
	query = firstNonEmptyString(clientQueryFromText(query), query)
	clientIDText := firstNonEmptyString(stringValue(input["client_id"]), stringValue(input["clientId"]))
	clientName := firstNonEmptyString(stringValue(input["client_name"]), stringValue(input["clientName"]))
	pageClientID, pageClientName := pageContextClient(mapFromAny(input["page_context"]))
	clientIDText = firstNonEmptyString(clientIDText, pageClientID)
	clientName = firstNonEmptyString(clientName, pageClientName)
	if clientIDText == "" {
		matches := t.findClients(ctx, projectID, userID, query, 3, false)
		if len(matches) > 1 {
			names := make([]string, 0, len(matches))
			for _, match := range matches {
				names = append(names, match.Name)
			}
			return "Нашел несколько клиентов по запросу «" + query + "»: " + strings.Join(names, ", ") + ". Уточните, по какому клиенту смотреть оплату.", map[string]any{"from_memory": false, "client_query": query, "matches": matches, "needs_choice": true}
		}
		if len(matches) == 1 {
			clientIDText = matches[0].ID
			clientName = matches[0].Name
		}
	}
	if clientIDText == "" {
		return "Активный клиент по запросу «" + query + "» не найден.", map[string]any{"from_memory": false, "client_query": query}
	}
	if t.Clients == nil {
		return "Модуль оплат недоступен в runtime AI.", map[string]any{"from_memory": false, "client_id": clientIDText}
	}
	clientID, _ := uuid.Parse(clientIDText)
	actions, err := t.Clients.ListClientPaymentActions(ctx, clientID, userID, &projectID)
	if err != nil {
		return "Не удалось получить оплаты клиента: доступ ограничен или данные недоступны.", map[string]any{"from_memory": false, "client_id": clientIDText}
	}
	return paymentActionAnswer(clientIDText, clientName, actions)
}

func (t *AIReadTools) projectPaymentLookup(ctx context.Context, projectID, userID uuid.UUID, limit int) (string, map[string]any) {
	if t.Clients == nil {
		return "Модуль оплат недоступен в runtime AI.", map[string]any{"from_memory": false}
	}
	if limit <= 0 {
		limit = 8
	}
	actions, err := t.Clients.ListPaymentActions(ctx, projectID, userID)
	if err != nil {
		return "Не удалось получить ближайшие оплаты проекта: доступ ограничен или данные недоступны.", map[string]any{"from_memory": false}
	}
	active := make([]clientsrepo.PaymentActionRow, 0, limit)
	for _, action := range actions {
		if !isActivePaymentStatus(action.Status) {
			continue
		}
		active = append(active, action)
		if len(active) >= limit {
			break
		}
	}
	return projectPaymentActionsAnswer(active)
}

func (t *AIReadTools) findClients(ctx context.Context, projectID, userID uuid.UUID, query string, limit int, allowEmpty bool) []aiClientMatch {
	query = strings.TrimSpace(query)
	if (!allowEmpty && query == "") || t == nil || t.P == nil || authz.RequireProject(ctx, t.P, projectID, authz.PermRead) != nil {
		return nil
	}
	if limit <= 0 {
		limit = 3
	}
	tokens := clientSearchTokens(query)
	rows, err := t.P.Query(ctx, `
WITH params AS (
  SELECT lower(trim($2::text)) AS query, $3::text[] AS tokens
)
SELECT c.client_id::text, COALESCE(c.name, c.full_name, '')
  FROM clients c
  JOIN client_projects cp ON cp.client_id = c.client_id
 CROSS JOIN params p
 WHERE cp.project_id = $1
   AND c.status <> 'deleted'
   AND (
     p.query = ''
     OR lower(COALESCE(c.name, '')) = p.query
     OR lower(COALESCE(c.full_name, '')) = p.query
     OR lower(COALESCE(c.short_name, '')) = p.query
     OR lower(COALESCE(c.name, '') || ' ' || COALESCE(c.full_name, '') || ' ' || COALESCE(c.short_name, '')) LIKE '%' || p.query || '%'
     OR EXISTS (
       SELECT 1 FROM unnest(p.tokens) token
        WHERE length(token) >= 3
          AND lower(COALESCE(c.name, '') || ' ' || COALESCE(c.full_name, '') || ' ' || COALESCE(c.short_name, '')) LIKE '%' || token || '%'
     )
     OR similarity(lower(COALESCE(c.name, '')), p.query) > 0.18
     OR word_similarity(p.query, lower(COALESCE(c.name, '') || ' ' || COALESCE(c.full_name, '') || ' ' || COALESCE(c.short_name, ''))) > 0.28
   )
 ORDER BY
   CASE
     WHEN p.query = '' THEN 10
     WHEN lower(COALESCE(c.name, '')) = p.query THEN 0
     WHEN lower(COALESCE(c.name, '')) LIKE p.query || '%' THEN 1
     WHEN lower(COALESCE(c.name, '') || ' ' || COALESCE(c.full_name, '') || ' ' || COALESCE(c.short_name, '')) LIKE '%' || p.query || '%' THEN 2
     ELSE 3
   END,
   GREATEST(
     similarity(lower(COALESCE(c.name, '')), p.query),
     word_similarity(p.query, lower(COALESCE(c.name, '') || ' ' || COALESCE(c.full_name, '') || ' ' || COALESCE(c.short_name, '')))
   ) DESC,
   c.updated_at DESC
 LIMIT $4`, projectID, query, tokens, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []aiClientMatch{}
	for rows.Next() {
		var match aiClientMatch
		if err := rows.Scan(&match.ID, &match.Name); err == nil {
			out = append(out, match)
		}
	}
	return out
}

func paymentActionAnswer(clientIDText, clientName string, actions []clientsrepo.PaymentActionRow) (string, map[string]any) {
	var selected *clientsrepo.PaymentActionRow
	for i := range actions {
		if isActivePaymentStatus(actions[i].Status) {
			selected = &actions[i]
			break
		}
	}
	if selected == nil && len(actions) > 0 {
		selected = &actions[0]
	}
	if selected == nil {
		return "По клиенту «" + clientName + "» активных оплат не найдено.", map[string]any{"from_memory": false, "client_id": clientIDText}
	}
	if clientName == "" {
		clientName = selected.ClientName
	}
	answer := fmt.Sprintf("Оплата по клиенту «%s»: срок %s, сумма %.2f %s, статус %s. Источник: payment_actions.", clientName, selected.DueDate, selected.TotalAmount, selected.Currency, selected.Status)
	return answer, map[string]any{
		"from_memory":       false,
		"client_id":         clientIDText,
		"client_name":       clientName,
		"payment_action_id": selected.PaymentActionID,
		"due_date":          selected.DueDate,
		"amount":            selected.TotalAmount,
		"currency":          selected.Currency,
		"status":            selected.Status,
		"url":               "/clients/" + clientIDText,
	}
}

func projectPaymentActionsAnswer(actions []clientsrepo.PaymentActionRow) (string, map[string]any) {
	if len(actions) == 0 {
		return "Ближайших активных оплат по проекту не найдено.", map[string]any{
			"from_memory": false,
			"items":       []map[string]any{},
		}
	}
	items := make([]map[string]any, 0, len(actions))
	parts := make([]string, 0, len(actions))
	for i, action := range actions {
		item := paymentActionMap(action)
		items = append(items, item)
		parts = append(
			parts,
			fmt.Sprintf("%d. %s — %s, %s", i+1, action.ClientName, formatPaymentDate(action.DueDate), formatPaymentAmount(action.TotalAmount, action.Currency)),
		)
	}
	return "Ближайшие оплаты:\n" + strings.Join(parts, "\n"), map[string]any{
		"from_memory": false,
		"items":       items,
	}
}

func paymentActionMap(action clientsrepo.PaymentActionRow) map[string]any {
	return map[string]any{
		"payment_action_id": action.PaymentActionID,
		"client_id":         action.ClientID,
		"client_name":       action.ClientName,
		"project_id":        action.ProjectID,
		"due_date":          action.DueDate,
		"amount":            action.TotalAmount,
		"currency":          action.Currency,
		"status":            action.Status,
		"url":               "/clients/" + action.ClientID,
	}
}

func isActivePaymentStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "paid", "cancelled", "canceled", "completed":
		return false
	default:
		return true
	}
}

func clientSearchSnippet(query string, matches []aiClientMatch) map[string]any {
	return map[string]any{"kind": "client_search", "query": query, "matches": matches}
}

func firstClientMap(matches []aiClientMatch) map[string]any {
	if len(matches) == 0 {
		return nil
	}
	return map[string]any{"client_id": matches[0].ID, "name": matches[0].Name}
}

func firstClientID(matches []aiClientMatch) string {
	if len(matches) == 0 {
		return ""
	}
	return matches[0].ID
}

func clientQueryFromText(text string) string {
	text = strings.TrimSpace(text)
	parts := strings.FieldsFunc(text, func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSpace(r) })
	for i, part := range parts {
		lower := strings.ToLower(part)
		if (lower == "клиент" || lower == "клиента" || lower == "client") && i+1 < len(parts) {
			return trimClientQueryTokens(parts[i+1:])
		}
		if strings.HasPrefix(lower, "оплат") && i+1 < len(parts) {
			next := i + 1
			if next < len(parts) && isClientQuerySkipWord(strings.ToLower(parts[next])) {
				next++
			}
			if next < len(parts) {
				return trimClientQueryTokens(parts[next:])
			}
		}
		if (lower == "у" || lower == "по" || lower == "для") && i+1 < len(parts) {
			return trimClientQueryTokens(parts[i+1:])
		}
	}
	return text
}

func clientSearchTokens(query string) []string {
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(query)), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	tokens := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len([]rune(part)) < 3 || seen[part] || isClientQueryStopWord(part) {
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

func trimClientQueryTokens(parts []string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.Trim(part, " \t\r\n.,!?\"'«»()[]{}")
		if clean == "" {
			continue
		}
		lower := strings.ToLower(clean)
		if isClientQueryStopWord(lower) {
			if len(out) > 0 {
				break
			}
			continue
		}
		out = append(out, clean)
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

func isClientQuerySkipWord(value string) bool {
	return value == "у" || value == "по" || value == "для"
}

func isClientQueryStopWord(value string) bool {
	switch value {
	case "задача", "задачу", "задачи", "создай", "создать", "поставь", "проверить", "оплату", "оплата", "оплаты", "завтра", "сегодня", "послезавтра", "клиент", "клиента", "найди", "найти", "покажи":
		return true
	default:
		return false
	}
}

func pageContextClient(pageContext map[string]any) (string, string) {
	return firstNonEmptyString(stringValue(pageContext["client_id"]), stringValue(pageContext["clientId"])), firstNonEmptyString(stringValue(pageContext["client_name"]), stringValue(pageContext["clientName"]))
}

func mapFromAny(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func formatPaymentDate(value string) string {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 3 || len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return strings.TrimSpace(value)
	}
	return parts[2] + "." + parts[1] + "." + parts[0]
}

func formatPaymentAmount(amount float64, currency string) string {
	value := fmt.Sprintf("%.2f", amount)
	value = strings.TrimSuffix(value, ".00")
	if parts := strings.SplitN(value, ".", 2); len(parts) == 2 {
		value = groupThousands(parts[0]) + "." + parts[1]
	} else {
		value = groupThousands(value)
	}
	return strings.TrimSpace(value + " " + strings.TrimSpace(currency))
}

func groupThousands(value string) string {
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign = "-"
		value = strings.TrimPrefix(value, "-")
	}
	if len(value) <= 3 {
		return sign + value
	}
	var groups []string
	for len(value) > 3 {
		groups = append([]string{value[len(value)-3:]}, groups...)
		value = value[:len(value)-3]
	}
	if value != "" {
		groups = append([]string{value}, groups...)
	}
	return sign + strings.Join(groups, " ")
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isProjectPaymentQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if !strings.Contains(lower, "оплат") && !strings.Contains(lower, "платеж") && !strings.Contains(lower, "платёж") {
		return false
	}
	if strings.Contains(lower, "у кого") || strings.Contains(lower, "кто должен") {
		return true
	}
	if strings.Contains(lower, "по проекту") || strings.Contains(lower, "в проекте") ||
		strings.Contains(lower, "по всем клиент") || strings.Contains(lower, "у всех клиент") {
		return true
	}
	if strings.Contains(lower, " у ") {
		return false
	}
	return strings.Contains(lower, "ближайш") ||
		strings.Contains(lower, "предстоящ") ||
		strings.Contains(lower, "следующ")
}
