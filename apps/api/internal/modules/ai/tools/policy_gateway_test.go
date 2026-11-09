package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPolicyGatewayListsDefaultTools(t *testing.T) {
	gateway := NewPolicyGateway(DefaultModuleCapabilities())
	catalog, err := gateway.ListTools(context.Background(), ToolCatalogRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Tools) == 0 {
		t.Fatal("expected registered tools")
	}
	if !catalog.RuntimePolicy.RequireIdempotencyForWrites {
		t.Fatal("expected write idempotency policy")
	}
	modules := map[string]bool{}
	for _, tool := range catalog.Tools {
		modules[tool.ModuleKey] = true
	}
	for _, moduleKey := range []string{"work.tasks", "workflows", "task_processes", "integrations.providers"} {
		if !modules[moduleKey] {
			t.Fatalf("expected module-owned tool descriptors for %s; got %#v", moduleKey, modules)
		}
	}
}

func TestPolicyGatewayRegistersLangGraphReadTools(t *testing.T) {
	gateway := NewPolicyGateway(DefaultModuleCapabilities())
	for _, toolKey := range []string{
		"context.read",
		"ui_context.read",
		"task.search",
		"task.read",
		"workflow.read_draft",
		"workflow.read_context",
		"workflow.compile_prompt",
		"task_process.read",
		"integration.catalog.list",
		"integration.dependency.read",
	} {
		result, err := gateway.ValidateToolCall(context.Background(), ToolCallSpec{
			ToolKey:   toolKey,
			RiskLevel: RiskRead,
			InputJSON: json.RawMessage(`{"query":"test"}`),
		})
		if err != nil {
			t.Fatalf("%s: %v", toolKey, err)
		}
		if !result.OK {
			t.Fatalf("%s validation = %#v", toolKey, result)
		}
	}
}

func TestPolicyGatewayRejectsDuplicateToolKeys(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected duplicate tool key panic")
		}
	}()
	NewPolicyGateway([]ModuleCapability{
		{ModuleKey: "one", Tools: []ToolDescriptor{tool("duplicate.read", "one", "Read", RiskRead, false, "read", nil)}},
		{ModuleKey: "two", Tools: []ToolDescriptor{tool("duplicate.read", "two", "Read", RiskRead, false, "read", nil)}},
	})
}

func TestPolicyGatewayRejectsWriteWithoutApprovalAndIdempotency(t *testing.T) {
	gateway := NewPolicyGateway(DefaultModuleCapabilities())
	result, err := gateway.ValidateToolCall(context.Background(), ToolCallSpec{
		ToolKey:   "task.create",
		RiskLevel: RiskHigh,
		InputJSON: json.RawMessage(`{"title":"x"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.ErrorCode != "approval_required" {
		t.Fatalf("result = %#v", result)
	}

	result, err = gateway.ValidateToolCall(context.Background(), ToolCallSpec{
		ToolKey:          "task.create",
		RiskLevel:        RiskHigh,
		RequiresApproval: true,
		InputJSON:        json.RawMessage(`{"title":"x"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.ErrorCode != "idempotency_required" {
		t.Fatalf("result = %#v", result)
	}
}

func TestPolicyGatewayCreatesProposalForApprovedWriteSpec(t *testing.T) {
	gateway := NewPolicyGateway(DefaultModuleCapabilities())
	proposal, err := gateway.CreateProposal(context.Background(), ToolCallSpec{
		ToolCallID:       "tool-1",
		ToolKey:          "task.create",
		RiskLevel:        RiskHigh,
		RequiresApproval: true,
		IdempotencyKey:   "project:session:run:tool-1:1",
		InputJSON:        json.RawMessage(`{"title":"x"}`),
		ExpectedCardType: "task_create_proposal",
	})
	if err != nil {
		t.Fatal(err)
	}
	if proposal.ToolCallID != "tool-1" || proposal.CardType != "task_create_proposal" || proposal.Status != "needs_approval" {
		t.Fatalf("proposal = %#v", proposal)
	}
}

func TestPolicyGatewayRejectsClientRiskDowngrade(t *testing.T) {
	gateway := NewPolicyGateway(DefaultModuleCapabilities())
	for _, risk := range []RiskLevel{RiskRead, "", "unknown"} {
		result, err := gateway.ValidateToolCall(context.Background(), ToolCallSpec{
			ToolKey: "task.create", RiskLevel: risk, InputJSON: json.RawMessage(`{"title":"x"}`),
		})
		if err != nil || result.OK || result.ErrorCode != "approval_required" {
			t.Fatalf("client risk %q bypassed registered policy: result=%+v err=%v", risk, result, err)
		}
	}
}
