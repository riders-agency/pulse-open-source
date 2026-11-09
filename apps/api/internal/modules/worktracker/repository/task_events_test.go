package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeTaskEventPublisher struct {
	called     bool
	companyID  uuid.UUID
	projectID  uuid.UUID
	workflowID uuid.UUID
	taskID     uuid.UUID
	eventKey   string
	payload    map[string]any
}

func (f *fakeTaskEventPublisher) EnqueueTaskModuleTriggers(_ context.Context, _ pgx.Tx, companyID, projectID, workflowID, taskID uuid.UUID, eventKey string, payload map[string]any) error {
	f.called = true
	f.companyID = companyID
	f.projectID = projectID
	f.workflowID = workflowID
	f.taskID = taskID
	f.eventKey = eventKey
	f.payload = payload
	return nil
}

func TestWorkTrackerTaskEventPublisherBridge(t *testing.T) {
	t.Parallel()

	publisher := &fakeTaskEventPublisher{}
	tracker := &WorkTracker{TaskEvents: publisher}
	companyID := uuid.New()
	projectID := uuid.New()
	workflowID := uuid.New()
	taskID := uuid.New()
	payload := map[string]any{"title": "prepare docs"}

	if err := tracker.enqueueTaskModuleTriggers(context.Background(), nil, companyID, projectID, workflowID, taskID, "task.updated", payload); err != nil {
		t.Fatalf("enqueue task module triggers: %v", err)
	}

	if !publisher.called {
		t.Fatal("expected task event publisher to be called")
	}
	if publisher.companyID != companyID || publisher.projectID != projectID || publisher.workflowID != workflowID || publisher.taskID != taskID {
		t.Fatalf("publisher received wrong ids: company=%s project=%s workflow=%s task=%s", publisher.companyID, publisher.projectID, publisher.workflowID, publisher.taskID)
	}
	if publisher.eventKey != "task.updated" {
		t.Fatalf("publisher event key = %q, want task.updated", publisher.eventKey)
	}
	if publisher.payload["title"] != "prepare docs" {
		t.Fatalf("publisher payload = %#v", publisher.payload)
	}
}

func TestWorkTrackerTaskEventPublisherNilNoops(t *testing.T) {
	t.Parallel()

	if err := (*WorkTracker)(nil).enqueueTaskModuleTriggers(context.Background(), nil, uuid.New(), uuid.New(), uuid.New(), uuid.New(), "task.updated", nil); err != nil {
		t.Fatalf("nil tracker enqueue task module triggers: %v", err)
	}
	if err := (&WorkTracker{}).enqueueTaskModuleTriggers(context.Background(), nil, uuid.New(), uuid.New(), uuid.New(), uuid.New(), "task.updated", nil); err != nil {
		t.Fatalf("nil publisher enqueue task module triggers: %v", err)
	}
}
