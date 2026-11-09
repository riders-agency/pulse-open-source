package runqueue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const CurrentRequestVersion = "run-request.v1"

const (
	SourceWebhook          = "webhook"
	SourceManual           = "manual"
	SourceSchedule         = "schedule"
	SourceDesktopMessage   = "desktop_message"
	SourceReplay           = "replay"
	SourceTest             = "test"
	SourceTelegramRelay    = "telegram_relay"
	SourceClientPaymentDue = "client_payment_due"
	SourceModuleTrigger    = "module_trigger"
	SourceTaskEvent        = "task_event"
)

type Request struct {
	ContractVersion           string
	WorkflowRunID             uuid.UUID
	ProjectID                 string
	WorkflowID                string
	WorkflowVersionID         string
	SourceType                string
	SourceWebhookEventID      string
	ReplayedFromWorkflowRunID string
	IdempotencyKey            string
	SourceMetadataJSON        json.RawMessage
	InputJSON                 json.RawMessage
	EventActionKey            string
	EventType                 string
	EventMessage              string
	EventPayloadJSON          json.RawMessage
}

type Result struct {
	WorkflowRunID uuid.UUID
	Inserted      bool
}

var allowedSourceTypes = map[string]struct{}{
	SourceWebhook:          {},
	SourceManual:           {},
	SourceSchedule:         {},
	SourceDesktopMessage:   {},
	SourceReplay:           {},
	SourceTest:             {},
	SourceTelegramRelay:    {},
	SourceClientPaymentDue: {},
	SourceModuleTrigger:    {},
	SourceTaskEvent:        {},
}

func normalizeRequest(req Request) (Request, error) {
	if req.WorkflowRunID == uuid.Nil {
		req.WorkflowRunID = uuid.New()
	}
	req.ContractVersion = strings.TrimSpace(req.ContractVersion)
	if req.ContractVersion == "" {
		req.ContractVersion = CurrentRequestVersion
	}
	if req.ContractVersion != CurrentRequestVersion {
		return Request{}, fmt.Errorf("unsupported run request version %q", req.ContractVersion)
	}
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.WorkflowID = strings.TrimSpace(req.WorkflowID)
	req.WorkflowVersionID = strings.TrimSpace(req.WorkflowVersionID)
	req.SourceType = strings.TrimSpace(req.SourceType)
	for name, value := range map[string]string{
		"project_id":          req.ProjectID,
		"workflow_id":         req.WorkflowID,
		"workflow_version_id": req.WorkflowVersionID,
		"source_type":         req.SourceType,
	} {
		if value == "" {
			return Request{}, fmt.Errorf("run queue %s is required", name)
		}
	}
	if _, ok := allowedSourceTypes[req.SourceType]; !ok {
		return Request{}, fmt.Errorf("run queue source_type %q is not allowed", req.SourceType)
	}
	req.IdempotencyKey = normalizeIdempotencyKey(req)
	var err error
	req.InputJSON, err = normalizeJSON(req.InputJSON, `{}`)
	if err != nil {
		return Request{}, fmt.Errorf("run queue input_json: %w", err)
	}
	req.EventPayloadJSON, err = normalizeJSON(req.EventPayloadJSON, string(req.InputJSON))
	if err != nil {
		return Request{}, fmt.Errorf("run queue event_payload_json: %w", err)
	}
	req.SourceMetadataJSON, err = normalizeJSON(req.SourceMetadataJSON, defaultSourceMetadata(req))
	if err != nil {
		return Request{}, fmt.Errorf("run queue source_metadata_json: %w", err)
	}
	req.EventActionKey = strings.TrimSpace(req.EventActionKey)
	req.EventType = strings.TrimSpace(req.EventType)
	if req.EventType == "" {
		req.EventType = "workflow.started"
	}
	req.EventMessage = strings.TrimSpace(req.EventMessage)
	if req.EventMessage == "" {
		req.EventMessage = "workflow run queued"
	}
	return req, nil
}

func normalizeIdempotencyKey(req Request) string {
	key := strings.TrimSpace(req.IdempotencyKey)
	if key != "" {
		return key
	}
	return "workflow_run:" + req.WorkflowRunID.String()
}

func normalizeJSON(raw json.RawMessage, fallback string) (json.RawMessage, error) {
	text := strings.TrimSpace(string(raw))
	if text == "" {
		text = fallback
	}
	if !json.Valid([]byte(text)) {
		return nil, fmt.Errorf("invalid JSON")
	}
	return json.RawMessage(text), nil
}

func defaultSourceMetadata(req Request) string {
	source := map[string]any{
		"contract_version": req.ContractVersion,
		"source_type":      req.SourceType,
	}
	if key := strings.TrimSpace(req.IdempotencyKey); key != "" {
		source["idempotency_key"] = key
	}
	if key := strings.TrimSpace(req.EventActionKey); key != "" {
		source["event_action_key"] = key
	}
	if id := strings.TrimSpace(req.SourceWebhookEventID); id != "" {
		source["source_webhook_event_id"] = id
	}
	if id := strings.TrimSpace(req.ReplayedFromWorkflowRunID); id != "" {
		source["replayed_from_workflow_run_id"] = id
	}
	raw, err := json.Marshal(source)
	if err != nil {
		return `{}`
	}
	return string(raw)
}
