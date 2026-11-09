package aicontext

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

type PageContext struct {
	Route             string            `json:"route"`
	Pathname          string            `json:"pathname"`
	PageKind          string            `json:"pageKind"`
	CompanyID         string            `json:"companyId,omitempty"`
	ProjectID         string            `json:"projectId,omitempty"`
	WorkflowID        string            `json:"workflowId,omitempty"`
	WorkflowRunID     string            `json:"workflowRunId,omitempty"`
	TaskID            string            `json:"taskId,omitempty"`
	ClientID          string            `json:"clientId,omitempty"`
	SelectedActionKey string            `json:"selectedActionKey,omitempty"`
	VisiblePanel      string            `json:"visiblePanel,omitempty"`
	Timestamp         string            `json:"timestamp,omitempty"`
	Correction        map[string]string `json:"correction,omitempty"`
	Extra             map[string]any    `json:"-"`
}

func NormalizePageContext(raw json.RawMessage) PageContext {
	frontend := pageContextFromRaw(raw)
	backend := ExtractIDsFromRoute(firstNonEmpty(frontend.Route, frontend.Pathname))
	return ReconcileFrontendAndBackendContext(frontend, backend)
}

func InferPageKindFromRoute(route string) string {
	pathname := routePathname(route)
	switch {
	case pathname == "/site-chat" || strings.HasPrefix(pathname, "/site-chat/"):
		return "site_chat"
	case pathname == "/status" || strings.HasPrefix(pathname, "/status/"):
		return "system_status"
	case strings.Contains(pathname, "/project-workflows") || strings.Contains(pathname, "/workflows/"):
		return "workflow_editor"
	case strings.Contains(pathname, "/runs/"):
		return "workflow_run"
	case strings.HasPrefix(pathname, "/tasks/boards"):
		return "task_board"
	case isTaskDetailPath(pathname):
		return "task"
	case strings.HasPrefix(pathname, "/tasks") || pathname == "/work" || strings.HasPrefix(pathname, "/work/"):
		return "tasks"
	case strings.Contains(pathname, "/project-integrations") || regexp.MustCompile(`/projects/[^/]+/integrations`).MatchString(pathname):
		return "integrations"
	case strings.Contains(pathname, "/clients/"):
		return "client"
	case strings.Contains(pathname, "/clients"):
		return "clients"
	case strings.Contains(pathname, "/payments"):
		return "payments"
	case strings.Contains(pathname, "/settings"):
		return "settings"
	case strings.Contains(pathname, "/projects/"):
		return "project"
	case pathname == "/dashboard" || strings.Contains(pathname, "/dashboard"):
		return "dashboard"
	default:
		return "unknown"
	}
}

func ExtractIDsFromRoute(route string) PageContext {
	pathname := routePathname(route)
	values := routeQuery(route)
	return PageContext{
		Route:             route,
		Pathname:          pathname,
		PageKind:          InferPageKindFromRoute(route),
		WorkflowID:        matchID(pathname, `/workflows/([^/?#]+)`),
		WorkflowRunID:     matchID(pathname, `/runs/([^/?#]+)`),
		TaskID:            taskIDFromPathname(pathname),
		ClientID:          matchID(pathname, `/clients/([^/?#]+)`),
		SelectedActionKey: values.Get("action"),
	}
}

func ReconcileFrontendAndBackendContext(frontend, backend PageContext) PageContext {
	out := frontend
	if out.Route == "" {
		out.Route = backend.Route
	}
	if out.Pathname == "" {
		out.Pathname = backend.Pathname
	}
	if out.PageKind == "" || out.PageKind == "unknown" || (backend.PageKind != "" && backend.PageKind != "unknown" && out.PageKind != backend.PageKind) {
		if out.PageKind != "" && out.PageKind != backend.PageKind {
			out.Correction = map[string]string{
				"field":    "pageKind",
				"frontend": out.PageKind,
				"backend":  backend.PageKind,
			}
		}
		out.PageKind = backend.PageKind
	}
	if out.WorkflowID == "" {
		out.WorkflowID = backend.WorkflowID
	}
	if out.WorkflowRunID == "" {
		out.WorkflowRunID = backend.WorkflowRunID
	}
	if out.TaskID == "" {
		out.TaskID = backend.TaskID
	}
	if out.ClientID == "" {
		out.ClientID = backend.ClientID
	}
	if out.SelectedActionKey == "" {
		out.SelectedActionKey = backend.SelectedActionKey
	}
	return out
}

func (p PageContext) RawJSON() json.RawMessage {
	raw, err := json.Marshal(p)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}

func pageContextFromRaw(raw json.RawMessage) PageContext {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return PageContext{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return PageContext{}
	}
	return PageContext{
		Route:             stringValue(payload["route"]),
		Pathname:          stringValue(payload["pathname"]),
		PageKind:          firstNonEmpty(stringValue(payload["pageKind"]), stringValue(payload["page_kind"])),
		CompanyID:         stringValue(payload["companyId"]),
		ProjectID:         stringValue(payload["projectId"]),
		WorkflowID:        stringValue(payload["workflowId"]),
		WorkflowRunID:     stringValue(payload["workflowRunId"]),
		TaskID:            stringValue(payload["taskId"]),
		ClientID:          stringValue(payload["clientId"]),
		SelectedActionKey: stringValue(payload["selectedActionKey"]),
		VisiblePanel:      stringValue(payload["visiblePanel"]),
		Timestamp:         stringValue(payload["timestamp"]),
	}
}

func routePathname(route string) string {
	route = strings.TrimSpace(route)
	if route == "" {
		return ""
	}
	if parsed, err := url.Parse(route); err == nil {
		if parsed.Path != "" {
			return parsed.Path
		}
	}
	if idx := strings.IndexAny(route, "?#"); idx >= 0 {
		return route[:idx]
	}
	return route
}

func routeQuery(route string) url.Values {
	if parsed, err := url.Parse(route); err == nil {
		return parsed.Query()
	}
	return url.Values{}
}

func isTaskDetailPath(pathname string) bool {
	if !strings.HasPrefix(pathname, "/tasks/") {
		return false
	}
	for _, prefix := range []string{
		"/tasks/boards",
		"/tasks/all",
		"/tasks/calendar",
		"/tasks/processes",
		"/tasks/time",
		"/tasks/timeline",
		"/tasks/reports",
		"/tasks/workload",
		"/tasks/employees",
	} {
		if strings.HasPrefix(pathname, prefix) {
			return false
		}
	}
	return true
}

func taskIDFromPathname(pathname string) string {
	if id := matchID(pathname, `^/tasks/([^/?#]+)`); id != "" && isTaskDetailPath(pathname) {
		return id
	}
	return matchID(pathname, `/work/tasks/([^/?#]+)`)
}

func matchID(value, pattern string) string {
	match := regexp.MustCompile(pattern).FindStringSubmatch(value)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringValue(value any) string {
	if out, ok := value.(string); ok {
		return strings.TrimSpace(out)
	}
	return ""
}
