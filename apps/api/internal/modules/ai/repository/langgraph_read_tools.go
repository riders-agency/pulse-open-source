package repository

import (
	"context"
	"fmt"

	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

	"strings"
	"unicode"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (a *AI) ReadLangGraphContext(ctx context.Context, projectID, sessionID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	if isProjectContextRead(input) {
		return a.readLangGraphProjectContext(ctx, projectID, input)
	}
	query := strings.TrimSpace(stringValue(input["query"]))
	tokens := contextSearchTokens(query)
	rows, err := a.P.Query(ctx, `
SELECT t.task_id::text,
       COALESCE(t.task_key, ''),
       t.title,
       COALESCE(t.description, ''),
       COALESCE(t.priority, ''),
       COALESCE(t.due_date::text, ''),
       COALESCE(s.name, ''),
       COALESCE(c.client_id::text, ''),
       COALESCE(c.name, ''),
       t.updated_at::text
  FROM pulse_read.tasks_v1 t
  JOIN pulse_read.task_workflow_statuses_v1 s ON s.task_workflow_status_id = t.task_workflow_status_id
  LEFT JOIN pulse_read.clients_v1 c ON c.client_id = t.client_id
 WHERE t.project_id = $1
   AND t.archived_at IS NULL
   AND (
     $2 = ''
     OR t.search_vector @@ websearch_to_tsquery('russian'::regconfig, $2)
     OR EXISTS (
       SELECT 1 FROM unnest($3::text[]) token
        WHERE length(token) >= 3
          AND (
            lower(t.title) LIKE '%' || lower(token) || '%'
            OR lower(COALESCE(t.description, '')) LIKE '%' || lower(token) || '%'
            OR lower(COALESCE(t.task_key, '')) LIKE '%' || lower(token) || '%'
            OR lower(COALESCE(c.name, '')) LIKE '%' || lower(token) || '%'
            OR lower(COALESCE(c.full_name, '')) LIKE '%' || lower(token) || '%'
          )
     )
   )
 ORDER BY t.updated_at DESC
 LIMIT 12`, projectID, query, tokens)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]map[string]any, 0)
	for rows.Next() {
		var id, key, title, description, priority, dueDate, statusName, clientID, clientName, updatedAt string
		if err := rows.Scan(&id, &key, &title, &description, &priority, &dueDate, &statusName, &clientID, &clientName, &updatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, map[string]any{
			"task_id":     id,
			"task_key":    key,
			"title":       title,
			"description": aiproviders.TruncateForPrompt(description, 240),
			"priority":    priority,
			"due_date":    dueDate,
			"status_name": statusName,
			"client_id":   clientID,
			"client_name": clientName,
			"url":         "/tasks/" + id,
			"updated_at":  updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	pageContext := mapFromAny(input["page_context"])
	route := firstNonEmptyString(stringValue(pageContext["route"]), stringValue(pageContext["pathname"]))
	pageKind := firstNonEmptyString(stringValue(pageContext["pageKind"]), stringValue(pageContext["page_kind"]))
	answer := "По этому запросу задач не найдено."
	if len(tasks) > 0 {
		parts := make([]string, 0, len(tasks))
		for _, task := range tasks {
			label := firstNonEmptyString(stringValue(task["task_key"]), stringValue(task["title"]))
			clientName := stringValue(task["client_name"])
			statusName := stringValue(task["status_name"])
			if clientName != "" {
				label += " · " + clientName
			}
			if statusName != "" {
				label += " · " + statusName
			}
			parts = append(parts, label)
		}
		answer = fmt.Sprintf("Нашел задачи по запросу: %s", strings.Join(parts, "; "))
	} else if route != "" || pageKind != "" {
		answer = fmt.Sprintf("Вы сейчас в проектном контексте: экран %s, путь %s. Данные читаю через project-scoped tools, не из памяти.", firstNonEmptyString(pageKind, "unknown"), firstNonEmptyString(route, "unknown"))
	}
	return map[string]any{
		"kind":         "task_search_result",
		"tool_key":     "context.read",
		"query":        query,
		"tasks":        tasks,
		"page_context": pageContext,
		"answer":       answer,
		"from_memory":  false,
		"runtime":      "langgraph",
	}, nil
}

func (a *AI) readLangGraphProjectContext(ctx context.Context, projectID uuid.UUID, input map[string]any) (map[string]any, error) {
	var projectName, projectSlug, companyName string
	if err := a.P.QueryRow(ctx, `
SELECT p.name, COALESCE(p.slug, ''), c.name
  FROM pulse_read.projects_v1 p
  JOIN pulse_read.companies_v1 c ON c.company_id = p.company_id
 WHERE p.project_id = $1`, projectID).Scan(&projectName, &projectSlug, &companyName); err != nil {
		return nil, err
	}
	var taskCount, clientCount, workflowCount int
	_ = a.P.QueryRow(ctx, `
SELECT count(*)
  FROM pulse_read.tasks_v1
 WHERE project_id = $1
   AND archived_at IS NULL`, projectID).Scan(&taskCount)
	_ = a.P.QueryRow(ctx, `
SELECT count(*)
  FROM pulse_read.client_projects_v1 cp
  JOIN pulse_read.clients_v1 c ON c.client_id = cp.client_id
 WHERE cp.project_id = $1
   AND c.status <> 'deleted'`, projectID).Scan(&clientCount)
	_ = a.P.QueryRow(ctx, `
SELECT count(*)
  FROM pulse_read.workflows_v1
 WHERE project_id = $1
   AND status <> 'deleted'`, projectID).Scan(&workflowCount)
	pageContext := mapFromAny(input["page_context"])
	route := firstNonEmptyString(stringValue(pageContext["route"]), stringValue(pageContext["pathname"]))
	pageKind := firstNonEmptyString(stringValue(pageContext["pageKind"]), stringValue(pageContext["page_kind"]))
	answer := fmt.Sprintf(
		"Текущий проект: %s. Компания: %s. Slug: %s. Клиентов: %d, задач: %d, workflow: %d.",
		projectName,
		companyName,
		firstNonEmptyString(projectSlug, "не задан"),
		clientCount,
		taskCount,
		workflowCount,
	)
	if route != "" || pageKind != "" {
		answer += fmt.Sprintf(" Текущий экран: %s, путь: %s.", firstNonEmptyString(pageKind, "unknown"), firstNonEmptyString(route, "unknown"))
	}
	return map[string]any{
		"kind":            "project_context",
		"tool_key":        "context.read",
		"project_id":      projectID.String(),
		"project_name":    projectName,
		"project_slug":    projectSlug,
		"company_name":    companyName,
		"tasks_count":     taskCount,
		"clients_count":   clientCount,
		"workflows_count": workflowCount,
		"page_context":    pageContext,
		"answer":          answer,
		"from_memory":     false,
		"runtime":         "langgraph",
	}, nil
}

func isProjectContextRead(input map[string]any) bool {
	intent := strings.TrimSpace(stringValue(input["intent"]))
	entityType := strings.TrimSpace(stringValue(input["entity_type"]))
	return intent == "project.explain" || entityType == "project"
}

func (a *AI) ReadLangGraphPageContext(ctx context.Context, projectID, sessionID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	pageContext := mapFromAny(input["page_context"])
	route := firstNonEmptyString(stringValue(pageContext["route"]), stringValue(pageContext["pathname"]))
	pageKind := firstNonEmptyString(stringValue(pageContext["pageKind"]), stringValue(pageContext["page_kind"]), "unknown")
	answer := fmt.Sprintf("Текущий экран: %s, путь: %s.", pageKind, firstNonEmptyString(route, "unknown"))
	if correction := mapFromAny(pageContext["correction"]); len(correction) > 0 {
		answer = fmt.Sprintf("Текущий экран: %s, путь: %s. Frontend pageKind был скорректирован backend normalizer-ом.", pageKind, firstNonEmptyString(route, "unknown"))
	}
	return map[string]any{
		"kind":         "page_context",
		"tool_key":     "ui_context.read",
		"page_kind":    pageKind,
		"route":        route,
		"page_context": pageContext,
		"answer":       answer,
		"from_memory":  false,
		"runtime":      "langgraph",
	}, nil
}

func (a *AI) ReadLangGraphTaskSearch(ctx context.Context, projectID, sessionID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	out, err := a.ReadLangGraphContext(ctx, projectID, sessionID, userID, input)
	if err != nil {
		return nil, err
	}
	out["kind"] = "task_search_result"
	out["tool_key"] = "task.search"
	return out, nil
}

func (a *AI) ReadLangGraphWorkflowDraft(ctx context.Context, projectID, sessionID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	pageContext := mapFromAny(input["page_context"])
	workflowID := firstNonEmptyString(stringValue(input["workflow_id"]), stringValue(pageContext["workflowId"]), stringValue(pageContext["workflow_id"]))
	if workflowID == "" {
		return map[string]any{
			"kind":         "workflow_read_draft",
			"tool_key":     "workflow.read_draft",
			"page_context": pageContext,
			"answer":       "Не вижу workflow_id в текущем контексте.",
			"runtime":      "langgraph",
		}, nil
	}
	wfID, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow_id: %w", err)
	}
	if a.Workflows == nil {
		return nil, fmt.Errorf("workflows repository is not configured")
	}
	detail, err := a.Workflows.Get(ctx, wfID, userID)
	if err != nil {
		return nil, err
	}
	def := workflowJSONSummary(detail.DraftDefinition)
	graph := workflowJSONSummary(detail.DraftGraph)
	actionsCount := firstPositiveInt(def.ActionsCount, graph.ActionsCount)
	edgesCount := firstPositiveInt(def.EdgesCount, graph.EdgesCount)
	triggerCount := firstPositiveInt(def.TriggerCount, graph.TriggerCount)
	actions := def.Actions
	if len(actions) == 0 {
		actions = graph.Actions
	}
	answer := fmt.Sprintf("Workflow %s: действий %d, связей %d, триггеров %d.", detail.Name, actionsCount, edgesCount, triggerCount)
	return map[string]any{
		"kind":                 "workflow_read_draft",
		"tool_key":             "workflow.read_draft",
		"workflow_id":          detail.ID,
		"name":                 detail.Name,
		"status":               detail.Status,
		"actions_count":        actionsCount,
		"edges_count":          edgesCount,
		"trigger_count":        triggerCount,
		"actions":              actions,
		"selected_action_key":  stringValue(pageContext["selectedActionKey"]),
		"duplicates":           duplicateWorkflowActions(actions),
		"orphan_actions":       []map[string]any{},
		"missing_dependencies": []map[string]any{},
		"summary":              answer,
		"answer":               answer,
		"page_context":         pageContext,
		"runtime":              "langgraph",
	}, nil
}

func contextSearchTokens(query string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, part := range strings.FieldsFunc(query, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	}) {
		part = strings.ToLower(strings.TrimSpace(part))
		if len([]rune(part)) < 3 || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return out
}
