package repository

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

	"github.com/google/uuid"
)

func selectEvaluationCandidates(models []ModelDescriptor, meta map[string]any) []ModelDescriptor {
	allowed := map[string]bool{}
	if raw, ok := meta["allowed_model_ids"].([]any); ok {
		for _, item := range raw {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				allowed[s] = true
			}
		}
	}
	filtered := make([]ModelDescriptor, 0, len(models))
	for _, model := range models {
		if len(allowed) > 0 && !allowed[model.ID] {
			continue
		}
		filtered = append(filtered, model)
	}
	if len(filtered) == 0 {
		filtered = models
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return aiproviders.ModelRank(filtered[i].ID) > aiproviders.ModelRank(filtered[j].ID)
	})
	if len(filtered) > 3 {
		filtered = filtered[:3]
	}
	return filtered
}

func scoreEvaluation(content string, err error, latency time.Duration) float64 {
	if err != nil {
		return 0
	}
	score := 20.0
	var parsed map[string]any
	if json.Unmarshal([]byte(extractFirstJSON(content)), &parsed) == nil {
		score += 35
		if stringValue(parsed["intent"]) != "" || stringValue(parsed["key"]) != "" {
			score += 15
		}
		if _, ok := parsed["patch"]; ok {
			score += 20
		}
	}
	latencyScore := math.Max(0, 10-float64(latency.Milliseconds())/1000)
	return score + latencyScore
}

func extractFirstJSON(text string) string {
	text = strings.TrimSpace(text)
	start := strings.IndexAny(text, "{[")
	if start < 0 {
		return text
	}
	open := text[start]
	close := byte('}')
	if open == '[' {
		close = ']'
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inString {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == open {
			depth++
		}
		if ch == close {
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return text[start:]
}

func proposalPayload(callID, intent, toolKey, risk string, fields map[string]any) map[string]any {
	display, missing := proposalDisplay(toolKey, fields)
	snippets := entitySnippetsFromFields(fields)
	risk = normalizeAIRiskLevel(risk)
	return map[string]any{
		"kind":                 "proposal",
		"intent":               intent,
		"tool_key":             toolKey,
		"risk_level":           risk,
		"requires_approval":    true,
		"ai_tool_call_id":      callID,
		"fields":               fields,
		"fields_json":          fields,
		"display_json":         display,
		"missing_fields_json":  missing,
		"entity_snippets_json": snippets,
		"developer_json": map[string]any{
			"input_json":      fields,
			"tool_key":        toolKey,
			"policy":          "typed tool approval required",
			"entity_snippets": snippets,
		},
		"confidence_json": map[string]any{
			"client_resolved": boolValue(fields["client_resolved"]),
			"date_resolved":   stringValue(fields["due_date"]) != "",
		},
	}
}

func proposalDisplay(toolKey string, fields map[string]any) (map[string]any, []map[string]any) {
	missing := []map[string]any{}
	var display map[string]any
	switch toolKey {
	case "task.create":
		clientName := stringValue(fields["client_name"])
		clientQuery := stringValue(fields["client_query"])
		missing = append(missing, mapSlice(fields["missing_fields"])...)
		if clientName == "" && clientQuery != "" {
			if len(missing) == 0 {
				missing = append(missing, map[string]any{
					"field":   "client_id",
					"label":   "Клиент",
					"query":   clientQuery,
					"message": "Клиент «" + clientQuery + "» пока не найден точно. Уточните клиента или создайте его.",
				})
			}
		}
		dueDate := stringValue(fields["due_date"])
		dueLabel := dueDisplayLabel(stringValue(fields["due_date_label"]), dueDate)
		display = map[string]any{
			"title":                    stringValue(fields["title"]),
			"description":              stringValue(fields["description"]),
			"client_name":              defaultString(clientName, clientQuery),
			"client_status":            map[bool]string{true: "найден", false: "нужно уточнить"}[clientName != ""],
			"entity_snippets":          entitySnippetsFromFields(fields),
			"due_date":                 dueDate,
			"due_date_label":           dueLabel,
			"status_id":                stringValue(fields["status_id"]),
			"priority":                 defaultString(stringValue(fields["priority"]), "medium"),
			"billable":                 boolValue(fields["billable"]),
			"assignee_name":            defaultString(stringValue(fields["assignee_name"]), "не назначен"),
			"labels":                   stringSlice(fields["labels"]),
			"checklist":                stringSlice(fields["checklist"]),
			"notification_note":        stringValue(fields["notification_intent"]),
			"start_timer_after_create": boolValue(fields["start_timer_after_create"]),
		}
	case "task.move":
		display = map[string]any{
			"title":              stringValue(fields["title"]),
			"task_id":            stringValue(fields["task_id"]),
			"task_key":           stringValue(fields["task_key"]),
			"from_status_id":     stringValue(fields["from_status_id"]),
			"to_status_id":       stringValue(fields["to_status_id"]),
			"to_status_name":     stringValue(fields["to_status_name"]),
			"comment":            stringValue(fields["comment"]),
			"human_action_label": defaultString(stringValue(fields["human_action_label"]), "Стартовать задачу"),
		}
	case "client.create":
		display = map[string]any{
			"name":        stringValue(fields["name"]),
			"full_name":   stringValue(fields["full_name"]),
			"client_type": defaultString(stringValue(fields["client_type"]), "company"),
			"status":      defaultString(stringValue(fields["status"]), "active"),
		}
	default:
		display = fields
	}
	return display, missing
}

func clientQueryFromText(text string) string {
	cleaned := strings.TrimSpace(text)
	patterns := []string{
		`(?i)(?:для|по)\s+кли[а-яёe]*\s+([\pL\pN\-_"«»\. ]{2,80})`,
		`(?i)кли[а-яёe]*\s+([\pL\pN\-_"«»\. ]{2,80})`,
		`(?i)оплат[а-яё]*\s+(?:у\s+|по\s+)?([\pL\pN\-_"«»\. ]{2,80})`,
		`(?i)(?:^|[\s,.;!?])у\s+([\pL\pN\-_"«»\. ]{2,80})`,
		`(?i)(?:^|[\s,.;!?])(?:для|по)\s+([\pL\pN\-_"«»\. ]{2,80})`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(cleaned); len(m) > 1 {
			return trimEntityName(m[1])
		}
	}
	return ""
}

func trimEntityName(value string) string {
	value = strings.Trim(value, " \t\r\n.,!?\"'«»")
	stopWords := []string{"завтра", "сегодня", "послезавтра", "пожалуйста", "сделать", "закрыть", "закрытыть", "закрытиь", "все", "всё", "для", "по", "правка", "правки", "правок", "сообщить", "запустить", "запусти", "стартовать", "стартани", "таймер", "ней", "нему", "сеошников", "сешоникрв", "сешо", "seo", "им", "тоже", "до", "после", "и", "email", "telegram"}
	parts := strings.Fields(value)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		lower := strings.ToLower(strings.Trim(part, ".,!?"))
		stop := false
		for _, sw := range stopWords {
			if lower == sw {
				stop = true
				break
			}
		}
		if !stop {
			out = append(out, part)
			continue
		}
		if len(out) > 0 {
			break
		}
	}
	return strings.Trim(strings.Join(out, " "), " \t\r\n.,!?\"'«»")
}

func bestClientQuery(text, modelQuery string) string {
	localQuery := clientQueryFromText(text)
	modelQuery = trimEntityName(modelQuery)
	if localQuery == "" {
		return modelQuery
	}
	if modelQuery == "" || noisyClientQuery(modelQuery) {
		return localQuery
	}
	if len([]rune(modelQuery)) > len([]rune(localQuery))+12 && strings.Contains(strings.ToLower(modelQuery), strings.ToLower(localQuery)) {
		return localQuery
	}
	return modelQuery
}

func noisyClientQuery(value string) bool {
	lower := strings.ToLower(value)
	noise := []string{"провер", "оплат", "созда", "задач", "сделать", "закрыть", "закрытыть", "правк", "сообщ"}
	for _, token := range noise {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return len(strings.Fields(value)) > 3
}

func cleanTaskTitle(title, clientQuery string) string {
	cleaned := strings.TrimSpace(title)
	if clientQuery != "" {
		cleaned = regexp.MustCompile(`(?i)\b`+regexp.QuoteMeta(clientQuery)+`\b`).ReplaceAllString(cleaned, " ")
	}
	if clientQuery != "" {
		for _, part := range strings.Fields(clientQuery) {
			if len([]rune(part)) >= 3 {
				cleaned = regexp.MustCompile(`(?i)\b`+regexp.QuoteMeta(part)+`\b`).ReplaceAllString(cleaned, " ")
			}
		}
	}
	cleaned = regexp.MustCompile(`(?i)\b(для|по)\s+кли[а-яёe]*\b`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`(?i)\b(сегодня|завтра|послезавтра)\b`).ReplaceAllString(cleaned, " ")
	cleaned = strings.Join(strings.Fields(strings.Trim(cleaned, " \t\r\n.,!?\"'«»")), " ")
	return cleaned
}

func isExplicitNewCommand(lower string) bool {
	return strings.Contains(lower, "созда") && (strings.Contains(lower, "зада") || strings.Contains(lower, "клиент"))
}

func dueDateFromTextWithBase(text, localDate, timezone string) string {
	lower := strings.ToLower(text)
	now := time.Now()
	if timezone != "" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			now = now.In(loc)
		}
	}
	if localDate != "" {
		if parsed, err := time.Parse("2006-01-02", localDate); err == nil {
			now = parsed
		}
	}
	switch {
	case strings.Contains(lower, "послезавтра"):
		return now.AddDate(0, 0, 2).Format("2006-01-02")
	case strings.Contains(lower, "завтра"):
		return now.AddDate(0, 0, 1).Format("2006-01-02")
	case strings.Contains(lower, "сегодня"):
		return now.Format("2006-01-02")
	case strings.Contains(lower, "через 2 дня"):
		return now.AddDate(0, 0, 2).Format("2006-01-02")
	case strings.Contains(lower, "через 3 дня"):
		return now.AddDate(0, 0, 3).Format("2006-01-02")
	default:
		return ""
	}
}

func dateLabel(value string) string {
	if value == "" {
		return "не указан"
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		now := time.Now()
		full := parsed.Format("02.01.2006")
		switch parsed.Format("2006-01-02") {
		case now.Format("2006-01-02"):
			return "Сегодня, " + full
		case now.AddDate(0, 0, 1).Format("2006-01-02"):
			return "Завтра, " + full
		default:
			return full
		}
	}
	return value
}

func dueDisplayLabel(relative, isoDate string) string {
	if isoDate == "" {
		return dateLabel(isoDate)
	}
	dateOnly := dateOnlyLabel(isoDate)
	switch strings.ToLower(strings.TrimSpace(relative)) {
	case "сегодня":
		return "Сегодня, " + dateOnly
	case "завтра":
		return "Завтра, " + dateOnly
	case "послезавтра":
		return "Послезавтра, " + dateOnly
	default:
		return dateLabel(isoDate)
	}
}

func dateOnlyLabel(value string) string {
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed.Format("02.01.2006")
	}
	return value
}

func taskTitleFromText(text string) string {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "сео") || strings.Contains(lower, "сешо") || strings.Contains(lower, "seo") {
		return "Внести правки для SEO-специалистов"
	}
	if strings.Contains(lower, "оплат") && strings.Contains(lower, "провер") {
		return "Проверить оплату"
	}
	cleaned := regexp.MustCompile(`(?i)^.*(?:созда[^\n]*|постав[^\n]*)задач[ауи]?\s*`).ReplaceAllString(text, "")
	if query := clientQueryFromText(text); query != "" {
		cleaned = cleanTaskTitle(cleaned, query)
	}
	cleaned = regexp.MustCompile(`(?i)\b(мне|сегодня|завтра|послезавтра)\b`).ReplaceAllString(cleaned, " ")
	cleaned = strings.Join(strings.Fields(strings.Trim(cleaned, " \t\r\n.,!?")), " ")
	if len([]rune(cleaned)) > 100 {
		cleaned = string([]rune(cleaned)[:100])
	}
	return strings.TrimSpace(cleaned)
}

func taskDescriptionFromText(text string) string {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "сео") || strings.Contains(lower, "seo") {
		return "Сделать все правки для SEO-специалистов и сообщить им о результате.\n\nСоздано AI-сотрудником из запроса: " + text
	}
	return "Создано AI-сотрудником из запроса: " + text
}

func isDateOnlyFollowUp(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	return lower == "сегодня" || lower == "давай сегодня" || lower == "на сегодня" || lower == "завтра" || lower == "давай завтра" || lower == "на завтра" || lower == "послезавтра"
}

func copyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func uuidOrNilValue(value string) uuid.UUID {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil
	}
	return parsed
}

func jsonDefault(raw json.RawMessage) string {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return "{}"
	}
	return string(raw)
}

func mustJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func stringEntity(intent aiIntent, key string) string {
	if intent.Entities == nil {
		return ""
	}
	return strings.TrimSpace(stringValue(intent.Entities[key]))
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		if value == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func boolValue(value any) bool {
	v, _ := value.(bool)
	return v
}

func floatValue(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		out, _ := v.Float64()
		return out
	case string:
		out, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return out
	default:
		return 0
	}
}

func stringSlice(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func mapSlice(value any) []map[string]any {
	switch v := value.(type) {
	case []map[string]any:
		return v
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if mapped, ok := item.(map[string]any); ok {
				out = append(out, mapped)
			}
		}
		return out
	default:
		return nil
	}
}

func workflowDefinitionFromRaw(raw json.RawMessage) map[string]any {
	var def map[string]any
	_ = json.Unmarshal(raw, &def)
	if def == nil {
		def = map[string]any{}
	}
	if _, ok := def["nodes"].([]any); !ok {
		def["nodes"] = []any{}
	}
	if _, ok := def["edges"].([]any); !ok {
		def["edges"] = []any{}
	}
	return def
}

func workflowGraphFromRaw(raw json.RawMessage, def map[string]any) map[string]any {
	var graph map[string]any
	_ = json.Unmarshal(raw, &graph)
	if graph == nil {
		graph = map[string]any{}
	}
	if _, ok := graph["nodes"].([]any); !ok {
		nodes := make([]any, 0)
		if defNodes, ok := def["nodes"].([]any); ok {
			for i, item := range defNodes {
				node, _ := item.(map[string]any)
				nodes = append(nodes, map[string]any{
					"id":       stringValue(node["action_key"]),
					"type":     "action",
					"position": map[string]any{"x": 80 + (i%3)*300, "y": 160 + (i/3)*170},
					"data": map[string]any{
						"label":           stringValue(node["label"]),
						"action_type_key": stringValue(node["action_type_key"]),
						"config":          node["config"],
					},
				})
			}
		}
		graph["nodes"] = nodes
	}
	if _, ok := graph["edges"].([]any); !ok {
		graph["edges"] = []any{}
	}
	return graph
}

func addOrUpdateWorkflowAction(def, graph map[string]any, action map[string]any) {
	key := stringValue(action["action_key"])
	if key == "" {
		return
	}
	defNodes, _ := def["nodes"].([]any)
	nextDefNode := map[string]any{
		"action_key":      key,
		"label":           stringValue(action["label"]),
		"action_type_key": stringValue(action["action_type_key"]),
		"config":          action["config"],
	}
	replaced := false
	for i, item := range defNodes {
		node, _ := item.(map[string]any)
		if stringValue(node["action_key"]) == key {
			defNodes[i] = nextDefNode
			replaced = true
		}
	}
	if !replaced {
		defNodes = append(defNodes, nextDefNode)
	}
	def["nodes"] = defNodes
	graphNodes, _ := graph["nodes"].([]any)
	pos := map[string]any{"x": 80 + (len(graphNodes)%3)*300, "y": 160 + (len(graphNodes)/3)*170}
	if ui, ok := action["ui"].(map[string]any); ok {
		if p, ok := ui["position"].(map[string]any); ok {
			pos = p
		}
	}
	nextGraphNode := map[string]any{
		"id":       key,
		"type":     "action",
		"position": pos,
		"data": map[string]any{
			"label":           nextDefNode["label"],
			"action_type_key": nextDefNode["action_type_key"],
			"config":          nextDefNode["config"],
		},
	}
	replaced = false
	for i, item := range graphNodes {
		node, _ := item.(map[string]any)
		if stringValue(node["id"]) == key {
			graphNodes[i] = nextGraphNode
			replaced = true
		}
	}
	if !replaced {
		graphNodes = append(graphNodes, nextGraphNode)
	}
	graph["nodes"] = graphNodes
}

func addWorkflowConnection(def, graph map[string]any, from, to, branch string) {
	defEdges, _ := def["edges"].([]any)
	exists := false
	for _, item := range defEdges {
		edge, _ := item.(map[string]any)
		if stringValue(edge["from"]) == from && stringValue(edge["to"]) == to && stringValue(edge["label"]) == branch {
			exists = true
			break
		}
	}
	if !exists {
		defEdges = append(defEdges, map[string]any{"from": from, "to": to, "label": branch})
	}
	def["edges"] = defEdges
	graphEdges, _ := graph["edges"].([]any)
	exists = false
	for _, item := range graphEdges {
		edge, _ := item.(map[string]any)
		if stringValue(edge["source"]) == from && stringValue(edge["target"]) == to && stringValue(edge["sourceHandle"]) == branch {
			exists = true
			break
		}
	}
	if !exists {
		graphEdges = append(graphEdges, map[string]any{
			"id":           fmt.Sprintf("ai-edge-%d-%s-%s", len(graphEdges), from, to),
			"source":       from,
			"target":       to,
			"sourceHandle": branch,
			"label":        branch,
			"markerEnd":    map[string]any{"type": "arrowclosed"},
		})
	}
	graph["edges"] = graphEdges
}
