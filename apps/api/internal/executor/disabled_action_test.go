package executor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDisabledActionNeverCallsProvider(t *testing.T) {
	engine := New(nil)
	nodes := flattenNodes([]byte(`{"nodes":[{"action_key":"mail","action_type_key":"email.send","config":{"disabled":true,"to":"nobody@example.test"}}]}`))
	if len(nodes) != 1 || !nodes[0].Config.Disabled {
		t.Fatal("disabled flag lost")
	}
	result := engine.executeAction(context.Background(), nodes[0], []byte(`{}`), nil, "run", "project", 1)
	if result.Err != nil || !strings.Contains(result.OutputJSON, `"disabled"`) {
		t.Fatalf("did not skip: %+v", result)
	}
	plan, err := engine.BuildRunPlan(context.Background(), []byte(`{"nodes":[{"action_key":"mail","action_type_key":"email.send","config":{"disabled":true}}]}`), []byte(`{}`), "project")
	if err != nil || len(plan.Execute) != 1 || !plan.Execute[0].Disabled {
		t.Fatalf("kernel flag lost: %+v %v", plan, err)
	}
	var output map[string]any
	if err = json.Unmarshal([]byte(result.OutputJSON), &output); err != nil || output["skipped"] != true {
		t.Fatalf("skip output: %s", result.OutputJSON)
	}
}

func TestDisabledNodeContinuesUnconditionalEdgesOnly(t *testing.T) {
	edges := []edgePlan{{From: "disabled", To: "next"}, {From: "disabled", To: "branch", Label: "true"}}
	if !shouldRunGraphNode(actionPlan{ActionKey: "next"}, edges, map[string]bool{"disabled": true}, map[string]string{}) {
		t.Fatal("following action blocked")
	}
	if shouldRunGraphNode(actionPlan{ActionKey: "branch"}, edges, map[string]bool{"disabled": true}, map[string]string{}) {
		t.Fatal("disabled condition chose a branch")
	}
}
