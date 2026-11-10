package workflows

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/outbox"
)

func TestValidateWorkflowRunRequestedPayload(t *testing.T) {
	payload := workflowRunRequestedPayload{
		EventVersion:      WorkflowRunRequestedEventVersion,
		ProjectID:         uuid.NewString(),
		WorkflowID:        uuid.NewString(),
		WorkflowVersionID: uuid.NewString(),
		SourceType:        "task_event",
		InputJSON:         []byte(`{"ok":true}`),
	}
	ev := outbox.Event{IdempotencyKey: "workflow.run.requested:" + uuid.NewString()}
	if err := validateWorkflowRunRequestedPayload(payload, ev); err != nil {
		t.Fatalf("valid payload failed: %v", err)
	}
}

func TestValidateWorkflowRunRequestedPayloadRejectsMissingFields(t *testing.T) {
	payload := workflowRunRequestedPayload{
		EventVersion:      WorkflowRunRequestedEventVersion,
		ProjectID:         uuid.NewString(),
		WorkflowVersionID: uuid.NewString(),
		SourceType:        "task_event",
	}
	err := validateWorkflowRunRequestedPayload(payload, outbox.Event{IdempotencyKey: "idem"})
	if err == nil || !strings.Contains(err.Error(), "workflow_id") {
		t.Fatalf("expected workflow_id error, got %v", err)
	}
}

func TestValidateWorkflowRunRequestedPayloadRejectsVersionDrift(t *testing.T) {
	payload := workflowRunRequestedPayload{
		EventVersion:      "workflow.run.requested.v0",
		ProjectID:         uuid.NewString(),
		WorkflowID:        uuid.NewString(),
		WorkflowVersionID: uuid.NewString(),
		SourceType:        "task_event",
	}
	err := validateWorkflowRunRequestedPayload(payload, outbox.Event{IdempotencyKey: "idem"})
	if err == nil || !strings.Contains(err.Error(), "unsupported event_version") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestWorkflowRunIDPrefersPayloadID(t *testing.T) {
	payloadID := "11111111-1111-1111-1111-111111111111"
	aggregateID := "22222222-2222-2222-2222-222222222222"
	got, err := workflowRunID(payloadID, aggregateID)
	if err != nil {
		t.Fatalf("workflowRunID: %v", err)
	}
	if got.String() != payloadID {
		t.Fatalf("run id = %s, want %s", got, payloadID)
	}
}

func TestWorkflowRunIDUsesAggregateID(t *testing.T) {
	aggregateID := "22222222-2222-2222-2222-222222222222"
	got, err := workflowRunID("", aggregateID)
	if err != nil {
		t.Fatalf("workflowRunID: %v", err)
	}
	if got.String() != aggregateID {
		t.Fatalf("run id = %s, want %s", got, aggregateID)
	}
}

func TestWorkflowRunIDRejectsBadExplicitID(t *testing.T) {
	if _, err := workflowRunID("bad", ""); err == nil {
		t.Fatal("expected bad id error")
	}
}
