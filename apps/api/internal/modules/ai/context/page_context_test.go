package aicontext

import (
	"encoding/json"
	"testing"
)

func TestInferPageKindFromRoute(t *testing.T) {
	tests := []struct {
		route string
		want  string
	}{
		{"/workflows/1f2b/edit?action=manual_payment_trigger", "workflow_editor"},
		{"/workflows/1f2b", "workflow_editor"},
		{"/site-chat", "site_chat"},
		{"/dashboard", "dashboard"},
		{"/clients", "clients"},
		{"/tasks/boards", "task_board"},
		{"/status", "system_status"},
	}

	for _, tt := range tests {
		t.Run(tt.route, func(t *testing.T) {
			if got := InferPageKindFromRoute(tt.route); got != tt.want {
				t.Fatalf("InferPageKindFromRoute(%q) = %q, want %q", tt.route, got, tt.want)
			}
		})
	}
}

func TestExtractIDsFromWorkflowRoute(t *testing.T) {
	ctx := ExtractIDsFromRoute("/workflows/1f2b/edit?action=manual_payment_trigger")
	if ctx.PageKind != "workflow_editor" {
		t.Fatalf("page kind = %q", ctx.PageKind)
	}
	if ctx.WorkflowID != "1f2b" {
		t.Fatalf("workflow id = %q", ctx.WorkflowID)
	}
	if ctx.SelectedActionKey != "manual_payment_trigger" {
		t.Fatalf("action = %q", ctx.SelectedActionKey)
	}
}

func TestNormalizePageContextBackendWinsWorkflowCorrection(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"route":    "/workflows/1f2b/edit?action=manual_payment_trigger",
		"pathname": "/workflows/1f2b/edit",
		"pageKind": "tasks",
	})

	ctx := NormalizePageContext(raw)
	if ctx.PageKind != "workflow_editor" {
		t.Fatalf("page kind = %q", ctx.PageKind)
	}
	if ctx.Correction["frontend"] != "tasks" || ctx.Correction["backend"] != "workflow_editor" {
		t.Fatalf("correction = %#v", ctx.Correction)
	}
}
