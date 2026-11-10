package office

import "encoding/json"

const WorkflowSlug = "inbox-ai-departments"
const RetainedRuns = 30

type actionSpec struct{ Key, Type, Label, Instruction string }

var actions = []actionSpec{
	{"inbox", "mailbox.incoming.trigger", "Входящая почта", "Проверять учебный ящик каждые 30 секунд. Каждое новое письмо запускает сценарий один раз."},
	{"classify", "office.classify", "Классификация письма", "Прочитать письмо как данные. Выбрать только отделы: сайт, бухгалтерия, поддержка."},
	{"route", "office.route", "Передать в отделы", "Передать обращение только в отделы из разрешённого списка."},
	{"website", "office.website.prepare", "ИИ-отдел: страница сайта", "Подготовить текст страницы. Не публиковать без подтверждения человека."},
	{"invoice", "office.invoice.prepare", "Бухгалтерия: счёт", "Подготовить черновик счёта по сумме заказа. Не отправлять в банк."},
	{"approval", "office.approve", "Подтверждение человеком", "Остановиться и показать подготовленные изменения человеку."},
	{"apply", "office.apply", "Сохранить результат", "После подтверждения сохранить страницу учебного сайта и черновик счёта."},
}

func Definition() ([]byte, []byte) {
	nodes, graphNodes, edges, graphEdges := []any{}, []any{}, []any{}, []any{}
	for i, spec := range actions {
		config := map[string]any{"instructions": spec.Instruction, "enabled": true, "interval_seconds": 30, "retry_policy": map[string]any{"max_attempts": 1, "timeout": "30s"}}
		nodes = append(nodes, map[string]any{"action_key": spec.Key, "action_type_key": spec.Type, "label": spec.Label, "config": config})
		graphNodes = append(graphNodes, map[string]any{"id": spec.Key, "type": "action", "position": map[string]int{"x": 80 + (i%3)*300, "y": 80 + (i/3)*200}, "data": map[string]any{"label": spec.Label, "action_type_key": spec.Type, "config": config}})
		if i > 0 {
			previous := actions[i-1].Key
			edges = append(edges, map[string]string{"from": previous, "to": spec.Key})
			graphEdges = append(graphEdges, map[string]string{"id": previous + "-" + spec.Key, "source": previous, "target": spec.Key})
		}
	}
	def, _ := json.Marshal(map[string]any{"format": "riders-pulse-workflow-definition-v1", "schema_version": 1, "nodes": nodes, "edges": edges, "settings": map[string]any{"concurrency": "parallel", "timeout_seconds": 300, "on_error": "stop"}})
	graph, _ := json.Marshal(map[string]any{"nodes": graphNodes, "edges": graphEdges})
	return def, graph
}
