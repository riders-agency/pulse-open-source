package registry

import (
	"context"
	"encoding/json"
	"testing"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

type fakeActionHandler struct {
	typeKey string
}

type fakeTriggerHandler struct {
	typeKey string
}

func (h fakeActionHandler) TypeKey() string {
	return h.typeKey
}

func (h fakeActionHandler) ValidateConfig(context.Context, json.RawMessage) error {
	return nil
}

func (h fakeActionHandler) Execute(context.Context, contracts.ActionRequest) contracts.ActionResult {
	return contracts.ActionResult{}
}

func (h fakeTriggerHandler) TypeKey() string {
	return h.typeKey
}

func (h fakeTriggerHandler) Sync(context.Context) error {
	return nil
}

func (h fakeTriggerHandler) ClaimDue(context.Context, int) ([]contracts.TriggerEvent, error) {
	return nil, nil
}

func (h fakeTriggerHandler) MarkEnqueued(context.Context, contracts.TriggerEvent, string) error {
	return nil
}

func TestRegisterActionRejectsEmptyTypeKey(t *testing.T) {
	reg := New()
	if err := reg.RegisterAction(fakeActionHandler{typeKey: "  "}); err == nil {
		t.Fatal("expected empty type key error")
	}
}

func TestRegisterActionRejectsDuplicateTypeKey(t *testing.T) {
	reg := New()
	if err := reg.RegisterAction(fakeActionHandler{typeKey: "demo.action"}); err != nil {
		t.Fatalf("register action failed: %v", err)
	}
	if err := reg.RegisterAction(fakeActionHandler{typeKey: "demo.action"}); err == nil {
		t.Fatal("expected duplicate action error")
	}
}

func TestActionLookupTrimsTypeKey(t *testing.T) {
	reg := New()
	want := fakeActionHandler{typeKey: "demo.action"}
	if err := reg.RegisterAction(want); err != nil {
		t.Fatalf("register action failed: %v", err)
	}
	got, ok := reg.Action(" demo.action ")
	if !ok {
		t.Fatal("expected action lookup hit")
	}
	if got.TypeKey() != want.TypeKey() {
		t.Fatalf("lookup returned %q, want %q", got.TypeKey(), want.TypeKey())
	}
}

func TestListActionsIsSorted(t *testing.T) {
	reg := New()
	for _, key := range []string{"z.action", "a.action", "m.action"} {
		if err := reg.RegisterAction(fakeActionHandler{typeKey: key}); err != nil {
			t.Fatalf("register action %q failed: %v", key, err)
		}
	}
	got := reg.ListActions()
	want := []string{"a.action", "m.action", "z.action"}
	if len(got) != len(want) {
		t.Fatalf("unexpected action count: got %d want %d", len(got), len(want))
	}
	for i, wantKey := range want {
		if got[i].TypeKey != wantKey {
			t.Fatalf("action[%d] = %q, want %q", i, got[i].TypeKey, wantKey)
		}
	}
}

func TestRegisterTriggerRejectsEmptyTypeKey(t *testing.T) {
	reg := New()
	if err := reg.RegisterTrigger(fakeTriggerHandler{typeKey: "  "}); err == nil {
		t.Fatal("expected empty trigger type key error")
	}
}

func TestRegisterTriggerRejectsDuplicateTypeKey(t *testing.T) {
	reg := New()
	if err := reg.RegisterTrigger(fakeTriggerHandler{typeKey: "demo.trigger"}); err != nil {
		t.Fatalf("register trigger failed: %v", err)
	}
	if err := reg.RegisterTrigger(fakeTriggerHandler{typeKey: "demo.trigger"}); err == nil {
		t.Fatal("expected duplicate trigger error")
	}
}

func TestTriggerLookupTrimsTypeKey(t *testing.T) {
	reg := New()
	want := fakeTriggerHandler{typeKey: "demo.trigger"}
	if err := reg.RegisterTrigger(want); err != nil {
		t.Fatalf("register trigger failed: %v", err)
	}
	got, ok := reg.Trigger(" demo.trigger ")
	if !ok {
		t.Fatal("expected trigger lookup hit")
	}
	if got.TypeKey() != want.TypeKey() {
		t.Fatalf("lookup returned %q, want %q", got.TypeKey(), want.TypeKey())
	}
}

func TestListTriggersIsSorted(t *testing.T) {
	reg := New()
	for _, key := range []string{"z.trigger", "a.trigger", "m.trigger"} {
		if err := reg.RegisterTrigger(fakeTriggerHandler{typeKey: key}); err != nil {
			t.Fatalf("register trigger %q failed: %v", key, err)
		}
	}
	got := reg.ListTriggers()
	want := []string{"a.trigger", "m.trigger", "z.trigger"}
	if len(got) != len(want) {
		t.Fatalf("unexpected trigger count: got %d want %d", len(got), len(want))
	}
	for i, wantKey := range want {
		if got[i].TypeKey != wantKey {
			t.Fatalf("trigger[%d] = %q, want %q", i, got[i].TypeKey, wantKey)
		}
	}
}
