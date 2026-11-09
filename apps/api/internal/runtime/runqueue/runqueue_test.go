package runqueue

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeRequestDefaults(t *testing.T) {
	req, err := normalizeRequest(Request{
		ProjectID:         " project ",
		WorkflowID:        "workflow",
		WorkflowVersionID: "version",
		SourceType:        " manual ",
	})
	if err != nil {
		t.Fatalf("normalize request: %v", err)
	}
	if req.ContractVersion != CurrentRequestVersion {
		t.Fatalf("expected current contract version, got %q", req.ContractVersion)
	}
	if req.WorkflowRunID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatal("expected generated workflow_run_id")
	}
	if req.ProjectID != "project" || req.SourceType != "manual" {
		t.Fatalf("expected trimmed identity, got %#v", req)
	}
	if !strings.HasPrefix(req.IdempotencyKey, "workflow_run:") {
		t.Fatalf("expected generated idempotency key, got %q", req.IdempotencyKey)
	}
	if string(req.InputJSON) != "{}" {
		t.Fatalf("expected empty input default, got %s", req.InputJSON)
	}
	if req.EventType != "workflow.started" {
		t.Fatalf("expected workflow.started, got %q", req.EventType)
	}
	if req.EventMessage != "workflow run queued" {
		t.Fatalf("expected default message, got %q", req.EventMessage)
	}
	if string(req.EventPayloadJSON) != "{}" {
		t.Fatalf("expected event payload default, got %s", req.EventPayloadJSON)
	}
	var meta map[string]any
	if err := json.Unmarshal(req.SourceMetadataJSON, &meta); err != nil {
		t.Fatalf("source metadata JSON: %v", err)
	}
	if meta["contract_version"] != CurrentRequestVersion || meta["source_type"] != SourceManual {
		t.Fatalf("unexpected source metadata: %#v", meta)
	}
}

func TestNormalizeRequestRequiresIdentity(t *testing.T) {
	_, err := normalizeRequest(Request{
		ProjectID:         "project",
		WorkflowVersionID: "version",
		SourceType:        "manual",
	})
	if err == nil {
		t.Fatal("expected missing workflow_id error")
	}
}

func TestNormalizeRequestRejectsUnsupportedVersionAndSource(t *testing.T) {
	base := Request{
		ProjectID:         "project",
		WorkflowID:        "workflow",
		WorkflowVersionID: "version",
		SourceType:        SourceManual,
	}
	badVersion := base
	badVersion.ContractVersion = "run-request.v2"
	if _, err := normalizeRequest(badVersion); err == nil {
		t.Fatal("expected unsupported version error")
	}
	badSource := base
	badSource.SourceType = "direct_insert"
	if _, err := normalizeRequest(badSource); err == nil {
		t.Fatal("expected unsupported source error")
	}
}

func TestNormalizeRequestRejectsInvalidJSON(t *testing.T) {
	_, err := normalizeRequest(Request{
		ProjectID:         "project",
		WorkflowID:        "workflow",
		WorkflowVersionID: "version",
		SourceType:        SourceManual,
		InputJSON:         []byte(`{bad`),
	})
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
