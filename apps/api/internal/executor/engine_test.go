package executor

import (
	"context"
	"encoding/json"
	"testing"

	"riders.agency/pulse/api/internal/runtime/contracts"
	runtimeregistry "riders.agency/pulse/api/internal/runtime/registry"
)

type captureRuntimeAction struct {
	req contracts.ActionRequest
}

func (a *captureRuntimeAction) TypeKey() string {
	return "custom.fake"
}

func (a *captureRuntimeAction) ValidateConfig(context.Context, json.RawMessage) error {
	return nil
}

func (a *captureRuntimeAction) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	a.req = req
	return contracts.ActionResult{
		InputJSON:            req.RunInputJSON,
		OutputJSON:           json.RawMessage(`{"ok":true,"source":"registry"}`),
		ProviderResponseJSON: json.RawMessage(`{"provider":"fake"}`),
	}
}

func TestExecuteActionUsesRegisteredRuntimeHandler(t *testing.T) {
	reg := runtimeregistry.New()
	handler := &captureRuntimeAction{}
	if err := reg.RegisterAction(handler); err != nil {
		t.Fatalf("register action failed: %v", err)
	}
	engine := &Engine{Registry: reg}
	result := engine.executeAction(context.Background(), actionPlan{
		ActionKey:     "custom_1",
		ActionTypeKey: "custom.fake",
		Config: actionConfig{
			Text:           "hello",
			TimeoutSeconds: 15,
		},
	}, []byte(`{"input":true}`), map[string]any{"prev": "ok"}, "run-1", "project-1", 2)

	if result.Err != nil {
		t.Fatalf("executeAction returned error: %v", result.Err)
	}
	if result.OutputJSON != `{"ok":true,"source":"registry"}` {
		t.Fatalf("unexpected output: %s", result.OutputJSON)
	}
	if handler.req.WorkflowRunID != "run-1" || handler.req.ProjectID != "project-1" {
		t.Fatalf("unexpected request ids: %#v", handler.req)
	}
	if handler.req.ActionKey != "custom_1" || handler.req.ActionTypeKey != "custom.fake" || handler.req.Attempt != 2 {
		t.Fatalf("unexpected action request: %#v", handler.req)
	}
	if string(handler.req.RunInputJSON) != `{"input":true}` {
		t.Fatalf("unexpected run input: %s", handler.req.RunInputJSON)
	}
	if handler.req.Steps["prev"] != "ok" {
		t.Fatalf("unexpected steps: %#v", handler.req.Steps)
	}
	if handler.req.Deadline.IsZero() {
		t.Fatal("expected timeout deadline to be set")
	}
	var cfg map[string]any
	if err := json.Unmarshal(handler.req.ConfigJSON, &cfg); err != nil {
		t.Fatalf("config json invalid: %v", err)
	}
	if cfg["text"] != "hello" {
		t.Fatalf("unexpected config json: %#v", cfg)
	}
}

func TestExecuteActionFailsUnregisteredActionType(t *testing.T) {
	engine := &Engine{Registry: runtimeregistry.New()}
	result := engine.executeAction(context.Background(), actionPlan{
		ActionKey:     "legacy_1",
		ActionTypeKey: "unknown.action",
	}, []byte(`{"input":true}`), nil, "run-1", "project-1", 1)

	if result.Err == nil {
		t.Fatal("expected missing handler error")
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &output); err != nil {
		t.Fatalf("output json invalid: %v", err)
	}
	if output["diagnostics"] == nil {
		t.Fatalf("expected diagnostics in missing handler output: %#v", output)
	}
}

func TestNewRegistersCoreRuntimeHandlers(t *testing.T) {
	engine := New(nil)
	for _, typeKey := range []string{"condition.branch", "stop", "delay.wait", "http.request"} {
		if _, ok := engine.Registry.Action(typeKey); !ok {
			t.Fatalf("expected core handler registered for %s", typeKey)
		}
	}
}

func TestNewRegistersExecutorOwnedRuntimeActions(t *testing.T) {
	engine := New(nil)
	for _, typeKey := range []string{
		"bot.set_state",
	} {
		if _, ok := engine.Registry.Action(typeKey); !ok {
			t.Fatalf("expected executor-owned handler registered for %s", typeKey)
		}
	}
	for _, typeKey := range []string{
		"telegram.send_message",
		"max.send_message",
		"email.send",
		"uon.create_lead",
		"clients.manual_payment.trigger",
		"clients.payment_due.trigger",
		"clients.payment_cycle.collect",
	} {
		if _, ok := engine.Registry.Action(typeKey); ok {
			t.Fatalf("expected module-owned handler not to be registered by executor for %s", typeKey)
		}
	}
}

func TestNewRegistersTriggerRuntimeHandlers(t *testing.T) {
	engine := New(nil)
	for _, typeKey := range []string{
		"schedule.trigger",
		"telegram.relay.trigger",
		"module.event.trigger",
	} {
		if _, ok := engine.Registry.Trigger(typeKey); !ok {
			t.Fatalf("expected trigger handler registered for %s", typeKey)
		}
	}
	if _, ok := engine.Registry.Trigger("clients.payment_due.trigger"); ok {
		t.Fatal("expected client payment trigger to be module-owned, not executor-registered")
	}
}

func TestMissingHandlerDoesNotExecuteProviderActions(t *testing.T) {
	engine := &Engine{Registry: runtimeregistry.New()}
	result := engine.executeAction(context.Background(), actionPlan{
		ActionKey:     "notify",
		ActionTypeKey: "telegram.send_message",
	}, []byte(`{"input":true}`), nil, "run-1", "project-1", 1)

	if result.Err == nil {
		t.Fatal("expected provider action without registered handler to fail")
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &output); err != nil {
		t.Fatalf("output json invalid: %v", err)
	}
	if output["diagnostics"] == nil {
		t.Fatalf("expected diagnostics, got %#v", output)
	}
}

func TestPlanRunNodesThroughActionSkipsDownstream(t *testing.T) {
	nodes := []actionPlan{
		{ActionKey: "trigger", ActionTypeKey: "webhook.trigger"},
		{ActionKey: "send", ActionTypeKey: "telegram.send_message"},
		{ActionKey: "stop", ActionTypeKey: "stop"},
	}
	input := []byte(`{"_pulse":{"test_action_key":"send","test_mode":"through_action"}}`)
	plan, err := planRunNodes(nodes, input)
	if err != nil {
		t.Fatalf("planRunNodes failed: %v", err)
	}
	if len(plan.Execute) != 2 || plan.Execute[1].ActionKey != "send" {
		t.Fatalf("unexpected execute nodes: %#v", plan.Execute)
	}
	if len(plan.Skipped) != 1 || plan.Skipped[0].ActionKey != "stop" {
		t.Fatalf("unexpected skipped nodes: %#v", plan.Skipped)
	}
}

func TestPlanRunNodesSelectedOnlyUsesStepOverrides(t *testing.T) {
	nodes := []actionPlan{
		{ActionKey: "one", ActionTypeKey: "webhook.trigger"},
		{ActionKey: "two", ActionTypeKey: "max.send_message"},
	}
	input := []byte(`{"_pulse":{"test_action_key":"two","test_mode":"selected_only","steps":{"one":{"message_id":"m1"}}}}`)
	plan, err := planRunNodes(nodes, input)
	if err != nil {
		t.Fatalf("planRunNodes failed: %v", err)
	}
	if len(plan.Execute) != 1 || plan.Execute[0].ActionKey != "two" {
		t.Fatalf("unexpected execute nodes: %#v", plan.Execute)
	}
	if got := renderTemplate("prev={{ steps.one.message_id }}", input, plan.InitialSteps, templateSystem{}); got != "prev=m1" {
		t.Fatalf("unexpected step template render: %q", got)
	}
}

func TestReplaySingleActionBypassesSatisfiedGraphPredecessors(t *testing.T) {
	definition := []byte(`{"nodes":[{"action_key":"trigger_1","action_type_key":"webhook.trigger"},{"action_key":"uon_create_lead","action_type_key":"uon.create_lead"}],"edges":[{"from":"trigger_1","to":"uon_create_lead"}]}`)
	engine := &Engine{}
	full, err := engine.BuildRunPlan(context.Background(), definition, []byte(`{"webhook":{}}`), "")
	if err != nil {
		t.Fatal(err)
	}
	if !full.UseGraphEdges {
		t.Fatal("full workflow must follow graph edges")
	}
	replay, err := engine.BuildRunPlan(context.Background(), definition, []byte(`{"webhook":{},"_pulse":{"retry_action_key":"uon_create_lead"}}`), "")
	if err != nil {
		t.Fatal(err)
	}
	if replay.UseGraphEdges || len(replay.Execute) != 1 || replay.Execute[0].ActionKey != "uon_create_lead" {
		t.Fatalf("single-action replay must execute the selected node: %#v", replay)
	}
}

func TestFlattenNodesExplicitEmptyWorkflowHasNoImplicitWebhook(t *testing.T) {
	nodes := flattenNodes([]byte(`{"nodes":[],"edges":[]}`))
	if len(nodes) != 0 {
		t.Fatalf("expected no implicit nodes for explicit empty workflow, got %#v", nodes)
	}
}

func TestFlattenNodesLegacyMissingNodesKeepsWebhookFallback(t *testing.T) {
	nodes := flattenNodes([]byte(`{}`))
	if len(nodes) != 1 || nodes[0].ActionKey != "entry" || nodes[0].ActionTypeKey != "webhook.trigger" {
		t.Fatalf("expected legacy webhook fallback, got %#v", nodes)
	}
}

func TestMergeActionExposeFeedsStepTemplates(t *testing.T) {
	steps := map[string]any{}
	out, _ := json.Marshal(map[string]any{"expose": map[string]any{"message_id": "tg-1"}})
	mergeActionExpose(steps, "notify", string(out))
	got := renderTemplate("sent={{ steps.notify.message_id }}", []byte(`{}`), steps, templateSystem{})
	if got != "sent=tg-1" {
		t.Fatalf("unexpected template render: %q", got)
	}
}

func TestRenderTemplateSystemRunURL(t *testing.T) {
	got := renderTemplate(
		"run={{ system.run_url }} id={{ run.id }}",
		[]byte(`{"order_id":"1001"}`),
		nil,
		templateSystem{RunID: "run-1", RunURL: "https://pulse.riders.agency/runs/run-1"},
	)
	if got != "run=https://pulse.riders.agency/runs/run-1 id=run-1" {
		t.Fatalf("unexpected system template render: %q", got)
	}
}

func TestRenderTemplateWebhookSlugNamespace(t *testing.T) {
	input := []byte(`{"_pulse":{"webhook_slug":"orders_paid"},"webhook":{"id":"1001","name":"Order 1001","total":12500,"customer":{"email":"client@example.com"}}}`)
	got := renderTemplate(
		"order={{ orders_paid.id }} name={{ orders_paid.name }} total={{ orders_paid.total }} email={{ orders_paid.customer.email }} link={{ system.link }} id={{ system.id }}",
		input,
		nil,
		templateSystem{RunID: "run-1", RunURL: "https://pulse.riders.agency/runs/run-1"},
	)
	want := "order=1001 name=Order 1001 total=12500 email=client@example.com link=https://pulse.riders.agency/runs/run-1 id=run-1"
	if got != want {
		t.Fatalf("unexpected slug render: got %q want %q", got, want)
	}
}

func TestRenderTemplateWebhookSlugNamespaceFlatPayload(t *testing.T) {
	input := []byte(`{"_pulse":{"webhook_slug":"bitrix-order-to-max"},"webhook":{"order_id":"123","price":1000,"currency":"RUB"}}`)
	got := renderTemplate(
		"order={{ bitrix-order-to-max.order_id }} total={{ bitrix-order-to-max.price }} {{ bitrix-order-to-max.currency }}",
		input,
		nil,
		templateSystem{},
	)
	want := "order=123 total=1000 RUB"
	if got != want {
		t.Fatalf("unexpected flat slug render: got %q want %q", got, want)
	}
}

func TestRenderTemplateJSONValueEscapesWebhookText(t *testing.T) {
	input := []byte(`{"telegram_relay":{"text":"Авария \"центр\"\nпроверка"}}`)
	got := renderTemplate(
		`{"text":{{ json telegram_relay.text }}}`,
		input,
		nil,
		templateSystem{},
	)
	if got != `{"text":"Авария \"центр\"\nпроверка"}` {
		t.Fatalf("unexpected json template render: %q", got)
	}
}

func TestShouldSkipTelegramBotReplyToChannel(t *testing.T) {
	channelInput := []byte(`{"telegram_relay":{"chat":{"type":"channel","id":-1001602381525},"update_type":"channel_post"}}`)
	groupInput := []byte(`{"telegram_relay":{"chat":{"type":"supergroup","id":-1002857094690},"update_type":"message"}}`)
	webhookInput := []byte(`{"webhook":{"order_id":"1"}}`)

	if !shouldSkipTelegramBotReplyToChannel(channelInput) {
		t.Fatal("expected skip for channel relay reply")
	}
	if shouldSkipTelegramBotReplyToChannel(groupInput) {
		t.Fatal("expected send for group relay reply")
	}
	if shouldSkipTelegramBotReplyToChannel(webhookInput) {
		t.Fatal("expected webhook sends unaffected")
	}
}

func TestRenderTemplateDoesNotHardcodeFlatAliases(t *testing.T) {
	input := []byte(`{"order_id":"123","price":1000}`)
	got := renderTemplate("{{ input.order.id }}|{{ input.order.total }}", input, nil, templateSystem{})
	if got != "|" {
		t.Fatalf("unexpected hardcoded alias render: %q", got)
	}
}
