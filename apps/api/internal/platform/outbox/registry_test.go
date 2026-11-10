package outbox

import (
	"context"
	"testing"
)

type fakeHandler struct {
	typeKey string
}

func (h fakeHandler) TypeKey() string {
	return h.typeKey
}

func (h fakeHandler) Handle(context.Context, Event) error {
	return nil
}

func TestRegistryRejectsEmptyTypeKey(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(fakeHandler{typeKey: " "}); err == nil {
		t.Fatal("expected empty type key error")
	}
}

func TestRegistryRejectsDuplicateTypeKey(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(fakeHandler{typeKey: "demo.event"}); err != nil {
		t.Fatalf("register handler failed: %v", err)
	}
	if err := reg.Register(fakeHandler{typeKey: "demo.event"}); err == nil {
		t.Fatal("expected duplicate handler error")
	}
}

func TestRegistryLookupAndList(t *testing.T) {
	reg := NewRegistry()
	for _, key := range []string{"z.event", "a.event"} {
		if err := reg.Register(fakeHandler{typeKey: key}); err != nil {
			t.Fatalf("register %s failed: %v", key, err)
		}
	}
	if got, ok := reg.Handler(" a.event "); !ok || got.TypeKey() != "a.event" {
		t.Fatal("expected trimmed lookup hit")
	}
	list := reg.List()
	if len(list) != 2 || list[0].TypeKey != "a.event" || list[1].TypeKey != "z.event" {
		t.Fatalf("unexpected sorted list: %#v", list)
	}
}
