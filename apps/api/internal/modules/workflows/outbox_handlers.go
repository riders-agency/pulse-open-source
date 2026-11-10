package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	outboxpkg "riders.agency/pulse/api/internal/platform/outbox"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

const WorkflowRunRequestedEvent = "workflow.run.requested"
const WorkflowRunRequestedEventVersion = WorkflowRunRequestedEvent + ".v1"

type NotifyFunc func()

func RegisterOutboxHandlers(reg *outboxpkg.Registry, pool *pgxpool.Pool, notify NotifyFunc) error {
	if reg == nil || pool == nil {
		return nil
	}
	return reg.Register(&workflowRunRequestedHandler{pool: pool, notify: notify})
}

type workflowRunRequestedHandler struct {
	pool   *pgxpool.Pool
	notify NotifyFunc
}

func (h *workflowRunRequestedHandler) TypeKey() string {
	return WorkflowRunRequestedEvent
}

type workflowRunRequestedPayload struct {
	EventVersion              string          `json:"event_version,omitempty"`
	WorkflowRunID             string          `json:"workflow_run_id,omitempty"`
	ProjectID                 string          `json:"project_id"`
	WorkflowID                string          `json:"workflow_id"`
	WorkflowVersionID         string          `json:"workflow_version_id"`
	SourceType                string          `json:"source_type"`
	SourceWebhookEventID      string          `json:"source_webhook_event_id,omitempty"`
	ReplayedFromWorkflowRunID string          `json:"replayed_from_workflow_run_id,omitempty"`
	IdempotencyKey            string          `json:"idempotency_key,omitempty"`
	SourceMetadataJSON        json.RawMessage `json:"source_metadata_json,omitempty"`
	InputJSON                 json.RawMessage `json:"input_json,omitempty"`
	EventActionKey            string          `json:"event_action_key,omitempty"`
	EventType                 string          `json:"event_type,omitempty"`
	EventMessage              string          `json:"event_message,omitempty"`
	EventPayloadJSON          json.RawMessage `json:"event_payload_json,omitempty"`
}

func (h *workflowRunRequestedHandler) Handle(ctx context.Context, ev outboxpkg.Event) error {
	if h == nil || h.pool == nil {
		return fmt.Errorf("workflow run requested handler is not configured")
	}
	var payload workflowRunRequestedPayload
	if err := json.Unmarshal(ev.PayloadJSON, &payload); err != nil {
		return fmt.Errorf("workflow run requested payload: %w", err)
	}
	if err := validateWorkflowRunRequestedPayload(payload, ev); err != nil {
		return err
	}
	runID, err := workflowRunID(payload.WorkflowRunID, ev.AggregateID)
	if err != nil {
		return err
	}
	input := payload.InputJSON
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	message := strings.TrimSpace(payload.EventMessage)
	if message == "" {
		message = "workflow run queued from outbox"
	}
	eventType := strings.TrimSpace(payload.EventType)
	if eventType == "" {
		eventType = "workflow.started"
	}
	eventPayload := payload.EventPayloadJSON
	if len(eventPayload) == 0 {
		eventPayload = input
	}
	actionKey := strings.TrimSpace(payload.EventActionKey)

	_, inserted, err := runqueue.RequestRun(ctx, h.pool, runqueue.Request{
		WorkflowRunID:             runID,
		ProjectID:                 payload.ProjectID,
		WorkflowID:                payload.WorkflowID,
		WorkflowVersionID:         payload.WorkflowVersionID,
		SourceType:                payload.SourceType,
		SourceWebhookEventID:      payload.SourceWebhookEventID,
		ReplayedFromWorkflowRunID: payload.ReplayedFromWorkflowRunID,
		IdempotencyKey:            firstNonEmpty(payload.IdempotencyKey, ev.IdempotencyKey),
		SourceMetadataJSON:        payload.SourceMetadataJSON,
		InputJSON:                 input,
		EventActionKey:            actionKey,
		EventType:                 eventType,
		EventMessage:              message,
		EventPayloadJSON:          eventPayload,
	})
	if err != nil {
		return err
	}
	if inserted {
		slog.Info("workflow run requested outbox processed",
			slog.String("workflow_run_id", runID.String()),
			slog.String("workflow_id", payload.WorkflowID),
			slog.String("workflow_version_id", payload.WorkflowVersionID),
			slog.String("project_id", payload.ProjectID),
			slog.String("source_type", payload.SourceType),
		)
		if h.notify != nil {
			h.notify()
		}
	}
	return nil
}

func validateWorkflowRunRequestedPayload(payload workflowRunRequestedPayload, ev outboxpkg.Event) error {
	if version := strings.TrimSpace(payload.EventVersion); version != "" && version != WorkflowRunRequestedEventVersion {
		return fmt.Errorf("workflow run requested unsupported event_version %q", version)
	}
	for name, value := range map[string]string{
		"project_id":          payload.ProjectID,
		"workflow_id":         payload.WorkflowID,
		"workflow_version_id": payload.WorkflowVersionID,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("workflow run requested %s is required", name)
		}
		if _, err := uuid.Parse(value); err != nil {
			return fmt.Errorf("workflow run requested %s is not a uuid", name)
		}
	}
	if strings.TrimSpace(payload.SourceType) == "" {
		return fmt.Errorf("workflow run requested source_type is required")
	}
	if strings.TrimSpace(ev.IdempotencyKey) == "" && strings.TrimSpace(payload.IdempotencyKey) == "" {
		return fmt.Errorf("workflow run requested idempotency key is required")
	}
	if len(payload.InputJSON) > 0 && !json.Valid(payload.InputJSON) {
		return fmt.Errorf("workflow run requested input_json is invalid")
	}
	if len(payload.SourceMetadataJSON) > 0 && !json.Valid(payload.SourceMetadataJSON) {
		return fmt.Errorf("workflow run requested source_metadata_json is invalid")
	}
	if len(payload.EventPayloadJSON) > 0 && !json.Valid(payload.EventPayloadJSON) {
		return fmt.Errorf("workflow run requested event_payload_json is invalid")
	}
	return nil
}

func workflowRunID(payloadID, aggregateID string) (uuid.UUID, error) {
	for _, value := range []string{payloadID, aggregateID} {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil {
			return uuid.Nil, fmt.Errorf("bad workflow_run_id %q", value)
		}
		return id, nil
	}
	return uuid.New(), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
