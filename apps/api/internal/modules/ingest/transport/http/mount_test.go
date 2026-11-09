package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestParseHookURLPath(t *testing.T) {
	cases := []struct {
		in, hk, ek, sfx string
		ok              bool
	}{
		{"/hooks/abc/def", "abc", "def", "", true},
		{"hooks/abc/def", "abc", "def", "", true},
		{"/hooks/abc/def/extra/more", "abc", "def", "extra/more", true},
		{"/hooks//abc///def/", "abc", "def", "", true},
		{"/hooks", "", "", "", false},
		{"/hooks/", "", "", "", false},
		{"/hooks/a", "", "", "", false},
		{"/api/v1/x", "", "", "", false},
	}
	for _, c := range cases {
		hk, ek, sfx, ok := parseHookURLPath(c.in)
		if ok != c.ok || hk != c.hk || ek != c.ek || sfx != c.sfx {
			t.Fatalf("parseHookURLPath(%q) = (%q,%q,%q,%v) want (%q,%q,%q,%v)", c.in, hk, ek, sfx, ok, c.hk, c.ek, c.sfx, c.ok)
		}
	}
}

func TestParseReceiptSuffix(t *testing.T) {
	runID := "4cf4116e-f306-4b7c-a7f4-9c1982d48ac3"
	for _, candidate := range []string{"runs/" + runID + "/receipt", "/runs/" + runID + "/receipt/"} {
		got, ok := parseReceiptSuffix(candidate)
		if !ok || got != runID {
			t.Fatalf("parseReceiptSuffix(%q) = (%q,%v)", candidate, got, ok)
		}
	}
	for _, candidate := range []string{"", "runs/receipt", "runs/x/actions", "other/x/receipt"} {
		if got, ok := parseReceiptSuffix(candidate); ok || got != "" {
			t.Fatalf("parseReceiptSuffix(%q) unexpectedly accepted", candidate)
		}
	}
}

func TestMountHookRoutesExposesReceiptRoute(t *testing.T) {
	router := chi.NewRouter()
	MountHookRoutes(router, Webhook{})
	request := httptest.NewRequest(http.MethodGet, "/hooks/hook/endpoint/runs/not-a-uuid/receipt", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("receipt route status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestWebhookRunTerminal(t *testing.T) {
	for _, status := range []string{"succeeded", "failed", "cancelled"} {
		if !webhookRunTerminal(status) {
			t.Fatalf("status %q must be terminal", status)
		}
	}
	for _, status := range []string{"pending", "running", "waiting", ""} {
		if webhookRunTerminal(status) {
			t.Fatalf("status %q must not be terminal", status)
		}
	}
}
