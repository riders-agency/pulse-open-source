package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

func TestCoreHandlersExposeExpectedTypes(t *testing.T) {
	handlers := CoreHandlers(nil)
	want := []string{"condition.branch", "stop", "delay.wait", "webhook.trigger", "manual.trigger", "schedule.trigger", "telegram.relay.trigger", "module.event.trigger", "site_chat.visitor_first_seen.trigger"}
	if len(handlers) != len(want) {
		t.Fatalf("handler count = %d, want %d", len(handlers), len(want))
	}
	for i, wantType := range want {
		if err := validateRegisteredType(handlers[i], wantType); err != nil {
			t.Fatalf("handler[%d]: %v", i, err)
		}
	}
}

func TestTriggerActionHandlerUsesRunInputAndScheduleKey(t *testing.T) {
	handler := triggerActionHandler{typeKey: "schedule.trigger"}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		ActionKey:     "every_hour",
		ActionTypeKey: "schedule.trigger",
		ConfigJSON:    json.RawMessage(`{"schedule_key":"hourly"}`),
		RunInputJSON:  json.RawMessage(`{"_pulse":{"trigger":"schedule"}}`),
	})
	if string(result.InputJSON) != `{"_pulse":{"trigger":"schedule"}}` {
		t.Fatalf("unexpected trigger input: %s", result.InputJSON)
	}
	var output map[string]any
	if err := json.Unmarshal(result.OutputJSON, &output); err != nil {
		t.Fatalf("output json invalid: %v", err)
	}
	if output["scheduled"] != true || output["schedule_key"] != "hourly" {
		t.Fatalf("unexpected schedule output: %#v", output)
	}
}

func TestConditionBranchHandlerRendersAndBranches(t *testing.T) {
	handler := conditionBranchHandler{
		renderer: func(template string, _ json.RawMessage, _ map[string]any, _ string) string {
			if template != "{{ input.status }} == paid" {
				t.Fatalf("unexpected template: %q", template)
			}
			return "paid == paid"
		},
	}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		WorkflowRunID: "run-1",
		ActionKey:     "check_payment",
		ActionTypeKey: "condition.branch",
		ConfigJSON:    json.RawMessage(`{"text":"{{ input.status }} == paid"}`),
		RunInputJSON:  json.RawMessage(`{"status":"paid"}`),
	})
	if result.Err != nil {
		t.Fatalf("execute failed: %v", result.Err)
	}
	var output map[string]any
	if err := json.Unmarshal(result.OutputJSON, &output); err != nil {
		t.Fatalf("output json invalid: %v", err)
	}
	if output["result"] != true || output["branch"] != "true" || output["condition"] != "paid == paid" {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestConditionBranchHandlerContainsEmptyNeedleIsFalse(t *testing.T) {
	handler := conditionBranchHandler{}
	for _, expr := range []string{
		",1,2, contains ,",
		",1343131,663122289, contains ,,",
	} {
		result := handler.Execute(context.Background(), contracts.ActionRequest{
			ActionKey:     "condition",
			ActionTypeKey: "condition.branch",
			ConfigJSON:    json.RawMessage(fmt.Sprintf(`{"text":%q}`, expr)),
		})
		var output map[string]any
		if err := json.Unmarshal(result.OutputJSON, &output); err != nil {
			t.Fatalf("output json invalid: %v", err)
		}
		if output["result"] != false {
			t.Fatalf("contains with empty user id must be false: %q -> %#v", expr, output)
		}
	}
}

func TestConditionBranchHandlerSupportsEqualityContainsAndNumbers(t *testing.T) {
	handler := conditionBranchHandler{}
	cases := map[string]bool{
		"alert == alert":           true,
		"alert != start":           true,
		",100,200, contains ,200,": true,
		",100,200, contains ,20,":  false,
		"12 >= 10":                 true,
		"12 < 10":                  false,
		"":                         false,
	}
	for expr, want := range cases {
		result := handler.Execute(context.Background(), contracts.ActionRequest{
			ActionKey:     "condition",
			ActionTypeKey: "condition.branch",
			ConfigJSON:    json.RawMessage(fmt.Sprintf(`{"text":%q}`, expr)),
		})
		var output map[string]any
		if err := json.Unmarshal(result.OutputJSON, &output); err != nil {
			t.Fatalf("output json invalid: %v", err)
		}
		if output["result"] != want {
			t.Fatalf("condition result for %q = %#v, want %v", expr, output["result"], want)
		}
	}
}

func TestStaticCoreHandlerMatchesLegacyMockShape(t *testing.T) {
	handler := staticCoreHandler{typeKey: "stop", outputKey: "stopped"}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		ActionKey:     "stop_1",
		ActionTypeKey: "stop",
		ConfigJSON:    json.RawMessage(`{"success_criteria":[{"kind":"ok"}]}`),
	})
	var output map[string]any
	if err := json.Unmarshal(result.OutputJSON, &output); err != nil {
		t.Fatalf("output json invalid: %v", err)
	}
	if output["mock"] != true || output["stopped"] != true || output["action_key"] != "stop_1" {
		t.Fatalf("unexpected output: %#v", output)
	}
	var input map[string]any
	if err := json.Unmarshal(result.InputJSON, &input); err != nil {
		t.Fatalf("input json invalid: %v", err)
	}
	if _, ok := input["config"].(map[string]any); !ok {
		t.Fatalf("expected config object in input: %#v", input)
	}
	if input["success_criteria"] == nil {
		t.Fatalf("expected success_criteria in input: %#v", input)
	}
}
