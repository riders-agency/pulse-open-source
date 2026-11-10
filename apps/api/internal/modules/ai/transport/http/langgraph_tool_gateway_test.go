package http

import "testing"

func TestDefaultToolCapabilitiesIncludeClientModuleDescriptors(t *testing.T) {
	modules := map[string]bool{}
	tools := map[string]bool{}
	for _, capability := range defaultToolCapabilities() {
		modules[capability.ModuleKey] = true
		for _, descriptor := range capability.Tools {
			tools[descriptor.ToolKey] = true
		}
	}
	for _, moduleKey := range []string{"clients", "payments"} {
		if !modules[moduleKey] {
			t.Fatalf("missing domain module capabilities for %s", moduleKey)
		}
	}
	for _, toolKey := range []string{"client.search", "payment.lookup_due"} {
		if !tools[toolKey] {
			t.Fatalf("missing domain-owned tool descriptor %s", toolKey)
		}
	}
}

func TestLangGraphReadToolWhitelistIncludesCatalogReadTools(t *testing.T) {
	for _, toolKey := range []string{
		"context.read",
		"ui_context.read",
		"task.search",
		"task.read",
		"client.search",
		"client.read",
		"client.count",
		"payment.lookup_due",
		"payment.schedule.read",
		"workflow.read_draft",
		"workflow.read_context",
		"workflow.compile_prompt",
		"task_process.read",
		"integration.catalog.list",
		"integration.dependency.read",
	} {
		if !isLangGraphReadTool(toolKey) {
			t.Fatalf("%s must be treated as a LangGraph read tool", toolKey)
		}
	}
}
