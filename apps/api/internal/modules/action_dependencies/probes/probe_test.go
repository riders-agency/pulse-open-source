package probes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func probeTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func probeJSONResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestConnectionProbeSupportedTypesHaveSuccessAndFailureFixtures(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	supported := []struct {
		typeKey     string
		providerKey string
		metadata    map[string]any
		secret      string
	}{
		{typeKey: "telegram.bot", providerKey: "telegram", secret: "telegram-secret"},
		{typeKey: "telegram.relay", providerKey: "telegram", secret: "telegram-secret"},
		{typeKey: "max.bot", providerKey: "max", secret: "max-secret"},
		{typeKey: "smtp.server", providerKey: "email", secret: "smtp-secret", metadata: map[string]any{"host": "smtp.example.test", "port": 587, "username": "user", "from_email": "robot@example.test", "encryption": "starttls"}},
		{typeKey: "dadata.party", providerKey: "dadata", secret: "dadata-secret"},
		{typeKey: "uon.account", providerKey: "uon", secret: "uon-secret"},
		{typeKey: "ai.apinet", providerKey: "ai.apinet", secret: "ai-secret"},
		{typeKey: "ai.openai", providerKey: "ai.openai", secret: "ai-secret"},
		{typeKey: "ai.openrouter", providerKey: "ai.openrouter", secret: "ai-secret"},
		{typeKey: "ai.google_gemini", providerKey: "ai.google_gemini", secret: "ai-secret"},
		{typeKey: "ai.anthropic", providerKey: "ai.anthropic", secret: "ai-secret"},
		{typeKey: "ai.mistral", providerKey: "ai.mistral", secret: "ai-secret"},
		{typeKey: "ai.groq", providerKey: "ai.groq", secret: "ai-secret"},
		{typeKey: "ai.xai", providerKey: "ai.xai", secret: "ai-secret"},
	}

	for _, tt := range supported {
		tt := tt
		t.Run(tt.typeKey+"/success", func(t *testing.T) {
			t.Parallel()
			deps := &Prober{
				HTTPClient: probeTestClient(func(req *http.Request) (*http.Response, error) {
					return probeJSONResponse(http.StatusOK, `{"ok":true,"data":[]}`), nil
				}),
				MaxHTTPClient: probeTestClient(func(req *http.Request) (*http.Response, error) {
					return probeJSONResponse(http.StatusOK, `{"id":123,"name":"bot"}`), nil
				}),
				smtpProbe: func(context.Context, smtpProbeConfig) probeResult {
					return probeResult{OK: true, Checked: true, Status: probeStatusReachable, Detail: "SMTP handshake succeeded"}
				},
			}
			result := deps.Run(ctx, dependencyProbeConfig{TypeKey: tt.typeKey, ProviderKey: tt.providerKey, Metadata: tt.metadata, Plaintext: tt.secret})
			if !result.OK || result.Status != probeStatusReachable {
				t.Fatalf("expected reachable success, got ok=%v status=%s detail=%q", result.OK, result.Status, result.Detail)
			}
			if result.ResponseMap()["ok"] != true {
				t.Fatalf("response ok should be true only for successful probe: %#v", result.ResponseMap())
			}
		})

		t.Run(tt.typeKey+"/auth_failure", func(t *testing.T) {
			t.Parallel()
			deps := &Prober{
				HTTPClient: probeTestClient(func(req *http.Request) (*http.Response, error) {
					return probeJSONResponse(http.StatusUnauthorized, `{"description":"Unauthorized"}`), nil
				}),
				MaxHTTPClient: probeTestClient(func(req *http.Request) (*http.Response, error) {
					return probeJSONResponse(http.StatusUnauthorized, `{"message":"Unauthorized"}`), nil
				}),
				smtpProbe: func(context.Context, smtpProbeConfig) probeResult {
					return probeResult{OK: false, Checked: true, Status: probeStatusAuthFailed, Detail: "SMTP authentication failed"}
				},
			}
			result := deps.Run(ctx, dependencyProbeConfig{TypeKey: tt.typeKey, ProviderKey: tt.providerKey, Metadata: tt.metadata, Plaintext: tt.secret})
			if result.OK || result.Status != probeStatusAuthFailed {
				t.Fatalf("expected auth_failed, got ok=%v status=%s detail=%q", result.OK, result.Status, result.Detail)
			}
			if result.ResponseMap()["ok"] != false {
				t.Fatalf("response ok must be false for failed probe: %#v", result.ResponseMap())
			}
		})
	}
}

func TestConnectionProbeBlocksWhenOutboundEffectsDisabled(t *testing.T) {
	t.Setenv("PULSE_OUTBOUND_EFFECTS_MODE", "disabled")
	result := (&Prober{}).Run(context.Background(), dependencyProbeConfig{
		TypeKey:     "telegram.bot",
		ProviderKey: "telegram",
		Plaintext:   "secret",
	})
	if result.OK || result.Checked || result.Status != probeStatusDisabled {
		t.Fatalf("expected disabled unchecked result, got %#v", result)
	}
}

func TestMAXProbeUsesGetMeAuthorizationAndDedicatedDomain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deps := &Prober{
		MaxHTTPClient: probeTestClient(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", req.Method)
			}
			if req.URL.Host != "platform-api2.max.ru" || req.URL.Path != "/me" {
				t.Fatalf("url = %s, want https://platform-api2.max.ru/me", req.URL.String())
			}
			if got := req.Header.Get("Authorization"); got != "max-secret" {
				t.Fatalf("Authorization = %q, want token without Bearer prefix", got)
			}
			return probeJSONResponse(http.StatusOK, `{"id":123}`), nil
		}),
	}
	result := deps.Run(ctx, dependencyProbeConfig{TypeKey: "max.bot", ProviderKey: "max", Plaintext: "Bearer max-secret"})
	if !result.OK {
		t.Fatalf("MAX probe failed: %#v", result)
	}
}

func TestProbeStatusesNeverReportOKForInvalidOrUnsupportedDependencies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deps := &Prober{}

	cases := []dependencyProbeConfig{
		{TypeKey: "telegram.bot", ProviderKey: "telegram"},
		{TypeKey: "max.bot", ProviderKey: "max"},
		{TypeKey: "smtp.server", ProviderKey: "email", Metadata: map[string]any{"port": 587}},
		{TypeKey: "email.template", ProviderKey: "email"},
	}
	for _, tc := range cases {
		result := deps.Run(ctx, tc)
		if result.OK {
			t.Fatalf("%s returned ok=true for invalid/unsupported probe: %#v", tc.TypeKey, result)
		}
		if result.Status != probeStatusMisconfigured && result.Status != probeStatusUnsupported {
			t.Fatalf("%s status = %s, want misconfigured or unsupported", tc.TypeKey, result.Status)
		}
	}
}

func TestProbeNetworkErrorsAreRedacted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secret := "uon-secret-value"
	deps := &Prober{
		HTTPClient: probeTestClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(`Get "` + req.URL.String() + `": i/o timeout`)
		}),
	}
	result := deps.Run(ctx, dependencyProbeConfig{TypeKey: "uon.account", ProviderKey: "uon", Plaintext: secret})
	if result.OK || result.Status != probeStatusUnreachable {
		t.Fatalf("expected unreachable failure, got ok=%v status=%s", result.OK, result.Status)
	}
	if strings.Contains(result.Detail, secret) {
		t.Fatalf("probe detail leaked secret: %q", result.Detail)
	}
}

func TestProbeBaseURLRejectsPlainHTTPAndCredentials(t *testing.T) {
	t.Parallel()
	bad := []string{
		"http://api.example.test",
		"https://user:pass@api.example.test",
		"://bad-url",
	}
	for _, raw := range bad {
		if _, err := probeBaseURL(map[string]any{"base_url": raw}, "base_url", "https://fallback.example.test"); err == nil {
			t.Fatalf("probeBaseURL(%q) unexpectedly succeeded", raw)
		}
	}
}
