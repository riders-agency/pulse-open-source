package demomode

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestDemoBlocksExternalOperationsAndBoundsBody(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "true")
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	for _, path := range []string{"/api/v1/projects/p/action-dependencies", "/api/v1/workflows/w/runs", "/api/v1/workflows/w/publish", "/api/v1/public/site-chat/messages", "/api/v1/work/tasks/t/attachments"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 403 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	for _, path := range []string{"/api/v1/auth/register", "/api/v1/companies", "/api/v1/companies/c/projects", "/api/v1/companies/c/switch", "/api/v1/projects/p/ai/employee/sessions", "/api/v1/projects/p/ai/tool-calls/t/approve", "/api/v1/work/tasks"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 204 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
}
func TestBudgetSurvivesIndependentConcurrentReservations(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "true")
	t.Setenv("PULSE_DEMO_BUDGET_PATH", filepath.Join(t.TempDir(), "budget.json"))
	var wg sync.WaitGroup
	var succeeded atomic.Int32
	for range 60 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ReserveModelCall("visitor") == nil {
				succeeded.Add(1)
			}
		}()
	}
	wg.Wait()
	if succeeded.Load() != 30 {
		t.Fatalf("accepted %d, expected 30", succeeded.Load())
	}
	if ReserveModelCall("visitor") == nil {
		t.Fatal("persisted limit lost")
	}
	if ReserveModelCall("other") != nil {
		t.Fatal("other visitor blocked")
	}
}
func TestDemoBudgetFailsClosedAndIsOptIn(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "true")
	t.Setenv("PULSE_DEMO_BUDGET_PATH", "")
	if ReserveModelCall("visitor") == nil {
		t.Fatal("missing budget storage must fail closed")
	}
	t.Setenv("PULSE_PUBLIC_DEMO", "false")
	if ReserveModelCall("") != nil {
		t.Fatal("normal installation changed")
	}
}
