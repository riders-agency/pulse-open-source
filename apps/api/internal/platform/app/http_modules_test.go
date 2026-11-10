package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"riders.agency/pulse/api/internal/platform/httpserver"
	"riders.agency/pulse/api/internal/runtime/registry"
)

func TestBuildModulesRegisterHTTPWithoutDuplicateMounts(t *testing.T) {
	var handler http.Handler
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("router construction panicked: %v", recovered)
			}
		}()
		handler = httpserver.NewRouter(httpserver.Dependencies{}, BuildModules())
	}()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestBuildModulesRegisterModuleOwnedRuntimeHandlers(t *testing.T) {
	reg := registry.New()

	BuildModules().RegisterRuntime(reg, Deps{HTTPClient: http.DefaultClient})

	for _, typeKey := range []string{
		"ai.apinet.normalize_text",
		"telegram.send_message",
		"telegram.edit_message_text",
		"telegram.answer_callback_query",
		"max.send_message",
		"email.send",
		"uon.create_lead",
		"clients.manual_payment.trigger",
		"clients.payment_due.trigger",
		"clients.payment_cycle.collect",
	} {
		if _, ok := reg.Action(typeKey); !ok {
			t.Fatalf("expected module-owned runtime handler registered for %s", typeKey)
		}
	}
	if _, ok := reg.Trigger("clients.payment_due.trigger"); !ok {
		t.Fatal("expected module-owned client payment trigger registered")
	}
}
