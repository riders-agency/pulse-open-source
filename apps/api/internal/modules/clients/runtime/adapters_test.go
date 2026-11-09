package runtime

import (
	"context"
	"slices"
	"testing"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

func TestRegisterHandlersUsesClientAdapters(t *testing.T) {
	reg := registry.New()
	if err := RegisterHandlers(reg, nil, nil); err != nil {
		t.Fatalf("register handlers: %v", err)
	}
	gotDescriptors := reg.ListActions()
	got := make([]string, 0, len(gotDescriptors))
	for _, desc := range gotDescriptors {
		got = append(got, desc.TypeKey)
	}
	want := []string{"clients.manual_payment.trigger", "clients.payment_cycle.collect", "clients.payment_due.trigger"}
	if !slices.Equal(got, want) {
		t.Fatalf("registered client handlers = %#v, want %#v", got, want)
	}
	for _, typeKey := range want {
		handler, ok := reg.Action(typeKey)
		if !ok {
			t.Fatalf("handler %s missing", typeKey)
		}
		result := handler.Execute(context.Background(), contracts.ActionRequest{
			ActionTypeKey: typeKey,
			ConfigJSON:    []byte(`{}`),
			RunInputJSON:  []byte(`{}`),
		})
		if result.ErrorType != "runtime" || result.Err == nil {
			t.Fatalf("handler %s should fail through runtime availability guard, got error_type=%q err=%v", typeKey, result.ErrorType, result.Err)
		}
	}
}
