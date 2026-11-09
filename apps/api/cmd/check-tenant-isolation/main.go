package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

type inventoryRow struct {
	Method      string
	Path        string
	Module      string
	OperationID string
}

type catalogRow struct {
	Surface       string
	Method        string
	Path          string
	OperationID   string
	Scope         string
	NegativeCase  string
	ExpectedGuard string
	Notes         string
}

func main() {
	inventoryPath := flag.String("inventory", "../../apps/api/internal/platform/httpserver/doc_inventory.tsv", "runtime HTTP inventory TSV")
	catalogPath := flag.String("catalog", "../../docs/architecture/tenant-isolation-catalog.tsv", "tenant isolation catalog TSV")
	writeBaseline := flag.Bool("write-baseline", false, "rewrite catalog with current scoped HTTP routes and AI tools")
	flag.Parse()

	inventory, err := readInventory(*inventoryPath)
	if err != nil {
		fail("read inventory: %v", err)
	}
	if *writeBaseline {
		if err := writeBaselineCatalog(*catalogPath, inventory); err != nil {
			fail("write baseline: %v", err)
		}
		fmt.Printf("Tenant isolation catalog baseline written: %s\n", *catalogPath)
		return
	}
	catalog, err := readCatalog(*catalogPath)
	if err != nil {
		fail("read catalog: %v", err)
	}

	violations := 0
	violations += checkHTTPCoverage(inventory, catalog)
	violations += checkAIToolCoverage(catalog)
	violations += checkCatalogReferences(inventory, catalog)
	if violations > 0 {
		fmt.Printf("Tenant isolation catalog check failed: violations=%d\n", violations)
		os.Exit(1)
	}
	fmt.Printf("Tenant isolation catalog check passed: http=%d ai_tools=%d\n", countSurface(catalog, "http"), countSurface(catalog, "ai_tool"))
}

func readInventory(path string) ([]inventoryRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rows []inventoryRow
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid inventory row: %q", line)
		}
		rows = append(rows, inventoryRow{
			Method:      strings.ToUpper(strings.TrimSpace(parts[0])),
			Path:        normalizePath(parts[1]),
			Module:      strings.TrimSpace(parts[2]),
			OperationID: strings.TrimSpace(parts[3]),
		})
	}
	return rows, scanner.Err()
}

func readCatalog(path string) ([]catalogRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	seen := map[string]struct{}{}
	var rows []catalogRow
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 8 {
			return nil, fmt.Errorf("catalog row must have 8 tab-separated fields: %q", line)
		}
		row := catalogRow{
			Surface:       strings.TrimSpace(parts[0]),
			Method:        strings.ToUpper(strings.TrimSpace(parts[1])),
			Path:          normalizePath(parts[2]),
			OperationID:   strings.TrimSpace(parts[3]),
			Scope:         strings.TrimSpace(parts[4]),
			NegativeCase:  strings.TrimSpace(parts[5]),
			ExpectedGuard: strings.TrimSpace(parts[6]),
			Notes:         strings.TrimSpace(parts[7]),
		}
		if row.Surface == "" || row.Path == "" || row.OperationID == "" || row.Scope == "" || row.NegativeCase == "" || row.ExpectedGuard == "" {
			return nil, fmt.Errorf("catalog row has empty required field: %q", line)
		}
		key := catalogKey(row.Surface, row.Method, row.Path)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate catalog row: %s", key)
		}
		seen[key] = struct{}{}
		rows = append(rows, row)
	}
	return rows, scanner.Err()
}

func checkHTTPCoverage(inventory []inventoryRow, catalog []catalogRow) int {
	catalogKeys := map[string]catalogRow{}
	for _, row := range catalog {
		if row.Surface == "http" {
			catalogKeys[catalogKey("http", row.Method, row.Path)] = row
		}
	}
	var missing []string
	for _, row := range inventory {
		if !requiresHTTPIsolation(row) {
			continue
		}
		key := catalogKey("http", row.Method, row.Path)
		if _, ok := catalogKeys[key]; !ok {
			missing = append(missing, fmt.Sprintf("%s %s operation_id=%s", row.Method, row.Path, row.OperationID))
		}
	}
	sort.Strings(missing)
	for _, item := range missing {
		fmt.Printf("VIOLATION tenant_isolation_http_fixture_missing %s\n", item)
	}
	return len(missing)
}

func checkAIToolCoverage(catalog []catalogRow) int {
	catalogKeys := map[string]catalogRow{}
	for _, row := range catalog {
		if row.Surface == "ai_tool" {
			catalogKeys[catalogKey("ai_tool", "-", row.Path)] = row
		}
	}
	var missing []string
	for _, key := range aiToolKeys() {
		if _, ok := catalogKeys[catalogKey("ai_tool", "-", key)]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	for _, key := range missing {
		fmt.Printf("VIOLATION tenant_isolation_ai_tool_fixture_missing tool_key=%s\n", key)
	}
	return len(missing)
}

func checkCatalogReferences(inventory []inventoryRow, catalog []catalogRow) int {
	httpRows := map[string]inventoryRow{}
	for _, row := range inventory {
		httpRows[catalogKey("http", row.Method, row.Path)] = row
	}
	toolRows := map[string]struct{}{}
	for _, key := range aiToolKeys() {
		toolRows[catalogKey("ai_tool", "-", key)] = struct{}{}
	}

	var violations []string
	for _, row := range catalog {
		switch row.Surface {
		case "http":
			key := catalogKey("http", row.Method, row.Path)
			inv, ok := httpRows[key]
			if !ok {
				violations = append(violations, fmt.Sprintf("unknown http route %s %s", row.Method, row.Path))
				continue
			}
			if row.OperationID != inv.OperationID {
				violations = append(violations, fmt.Sprintf("operation_id drift %s %s catalog=%s inventory=%s", row.Method, row.Path, row.OperationID, inv.OperationID))
			}
		case "ai_tool":
			if row.Method != "-" {
				violations = append(violations, fmt.Sprintf("ai_tool row must use method '-' tool=%s", row.Path))
				continue
			}
			if _, ok := toolRows[catalogKey("ai_tool", "-", row.Path)]; !ok {
				violations = append(violations, fmt.Sprintf("unknown ai tool %s", row.Path))
			}
		default:
			violations = append(violations, fmt.Sprintf("unknown surface %s path=%s", row.Surface, row.Path))
		}
	}
	sort.Strings(violations)
	for _, item := range violations {
		fmt.Printf("VIOLATION tenant_isolation_catalog_reference_invalid %s\n", item)
	}
	return len(violations)
}

func requiresHTTPIsolation(row inventoryRow) bool {
	if strings.HasPrefix(row.Path, "/api/v1/docs/") || strings.HasPrefix(row.Path, "/api/v1/public/site-chat/") {
		return scopedPublicPath(row.Path)
	}
	if row.Path == "/healthz" || row.Path == "/site-chat/widget.js" {
		return false
	}
	if row.Module == "catalog" || row.Module == "docs" {
		return false
	}
	if row.Module == "worktracker" {
		return true
	}
	for _, marker := range scopedPathMarkers {
		if strings.Contains(row.Path, marker) {
			return true
		}
	}
	switch row.Path {
	case "/hooks/{hook_key}/{endpoint_key}",
		"/api/v1/companies",
		"/api/v1/client-service-types",
		"/api/v1/action-types",
		"/api/v1/action-dependency-types":
		return false
	default:
		return false
	}
}

func scopedPublicPath(path string) bool {
	for _, marker := range []string{"{public_key}", "{conversation_id}", "{call_session_id}"} {
		if strings.Contains(path, marker) {
			return true
		}
	}
	return false
}

var scopedPathMarkers = []string{
	"{action_dependency_id}",
	"{api_key_id}",
	"{attachment_id}",
	"{call_session_id}",
	"{chat_widget_id}",
	"{client_contact_id}",
	"{client_id}",
	"{client_responsible_id}",
	"{client_service_id}",
	"{client_service_type_id}",
	"{company_id}",
	"{company_legal_entity_id}",
	"{conversation_id}",
	"{desktop_message_id}",
	"{download_token}",
	"{email_template_id}",
	"{message_template_id}",
	"{payment_action_document_id}",
	"{payment_action_id}",
	"{project_id}",
	"{relay_key}",
	"{session_id}",
	"{task_attachment_id}",
	"{task_board_id}",
	"{task_checklist_item_id}",
	"{task_id}",
	"{task_process_role_id}",
	"{task_workflow_id}",
	"{tool_call_id}",
	"{visit_session_id}",
	"{webhook_endpoint_id}",
	"{webhook_event_id}",
	"{workflow_id}",
	"{workflow_run_id}",
	"{workflow_version_id}",
}

func aiToolKeys() []string {
	seen := map[string]struct{}{}
	var keys []string
	for _, capability := range aitools.DefaultModuleCapabilities() {
		for _, descriptor := range capability.Tools {
			key := strings.TrimSpace(descriptor.ToolKey)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func countSurface(rows []catalogRow, surface string) int {
	count := 0
	for _, row := range rows {
		if row.Surface == surface {
			count++
		}
	}
	return count
}

func writeBaselineCatalog(path string, inventory []inventoryRow) error {
	var rows []catalogRow
	for _, row := range inventory {
		if !requiresHTTPIsolation(row) {
			continue
		}
		rows = append(rows, catalogRow{
			Surface:       "http",
			Method:        row.Method,
			Path:          row.Path,
			OperationID:   row.OperationID,
			Scope:         inferHTTPScope(row.Path),
			NegativeCase:  inferHTTPNegativeCase(row),
			ExpectedGuard: inferHTTPExpectedGuard(row),
			Notes:         "seeded_from_runtime_inventory",
		})
	}
	for _, key := range aiToolKeys() {
		rows = append(rows, catalogRow{
			Surface:       "ai_tool",
			Method:        "-",
			Path:          key,
			OperationID:   key,
			Scope:         inferAIToolScope(key),
			NegativeCase:  "forged_project_or_entity_ref",
			ExpectedGuard: inferAIToolExpectedGuard(key),
			Notes:         "seeded_from_ai_tool_registry",
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return catalogKey(rows[i].Surface, rows[i].Method, rows[i].Path) < catalogKey(rows[j].Surface, rows[j].Method, rows[j].Path)
	})

	var out strings.Builder
	out.WriteString("# surface\tmethod\tpath_or_tool_key\toperation_id\tscope\tnegative_case\texpected_guard\tnotes\n")
	out.WriteString("# This is the exhaustive tenant/project isolation fixture inventory. New scoped HTTP routes or AI tools must add a row here before CI passes.\n")
	for _, row := range rows {
		out.WriteString(strings.Join([]string{
			row.Surface,
			row.Method,
			row.Path,
			row.OperationID,
			row.Scope,
			row.NegativeCase,
			row.ExpectedGuard,
			row.Notes,
		}, "\t"))
		out.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(out.String()), 0o644)
}

func inferHTTPScope(path string) string {
	switch {
	case strings.Contains(path, "{company_id}") || strings.Contains(path, "{company_legal_entity_id}"):
		return "company"
	case strings.Contains(path, "{project_id}") || strings.HasPrefix(path, "/api/v1/work/"):
		return "project"
	case strings.Contains(path, "{workflow_id}") || strings.Contains(path, "{workflow_version_id}") || strings.Contains(path, "{workflow_run_id}"):
		return "workflow_project"
	case strings.Contains(path, "{client_id}") || strings.Contains(path, "{client_service_id}") || strings.Contains(path, "{payment_action_id}") || strings.Contains(path, "{payment_action_document_id}"):
		return "client_project"
	case strings.Contains(path, "{conversation_id}") || strings.Contains(path, "{visit_session_id}") || strings.Contains(path, "{chat_widget_id}") || strings.Contains(path, "{call_session_id}") || strings.Contains(path, "{attachment_id}"):
		return "site_chat_project"
	case strings.Contains(path, "{webhook_endpoint_id}") || strings.Contains(path, "{webhook_event_id}"):
		return "webhook_workflow_project"
	case strings.Contains(path, "{relay_key}"):
		return "telegram_relay"
	case strings.Contains(path, "{api_key_id}") || strings.Contains(path, "{session_id}"):
		return "user_session"
	default:
		return "scoped_resource"
	}
}

func inferHTTPNegativeCase(row inventoryRow) string {
	switch {
	case strings.Contains(row.Path, "{project_id}"):
		return "foreign_project_id"
	case strings.Contains(row.Path, "{company_id}"):
		return "foreign_company_id"
	case strings.Contains(row.Path, "{client_id}"):
		return "foreign_client_id"
	case strings.Contains(row.Path, "{workflow_id}") || strings.Contains(row.Path, "{workflow_run_id}") || strings.Contains(row.Path, "{workflow_version_id}"):
		return "foreign_workflow_or_run_id"
	case strings.Contains(row.Path, "{webhook_endpoint_id}") || strings.Contains(row.Path, "{webhook_event_id}"):
		return "foreign_webhook_id"
	case strings.Contains(row.Path, "{relay_key}"):
		return "foreign_relay_key"
	case row.Module == "worktracker":
		return "foreign_project_or_task_id"
	default:
		return "foreign_resource_id"
	}
}

func inferHTTPExpectedGuard(row inventoryRow) string {
	switch row.Module {
	case "company":
		return "authz.RequireCompany_or_membership"
	case "project":
		return "authz.RequireProject_or_company_membership"
	case "workflow":
		return "authz.RequireWorkflow_or_ProjectProcess"
	case "runs":
		return "authz.RequireRun"
	case "webhooks":
		return "authz.RequireWebhookEndpoint_or_Event"
	case "clients":
		return "authz.RequireClient_or_ClientService_or_PaymentAction"
	case "deps":
		return "action_dependencies.requireProjectOrDependencyAccess"
	case "site_chat":
		return "site_chat.project_or_public_token_scope"
	case "worktracker":
		return "worktracker.AccessForProject_or_Task"
	case "ai":
		return "ai.project_session_tool_scope"
	case "auth":
		return "auth.user_owns_session_or_api_key"
	default:
		return "module_owned_scope_guard"
	}
}

func inferAIToolScope(key string) string {
	switch {
	case strings.HasPrefix(key, "task."):
		return "project_task"
	case strings.HasPrefix(key, "workflow."):
		return "project_workflow"
	case strings.HasPrefix(key, "task_process."):
		return "project_task_process"
	case strings.HasPrefix(key, "integration."):
		return "project_integration"
	case strings.Contains(key, "context") || strings.HasPrefix(key, "route."):
		return "project_ui_context"
	default:
		return "project"
	}
}

func inferAIToolExpectedGuard(key string) string {
	switch {
	case strings.HasPrefix(key, "task."):
		return "worktracker.AccessForProject_or_Task"
	case strings.HasPrefix(key, "workflow."):
		return "authz.RequireWorkflow_or_ProjectProcess"
	case strings.HasPrefix(key, "task_process."):
		return "authz.RequireProjectProcess"
	case strings.HasPrefix(key, "integration."):
		return "authz.RequireProject_and_dependency_project_match"
	default:
		return "tool_gateway_project_context"
	}
}

func catalogKey(surface, method, path string) string {
	return surface + "\t" + strings.ToUpper(strings.TrimSpace(method)) + "\t" + normalizePath(path)
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "-" {
		return path
	}
	if !strings.HasPrefix(path, "/") && strings.Contains(path, ".") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
