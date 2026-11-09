package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	"riders.agency/pulse/api/internal/platform/httptransport"
	"riders.agency/pulse/api/internal/platform/outbound"
	"riders.agency/pulse/api/internal/runtime/contracts"
)

type TemplateRenderer func(template string, input json.RawMessage, steps map[string]any, workflowRunID string) string

type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type Action struct {
	HTTPClient *http.Client
	Renderer   TemplateRenderer
	Resolver   Resolver
}

type actionConfig struct {
	Method          string      `json:"method"`
	URL             string      `json:"url"`
	HeadersJSON     string      `json:"headers_json"`
	Body            string      `json:"body"`
	TimeoutSeconds  int         `json:"timeout_seconds"`
	RetryPolicy     retryPolicy `json:"retry_policy"`
	SuccessCriteria any         `json:"success_criteria"`
}

type retryPolicy struct {
	Timeout string `json:"timeout"`
}

func (a Action) TypeKey() string {
	return "http.request"
}

func (a Action) ValidateConfig(_ context.Context, cfg json.RawMessage) error {
	var config actionConfig
	if err := json.Unmarshal(cfg, &config); err != nil {
		return fmt.Errorf("http.request config must be JSON: %w", err)
	}
	if strings.TrimSpace(config.URL) == "" {
		return errors.New("http request url is required")
	}
	return nil
}

func (a Action) Execute(ctx context.Context, req contracts.ActionRequest) contracts.ActionResult {
	var config actionConfig
	_ = json.Unmarshal(req.ConfigJSON, &config)
	if outbound.Disabled() {
		return failedResult(actionInput(req.ConfigJSON), "{}", map[string]any{
			"provider":              "http",
			"ok":                    false,
			"outbound_effects_mode": "disabled",
			"error":                 outbound.ErrDisabled.Error(),
		}, false, "outbound_disabled", outbound.ErrDisabled)
	}
	method := strings.ToUpper(strings.TrimSpace(firstNonEmpty(config.Method, http.MethodGet)))
	if method == "" {
		method = http.MethodGet
	}
	rawURL := strings.TrimSpace(a.render(config.URL, req))
	if rawURL == "" {
		err := errors.New("http request url is required")
		return failedResult(actionInput(req.ConfigJSON), "{}", map[string]any{"provider": "http", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	policy := URLPolicy{Resolver: a.Resolver}
	if _, err := policy.ValidateRaw(ctx, rawURL); err != nil {
		return failedResult(actionInput(req.ConfigJSON), "{}", map[string]any{"provider": "http", "ok": false, "error": err.Error()}, false, "validation", err)
	}

	headers := renderHTTPHeaders(config.HeadersJSON, req, a.Renderer)
	body := a.render(config.Body, req)
	var bodyReader io.Reader
	if body != "" && method != http.MethodGet && method != http.MethodHead {
		bodyReader = strings.NewReader(body)
	}
	requestInput, _ := json.Marshal(map[string]any{
		"method":           method,
		"url":              rawURL,
		"headers":          redactHeaders(headers),
		"body":             body,
		"success_criteria": config.SuccessCriteria,
	})
	timeout := durationOrDefault(config.RetryPolicy.Timeout, 30*time.Second)
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(reqCtx, method, rawURL, bodyReader)
	if err != nil {
		return failedResult(requestInput, "{}", map[string]any{"provider": "http", "ok": false, "error": err.Error()}, false, "request", err)
	}
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}
	if bodyReader != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.clientWithRedirectPolicy(policy).Do(httpReq)
	if err != nil {
		perr := fmt.Errorf("http request failed: %w", err)
		return failedResult(requestInput, "{}", map[string]any{"provider": "http", "ok": false, "error": "request failed"}, true, "network", perr)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	provider := map[string]any{
		"provider":                   "http",
		"ok":                         resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status_code":                resp.StatusCode,
		"response_headers":           redactHeaders(firstHeaderValues(resp.Header)),
		"provider_response_redacted": true,
	}
	providerJSON, _ := json.Marshal(provider)
	output := httpOutput(resp.StatusCode, raw)
	output["action_key"] = req.ActionKey
	output["action_type_key"] = req.ActionTypeKey
	outputJSON, _ := json.Marshal(output)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		perr := fmt.Errorf("http request returned HTTP %d", resp.StatusCode)
		return contracts.ActionResult{
			InputJSON:            requestInput,
			OutputJSON:           outputJSON,
			ProviderResponseJSON: providerJSON,
			Retryable:            retryable,
			ErrorType:            "provider_http",
			Err:                  perr,
		}
	}
	return contracts.ActionResult{
		InputJSON:            requestInput,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: providerJSON,
	}
}

func (a Action) render(template string, req contracts.ActionRequest) string {
	if a.Renderer == nil {
		return template
	}
	return a.Renderer(template, req.RunInputJSON, req.Steps, req.WorkflowRunID)
}

func (a Action) clientWithRedirectPolicy(policy URLPolicy) *http.Client {
	base := a.HTTPClient
	if base == nil {
		base = httptransport.NewDefaultClient(0)
	}
	client := *base
	if _, isStandardTransport := client.Transport.(*http.Transport); client.Transport == nil || isStandardTransport {
		client.Transport = safeTransport(policy)
	}
	previous := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if _, err := policy.ValidateURL(req.Context(), req.URL); err != nil {
			return err
		}
		if previous != nil {
			return previous(req, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	return &client
}

func safeTransport(policy URLPolicy) *http.Transport {
	transport := httptransport.NewTransport(nil)
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		if isLocalhostName(host) {
			return nil, errors.New("http request host is not allowed")
		}
		ips, err := policy.resolveHost(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("http request host resolution failed: %w", err)
		}
		for _, ip := range ips {
			if isBlockedIP(ip) {
				return nil, fmt.Errorf("http request host resolves to blocked address %s", ip.String())
			}
		}
		if len(ips) == 0 {
			return nil, errors.New("http request host did not resolve")
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
	}
	return transport
}

type URLPolicy struct {
	Resolver Resolver
}

func (p URLPolicy) ValidateRaw(ctx context.Context, rawURL string) (*neturl.URL, error) {
	parsed, err := neturl.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("http request url must be absolute")
	}
	return p.ValidateURL(ctx, parsed)
}

func (p URLPolicy) ValidateURL(ctx context.Context, parsed *neturl.URL) (*neturl.URL, error) {
	if parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("http request url must be absolute")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("http request url must use http or https")
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return nil, errors.New("http request url must include host")
	}
	if isLocalhostName(host) {
		return nil, errors.New("http request host is not allowed")
	}
	ips, err := p.resolveHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("http request host resolution failed: %w", err)
	}
	if len(ips) == 0 {
		return nil, errors.New("http request host did not resolve")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("http request host resolves to blocked address %s", ip.String())
		}
	}
	return parsed, nil
}

func (p URLPolicy) resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	resolver := p.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	out := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if addr.IP != nil {
			out = append(out, addr.IP)
		}
	}
	return out, nil
}

func isLocalhostName(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	for _, cidr := range blockedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

var blockedCIDRs = mustParseCIDRs([]string{
	"0.0.0.0/8",
	"100.64.0.0/10",
	"169.254.169.254/32",
	"198.18.0.0/15",
	"::/128",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
	"fd00:ec2::254/128",
})

func mustParseCIDRs(values []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		_, network, err := net.ParseCIDR(value)
		if err != nil {
			panic(err)
		}
		out = append(out, network)
	}
	return out
}

func actionInput(configJSON json.RawMessage) json.RawMessage {
	var config actionConfig
	_ = json.Unmarshal(configJSON, &config)
	out, _ := json.Marshal(map[string]any{
		"config":           jsonRawObjectOrEmpty(configJSON),
		"success_criteria": config.SuccessCriteria,
	})
	return out
}

func jsonRawObjectOrEmpty(raw json.RawMessage) any {
	var out any
	if len(strings.TrimSpace(string(raw))) == 0 || json.Unmarshal(raw, &out) != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func failedResult(inputJSON json.RawMessage, outputJSON string, provider map[string]any, retryable bool, errorType string, err error) contracts.ActionResult {
	provider["provider_response_redacted"] = true
	providerJSON, _ := json.Marshal(provider)
	report := diagnosticsForFailure(provider, retryable, errorType, err)
	output := providerdiagnostics.OutputWithReport(outputJSON, report)
	return contracts.ActionResult{
		InputJSON:            inputJSON,
		OutputJSON:           json.RawMessage(output),
		ProviderResponseJSON: providerJSON,
		Retryable:            retryable,
		ErrorType:            errorType,
		Err:                  err,
	}
}

func diagnosticsForFailure(provider map[string]any, retryable bool, errorType string, err error) providerdiagnostics.Report {
	if errorType == "validation" {
		return providerdiagnostics.ValidationFailure(provider)
	}
	base := providerdiagnostics.BaseFailure(provider, retryable, err)
	if errorType == "network" {
		base.ReasonCode = "network_error"
		base.HumanReason = "Не удалось выполнить сетевой запрос к провайдеру."
		base.NextAction = "Проверьте доступность провайдера и повторите действие."
	}
	return base
}

func renderHTTPHeaders(raw string, req contracts.ActionRequest, renderer TemplateRenderer) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	rendered := raw
	if renderer != nil {
		rendered = renderer(raw, req.RunInputJSON, req.Steps, req.WorkflowRunID)
	}
	values := map[string]any{}
	if err := json.Unmarshal([]byte(rendered), &values); err != nil {
		return map[string]string{}
	}
	out := map[string]string{}
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" || value == nil {
			continue
		}
		out[key] = strings.TrimSpace(fmt.Sprint(value))
	}
	return out
}

func firstHeaderValues(headers http.Header) map[string]string {
	out := map[string]string{}
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		out[key] = values[0]
	}
	return out
}

func redactHeaders(headers map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range headers {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "authorization") ||
			strings.Contains(lower, "api_key") ||
			strings.Contains(lower, "token") ||
			strings.Contains(lower, "secret") ||
			strings.Contains(lower, "vless") ||
			strings.Contains(lower, "password") ||
			strings.Contains(lower, "cookie") {
			out[key] = "***"
			continue
		}
		out[key] = value
	}
	return out
}

func httpOutput(statusCode int, raw []byte) map[string]any {
	body := string(raw)
	out := map[string]any{
		"ok":          statusCode >= 200 && statusCode < 300,
		"status_code": statusCode,
		"body":        body,
	}
	expose := map[string]any{
		"ok":          out["ok"],
		"status_code": statusCode,
		"body":        body,
	}
	var parsed any
	if len(raw) > 0 && json.Unmarshal(raw, &parsed) == nil {
		out["json"] = parsed
		expose["json"] = parsed
	}
	out["expose"] = expose
	return out
}

func durationOrDefault(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
