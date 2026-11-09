package main

import "testing"

func TestCheckHTTPCoverageRequiresScopedRoutesOnly(t *testing.T) {
	inventory := []inventoryRow{
		{Method: "GET", Path: "/healthz", Module: "health", OperationID: "healthz"},
		{Method: "GET", Path: "/api/v1/projects/{project_id}/workflows", Module: "workflow", OperationID: "workflowList"},
	}

	if got := checkHTTPCoverage(inventory, nil); got != 1 {
		t.Fatalf("missing fixture violations = %d, want 1", got)
	}

	catalog := []catalogRow{{
		Surface:       "http",
		Method:        "GET",
		Path:          "/api/v1/projects/{project_id}/workflows",
		OperationID:   "workflowList",
		Scope:         "project",
		NegativeCase:  "foreign_project_id",
		ExpectedGuard: "authz.RequireProjectProcess",
		Notes:         "test",
	}}
	if got := checkHTTPCoverage(inventory, catalog); got != 0 {
		t.Fatalf("covered route violations = %d, want 0", got)
	}
}

func TestCheckAIToolCoverageRequiresEveryRegisteredTool(t *testing.T) {
	if got := checkAIToolCoverage(nil); got != len(aiToolKeys()) {
		t.Fatalf("missing AI tool violations = %d, want %d", got, len(aiToolKeys()))
	}

	var catalog []catalogRow
	for _, key := range aiToolKeys() {
		catalog = append(catalog, catalogRow{
			Surface:       "ai_tool",
			Method:        "-",
			Path:          key,
			OperationID:   key,
			Scope:         "project",
			NegativeCase:  "forged_project_or_entity_ref",
			ExpectedGuard: "tool_gateway_project_context",
			Notes:         "test",
		})
	}
	if got := checkAIToolCoverage(catalog); got != 0 {
		t.Fatalf("covered AI tool violations = %d, want 0", got)
	}
}

func TestCatalogReferencesDetectOperationIDDrift(t *testing.T) {
	inventory := []inventoryRow{{
		Method:      "GET",
		Path:        "/api/v1/projects/{project_id}",
		Module:      "project",
		OperationID: "projectGet",
	}}
	catalog := []catalogRow{{
		Surface:       "http",
		Method:        "GET",
		Path:          "/api/v1/projects/{project_id}",
		OperationID:   "wrongOperation",
		Scope:         "project",
		NegativeCase:  "foreign_project_id",
		ExpectedGuard: "authz.RequireProject",
		Notes:         "test",
	}}
	if got := checkCatalogReferences(inventory, catalog); got != 1 {
		t.Fatalf("reference violations = %d, want 1", got)
	}
}
