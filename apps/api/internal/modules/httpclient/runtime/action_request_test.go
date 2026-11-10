package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"riders.agency/pulse/api/internal/platform/httptransport"
	"riders.agency/pulse/api/internal/runtime/contracts"
)

type fakeResolver map[string][]net.IP

func (r fakeResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	ips := r[host]
	out := make([]net.IPAddr, 0, len(ips))
	for _, ip := range ips {
		out = append(out, net.IPAddr{IP: ip})
	}
	return out, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestURLPolicyBlocksPrivateAndMetadataAddresses(t *testing.T) {
	policy := URLPolicy{Resolver: fakeResolver{
		"private.example":  {net.ParseIP("10.0.0.5")},
		"metadata.example": {net.ParseIP("169.254.169.254")},
	}}
	for _, rawURL := range []string{
		"http://127.0.0.1/status",
		"http://localhost/status",
		"http://private.example/status",
		"http://metadata.example/status",
	} {
		if _, err := policy.ValidateRaw(context.Background(), rawURL); err == nil {
			t.Fatalf("expected %s to be blocked", rawURL)
		}
	}
}

func TestURLPolicyAllowsResolvedPublicAddress(t *testing.T) {
	policy := URLPolicy{Resolver: fakeResolver{"api.example": {net.ParseIP("93.184.216.34")}}}
	if _, err := policy.ValidateRaw(context.Background(), "https://api.example/orders"); err != nil {
		t.Fatalf("expected public URL to pass: %v", err)
	}
}

func TestHTTPActionRendersRequestAndRedactsHeaders(t *testing.T) {
	handler := Action{
		Resolver: fakeResolver{"api.example": {net.ParseIP("93.184.216.34")}},
		Renderer: func(template string, _ json.RawMessage, _ map[string]any, _ string) string {
			return strings.ReplaceAll(template, "{{ input.order_id }}", "1001")
		},
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.example/orders/1001" {
				t.Fatalf("unexpected URL: %s", req.URL.String())
			}
			if req.Header.Get("Authorization") != "Bearer secret" {
				t.Fatalf("authorization header was not sent")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
				Request:    req,
			}, nil
		})},
	}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		WorkflowRunID: "run-1",
		ActionKey:     "http_1",
		ActionTypeKey: "http.request",
		ConfigJSON: json.RawMessage(`{
			"method":"GET",
			"url":"https://api.example/orders/{{ input.order_id }}",
			"headers_json":"{\"Authorization\":\"Bearer secret\"}"
		}`),
		RunInputJSON: json.RawMessage(`{"order_id":"1001"}`),
	})
	if result.Err != nil {
		t.Fatalf("execute failed: %v", result.Err)
	}
	var input map[string]any
	if err := json.Unmarshal(result.InputJSON, &input); err != nil {
		t.Fatalf("input json invalid: %v", err)
	}
	headers := input["headers"].(map[string]any)
	if headers["Authorization"] != "***" {
		t.Fatalf("expected redacted authorization in input: %#v", headers)
	}
	var output map[string]any
	if err := json.Unmarshal(result.OutputJSON, &output); err != nil {
		t.Fatalf("output json invalid: %v", err)
	}
	if output["status_code"] != float64(200) || output["action_key"] != "http_1" {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestHTTPActionBlocksWhenOutboundEffectsDisabled(t *testing.T) {
	t.Setenv("PULSE_OUTBOUND_EFFECTS_MODE", "disabled")
	handler := Action{}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		ActionKey:     "http_1",
		ActionTypeKey: "http.request",
		ConfigJSON:    json.RawMessage(`{"url":"https://api.example/orders"}`),
	})
	if result.ErrorType != "outbound_disabled" || result.Err == nil {
		t.Fatalf("expected outbound_disabled, got error_type=%q err=%v", result.ErrorType, result.Err)
	}
}

func TestRedactHeadersCoversSensitiveNames(t *testing.T) {
	headers := map[string]string{
		"api_key":               "secret-api-key",
		"Authorization":         "Bearer secret",
		"Cookie":                "rp_access=secret",
		"password":              "secret-password",
		"X-Secret":              "secret",
		"X-Token":               "secret-token",
		"X-Pulse-Service-Token": "service-secret",
		"vless_url":             "vless://secret",
		"X-Trace-ID":            "trace-1",
	}
	redacted := redactHeaders(headers)
	for _, key := range []string{"api_key", "Authorization", "Cookie", "password", "X-Secret", "X-Token", "X-Pulse-Service-Token", "vless_url"} {
		if redacted[key] != "***" {
			t.Fatalf("%s not redacted: %#v", key, redacted)
		}
	}
	if redacted["X-Trace-ID"] != "trace-1" {
		t.Fatalf("non-secret header should pass through: %#v", redacted)
	}
}

func TestHTTPActionBlocksUnsafeRedirect(t *testing.T) {
	handler := Action{
		Resolver: fakeResolver{"api.example": {net.ParseIP("93.184.216.34")}},
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{"http://127.0.0.1/admin"}},
				Body:       io.NopCloser(bytes.NewBuffer(nil)),
				Request:    req,
			}, nil
		})},
	}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		ActionKey:     "http_1",
		ActionTypeKey: "http.request",
		ConfigJSON:    json.RawMessage(`{"url":"https://api.example/start"}`),
	})
	if result.Err == nil {
		t.Fatal("expected unsafe redirect error")
	}
	if result.ErrorType != "network" {
		t.Fatalf("unexpected error type: %s", result.ErrorType)
	}
}

func TestHTTPActionReplacesStandardClientTransportWithSafeTransport(t *testing.T) {
	base := httptransport.NewDefaultClient(0)
	client := Action{HTTPClient: base}.clientWithRedirectPolicy(URLPolicy{})
	if client == base {
		t.Fatal("expected a per-action client copy")
	}
	if client.Transport == base.Transport {
		t.Fatal("expected standard transport to be replaced with safe transport")
	}
}
