package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/httptransport"
)

type Scenario struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Route string `json:"route"`
	Turns []Turn `json:"turns"`
}

type Turn struct {
	Route  string     `json:"route,omitempty"`
	User   string     `json:"user"`
	Expect TurnExpect `json:"expect"`
}

type TurnExpect struct {
	IntentKey         string   `json:"intent_key,omitempty"`
	ProposalGate      string   `json:"proposal_gate,omitempty"`
	ToolKeys          []string `json:"tool_keys,omitempty"`
	DeniedToolKeys    []string `json:"denied_tool_keys,omitempty"`
	CardType          string   `json:"card_type,omitempty"`
	AssistantLike     string   `json:"assistant_like,omitempty"`
	AssistantExcludes []string `json:"assistant_excludes,omitempty"`
	NoSideEffects     bool     `json:"no_side_effects,omitempty"`
	ApprovalRequired  bool     `json:"approval_required,omitempty"`
}

func main() {
	baseURL := flag.String("base-url", env("PULSE_BASE_URL", "http://127.0.0.1:9810"), "Go API base URL")
	projectID := flag.String("project-id", os.Getenv("PULSE_PROJECT_ID"), "project id")
	token := flag.String("token", os.Getenv("PULSE_AUTH_TOKEN"), "bearer token")
	cookie := flag.String("cookie", os.Getenv("PULSE_AUTH_COOKIE"), "raw Cookie header")
	run := flag.Bool("run", false, "execute scenarios through HTTP API")
	jsonOut := flag.Bool("json", false, "print scenario catalog as JSON")
	limit := flag.Int("limit", 30, "maximum scenarios to execute")
	flag.Parse()

	scenarios := scenarioCatalog()
	if *jsonOut {
		writeJSON(scenarios)
		return
	}
	if len(scenarios) != 30 {
		fail("scenario catalog must contain exactly 30 scenarios, got %d", len(scenarios))
	}
	if err := validateScenarioCatalog(scenarios); err != nil {
		fail("scenario catalog invalid: %v", err)
	}
	if !*run {
		fmt.Printf("AI scenario catalog valid: scenarios=%d\n", len(scenarios))
		return
	}
	if strings.TrimSpace(*projectID) == "" {
		fail("--project-id or PULSE_PROJECT_ID is required")
	}
	if strings.TrimSpace(*token) == "" && strings.TrimSpace(*cookie) == "" {
		fail("--token/PULSE_AUTH_TOKEN or --cookie/PULSE_AUTH_COOKIE is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := runScenarios(ctx, *baseURL, *projectID, *token, *cookie, scenarios[:min(*limit, len(scenarios))]); err != nil {
		fail("%v", err)
	}
}

func validateScenarioCatalog(scenarios []Scenario) error {
	seen := map[string]struct{}{}
	required := map[string]struct{}{
		"scenario_01": {},
		"scenario_03": {},
		"scenario_13": {},
		"scenario_22": {},
		"scenario_30": {},
	}
	for _, scenario := range scenarios {
		if strings.TrimSpace(scenario.ID) == "" {
			return fmt.Errorf("empty scenario id")
		}
		if _, ok := seen[scenario.ID]; ok {
			return fmt.Errorf("duplicate scenario id %s", scenario.ID)
		}
		seen[scenario.ID] = struct{}{}
		delete(required, scenario.ID)
		if strings.TrimSpace(scenario.Route) == "" {
			return fmt.Errorf("%s route is required", scenario.ID)
		}
		if len(scenario.Turns) == 0 {
			return fmt.Errorf("%s has no turns", scenario.ID)
		}
		for index, turn := range scenario.Turns {
			if strings.TrimSpace(turn.User) == "" {
				return fmt.Errorf("%s turn %d user text is required", scenario.ID, index+1)
			}
			if strings.TrimSpace(turn.Expect.IntentKey) == "" {
				return fmt.Errorf("%s turn %d expected intent_key is required", scenario.ID, index+1)
			}
			if !containsStringFold(turn.Expect.AssistantExcludes, "Не увидел точное изменение") {
				return fmt.Errorf("%s turn %d must guard stale proposal-edit copy", scenario.ID, index+1)
			}
		}
	}
	if len(required) > 0 {
		keys := make([]string, 0, len(required))
		for key := range required {
			keys = append(keys, key)
		}
		return fmt.Errorf("required scenarios missing: %s", strings.Join(keys, ","))
	}
	return nil
}

func runScenarios(ctx context.Context, baseURL, projectID, token, cookie string, scenarios []Scenario) error {
	client := httptransport.NewDefaultClient(90 * time.Second)
	for _, scenario := range scenarios {
		sessionID, err := createSession(ctx, client, baseURL, projectID, token, cookie, scenario.Route)
		if err != nil {
			return fmt.Errorf("%s create session: %w", scenario.ID, err)
		}
		activeToolCallID := ""
		for index, turn := range scenario.Turns {
			route := firstNonEmpty(turn.Route, scenario.Route)
			events, err := postTurn(ctx, client, baseURL, projectID, sessionID, token, cookie, route, activeToolCallID, turn.User)
			if err != nil {
				return fmt.Errorf("%s turn %d: %w", scenario.ID, index+1, err)
			}
			if err := assertTurn(turn.Expect, events); err != nil {
				return fmt.Errorf(
					"%s turn %d active_tool_call_id=%q events=%s: %w",
					scenario.ID,
					index+1,
					activeToolCallID,
					truncate(compactCurrentTurnEvents(events), 6000),
					err,
				)
			}
			activeToolCallID = activeToolCallIDFromEvents(events, activeToolCallID)
			if activeToolCallID != "" && !isLikelyUUID(activeToolCallID) {
				activeToolCallID = ""
			}
			if activeToolCallID == "" {
				restored, restoreErr := getSessionActiveToolCallID(ctx, client, baseURL, projectID, sessionID, token, cookie)
				if restoreErr != nil {
					return fmt.Errorf("%s turn %d restore session active tool call: %w", scenario.ID, index+1, restoreErr)
				}
				activeToolCallID = restored
			}
		}
		fmt.Printf("PASS %s %s\n", scenario.ID, scenario.Title)
	}
	return nil
}

func createSession(ctx context.Context, client *http.Client, baseURL, projectID, token, cookie, route string) (string, error) {
	payload := map[string]any{
		"screen_scope": "scenario_runner",
		"page_context_json": map[string]any{
			"route":     route,
			"pathname":  strings.Split(route, "?")[0],
			"pageKind":  "",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	var out map[string]any
	if err := requestJSON(ctx, client, http.MethodPost, baseURL+"/api/v1/projects/"+projectID+"/ai/employee/sessions", token, cookie, payload, &out); err != nil {
		return "", err
	}
	id, _ := out["ai_chat_session_id"].(string)
	if id == "" {
		return "", fmt.Errorf("session id missing")
	}
	return id, nil
}

func getSessionActiveToolCallID(ctx context.Context, client *http.Client, baseURL, projectID, sessionID, token, cookie string) (string, error) {
	var out map[string]any
	if err := requestJSON(ctx, client, http.MethodGet, baseURL+"/api/v1/projects/"+projectID+"/ai/employee/sessions/"+sessionID, token, cookie, nil, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(stringFromAny(out["active_tool_call_id"])), nil
}

func postTurn(ctx context.Context, client *http.Client, baseURL, projectID, sessionID, token, cookie, route, activeToolCallID, body string) ([]map[string]any, error) {
	payload := map[string]any{
		"body":                body,
		"active_tool_call_id": activeToolCallID,
		"page_context_json": map[string]any{
			"route":     route,
			"pathname":  strings.Split(route, "?")[0],
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
		"client_context_json": map[string]any{"source_surface": "scenario_runner", "active_tool_call_id": activeToolCallID},
		"ui_state_json":       map[string]any{"open": true, "pathname": strings.Split(route, "?")[0]},
		"local_date":          "2026-06-15",
		"timezone":            "Europe/Moscow",
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/projects/"+projectID+"/ai/employee/sessions/"+sessionID+"/messages/stream", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	authorize(req, token, cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, fmt.Errorf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return readSSE(resp.Body), nil
}

func assertTurn(expect TurnExpect, events []map[string]any) error {
	joined := compactEvents(events)
	currentToolKeys := currentTurnToolKeys(events)
	hasApprovalRequired := currentTurnHasEventType(events, "APPROVAL_REQUIRED") || currentTurnHasEventType(events, "approval_required")
	if expect.IntentKey != "" && !strings.Contains(joined, "intent_key="+expect.IntentKey) {
		return fmt.Errorf("intent_key %q not found", expect.IntentKey)
	}
	if expect.ProposalGate != "" && !strings.Contains(joined, "proposal_gate="+expect.ProposalGate) {
		return fmt.Errorf("proposal_gate %q not found", expect.ProposalGate)
	}
	if expect.AssistantLike != "" && !strings.Contains(strings.ToLower(joined), strings.ToLower(expect.AssistantLike)) {
		return fmt.Errorf("assistant_like %q not found", expect.AssistantLike)
	}
	for _, denied := range expect.AssistantExcludes {
		if strings.Contains(strings.ToLower(joined), strings.ToLower(denied)) {
			return fmt.Errorf("assistant_excludes %q was found", denied)
		}
	}
	if expect.CardType != "" && !strings.Contains(joined, `"card_type":"`+expect.CardType+`"`) {
		return fmt.Errorf("card_type %q not found", expect.CardType)
	}
	for _, tool := range expect.ToolKeys {
		if !currentToolKeys[tool] {
			return fmt.Errorf("tool %q not found", tool)
		}
	}
	for _, tool := range expect.DeniedToolKeys {
		if currentToolKeys[tool] {
			return fmt.Errorf("denied tool %q was called", tool)
		}
	}
	if expect.ApprovalRequired && !hasApprovalRequired {
		return fmt.Errorf("approval event not found")
	}
	if !expect.ApprovalRequired && hasApprovalRequired {
		return fmt.Errorf("unexpected approval-required event")
	}
	if expect.NoSideEffects && currentTurnHasWriteRisk(events) {
		return fmt.Errorf("read-only turn emitted write/critical risk")
	}
	return nil
}

func scenarioCatalog() []Scenario {
	return []Scenario{
		s("01", "phatic no-tool", "/tasks/boards", t("ты тут?", e("conversation.smalltalk", "", nil, []string{"task.create"}, "", "на связи", true, false))),
		s("02", "current location", "/tasks/boards", t("Где я сейчас?", e("page.where_am_i", "", []string{"ui_context.read"}, []string{"task.create"}, "", "экран", true, false))),
		s("03", "simple task create from board", "/tasks/boards", t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, []string{"task.update"}, "task_create_proposal", "", false, true))),
		s("04", "approve created task", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("да, создавай", e("approval.apply", "approval_bridge", []string{"approval.apply"}, []string{"task.create"}, "", "создан", false, false)),
		),
		s("05", "reject created task", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("не надо, отмена", e("approval.reject", "approval_bridge", []string{"approval.reject"}, []string{"task.create.commit"}, "", "отмен", true, false)),
		),
		s("06", "pending draft edit survives next turn", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("на сегодня", e("proposal.edit", "patch_current", []string{"proposal.edit"}, []string{"task.create"}, "", "сегодня", false, true)),
		),
		s("07", "route switch while approval pending", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			rt("/clients", "Найди клиента Орион", e("client.search", "run_new_keep_pending", []string{"client.search"}, []string{"proposal.edit", "task.create"}, "", "Орион", true, false)),
		),
		s("08", "explicit create ignores selected card", "/tasks/boards", t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"task.create"}, []string{"task.update"}, "task_create_proposal", "", false, true))),
		s("09", "deictic update uses selected card", "/tasks/00000000-0000-0000-0000-000000000001", t("Запусти эту задачу", e("task.move", "", []string{"task.move"}, []string{"task.create"}, "task_move_proposal", "", false, true))),
		s("10", "deictic update without selected target", "/tasks/boards", t("Перенеси ее на завтра", e("unknown", "", nil, []string{"task.move"}, "", "уточ", true, false))),
		s("11", "continue previous context priority", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("сделай приоритет высокий", e("proposal.edit", "patch_current", []string{"proposal.edit"}, []string{"task.create"}, "", "высок", true, true)),
		),
		s("12", "pending create to new explicit create", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("Нет, лучше создай отдельную задачу позвонить Орион в пятницу", e("task.create", "conflict_choice", nil, []string{"proposal.edit"}, "conflict_choice", "", true, false)),
		),
		s("13", "mixed multi-action sequential", "/tasks/boards", t("Создай клиента Орион и поставь ему задачу подготовить документы для начала сотрудничества и стартани эту задачу", e("multi_step.plan", "", []string{"multi_step.plan"}, nil, "multi_step_plan", "", false, true))),
		s("14", "read-only parallel tools", "/clients", t("Покажи что у Орион по задачам и оплатам", e("project.explain", "", []string{"task.search", "payment.lookup_due"}, []string{"task.create"}, "", "Орион", true, false))),
		s("15", "partial client query stays proposal-only", "/tasks/boards", t("Создай задачу проверить оплату Эдвен завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "Эдвен", true, true))),
		s("16", "client alias memory", "/tasks/boards", t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search"}, nil, "task_create_proposal", "Орион", false, true))),
		s("17", "temporal today", "/tasks/boards", t("Создай задачу проверить оплату Орион сегодня", e("task.create", "", []string{"task.create"}, nil, "task_create_proposal", "15.06.2026", false, true))),
		s("18", "temporal tomorrow", "/tasks/boards", t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"task.create"}, nil, "task_create_proposal", "16.06.2026", false, true))),
		s("19", "temporal next monday", "/tasks/boards", t("Создай задачу созвониться с Орион в следующий понедельник", e("task.create", "", []string{"task.create"}, nil, "task_create_proposal", "понедельник", false, true))),
		s("20", "read task search without write", "/tasks/boards", t("Покажи большой отчет по задачам", e("task.search", "", []string{"task.search"}, []string{"task.create"}, "", "", true, false))),
		s("21", "payment lookup read-only", "/clients", t("Когда оплата у Орион?", e("payment.lookup_due", "", []string{"payment.lookup_due"}, []string{"task.create"}, "", "оплат", true, false))),
		s("22", "workflow URL question stays read", "/workflows/cf59e001-edb7-4e95-9efb-ea1248215ab1/edit?action=manual_payment_trigger", t("сколько здесь действий https://pulse.riders.agency/workflows/cf59e001-edb7-4e95-9efb-ea1248215ab1/edit?action=manual_payment_trigger", e("workflow.answer_question", "", []string{"workflow.read_draft"}, []string{"workflow.patch"}, "", "действ", true, false))),
		s("23", "duplicate approve protection through same active proposal", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("да, создавай", e("approval.apply", "approval_bridge", []string{"approval.apply"}, nil, "", "создан", false, false)),
			t("да, создавай", e("unknown", "", nil, []string{"task.create", "approval.apply"}, "", "Уточните", true, false)),
		),
		s("24", "conversation memory phrase does not create write", "/tasks/boards", t("Запомни, что сейчас работаем с Орион", e("unknown", "", nil, []string{"task.create"}, "", "Уточните", true, false))),
		s("25", "context switch to another client read", "/tasks/boards", t("Найди клиента Ромашка", e("client.search", "", []string{"client.search"}, []string{"task.create"}, "", "Ромашка", true, false))),
		s("26", "page context after conversation", "/tasks/boards",
			t("ты тут?", e("conversation.smalltalk", "", nil, []string{"task.create"}, "", "на связи", true, false)),
			t("Где я сейчас?", e("page.where_am_i", "", []string{"ui_context.read"}, []string{"task.create"}, "", "экран", true, false)),
		),
		s("27", "approval natural language", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("да, создавай", e("approval.apply", "approval_bridge", []string{"approval.apply"}, nil, "", "создан", false, false)),
		),
		s("28", "natural language rejection", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("не надо, отмена", e("approval.reject", "approval_bridge", []string{"approval.reject"}, []string{"task.create.commit"}, "", "отмен", true, false)),
		),
		s("29", "update pending draft natural language", "/tasks/boards",
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("срок поставь на пятницу, приоритет низкий", e("proposal.edit", "patch_current", []string{"proposal.edit"}, []string{"task.create"}, "", "низкий", true, true)),
		),
		s("30", "long sequential mixed session", "/tasks/boards",
			t("ты тут?", e("conversation.smalltalk", "", nil, []string{"task.create"}, "", "на связи", true, false)),
			t("Где я сейчас?", e("page.where_am_i", "", []string{"ui_context.read"}, []string{"task.create"}, "", "экран", true, false)),
			t("Создай задачу проверить оплату Орион завтра", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "", false, true)),
			t("сделай приоритет высокий", e("proposal.edit", "patch_current", []string{"proposal.edit"}, []string{"task.create"}, "", "высок", true, true)),
			t("да, создавай", e("approval.apply", "approval_bridge", []string{"approval.apply"}, nil, "", "создан", false, false)),
			rt("/clients", "Создай задачу позвонить Орион в пятницу", e("task.create", "", []string{"client.search", "task.create"}, nil, "task_create_proposal", "пятницу", false, true)),
			t("не надо, отмена", e("approval.reject", "approval_bridge", []string{"approval.reject"}, nil, "", "отмен", true, false)),
			t("Что у Орион осталось по задачам?", e("task.search", "", []string{"task.search"}, []string{"task.create"}, "", "Орион", true, false)),
		),
	}
}

func s(id, title, route string, turns ...Turn) Scenario {
	return Scenario{ID: "scenario_" + id, Title: title, Route: route, Turns: turns}
}

func t(user string, expect TurnExpect) Turn {
	return Turn{User: user, Expect: expect}
}

func rt(route, user string, expect TurnExpect) Turn {
	return Turn{Route: route, User: user, Expect: expect}
}

func e(intent, gate string, tools, denied []string, card, like string, noSideEffects, approval bool) TurnExpect {
	return TurnExpect{
		IntentKey:         intent,
		ProposalGate:      gate,
		ToolKeys:          tools,
		DeniedToolKeys:    denied,
		CardType:          card,
		AssistantLike:     like,
		AssistantExcludes: []string{"Не увидел точное изменение"},
		NoSideEffects:     noSideEffects,
		ApprovalRequired:  approval,
	}
}

func requestJSON(ctx context.Context, client *http.Client, method, urlValue, token, cookie string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, _ := json.Marshal(payload)
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, urlValue, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	authorize(req, token, cookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return json.Unmarshal(responseBody, out)
}

func authorize(req *http.Request, token, cookie string) {
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	if strings.TrimSpace(cookie) != "" {
		req.Header.Set("Cookie", strings.TrimSpace(cookie))
	}
}

func readSSE(reader io.Reader) []map[string]any {
	scanner := bufio.NewScanner(reader)
	events := make([]map[string]any, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var event map[string]any
		if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &event) == nil {
			events = append(events, event)
		}
	}
	return events
}

func compactEvents(events []map[string]any) string {
	raw, _ := json.Marshal(events)
	return string(raw)
}

func compactCurrentTurnEvents(events []map[string]any) string {
	current := make([]map[string]any, 0, len(events))
	for _, event := range events {
		if _, ok := event["session"]; ok {
			continue
		}
		current = append(current, event)
	}
	raw, _ := json.Marshal(current)
	return string(raw)
}

func currentTurnToolKeys(events []map[string]any) map[string]bool {
	out := map[string]bool{}
	for _, event := range events {
		if _, ok := event["session"]; ok {
			continue
		}
		if toolKey := strings.TrimSpace(stringFromAny(event["tool_key"])); toolKey != "" {
			out[toolKey] = true
		}
		if toolCall, ok := event["tool_call"].(map[string]any); ok {
			if toolKey := strings.TrimSpace(stringFromAny(toolCall["tool_key"])); toolKey != "" {
				out[toolKey] = true
			}
		}
	}
	return out
}

func currentTurnHasEventType(events []map[string]any, eventType string) bool {
	for _, event := range events {
		if _, ok := event["session"]; ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(stringFromAny(event["type"])), eventType) {
			return true
		}
	}
	return false
}

func currentTurnHasWriteRisk(events []map[string]any) bool {
	for _, event := range events {
		if _, ok := event["session"]; ok {
			continue
		}
		if toolCall, ok := event["tool_call"].(map[string]any); ok {
			if !strings.Contains(strings.ToUpper(stringFromAny(event["type"])), "TOOL_STARTED") {
				continue
			}
			switch stringFromAny(toolCall["tool_key"]) {
			case "approval.reject", "proposal.edit":
				continue
			}
			if isWriteRisk(stringFromAny(toolCall["risk_level"])) {
				return true
			}
		}
	}
	return false
}

func isWriteRisk(risk string) bool {
	switch strings.ToLower(strings.TrimSpace(risk)) {
	case "high", "critical", "write":
		return true
	default:
		return false
	}
}

func activeToolCallIDFromEvents(events []map[string]any, fallback string) string {
	for i := len(events) - 1; i >= 0; i-- {
		session, ok := events[i]["session"].(map[string]any)
		if !ok {
			continue
		}
		if active, ok := session["active_tool_call_id"].(string); ok {
			return strings.TrimSpace(active)
		}
		if active, ok := session["active_tool_call_id"]; ok && active == nil {
			return ""
		}
	}
	for i := len(events) - 1; i >= 0; i-- {
		if card, ok := events[i]["card"].(map[string]any); ok {
			if id := strings.TrimSpace(stringFromAny(card["tool_call_id"])); id != "" {
				if isLikelyUUID(id) {
					return id
				}
			}
		}
		if id := strings.TrimSpace(stringFromAny(events[i]["tool_call_id"])); id != "" {
			if isLikelyUUID(id) {
				return id
			}
		}
	}
	return fallback
}

func isLikelyUUID(value string) bool {
	value = strings.TrimSpace(value)
	return len(value) == 36 && strings.Count(value, "-") == 4
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func containsStringFold(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}

func stringFromAny(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func truncate(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "...[truncated]"
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
