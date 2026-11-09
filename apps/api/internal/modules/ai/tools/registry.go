package tools

import "encoding/json"

type RiskLevel string

const (
	RiskRead     RiskLevel = "read"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type ToolDescriptor struct {
	ToolKey              string          `json:"tool_key"`
	ModuleKey            string          `json:"module_key"`
	DisplayName          string          `json:"display_name"`
	Description          string          `json:"description"`
	RiskLevel            RiskLevel       `json:"risk_level"`
	RequiresApproval     bool            `json:"requires_approval"`
	InputSchema          json.RawMessage `json:"input_schema"`
	OutputSchema         json.RawMessage `json:"output_schema"`
	Permission           string          `json:"permission"`
	IdempotencyKeyFields []string        `json:"idempotency_key_fields"`
	SideEffects          []string        `json:"side_effects"`
}

type IntentDescriptor struct {
	Intent             string    `json:"intent"`
	ModuleKey          string    `json:"module_key"`
	RiskLevel          RiskLevel `json:"risk_level"`
	RequiresApproval   bool      `json:"requires_approval"`
	AllowedPageKinds   []string  `json:"allowed_page_kinds"`
	RequiredPermission string    `json:"required_permission"`
}

type ModuleCapability struct {
	ModuleKey         string             `json:"module_key"`
	ContextDescriptor string             `json:"context_descriptor"`
	PermissionScopes  []string           `json:"permission_scopes"`
	Tools             []ToolDescriptor   `json:"tools"`
	Intents           []IntentDescriptor `json:"intents"`
}

func DefaultModuleCapabilities() []ModuleCapability {
	return []ModuleCapability{
		{
			ModuleKey:         "platform.ui_context",
			ContextDescriptor: "Current Pulse UI route, page kind, selected entities and normalized frontend context.",
			PermissionScopes:  []string{"ui_context:read"},
			Tools: []ToolDescriptor{
				tool("ui_context.read", "platform.ui_context", "Прочитать текущий экран", RiskRead, false, "ui_context:read", nil),
				tool("context.read", "platform.ui_context", "Прочитать контекст страницы", RiskRead, false, "ui_context:read", nil),
				tool("route.explain", "platform.ui_context", "Объяснить текущий маршрут", RiskRead, false, "ui_context:read", nil),
			},
			Intents: []IntentDescriptor{
				intent("page.where_am_i", "platform.ui_context", RiskRead, false, []string{"dashboard", "project", "tasks", "task_board", "workflow_editor", "site_chat", "status"}, "ui_context:read"),
				intent("page.verify_context", "platform.ui_context", RiskRead, false, []string{"dashboard", "project", "tasks", "task_board", "workflow_editor", "site_chat", "status"}, "ui_context:read"),
				intent("project.explain", "platform.ui_context", RiskRead, false, []string{"dashboard", "project", "tasks", "task_board", "workflow_editor", "site_chat", "status"}, "ui_context:read"),
			},
		},
		{
			ModuleKey:         "work.tasks",
			ContextDescriptor: "Project-scoped tasks, boards, task workflows, timers and activity events.",
			PermissionScopes:  []string{"tasks:read", "tasks:create", "tasks:update", "tasks:move"},
			Tools: []ToolDescriptor{
				tool("task.search", "work.tasks", "Найти задачи", RiskRead, false, "tasks:read", nil),
				tool("task.read", "work.tasks", "Прочитать задачу", RiskRead, false, "tasks:read", nil),
				tool("task.create", "work.tasks", "Создать задачу", RiskMedium, true, "tasks:create", []string{"project_id", "title", "created_by_user_id"}),
				tool("task.update", "work.tasks", "Изменить задачу", RiskMedium, true, "tasks:update", []string{"task_id", "updated_by_user_id"}),
				tool("task.move", "work.tasks", "Переместить задачу", RiskMedium, true, "tasks:move", []string{"task_id", "to_status_id", "updated_by_user_id"}),
			},
			Intents: []IntentDescriptor{
				intent("task.search", "work.tasks", RiskRead, false, []string{"dashboard", "project", "tasks", "task_board", "client_detail"}, "tasks:read"),
				intent("task.read", "work.tasks", RiskRead, false, []string{"tasks", "task", "task_detail", "task_board"}, "tasks:read"),
				intent("task.create", "work.tasks", RiskMedium, true, []string{"dashboard", "project", "tasks", "task_board", "client_detail"}, "tasks:create"),
				intent("task.update", "work.tasks", RiskMedium, true, []string{"tasks", "task", "task_detail"}, "tasks:update"),
				intent("task.move", "work.tasks", RiskMedium, true, []string{"tasks", "task_board", "task_detail"}, "tasks:move"),
			},
		},
		{
			ModuleKey:         "workflows",
			ContextDescriptor: "Workflow drafts, versions, graph patch validation and test/manual runs.",
			PermissionScopes:  []string{"workflows:read", "workflows:update", "workflows:run", "workflows:publish"},
			Tools: []ToolDescriptor{
				tool("workflow.read_draft", "workflows", "Прочитать draft workflow", RiskRead, false, "workflows:read", nil),
				tool("workflow.read_context", "workflows", "Прочитать workflow context", RiskRead, false, "workflows:read", nil),
				tool("workflow.compile_prompt", "workflows", "Скомпилировать prompt", RiskRead, false, "workflows:read", nil),
				tool("workflow.patch.propose", "workflows", "Предложить patch workflow", RiskRead, false, "workflows:read", nil),
				tool("workflow.patch", "workflows", "Предложить и применить patch workflow", RiskCritical, true, "workflows:update", []string{"workflow_id", "patch_hash", "applied_by_user_id"}),
				tool("workflow.patch.apply", "workflows", "Применить patch workflow", RiskCritical, true, "workflows:update", []string{"workflow_id", "patch_hash", "applied_by_user_id"}),
				tool("workflow.run.test", "workflows", "Запустить test run", RiskHigh, true, "workflows:run", []string{"workflow_id", "test_input_hash", "run_by_user_id"}),
			},
			Intents: []IntentDescriptor{
				intent("workflow.answer_question", "workflows", RiskRead, false, []string{"workflow_editor", "workflow_run", "project"}, "workflows:read"),
				intent("workflow.read_draft", "workflows", RiskRead, false, []string{"workflow_editor", "workflow_run", "project"}, "workflows:read"),
				intent("workflow.read_context", "workflows", RiskRead, false, []string{"workflow_editor", "workflow_run", "project"}, "workflows:read"),
				intent("workflow.compile_prompt", "workflows", RiskRead, false, []string{"workflow_editor", "workflow_run", "project"}, "workflows:read"),
				intent("workflow.patch", "workflows", RiskCritical, true, []string{"workflow_editor", "workflow_run", "project"}, "workflows:update"),
				intent("workflow.run_test", "workflows", RiskHigh, true, []string{"workflow_editor", "workflow_run"}, "workflows:run"),
				intent("prompt.copy", "workflows", RiskRead, false, []string{"workflow_editor", "workflow_run", "project"}, "workflows:read"),
			},
		},
		{
			ModuleKey:         "task_processes",
			ContextDescriptor: "Task workflow statuses, transitions, roles, boards and required fields.",
			PermissionScopes:  []string{"task_processes:read", "task_processes:update"},
			Tools: []ToolDescriptor{
				tool("task_process.read", "task_processes", "Прочитать процесс задач", RiskRead, false, "task_processes:read", nil),
				tool("task_process.patch.propose", "task_processes", "Предложить patch процесса задач", RiskRead, false, "task_processes:read", nil),
				tool("task_process.patch.apply", "task_processes", "Применить patch процесса задач", RiskCritical, true, "task_processes:update", []string{"task_workflow_id", "patch_hash", "applied_by_user_id"}),
			},
			Intents: []IntentDescriptor{
				intent("task_process.read", "task_processes", RiskRead, false, []string{"tasks", "task_board", "settings"}, "task_processes:read"),
				intent("task_process.patch", "task_processes", RiskCritical, true, []string{"tasks", "task_board", "settings"}, "task_processes:update"),
			},
		},
		{
			ModuleKey:         "integrations.providers",
			ContextDescriptor: "Action dependencies, provider catalog, integration diagnostics and secret-scoped provider configuration.",
			PermissionScopes:  []string{"integrations:read", "integrations:update"},
			Tools: []ToolDescriptor{
				tool("integration.catalog.list", "integrations.providers", "Прочитать каталог интеграций", RiskRead, false, "integrations:read", nil),
				tool("integration.dependency.read", "integrations.providers", "Прочитать интеграцию", RiskRead, false, "integrations:read", nil),
				tool("integration.diagnostics.run", "integrations.providers", "Проверить интеграцию", RiskHigh, true, "integrations:update", []string{"action_dependency_id", "checked_by_user_id"}),
				tool("integration.secret.rotate", "integrations.providers", "Ротировать секрет интеграции", RiskCritical, true, "integrations:update", []string{"action_dependency_id", "rotated_by_user_id"}),
			},
			Intents: []IntentDescriptor{
				intent("integration.catalog.list", "integrations.providers", RiskRead, false, []string{"integrations", "settings", "project"}, "integrations:read"),
				intent("integration.dependency.read", "integrations.providers", RiskRead, false, []string{"integrations", "settings", "project"}, "integrations:read"),
				intent("integration.diagnostics", "integrations.providers", RiskHigh, true, []string{"integrations", "settings"}, "integrations:update"),
				intent("integration.secret.rotate", "integrations.providers", RiskCritical, true, []string{"integrations", "settings"}, "integrations:update"),
			},
		},
	}
}

func tool(key, module, name string, risk RiskLevel, approval bool, permission string, idempotency []string) ToolDescriptor {
	return ToolDescriptor{
		ToolKey:              key,
		ModuleKey:            module,
		DisplayName:          name,
		Description:          name + " через доменный модуль " + module,
		RiskLevel:            risk,
		RequiresApproval:     approval,
		InputSchema:          json.RawMessage(`{"type":"object","additionalProperties":true}`),
		OutputSchema:         json.RawMessage(`{"type":"object","additionalProperties":true}`),
		Permission:           permission,
		IdempotencyKeyFields: idempotency,
		SideEffects:          sideEffects(risk),
	}
}

func intent(key, module string, risk RiskLevel, approval bool, pages []string, permission string) IntentDescriptor {
	return IntentDescriptor{
		Intent:             key,
		ModuleKey:          module,
		RiskLevel:          risk,
		RequiresApproval:   approval,
		AllowedPageKinds:   pages,
		RequiredPermission: permission,
	}
}

func sideEffects(risk RiskLevel) []string {
	switch risk {
	case RiskRead:
		return nil
	case RiskCritical:
		return []string{"db_write", "audit_event", "approval_required", "transactional_patch"}
	default:
		return []string{"db_write", "audit_event", "approval_required"}
	}
}
