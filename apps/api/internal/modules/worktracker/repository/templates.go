package repository

func WorkflowTemplates() []WorkflowTemplate {
	return []WorkflowTemplate{
		{
			Key:         "client_request",
			Name:        "Client request",
			Description: "Запрос клиента с оценкой, согласованием, внутренним ревью и приемкой клиента.",
			TaskTypes:   []string{"request", "task", "support"},
			Statuses: []TemplateStatus{
				statusTpl("New request", "backlog", "slate", nil, false, false, false, false),
				statusTpl("Triage", "todo", "blue", nil, false, false, false, false),
				statusTpl("Need client info", "todo", "amber", nil, false, false, false, false),
				statusTpl("Estimate", "todo", "blue", nil, false, false, false, false),
				statusTpl("Waiting approval", "todo", "violet", nil, false, true, false, false),
				statusTpl("Planned", "todo", "blue", nil, false, true, false, false),
				statusTpl("In progress", "in_progress", "amber", intPtr(8), true, false, true, false),
				statusTpl("Internal review", "review", "violet", intPtr(6), true, false, false, false),
				statusTpl("Client acceptance", "review", "teal", intPtr(5), false, false, false, false),
				statusTpl("Done", "done", "green", nil, false, false, false, true),
				statusTpl("Canceled", "canceled", "slate", nil, false, false, false, true),
			},
			Transitions: []TemplateMove{
				moveTpl("New request", "Triage", "Triage request", []string{"process_owner", "process_manager", "executor", "reviewer"}, nil, ""),
				moveTpl("Triage", "Need client info", "Ask client info", []string{"process_owner", "process_manager"}, nil, `{"notify_role":"process_manager"}`),
				moveTpl("Triage", "Estimate", "Estimate", []string{"process_owner", "process_manager", "executor"}, nil, ""),
				moveTpl("Estimate", "Waiting approval", "Send approval", []string{"process_owner", "process_manager"}, []string{"estimate_minutes"}, ""),
				moveTpl("Waiting approval", "Planned", "Approve plan", []string{"process_owner", "process_manager"}, []string{"estimate_minutes"}, ""),
				moveTpl("Planned", "In progress", "Start work", []string{"process_owner", "process_manager", "executor"}, []string{"assignee_user_id"}, `{"start_timer":true}`),
				moveTpl("In progress", "Internal review", "Request review", []string{"process_owner", "process_manager", "executor"}, nil, ""),
				moveTpl("Internal review", "Client acceptance", "Send to client", []string{"process_owner", "process_manager", "reviewer"}, nil, `{"set_client_visible":true}`),
				moveTpl("Client acceptance", "Done", "Complete", []string{"process_owner", "process_manager", "reviewer"}, nil, `{"stop_timer":true}`),
			},
			Boards: defaultBoardTemplates(),
			Fields: []CustomFieldSpec{
				fieldTpl("Client priority", "client_priority", "select", false, []string{"low", "normal", "high", "urgent"}, []string{"request", "support"}),
				fieldTpl("Request channel", "request_channel", "select", false, []string{"email", "telegram", "phone", "meeting", "portal"}, []string{"request", "support"}),
			},
		},
		{
			Key:         "bug",
			Name:        "Bug",
			Description: "Баги с воспроизведением, разработкой, code review, QA и деплоем.",
			TaskTypes:   []string{"bug"},
			Statuses: []TemplateStatus{
				statusTpl("Reported", "backlog", "slate", nil, false, false, false, false),
				statusTpl("Reproducing", "todo", "blue", nil, false, false, false, false),
				statusTpl("Confirmed", "todo", "amber", nil, false, false, false, false),
				statusTpl("In progress", "in_progress", "amber", intPtr(8), true, false, true, false),
				statusTpl("Code review", "review", "violet", intPtr(6), true, false, false, false),
				statusTpl("QA", "review", "teal", intPtr(6), true, false, false, false),
				statusTpl("Ready for deploy", "review", "blue", nil, false, false, false, false),
				statusTpl("Done", "done", "green", nil, false, false, false, true),
				statusTpl("Won't fix", "canceled", "slate", nil, false, false, false, true),
			},
			Transitions: linearMoves([]string{"Reported", "Reproducing", "Confirmed", "In progress", "Code review", "QA", "Ready for deploy", "Done"}),
			Boards:      defaultBoardTemplates(),
			Fields: []CustomFieldSpec{
				fieldTpl("Environment", "environment", "select", false, []string{"production", "staging", "local", "mobile"}, []string{"bug"}),
				fieldTpl("Steps to reproduce", "steps_to_reproduce", "text", true, nil, []string{"bug"}),
				fieldTpl("Expected result", "expected_result", "text", false, nil, []string{"bug"}),
				fieldTpl("Actual result", "actual_result", "text", false, nil, []string{"bug"}),
			},
		},
		{
			Key:         "internal_task",
			Name:        "Internal task",
			Description: "Простой внутренний процесс без клиентской приемки.",
			TaskTypes:   []string{"task", "idea", "document"},
			Statuses: []TemplateStatus{
				statusTpl("Backlog", "backlog", "slate", nil, false, false, false, false),
				statusTpl("Todo", "todo", "blue", nil, false, false, false, false),
				statusTpl("In progress", "in_progress", "amber", intPtr(8), true, false, true, false),
				statusTpl("Review", "review", "violet", intPtr(6), true, false, false, false),
				statusTpl("Done", "done", "green", nil, false, false, false, true),
			},
			Transitions: linearMoves([]string{"Backlog", "Todo", "In progress", "Review", "Done"}),
			Boards:      defaultBoardTemplates(),
		},
		{
			Key:         "support_sla",
			Name:        "Support / SLA",
			Description: "Поддержка с first response, ожиданием клиента, эскалацией и SLA.",
			TaskTypes:   []string{"support", "request"},
			Statuses: []TemplateStatus{
				statusTpl("New", "backlog", "slate", nil, false, false, false, false),
				statusTpl("First response needed", "todo", "red", intPtr(12), false, false, false, false),
				statusTpl("Waiting customer", "todo", "blue", nil, false, false, false, false),
				statusTpl("In progress", "in_progress", "amber", intPtr(12), true, false, true, false),
				statusTpl("Escalated", "review", "red", nil, true, false, false, false),
				statusTpl("Resolved", "done", "green", nil, false, false, false, true),
				statusTpl("Closed", "done", "slate", nil, false, false, false, true),
			},
			Transitions: linearMoves([]string{"New", "First response needed", "In progress", "Resolved", "Closed"}),
			Boards:      defaultBoardTemplates(),
			Fields: []CustomFieldSpec{
				fieldTpl("SLA tier", "sla_tier", "select", true, []string{"standard", "priority", "critical"}, []string{"support"}),
				fieldTpl("First response due", "first_response_due", "datetime", false, nil, []string{"support"}),
				fieldTpl("Requester", "requester", "text", false, nil, []string{"support"}),
			},
		},
		{
			Key:         "content_design",
			Name:        "Content / Design",
			Description: "Производство материалов с ревью и клиентским согласованием.",
			TaskTypes:   []string{"content", "design"},
			Statuses: []TemplateStatus{
				statusTpl("Brief", "backlog", "slate", nil, false, false, false, false),
				statusTpl("Draft", "todo", "blue", nil, true, false, false, false),
				statusTpl("Production", "in_progress", "amber", intPtr(8), true, false, true, false),
				statusTpl("Internal review", "review", "violet", intPtr(6), true, false, false, false),
				statusTpl("Client approval", "review", "teal", nil, false, false, false, false),
				statusTpl("Published", "done", "green", nil, false, false, false, true),
				statusTpl("Done", "done", "green", nil, false, false, false, true),
			},
			Transitions: linearMoves([]string{"Brief", "Draft", "Production", "Internal review", "Client approval", "Published", "Done"}),
			Boards:      defaultBoardTemplates(),
		},
		{
			Key:         "finance_legal_hr",
			Name:        "Finance / Legal / HR",
			Description: "Внутренние согласования для финансов, юристов и HR.",
			TaskTypes:   []string{"finance", "legal", "hr"},
			Statuses: []TemplateStatus{
				statusTpl("Request", "backlog", "slate", nil, false, false, false, false),
				statusTpl("Review", "todo", "blue", nil, false, false, false, false),
				statusTpl("Waiting info", "todo", "amber", nil, false, false, false, false),
				statusTpl("In progress", "in_progress", "amber", intPtr(6), true, false, true, false),
				statusTpl("Approval", "review", "violet", nil, true, false, false, false),
				statusTpl("Done", "done", "green", nil, false, false, false, true),
				statusTpl("Canceled", "canceled", "slate", nil, false, false, false, true),
			},
			Transitions: linearMoves([]string{"Request", "Review", "In progress", "Approval", "Done"}),
			Boards:      defaultBoardTemplates(),
		},
	}
}

func WorkflowTemplateByKey(key string) (WorkflowTemplate, bool) {
	for _, tpl := range WorkflowTemplates() {
		if tpl.Key == key {
			return tpl, true
		}
	}
	return WorkflowTemplate{}, false
}

func BoardTemplateByKey(key string) (BoardTemplate, bool) {
	for _, tpl := range defaultBoardTemplates() {
		if tpl.Key == key {
			return tpl, true
		}
	}
	return BoardTemplate{}, false
}

func defaultBoardTemplates() []BoardTemplate {
	return []BoardTemplate{
		{Key: "kanban_status", Name: "Канбан по статусам", Description: "Классическая доска по колонкам workflow.", Type: "kanban", GroupBy: "status", SwimlaneBy: "none", VisibleFields: []string{"client", "project", "assignee", "estimate", "spent", "due_date", "labels"}, Density: "normal"},
		{Key: "kanban_assignee", Name: "По исполнителям", Description: "Задачи сгруппированы по статусам со swimlanes по ответственным.", Type: "kanban", GroupBy: "status", SwimlaneBy: "assignee", VisibleFields: []string{"client", "assignee", "estimate", "spent", "due_date"}, Density: "compact"},
		{Key: "priority_triage", Name: "Приоритеты", Description: "Быстрая сортировка по срочности и рискам.", Type: "custom", GroupBy: "priority", SwimlaneBy: "deadline", VisibleFields: []string{"client", "project", "assignee", "due_date", "labels"}, Density: "compact"},
		{Key: "client_acceptance", Name: "Клиентская приемка", Description: "Фокус на задачах, которые можно показывать клиенту.", Type: "custom", GroupBy: "client", SwimlaneBy: "none", VisibleFields: []string{"client", "project", "assignee", "spent", "due_date"}, Density: "normal"},
	}
}

func statusTpl(name, category, color string, wip *int, assignee, estimate, startTimer, stopTimer bool) TemplateStatus {
	return TemplateStatus{Name: name, Category: category, Color: color, WIPLimit: wip, RequiresAssignee: assignee, RequiresEstimate: estimate, AutoStartTimer: startTimer, AutoStopTimer: stopTimer}
}

func moveTpl(from, to, name string, roles []string, required []string, automation string) TemplateMove {
	if roles == nil {
		roles = []string{"process_owner", "process_manager", "executor", "reviewer"}
	}
	if required == nil {
		required = []string{}
	}
	if automation == "" {
		automation = "{}"
	}
	return TemplateMove{From: from, To: to, Name: name, AllowedRoles: roles, RequiredFields: required, AutomationJSON: automation}
}

func linearMoves(names []string) []TemplateMove {
	out := make([]TemplateMove, 0, len(names)-1)
	for i := 0; i < len(names)-1; i++ {
		required := []string{}
		automation := "{}"
		if names[i+1] == "In progress" {
			required = []string{"assignee_user_id"}
			automation = `{"start_timer":true}`
		}
		if names[i+1] == "Done" || names[i+1] == "Closed" || names[i+1] == "Resolved" {
			automation = `{"stop_timer":true}`
		}
		out = append(out, moveTpl(names[i], names[i+1], "Move to "+names[i+1], nil, required, automation))
	}
	return out
}

func fieldTpl(name, key, typ string, required bool, options []string, taskTypes []string) CustomFieldSpec {
	return CustomFieldSpec{Name: name, Key: key, Type: typ, Required: required, Options: options, VisibleByDefault: true, AppliesToTaskTypes: taskTypes}
}

func intPtr(value int) *int {
	return &value
}
