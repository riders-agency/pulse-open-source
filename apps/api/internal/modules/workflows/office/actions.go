package office

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

type identity struct{ RunID, ProjectID, CompanyID, UserID string }
type Handler struct {
	P    *pgxpool.Pool
	Kind string
}

func (h Handler) TypeKey() string                                       { return h.Kind }
func (h Handler) ValidateConfig(context.Context, json.RawMessage) error { return nil }
func Register(reg *registry.Registry, pool *pgxpool.Pool) {
	for _, spec := range actions {
		reg.MustRegisterAction(Handler{P: pool, Kind: spec.Type})
	}
	reg.MustRegisterTrigger(&Trigger{P: pool})
}

func (h Handler) Execute(ctx context.Context, req contracts.ActionRequest) contracts.ActionResult {
	if h.P == nil {
		return failure(fmt.Errorf("учебный сценарий недоступен"))
	}
	identity := identity{RunID: req.WorkflowRunID, ProjectID: req.ProjectID}
	var subject, body, sender string
	var amount int64
	err := h.P.QueryRow(ctx, `SELECT p.company_id::text,o.owner_user_id::text,m.subject,m.body,m.sender,m.amount_minor
 FROM workflow_office_messages m JOIN workflow_office_workspaces o USING(project_id) JOIN pulse_read.projects_v1 p USING(project_id)
 JOIN pulse_read.workflow_runs_v1 r ON r.workflow_run_id=m.workflow_run_id AND r.workflow_id=o.workflow_id
 WHERE m.workflow_run_id=$1 AND m.project_id=$2`, req.WorkflowRunID, req.ProjectID).Scan(&identity.CompanyID, &identity.UserID, &subject, &body, &sender, &amount)
	if err != nil {
		return failure(fmt.Errorf("письмо не принадлежит этому запуску"))
	}
	mail := map[string]any{"subject": subject, "body": body, "sender": sender}
	out, provider, waiting, err := h.execute(ctx, req, identity, mail, amount)
	if err != nil {
		return failure(err)
	}
	raw, _ := json.Marshal(map[string]any{"expose": out, "result": out})
	prov, _ := json.Marshal(provider)
	result := contracts.ActionResult{InputJSON: req.RunInputJSON, OutputJSON: raw, ProviderResponseJSON: prov, Waiting: waiting}
	if waiting {
		result.WaitingReason = "office_approval_required"
	}
	return result
}

func (h Handler) execute(ctx context.Context, req contracts.ActionRequest, id identity, mail map[string]any, amount int64) (map[string]any, map[string]any, bool, error) {
	provider := map[string]any{"provider": "pulse", "mode": "demo", "ok": true}
	switch h.Kind {
	case "mailbox.incoming.trigger":
		return mail, provider, false, nil
	case "office.classify":
		out, p, err := prepareWithAgent(ctx, id, "classify", mail)
		return out, p, false, err
	case "office.route":
		route, err := departmentRoute(req.Steps)
		return route, provider, false, err
	case "office.website.prepare":
		if !hasDepartment(req.Steps, "content") {
			return map[string]any{"skipped": true, "reason": "Письмо не адресовано ИИ-отделу"}, provider, false, nil
		}
		out, p, err := prepareWithAgent(ctx, id, "website", mail)
		return out, p, false, err
	case "office.invoice.prepare":
		if !hasDepartment(req.Steps, "accounting") {
			return map[string]any{"skipped": true, "reason": "Счёт не запрошен"}, provider, false, nil
		}
		return map[string]any{"number": "DEMO-" + id.RunID[:8], "customer": "ООО «Демо»", "amount_minor": amount, "currency": "RUB", "status": "draft", "bank_connected": false}, provider, false, nil
	case "office.approve":
		var approved bool
		err := h.P.QueryRow(ctx, `SELECT approved FROM workflow_office_decisions WHERE workflow_run_id=$1`, id.RunID).Scan(&approved)
		if isMissing(err) {
			return map[string]any{"page": step(req.Steps, "website"), "invoice": step(req.Steps, "invoice"), "message": "Проверьте страницу и счёт. До вашего решения изменения не сохраняются."}, provider, true, nil
		}
		return map[string]any{"approved": approved}, provider, false, err
	case "office.apply":
		out, err := h.apply(ctx, id, req.Steps)
		return out, provider, false, err
	}
	return nil, nil, false, fmt.Errorf("неизвестное действие")
}

func failure(err error) contracts.ActionResult {
	return contracts.ActionResult{OutputJSON: json.RawMessage(`{}`), ProviderResponseJSON: json.RawMessage(`{"ok":false}`), ErrorType: "office", Err: err}
}
func step(steps map[string]any, key string) map[string]any {
	out, _ := steps[key].(map[string]any)
	return out
}
func departmentRoute(steps map[string]any) (map[string]any, error) {
	value := step(steps, "classify")
	items, ok := value["departments"].([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("классификация отсутствует")
	}
	for _, item := range items {
		if item != "content" && item != "accounting" && item != "support" {
			return nil, fmt.Errorf("отдел не разрешён")
		}
	}
	return value, nil
}
func hasDepartment(steps map[string]any, want string) bool {
	items, _ := step(steps, "route")["departments"].([]any)
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
