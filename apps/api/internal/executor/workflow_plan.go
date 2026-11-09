package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/runtime/contracts"
	runtimeengine "riders.agency/pulse/api/internal/runtime/engine"
)

type actionPlan struct {
	ActionKey     string       `json:"action_key"`
	ActionTypeKey string       `json:"action_type_key"`
	Config        actionConfig `json:"config"`
}

type actionConfig struct {
	Disabled           bool   `json:"disabled,omitempty"`
	ActionDependencyID string `json:"action_dependency_id"`
	yclientsActionConfig
	SMTPActionDependencyID string             `json:"smtp_action_dependency_id"`
	MessageTemplateID      string             `json:"message_template_id"`
	EmailTemplateID        string             `json:"email_template_id"`
	RetryPolicy            retryPolicy        `json:"retry_policy"`
	SuccessCriteria        []successCriterion `json:"success_criteria"`
	Enabled                *bool              `json:"enabled"`
	ActiveFrom             string             `json:"active_from"`
	ActiveUntil            string             `json:"active_until"`
	ScheduleKey            string             `json:"schedule_key"`
	IntervalSeconds        int                `json:"interval_seconds"`
	RemindEveryDays        int                `json:"remind_every_days"`
	DaysBeforeDue          int                `json:"days_before_due"`
	Timezone               string             `json:"timezone"`
	SamplePayloadJSON      string             `json:"sample_payload_json"`
	Method                 string             `json:"method"`
	URL                    string             `json:"url"`
	HeadersJSON            string             `json:"headers_json"`
	TimeoutSeconds         int                `json:"timeout_seconds"`
	Text                   string             `json:"text"`
	SourceText             string             `json:"source_text"`
	Instructions           string             `json:"instructions"`
	Model                  string             `json:"model"`
	Temperature            *float64           `json:"temperature"`
	MaxTokens              int                `json:"max_tokens"`
	ChatID                 string             `json:"chat_id"`
	MessageID              string             `json:"message_id"`
	ReplyToMessageID       string             `json:"reply_to_message_id"`
	CallbackQueryID        string             `json:"callback_query_id"`
	ParseMode              string             `json:"parse_mode"`
	ReplyMarkupJSON        string             `json:"reply_markup_json"`
	DisableNotification    bool               `json:"disable_notification"`
	ShowAlert              bool               `json:"show_alert"`
	To                     string             `json:"to"`
	Subject                string             `json:"subject"`
	Body                   string             `json:"body"`
	BodyFormat             string             `json:"body_format"`
	StateKey               string             `json:"state_key"`
	ContextJSON            string             `json:"context_json"`
	MergeContext           bool               `json:"merge_context"`
	ContinueOnFailure      bool               `json:"continue_on_failure"`
	FailureRequired        bool               `json:"failure_required"`
	// Webhook trigger
	AuthType    string          `json:"auth_type"`
	EndpointKey string          `json:"endpoint_key"`
	JSONFields  json.RawMessage `json:"json_fields"`
	// Person normalize
	Input        string            `json:"input"`
	OutputFields map[string]string `json:"output_fields"`
	// U-ON
	Channel         string `json:"channel"`
	Note            string `json:"note"`
	UName           string `json:"u_name"`
	USurname        string `json:"u_surname"`
	USname          string `json:"u_sname"`
	UPhone          string `json:"u_phone"`
	UEmail          string `json:"u_email"`
	ExtraFieldsJSON string `json:"extra_fields_json"`
	Phone           string `json:"phone"`
	// Message documents
	DocumentURL          string   `json:"document_url"`
	DocumentURLsJSON     string   `json:"document_urls_json"`
	DocumentAllowedHosts []string `json:"document_allowed_hosts"`
	DocumentRequired     bool     `json:"document_required"`
	Caption              string   `json:"caption"`
}

func (e *Engine) executeAction(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID, projectID string, attempt int) actionExecution {
	if node.Config.Disabled {
		return actionExecution{InputJSON: actionInput(node, runInput), OutputJSON: skippedOutputWithReason(node, "disabled"), ProviderResponseJSON: "{}"}
	}
	if final, ok := e.executeRegisteredAction(ctx, node, runInput, steps, runID, projectID, attempt); ok {
		return final
	}
	return missingActionHandler(node, runInput)
}

func (e *Engine) executeRegisteredAction(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID, projectID string, attempt int) (actionExecution, bool) {
	if e.Registry == nil {
		return actionExecution{}, false
	}
	atk := strings.TrimSpace(node.ActionTypeKey)
	handler, ok := e.Registry.Action(atk)
	if !ok {
		return actionExecution{}, false
	}
	configJSON, err := json.Marshal(node.Config)
	if err != nil {
		return failedAction(actionInput(node, runInput), "{}", map[string]any{
			"provider": "runtime",
			"error":    "action config serialization failed",
		}, false, "config", err), true
	}
	deadline := time.Time{}
	if node.Config.TimeoutSeconds > 0 {
		deadline = time.Now().UTC().Add(time.Duration(node.Config.TimeoutSeconds) * time.Second)
	}
	result := handler.Execute(ctx, contracts.ActionRequest{
		WorkflowRunID: runID,
		ProjectID:     projectID,
		ActionKey:     node.ActionKey,
		ActionTypeKey: atk,
		ConfigJSON:    configJSON,
		RunInputJSON:  json.RawMessage(runInput),
		Steps:         steps,
		Attempt:       attempt,
		Deadline:      deadline,
	})
	return actionExecution{
		InputJSON:            runtimeengine.RawJSONOrDefault(result.InputJSON, actionInput(node, runInput)),
		OutputJSON:           runtimeengine.RawJSONOrDefault(result.OutputJSON, "{}"),
		ProviderResponseJSON: runtimeengine.RawJSONOrDefault(result.ProviderResponseJSON, "{}"),
		Waiting:              result.Waiting,
		WaitingReason:        result.WaitingReason,
		WaitingDeadline:      result.WaitingDeadline,
		Retryable:            result.Retryable,
		ErrorType:            result.ErrorType,
		Err:                  result.Err,
	}, true
}

func missingActionHandler(node actionPlan, runInput []byte) actionExecution {
	atk := strings.TrimSpace(node.ActionTypeKey)
	if atk == "" {
		atk = "unknown"
	}
	err := fmt.Errorf("runtime action handler is not registered: %s", atk)
	slog.Warn("runtime action handler missing",
		slog.String("action_type_key", atk),
		slog.String("action_key", strings.TrimSpace(node.ActionKey)),
	)
	return failedAction(actionInput(node, runInput), "{}", map[string]any{
		"provider":        "runtime",
		"ok":              false,
		"action_type_key": atk,
		"error":           err.Error(),
	}, false, "missing_handler", err)
}

type retryPolicy struct {
	MaxAttempts     int    `json:"max_attempts"`
	Strategy        string `json:"strategy"`
	InitialInterval string `json:"initial_interval"`
	MaxInterval     string `json:"max_interval"`
	MaxElapsed      string `json:"max_elapsed"`
	Timeout         string `json:"timeout"`
}

type successCriterion struct {
	Type     string `json:"type"`
	JSONPath string `json:"json_path"`
	Equals   string `json:"equals"`
}

type wfDef struct {
	Nodes []actionPlan `json:"nodes"`
	Edges []edgePlan   `json:"edges"`
}

type edgePlan struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Branch string `json:"branch"`
	Label  string `json:"label"`
}

func flattenNodes(def []byte) []actionPlan {
	var d wfDef
	var root map[string]json.RawMessage
	hasNodesField := false
	if err := json.Unmarshal(def, &root); err == nil {
		_, hasNodesField = root["nodes"]
	}
	_ = json.Unmarshal(def, &d)
	out := make([]actionPlan, 0, len(d.Nodes))
	for i := range d.Nodes {
		node := &d.Nodes[i]
		node.ActionKey = strings.TrimSpace(node.ActionKey)
		node.ActionTypeKey = strings.TrimSpace(node.ActionTypeKey)
		if node.ActionKey == "" {
			node.ActionKey = node.ActionTypeKey
		}
		if node.ActionTypeKey == "" {
			node.ActionTypeKey = "unknown"
		}
		if node.ActionKey != "" {
			out = append(out, *node)
		}
	}
	if len(out) == 0 && !hasNodesField {
		out = []actionPlan{{ActionKey: "entry", ActionTypeKey: "webhook.trigger"}}
	}
	return out
}

func flattenEdges(def []byte) []edgePlan {
	var d wfDef
	_ = json.Unmarshal(def, &d)
	out := make([]edgePlan, 0, len(d.Edges))
	for _, edge := range d.Edges {
		from := strings.TrimSpace(edge.From)
		to := strings.TrimSpace(edge.To)
		if from == "" || to == "" {
			continue
		}
		branch := strings.TrimSpace(edge.Branch)
		if branch == "" {
			branch = strings.TrimSpace(edge.Label)
		}
		out = append(out, edgePlan{From: from, To: to, Branch: branch})
	}
	return out
}

type runNodePlan struct {
	Execute          []actionPlan
	Skipped          []actionPlan
	InitialSteps     map[string]any
	IgnoreGraphEdges bool
}

func planRunNodes(nodes []actionPlan, runInput []byte) (runNodePlan, error) {
	plan := runNodePlan{Execute: nodes, InitialSteps: map[string]any{}}
	if len(runInput) == 0 {
		return plan, nil
	}
	var root map[string]any
	if err := json.Unmarshal(runInput, &root); err != nil || root == nil {
		return plan, nil
	}
	pulse, ok := root["_pulse"].(map[string]any)
	if !ok || pulse == nil {
		return plan, nil
	}
	plan.InitialSteps = pulseSteps(pulse)
	if sk, ok := pulse["retry_action_key"].(string); ok {
		return onlyActionPlan(nodes, sk, plan.InitialSteps)
	}
	actionKey, _ := pulse["test_action_key"].(string)
	actionKey = strings.TrimSpace(actionKey)
	if actionKey == "" {
		return plan, nil
	}
	mode, _ := pulse["test_mode"].(string)
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "through_action"
	}
	idx := actionIndex(nodes, actionKey)
	if idx < 0 {
		return plan, fmt.Errorf("test_action_key %q not found in workflow definition", actionKey)
	}
	switch mode {
	case "selected_only":
		skipped := append([]actionPlan{}, nodes[:idx]...)
		skipped = append(skipped, nodes[idx+1:]...)
		return runNodePlan{Execute: []actionPlan{nodes[idx]}, Skipped: skipped, InitialSteps: plan.InitialSteps}, nil
	case "through_action":
		return runNodePlan{Execute: nodes[:idx+1], Skipped: nodes[idx+1:], InitialSteps: plan.InitialSteps}, nil
	case "full":
		return plan, nil
	default:
		return plan, fmt.Errorf("unsupported test_mode %q", mode)
	}
}

func onlyActionPlan(nodes []actionPlan, actionKey string, steps map[string]any) (runNodePlan, error) {
	actionKey = strings.TrimSpace(actionKey)
	if actionKey == "" {
		return runNodePlan{Execute: nodes, InitialSteps: steps}, nil
	}
	idx := actionIndex(nodes, actionKey)
	if idx < 0 {
		return runNodePlan{InitialSteps: steps}, fmt.Errorf("retry_action_key %q not found in workflow definition", actionKey)
	}
	return runNodePlan{Execute: []actionPlan{nodes[idx]}, InitialSteps: steps, IgnoreGraphEdges: true}, nil
}

func actionIndex(nodes []actionPlan, actionKey string) int {
	for i, node := range nodes {
		if node.ActionKey == actionKey {
			return i
		}
	}
	return -1
}

func shouldRunGraphNode(node actionPlan, edges []edgePlan, completed map[string]bool, branchOutputs map[string]string) bool {
	incoming := make([]edgePlan, 0, 1)
	for _, edge := range edges {
		if edge.To == node.ActionKey {
			incoming = append(incoming, edge)
		}
	}
	if len(incoming) == 0 {
		return true
	}
	for _, edge := range incoming {
		if !completed[edge.From] {
			continue
		}
		wantBranch := strings.TrimSpace(edge.Branch)
		if wantBranch == "" {
			wantBranch = strings.TrimSpace(edge.Label)
		}
		if wantBranch == "" {
			return true
		}
		if strings.EqualFold(branchOutputs[edge.From], wantBranch) {
			return true
		}
	}
	return false
}

func branchFromOutput(outputJSON string) string {
	var output map[string]any
	if err := json.Unmarshal([]byte(outputJSON), &output); err != nil || output == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(output["branch"]))
}
