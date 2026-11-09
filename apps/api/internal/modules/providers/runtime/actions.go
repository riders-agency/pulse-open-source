package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	emailprovider "riders.agency/pulse/api/internal/modules/providers/email"
	maxprovider "riders.agency/pulse/api/internal/modules/providers/max"
	telegramprovider "riders.agency/pulse/api/internal/modules/providers/telegram"
	uonprovider "riders.agency/pulse/api/internal/modules/providers/uon"
	yclientsprovider "riders.agency/pulse/api/internal/modules/providers/yclients"
	"riders.agency/pulse/api/internal/platform/httptransport"
	"riders.agency/pulse/api/internal/platform/outbound"
	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

var errProviderRuntimeUnavailable = errors.New("provider runtime unavailable")

type Runner struct {
	P                 *pgxpool.Pool
	MasterHEX         string
	HTTPClient        *http.Client
	MaxHTTPClient     *http.Client
	DocumentBasicAuth map[string]DocumentBasicAuth
}

func RegisterHandlers(reg *registry.Registry, runner *Runner) error {
	for _, adapter := range providerActionAdapters {
		if err := reg.RegisterAction(handler{runner: runner, adapter: adapter}); err != nil {
			return err
		}
	}
	return nil
}

type handler struct {
	runner  *Runner
	adapter providerActionAdapter
}

func (h handler) TypeKey() string { return h.adapter.typeKey }

func (h handler) ValidateConfig(context.Context, json.RawMessage) error { return nil }

func (h handler) Execute(ctx context.Context, req contracts.ActionRequest) contracts.ActionResult {
	if outbound.Disabled() {
		var cfg actionConfig
		_ = json.Unmarshal(req.ConfigJSON, &cfg)
		input := actionInput(actionPlan{ActionKey: req.ActionKey, ActionTypeKey: req.ActionTypeKey, Config: cfg}, req.RunInputJSON)
		return failedAction(input, "{}", map[string]any{
			"provider":              providerNameForAction(req.ActionTypeKey),
			"ok":                    false,
			"outbound_effects_mode": "disabled",
			"error":                 outbound.ErrDisabled.Error(),
		}, false, "outbound_disabled", outbound.ErrDisabled).contractResult()
	}
	if h.runner == nil || h.runner.P == nil {
		return contracts.ActionResult{
			OutputJSON:           json.RawMessage(`{}`),
			ProviderResponseJSON: json.RawMessage(`{"provider":"runtime","ok":false}`),
			ErrorType:            "runtime",
			Err:                  errProviderRuntimeUnavailable,
		}
	}
	if h.adapter.execute == nil {
		err := fmt.Errorf("provider action handler is not registered: %s", h.adapter.typeKey)
		return failedAction("{}", "{}", map[string]any{
			"provider":        "runtime",
			"ok":              false,
			"action_type_key": h.adapter.typeKey,
			"error":           err.Error(),
		}, false, "missing_handler", err).contractResult()
	}
	return h.adapter.execute(ctx, h.runner, req)
}

type providerActionAdapter struct {
	typeKey string
	execute func(context.Context, *Runner, contracts.ActionRequest) contracts.ActionResult
}

type providerExecutor func(context.Context, actionPlan, []byte, map[string]any, string) actionExecution

func (r *Runner) executeProviderAction(ctx context.Context, req contracts.ActionRequest, execute providerExecutor) contracts.ActionResult {
	var cfg actionConfig
	_ = json.Unmarshal(req.ConfigJSON, &cfg)
	node := actionPlan{ActionKey: req.ActionKey, ActionTypeKey: req.ActionTypeKey, Config: cfg}
	return execute(ctx, node, req.RunInputJSON, req.Steps, req.WorkflowRunID).contractResult()
}

func (final actionExecution) contractResult() contracts.ActionResult {
	return contracts.ActionResult{
		InputJSON:            json.RawMessage(final.InputJSON),
		OutputJSON:           json.RawMessage(final.OutputJSON),
		ProviderResponseJSON: json.RawMessage(final.ProviderResponseJSON),
		Waiting:              final.Waiting,
		WaitingReason:        final.WaitingReason,
		WaitingDeadline:      final.WaitingDeadline,
		Retryable:            final.Retryable,
		ErrorType:            final.ErrorType,
		Err:                  final.Err,
	}
}

func (r *Runner) httpClient() *http.Client {
	if r != nil && r.HTTPClient != nil {
		return r.HTTPClient
	}
	return httptransport.NewDefaultClient(30 * time.Second)
}

func (r *Runner) maxHTTPClient() *http.Client {
	if r != nil && r.MaxHTTPClient != nil {
		return r.MaxHTTPClient
	}
	return r.httpClient()
}

type actionPlan struct {
	ActionKey     string       `json:"action_key"`
	ActionTypeKey string       `json:"action_type_key"`
	Config        actionConfig `json:"config"`
}

type actionConfig struct {
	ActionDependencyID     string             `json:"action_dependency_id"`
	OperationKey           string             `json:"operation_key"`
	OrderID                string             `json:"order_id"`
	OrderItemID            string             `json:"order_item_id"`
	InstanceID             string             `json:"instance_id"`
	CertificateNumber      string             `json:"certificate_number"`
	ProductKind            string             `json:"product_kind"`
	ProductTitle           string             `json:"product_title"`
	AmountRub              string             `json:"amount_rub"`
	ServiceIDsJSON         string             `json:"service_ids_json"`
	SMTPActionDependencyID string             `json:"smtp_action_dependency_id"`
	MessageTemplateID      string             `json:"message_template_id"`
	EmailTemplateID        string             `json:"email_template_id"`
	RetryPolicy            retryPolicy        `json:"retry_policy"`
	SuccessCriteria        []successCriterion `json:"success_criteria"`
	TimeoutSeconds         int                `json:"timeout_seconds"`
	Text                   string             `json:"text"`
	SourceText             string             `json:"source_text"`
	Instructions           string             `json:"instructions"`
	Model                  string             `json:"model"`
	Temperature            *float64           `json:"temperature"`
	MaxTokens              int                `json:"max_tokens"`
	Channel                string             `json:"channel"`
	Note                   string             `json:"note"`
	UName                  string             `json:"u_name"`
	USurname               string             `json:"u_surname"`
	USname                 string             `json:"u_sname"`
	UPhone                 string             `json:"u_phone"`
	UEmail                 string             `json:"u_email"`
	DocumentURL            string             `json:"document_url"`
	DocumentURLsJSON       string             `json:"document_urls_json"`
	DocumentAllowedHosts   []string           `json:"document_allowed_hosts"`
	DocumentRequired       bool               `json:"document_required"`
	Caption                string             `json:"caption"`
	Phone                  string             `json:"phone"`
	ExtraFieldsJSON        string             `json:"extra_fields_json"`
	ChatID                 string             `json:"chat_id"`
	UserID                 string             `json:"user_id"`
	MessageID              string             `json:"message_id"`
	ReplyToMessageID       string             `json:"reply_to_message_id"`
	CallbackQueryID        string             `json:"callback_query_id"`
	Format                 string             `json:"format"`
	Notify                 *bool              `json:"notify"`
	DisableLinkPreview     *bool              `json:"disable_link_preview"`
	ParseMode              string             `json:"parse_mode"`
	ReplyMarkupJSON        string             `json:"reply_markup_json"`
	DisableNotification    bool               `json:"disable_notification"`
	ShowAlert              bool               `json:"show_alert"`
	To                     string             `json:"to"`
	Subject                string             `json:"subject"`
	Body                   string             `json:"body"`
	BodyFormat             string             `json:"body_format"`
}

type retryPolicy struct {
	Timeout string `json:"timeout"`
}

type successCriterion struct {
	Type     string `json:"type"`
	JSONPath string `json:"json_path"`
	Equals   string `json:"equals"`
}

type actionExecution struct {
	InputJSON            string
	OutputJSON           string
	ProviderResponseJSON string
	Waiting              bool
	WaitingReason        string
	WaitingDeadline      time.Time
	Retryable            bool
	ErrorType            string
	Err                  error
}

type dependencyRecord struct {
	TypeKey   string
	Status    string
	Metadata  map[string]any
	Plaintext string
}

func (r *Runner) loadMessageTemplate(ctx context.Context, templateID string) (string, error) {
	var body string
	err := r.P.QueryRow(ctx, `SELECT body_template::text FROM pulse_read.message_templates_v1 WHERE message_template_id = $1`, templateID).Scan(&body)
	return body, err
}

func (r *Runner) loadEmailTemplate(ctx context.Context, templateID string) (string, string, string, error) {
	var subject, body, bodyFormat string
	err := r.P.QueryRow(ctx, `
SELECT subject_template::text, body_template::text, body_format::text
  FROM pulse_read.email_templates_v1 WHERE email_template_id = $1`, templateID).Scan(&subject, &body, &bodyFormat)
	return subject, body, bodyFormat, err
}

func actionInput(node actionPlan, runInput []byte) string {
	b, _ := json.Marshal(map[string]any{
		"config":           node.Config,
		"success_criteria": node.Config.SuccessCriteria,
	})
	return string(b)
}

func skippedOutputWithReason(node actionPlan, reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "skipped"
	}
	output := map[string]any{
		"skipped":         true,
		"reason":          reason,
		"action_key":      node.ActionKey,
		"action_type_key": node.ActionTypeKey,
	}
	b, _ := json.Marshal(output)
	return string(b)
}

func failedAction(inputJSON, outputJSON string, provider map[string]any, retryable bool, errorType string, err error) actionExecution {
	provider["provider_response_redacted"] = true
	providerJSON, _ := json.Marshal(provider)
	report := diagnosticsForFailure("", provider, retryable, errorType, err)
	output := providerdiagnostics.OutputWithReport(outputJSON, report)
	return actionExecution{InputJSON: inputJSON, OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: errorType, Err: err}
}

func providerNameForAction(actionTypeKey string) string {
	switch actionTypeKey {
	case "telegram.send_message", "telegram.send_document", "telegram.edit_message_text", "telegram.answer_callback_query", "telegram.pin_message", "telegram.unpin_message":
		return "telegram"
	case "max.send_message":
		return "max"
	case "email.send":
		return "smtp"
	case "uon.create_lead":
		return "uon"
	case "yclients.certificate.issue":
		return "yclients"
	case "ai.apinet.normalize_text":
		return "ai.apinet"
	default:
		return "runtime"
	}
}

func diagnosticsForFailure(actionTypeKey string, provider map[string]any, retryable bool, errorType string, err error) providerdiagnostics.Report {
	if errorType == "dependency" || errorType == "dependency_type" || errorType == "dependency_status" {
		return providerdiagnostics.DependencyFailure(provider)
	}
	if errorType == "validation" {
		return providerdiagnostics.ValidationFailure(provider)
	}
	base := providerdiagnostics.BaseFailure(provider, retryable, err)
	providerKey := providerdiagnostics.ProviderName(provider)
	switch {
	case actionTypeKey == "telegram.send_message" || providerKey == "telegram":
		return telegramprovider.Analyze(provider, retryable, base)
	case actionTypeKey == "max.send_message" || providerKey == "max":
		return maxprovider.Analyze(provider, retryable, base)
	case actionTypeKey == "email.send" || providerKey == "smtp":
		return emailprovider.AnalyzeSMTP(provider, retryable, base)
	case actionTypeKey == "uon.create_lead" || providerKey == "uon":
		return uonprovider.Analyze(provider, retryable, base)
	case actionTypeKey == "yclients.certificate.issue" || providerKey == "yclients":
		return yclientsprovider.Analyze(provider, retryable, base)
	default:
		if errorType == "network" {
			base.ReasonCode = "network_error"
			base.HumanReason = "Не удалось выполнить сетевой запрос к провайдеру."
			base.NextAction = "Проверьте доступность провайдера и повторите действие."
		}
		return base
	}
}

type CriterionResult = providerdiagnostics.CriterionResult

func defaultSuccessCriteria(actionTypeKey string) []successCriterion {
	switch actionTypeKey {
	case "telegram.send_message", "telegram.edit_message_text", "telegram.pin_message", "telegram.unpin_message":
		return specsToCriteria(telegramprovider.Criteria())
	case "telegram.answer_callback_query":
		return specsToCriteria(telegramprovider.CallbackCriteria())
	case "max.send_message":
		return specsToCriteria(maxprovider.Criteria())
	case "email.send":
		return specsToCriteria(emailprovider.Criteria())
	case "uon.create_lead":
		return specsToCriteria(uonprovider.Criteria())
	case "yclients.certificate.issue":
		return specsToCriteria(yclientsprovider.Criteria())
	default:
		return nil
	}
}

func evaluateSuccessCriteria(actionTypeKey string, criteria []successCriterion, provider map[string]any) ([]CriterionResult, error) {
	if results, err, ok := evaluateProviderCriteria(actionTypeKey, provider); ok {
		return results, err
	}
	list := criteria
	if len(list) == 0 {
		list = defaultSuccessCriteria(actionTypeKey)
	}
	if len(list) == 0 {
		return nil, nil
	}
	var out []CriterionResult
	for _, c := range list {
		passed, detail, expected, actual, source := evalCriterion(c, provider)
		out = append(out, CriterionResult{Type: c.Type, Label: criterionLabel(c.Type), Passed: passed, Expected: expected, Actual: actual, Source: source, Detail: detail})
		if !passed {
			return out, fmt.Errorf("%s: %s", c.Type, detail)
		}
	}
	return out, nil
}

func evaluateProviderCriteria(actionTypeKey string, provider map[string]any) ([]CriterionResult, error, bool) {
	switch actionTypeKey {
	case "telegram.send_message", "telegram.edit_message_text", "telegram.pin_message", "telegram.unpin_message":
		results, err := telegramprovider.Evaluate(provider)
		return results, err, true
	case "telegram.answer_callback_query":
		results, err := telegramprovider.EvaluateCallback(provider)
		return results, err, true
	case "max.send_message":
		results, err := maxprovider.Evaluate(provider)
		return results, err, true
	case "email.send":
		results, err := emailprovider.Evaluate(provider)
		return results, err, true
	case "uon.create_lead":
		results, err := uonprovider.Evaluate(provider)
		return results, err, true
	case "yclients.certificate.issue":
		results, err := yclientsprovider.Evaluate(provider)
		return results, err, true
	default:
		return nil, nil, false
	}
}

func specsToCriteria(specs []providerdiagnostics.CriterionSpec) []successCriterion {
	out := make([]successCriterion, 0, len(specs))
	for _, spec := range specs {
		out = append(out, successCriterion{Type: spec.Type})
	}
	return out
}

func criterionLabel(kind string) string {
	switch strings.TrimSpace(kind) {
	case "http_status_2xx":
		return "HTTP ответ 2xx"
	case "telegram_api_ok":
		return "Telegram API ok=true"
	case "max_api_ok":
		return "MAX API принял сообщение"
	case "smtp_session_ok":
		return "SMTP сервер принял письмо"
	case "json_path_equals":
		return "JSON path равен значению"
	default:
		return kind
	}
}

func evalCriterion(c successCriterion, provider map[string]any) (bool, string, string, string, string) {
	switch strings.TrimSpace(c.Type) {
	case "http_status_2xx":
		return providerdiagnostics.HTTPStatus2xx(provider)
	case "telegram_api_ok":
		if v, ok := provider["ok"].(bool); ok && v {
			return true, "telegram ok=true", "true", "true", "provider_response_json.ok"
		}
		return false, "telegram ok is not true", "true", providerdiagnostics.StringifyJSONValue(provider["ok"]), "provider_response_json.ok"
	case "max_api_ok":
		if v, exists := provider["ok"]; exists {
			if b, ok := v.(bool); ok && !b {
				return false, "MAX ok=false", "true или отсутствие ok при HTTP 2xx", "false", "provider_response_json.ok"
			}
			if b, ok := v.(bool); ok && b {
				return true, "MAX ok=true", "true или отсутствие ok при HTTP 2xx", "true", "provider_response_json.ok"
			}
		}
		return true, "MAX response accepted (no ok field)", "true или отсутствие ok при HTTP 2xx", "ok отсутствует", "provider_response_json.ok"
	case "smtp_session_ok":
		smtp, ok := provider["smtp"].(map[string]any)
		if !ok {
			return false, "missing smtp envelope", "smtp.accepted=true", "smtp envelope отсутствует", "provider_response_json.smtp.accepted"
		}
		if accepted, ok := smtp["accepted"].(bool); ok && accepted {
			return true, "SMTP session accepted", "smtp.accepted=true", "true", "provider_response_json.smtp.accepted"
		}
		return false, "SMTP session not accepted", "smtp.accepted=true", providerdiagnostics.StringifyJSONValue(smtp["accepted"]), "provider_response_json.smtp.accepted"
	case "json_path_equals":
		path := strings.TrimSpace(c.JSONPath)
		want := strings.TrimSpace(c.Equals)
		if path == "" {
			return false, "json_path is empty", want, "", "provider_response_json"
		}
		got := providerdiagnostics.StringifyAtPath(provider, path)
		if got == want {
			return true, fmt.Sprintf("%q equals %q", path, want), want, got, "provider_response_json." + path
		}
		return false, fmt.Sprintf("%q got %q want %q", path, got, want), want, got, "provider_response_json." + path
	default:
		return false, fmt.Sprintf("unknown criterion type %q", c.Type), "", "", "success_criteria"
	}
}

func finalizeMessagingOutput(actionTypeKey, actionKey string, successCriteria []successCriterion, requestInputJSON string, provider map[string]any, deliveredPayload map[string]any, expose map[string]any) actionExecution {
	providerJSON, _ := json.Marshal(provider)
	cr, cerr := evaluateSuccessCriteria(actionTypeKey, successCriteria, provider)
	deliveredPayload["action_key"] = actionKey
	deliveredPayload["action_type_key"] = actionTypeKey
	if len(cr) > 0 {
		deliveredPayload["criteria_results"] = cr
		deliveredPayload["criteria_passed"] = cerr == nil
	}
	if len(expose) > 0 {
		deliveredPayload["expose"] = expose
	}
	if cerr != nil {
		deliveredPayload["diagnostics"] = providerdiagnostics.CriteriaFailure(actionTypeKey, provider, cr, cerr)
	} else {
		deliveredPayload["diagnostics"] = providerdiagnostics.Success(actionTypeKey, provider, cr)
	}
	outJSON, _ := json.Marshal(deliveredPayload)
	ex := actionExecution{InputJSON: requestInputJSON, OutputJSON: string(outJSON), ProviderResponseJSON: string(providerJSON)}
	if cerr != nil {
		ex.Retryable = false
		ex.ErrorType = "success_criteria"
		ex.Err = cerr
	}
	return ex
}
