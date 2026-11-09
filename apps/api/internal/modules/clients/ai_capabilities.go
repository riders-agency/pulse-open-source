package clients

import (
	"encoding/json"

	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

func AIToolCapabilities() []aitools.ModuleCapability {
	return []aitools.ModuleCapability{
		{
			ModuleKey:         "clients",
			ContextDescriptor: "Company and project clients, contacts, services, payment schedule and payment actions.",
			PermissionScopes:  []string{"clients:read", "clients:create", "clients:update"},
			Tools: []aitools.ToolDescriptor{
				tool("client.search", "clients", "Найти клиента", aitools.RiskRead, false, "clients:read", nil),
				tool("client.count", "clients", "Посчитать клиентов проекта", aitools.RiskRead, false, "clients:read", nil),
				tool("client.read", "clients", "Прочитать клиента", aitools.RiskRead, false, "clients:read", nil),
				tool("client.create", "clients", "Создать клиента", aitools.RiskMedium, true, "clients:create", []string{"project_id", "name", "created_by_user_id"}),
				tool("client.update", "clients", "Изменить клиента", aitools.RiskMedium, true, "clients:update", []string{"client_id", "updated_by_user_id"}),
			},
			Intents: []aitools.IntentDescriptor{
				intent("client.search", "clients", aitools.RiskRead, false, []string{"dashboard", "project", "clients", "client_detail"}, "clients:read"),
				intent("client.read", "clients", aitools.RiskRead, false, []string{"clients", "client", "client_detail"}, "clients:read"),
				intent("client.count", "clients", aitools.RiskRead, false, []string{"dashboard", "project", "clients"}, "clients:read"),
				intent("client.create", "clients", aitools.RiskMedium, true, []string{"dashboard", "project", "clients"}, "clients:create"),
				intent("client.update", "clients", aitools.RiskMedium, true, []string{"clients", "client", "client_detail"}, "clients:update"),
			},
		},
		{
			ModuleKey:         "payments",
			ContextDescriptor: "Payment actions, client services billing cycles, due reminders and payment facts.",
			PermissionScopes:  []string{"payments:read", "payments:create", "payments:update"},
			Tools: []aitools.ToolDescriptor{
				tool("payment.lookup_due", "payments", "Найти ближайшую оплату", aitools.RiskRead, false, "payments:read", nil),
				tool("payment.schedule.read", "payments", "Прочитать график оплат", aitools.RiskRead, false, "payments:read", nil),
				tool("payment.action.create", "payments", "Создать действие оплаты", aitools.RiskMedium, true, "payments:create", []string{"project_id", "client_id", "period_start", "created_by_user_id"}),
				tool("payment.action.confirm", "payments", "Подтвердить оплату", aitools.RiskHigh, true, "payments:update", []string{"payment_action_id", "paid_at", "confirmed_by_user_id"}),
			},
			Intents: []aitools.IntentDescriptor{
				intent("payment.lookup_due", "payments", aitools.RiskRead, false, []string{"dashboard", "clients", "client", "client_detail", "payments"}, "payments:read"),
				intent("payment.schedule.read", "payments", aitools.RiskRead, false, []string{"dashboard", "clients", "client", "client_detail", "payments"}, "payments:read"),
				intent("payment.action.create", "payments", aitools.RiskMedium, true, []string{"clients", "client_detail", "payments"}, "payments:create"),
				intent("payment.confirm", "payments", aitools.RiskHigh, true, []string{"clients", "client_detail", "payments"}, "payments:update"),
			},
		},
	}
}

func tool(key, module, name string, risk aitools.RiskLevel, approval bool, permission string, idempotency []string) aitools.ToolDescriptor {
	return aitools.ToolDescriptor{
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

func intent(key, module string, risk aitools.RiskLevel, approval bool, pages []string, permission string) aitools.IntentDescriptor {
	return aitools.IntentDescriptor{
		Intent:             key,
		ModuleKey:          module,
		RiskLevel:          risk,
		RequiresApproval:   approval,
		AllowedPageKinds:   pages,
		RequiredPermission: permission,
	}
}

func sideEffects(risk aitools.RiskLevel) []string {
	switch risk {
	case aitools.RiskRead:
		return nil
	case aitools.RiskCritical:
		return []string{"db_write", "audit_event", "approval_required", "transactional_patch"}
	default:
		return []string{"db_write", "audit_event", "approval_required"}
	}
}
