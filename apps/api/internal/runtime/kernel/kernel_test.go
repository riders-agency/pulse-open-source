package kernel

import (
	"context"
	"testing"
	"time"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

type fakeRunner struct {
	ran      bool
	notified bool
}

func TestScheduleRunInputHasNamespaceAndSystemData(t *testing.T) {
	at := time.Date(2026, 5, 19, 10, 30, 0, 0, time.UTC)
	input := scheduleRunInput("poll_api", "schedule_1", "version-1", at)
	pulse, _ := input["_pulse"].(map[string]any)
	if pulse["source_type"] != "schedule" || pulse["schedule_key"] != "poll_api" {
		t.Fatalf("unexpected pulse schedule input: %#v", pulse)
	}
	schedule, _ := input["poll_api"].(map[string]any)
	if schedule["scheduled_at"] != "2026-05-19T10:30:00Z" {
		t.Fatalf("unexpected schedule namespace: %#v", schedule)
	}
}

func (f *fakeRunner) Run(ctx context.Context) {
	f.ran = ctx != nil
}

func (f *fakeRunner) Notify() {
	f.notified = true
}

func TestKernelDelegatesToRunner(t *testing.T) {
	runner := &fakeRunner{}
	kernel := NewLegacy(runner)

	kernel.Notify()
	kernel.Run(context.Background())

	if !runner.notified {
		t.Fatal("Notify was not delegated")
	}
	if !runner.ran {
		t.Fatal("Run was not delegated")
	}
}

func TestNilKernelIsNoop(t *testing.T) {
	var kernel *Kernel
	kernel.Notify()
	kernel.Run(context.Background())
}

func TestWorkflowCompletionOutputMarksRequiredFailures(t *testing.T) {
	out := workflowCompletionOutput(
		[]string{"trigger"},
		[]string{"max_notify", "email_buyer"},
		[]string{"max_notify", "email_buyer"},
	)
	if out["status_detail"] != "failed_required_actions" {
		t.Fatalf("unexpected status detail: %#v", out)
	}
	required, ok := out["required_failures"].([]string)
	if !ok || len(required) != 2 {
		t.Fatalf("required failures were not preserved: %#v", out)
	}
}

func TestWorkflowCompletionOutputKeepsOptionalPartialSuccess(t *testing.T) {
	out := workflowCompletionOutput([]string{"trigger"}, []string{"optional"}, nil)
	if out["status_detail"] != "partial_success" {
		t.Fatalf("optional failure semantics changed: %#v", out)
	}
	if _, exists := out["required_failures"]; exists {
		t.Fatalf("optional failure was marked as required: %#v", out)
	}
}

type fakeDirectTrigger struct {
	synced   bool
	enqueued int
}

func (f *fakeDirectTrigger) TypeKey() string { return "demo.trigger" }

func (f *fakeDirectTrigger) Sync(context.Context) error {
	f.synced = true
	return nil
}

func (f *fakeDirectTrigger) ClaimDue(context.Context, int) ([]contracts.TriggerEvent, error) {
	return nil, nil
}

func (f *fakeDirectTrigger) MarkEnqueued(context.Context, contracts.TriggerEvent, string) error {
	return nil
}

func (f *fakeDirectTrigger) EnqueueDue(context.Context, int) (int, error) {
	f.enqueued++
	return 1, nil
}

func TestKernelEnqueueDueUsesRegisteredDirectTriggers(t *testing.T) {
	reg := registry.New()
	trigger := &fakeDirectTrigger{}
	if err := reg.RegisterTrigger(trigger); err != nil {
		t.Fatalf("register trigger failed: %v", err)
	}
	kernel := New(nil, nil, WithTriggerRegistry(reg))

	kernel.EnqueueDue(context.Background())

	if !trigger.synced {
		t.Fatal("expected trigger Sync to run")
	}
	if trigger.enqueued != 1 {
		t.Fatalf("expected direct trigger enqueue once, got %d", trigger.enqueued)
	}
}
