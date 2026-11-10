package probes

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/modules/action_dependencies/contracts"

	emailprovider "riders.agency/pulse/api/internal/modules/providers/email"
	maxprovider "riders.agency/pulse/api/internal/modules/providers/max"
	"riders.agency/pulse/api/internal/platform/httptransport"
	"riders.agency/pulse/api/internal/platform/outbound"
)

const dependencyProbeTimeout = 5 * time.Second

type probeStatus string

const (
	probeStatusReachable     probeStatus = "reachable"
	probeStatusAuthFailed    probeStatus = "auth_failed"
	probeStatusMisconfigured probeStatus = "misconfigured"
	probeStatusRateLimited   probeStatus = "rate_limited"
	probeStatusDisabled      probeStatus = "disabled"
	probeStatusUnsupported   probeStatus = "unsupported"
	probeStatusUnreachable   probeStatus = "unreachable"
)

type dependencyProbeConfig = contracts.ProbeConfig

type probeResult struct {
	OK         bool
	Checked    bool
	Status     probeStatus
	Provider   string
	TypeKey    string
	Detail     string
	Endpoint   string
	StatusCode int
	Retryable  bool
	LatencyMS  int64
	Stages     []string
	Extra      map[string]any
}

type smtpProbeConfig struct {
	Host       string
	Port       int
	Encryption string
	Username   string
	Password   string
	Timeout    time.Duration
}

func (d *Prober) Run(ctx context.Context, dep dependencyProbeConfig) probeResult {
	dep.TypeKey = strings.TrimSpace(dep.TypeKey)
	if outbound.Disabled() {
		return probeResult{
			OK:       false,
			Checked:  false,
			Status:   probeStatusDisabled,
			Provider: dep.ProviderKey,
			TypeKey:  dep.TypeKey,
			Detail:   outbound.ErrDisabled.Error(),
			Extra: map[string]any{
				"outbound_effects_mode": "disabled",
			},
		}
	}
	switch dep.TypeKey {
	case "telegram.bot", "telegram.relay":
		return d.probeTelegram(ctx, dep)
	case "max.bot":
		return d.probeMAX(ctx, dep)
	case "smtp.server":
		return d.probeSMTP(ctx, dep)
	case "dadata.party":
		return d.probeDaData(ctx, dep)
	case "uon.account":
		return d.probeUON(ctx, dep)
	case "ai.apinet", "ai.openai", "ai.openrouter", "ai.google_gemini", "ai.anthropic", "ai.mistral", "ai.groq", "ai.xai":
		return d.probeAIModelProvider(ctx, dep)
	default:
		return probeResult{
			OK:       false,
			Checked:  true,
			Status:   probeStatusUnsupported,
			Provider: dep.ProviderKey,
			TypeKey:  dep.TypeKey,
			Detail:   "connection probe is not implemented for this dependency type",
		}
	}
}

func (r probeResult) ResponseMap() map[string]any {
	status := r.Status
	if status == "" {
		status = probeStatusUnsupported
	}
	out := map[string]any{
		"ok":                         r.OK,
		"checked":                    r.Checked,
		"status":                     string(status),
		"provider":                   r.Provider,
		"dependency_type_key":        r.TypeKey,
		"detail":                     r.Detail,
		"provider_response_redacted": true,
	}
	if r.Endpoint != "" {
		out["endpoint"] = r.Endpoint
	}
	if r.StatusCode > 0 {
		out["status_code"] = r.StatusCode
	}
	if r.Retryable {
		out["retryable"] = true
	}
	if r.LatencyMS > 0 {
		out["latency_ms"] = r.LatencyMS
	}
	if len(r.Stages) > 0 {
		out["stages"] = append([]string{}, r.Stages...)
	}
	for key, value := range r.Extra {
		if _, exists := out[key]; !exists {
			out[key] = value
		}
	}
	return out
}

func (d *Prober) probeTelegram(ctx context.Context, dep dependencyProbeConfig) probeResult {
	token := strings.TrimSpace(dep.Plaintext)
	if token == "" {
		return misconfiguredProbe("telegram", dep.TypeKey, "telegram bot token is required")
	}
	reqCtx, cancel := context.WithTimeout(ctx, dependencyProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://api.telegram.org/bot"+token+"/getMe", nil)
	if err != nil {
		return misconfiguredProbe("telegram", dep.TypeKey, "telegram probe request could not be built")
	}
	req.Header.Set("Accept", "application/json")
	start := time.Now()
	resp, err := d.defaultHTTPClient().Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return networkProbe("telegram", dep.TypeKey, "telegram request failed", latency)
	}
	defer resp.Body.Close()
	provider := readProbeResponse(resp)
	result := httpProbeResult("telegram", dep.TypeKey, "GET /bot<token>/getMe", provider, latency)
	if result.OK {
		if ok, _ := provider.JSON["ok"].(bool); !ok {
			result.OK = false
			result.Status = probeStatusMisconfigured
			result.Detail = "telegram response did not contain ok=true"
		}
	}
	return result
}

func (d *Prober) probeMAX(ctx context.Context, dep dependencyProbeConfig) probeResult {
	token := strings.TrimSpace(dep.Plaintext)
	if token == "" {
		return misconfiguredProbe("max", dep.TypeKey, "MAX bot token is required")
	}
	base, err := probeBaseURL(dep.Metadata, "api_base", "https://platform-api2.max.ru")
	if err != nil {
		return misconfiguredProbe("max", dep.TypeKey, "MAX api_base must be a valid https URL")
	}
	reqCtx, cancel := context.WithTimeout(ctx, dependencyProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, strings.TrimRight(base, "/")+"/me", nil)
	if err != nil {
		return misconfiguredProbe("max", dep.TypeKey, "MAX probe request could not be built")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", probeMaxAuthorizationHeader(token))
	client, err := d.maxProbeHTTPClient()
	if err != nil {
		return probeResult{
			OK:       false,
			Checked:  true,
			Status:   probeStatusMisconfigured,
			Provider: "max",
			TypeKey:  dep.TypeKey,
			Detail:   "MAX TLS/CA transport configuration is invalid",
			Endpoint: "GET /me",
		}
	}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return networkProbe("max", dep.TypeKey, "MAX request failed", latency)
	}
	defer resp.Body.Close()
	provider := readProbeResponse(resp)
	out := httpProbeResult("max", dep.TypeKey, "GET /me", provider, latency)
	out.Extra = mergeProbeExtra(out.Extra, map[string]any{"tls_trust": "system+mincifry_ca"})
	return out
}

func (d *Prober) probeSMTP(ctx context.Context, dep dependencyProbeConfig) probeResult {
	cfg, err := smtpProbeConfigFromMetadata(dep.Metadata, dep.Plaintext)
	if err != nil {
		return misconfiguredProbe("smtp", dep.TypeKey, err.Error())
	}
	prober := d.smtpProbe
	if prober == nil {
		prober = defaultSMTPProbe
	}
	out := prober(ctx, cfg)
	out.Provider = "smtp"
	out.TypeKey = dep.TypeKey
	out.Endpoint = "SMTP handshake"
	return out
}

func (d *Prober) probeDaData(ctx context.Context, dep dependencyProbeConfig) probeResult {
	token := strings.TrimSpace(dep.Plaintext)
	if token == "" {
		return misconfiguredProbe("dadata", dep.TypeKey, "DaData api_key is required")
	}
	base, err := probeBaseURL(dep.Metadata, "base_url", "https://suggestions.dadata.ru/suggestions/api/4_1/rs")
	if err != nil {
		return misconfiguredProbe("dadata", dep.TypeKey, "DaData base_url must be a valid https URL")
	}
	body, _ := json.Marshal(map[string]string{"query": "7707083893"})
	reqCtx, cancel := context.WithTimeout(ctx, dependencyProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, strings.TrimRight(base, "/")+"/findById/party", bytes.NewReader(body))
	if err != nil {
		return misconfiguredProbe("dadata", dep.TypeKey, "DaData probe request could not be built")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+token)
	start := time.Now()
	resp, err := d.defaultHTTPClient().Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return networkProbe("dadata", dep.TypeKey, "DaData request failed", latency)
	}
	defer resp.Body.Close()
	provider := readProbeResponse(resp)
	return httpProbeResult("dadata", dep.TypeKey, "POST /findById/party", provider, latency)
}

func (d *Prober) probeUON(ctx context.Context, dep dependencyProbeConfig) probeResult {
	apiKey := strings.TrimSpace(dep.Plaintext)
	if apiKey == "" {
		return misconfiguredProbe("uon", dep.TypeKey, "U-ON api_key is required")
	}
	base, err := probeBaseURL(dep.Metadata, "api_base", "https://api.u-on.ru")
	if err != nil {
		return misconfiguredProbe("uon", dep.TypeKey, "U-ON api_base must be a valid https URL")
	}
	reqCtx, cancel := context.WithTimeout(ctx, dependencyProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, strings.TrimRight(base, "/")+"/"+url.PathEscape(apiKey)+"/user/list.json", nil)
	if err != nil {
		return misconfiguredProbe("uon", dep.TypeKey, "U-ON probe request could not be built")
	}
	req.Header.Set("Accept", "application/json")
	start := time.Now()
	resp, err := d.defaultHTTPClient().Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return networkProbe("uon", dep.TypeKey, "U-ON request failed", latency)
	}
	defer resp.Body.Close()
	provider := readProbeResponse(resp)
	return httpProbeResult("uon", dep.TypeKey, "GET /<api_key>/user/list.json", provider, latency)
}

func (d *Prober) probeAIModelProvider(ctx context.Context, dep dependencyProbeConfig) probeResult {
	apiKey := strings.TrimSpace(dep.Plaintext)
	if apiKey == "" {
		return misconfiguredProbe(dep.ProviderKey, dep.TypeKey, "AI provider api_key is required")
	}
	spec, ok := aiProbeSpecs[dep.TypeKey]
	if !ok {
		return probeResult{OK: false, Checked: true, Status: probeStatusUnsupported, Provider: dep.ProviderKey, TypeKey: dep.TypeKey, Detail: "AI provider probe is not implemented"}
	}
	base, err := probeBaseURL(dep.Metadata, "base_url", spec.BaseURL)
	if err != nil {
		return misconfiguredProbe(dep.ProviderKey, dep.TypeKey, "AI provider base_url must be a valid https URL")
	}
	reqCtx, cancel := context.WithTimeout(ctx, dependencyProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, strings.TrimRight(base, "/")+spec.ModelsPath, nil)
	if err != nil {
		return misconfiguredProbe(dep.ProviderKey, dep.TypeKey, "AI provider probe request could not be built")
	}
	req.Header.Set("Accept", "application/json")
	spec.Auth(req, apiKey)
	start := time.Now()
	resp, err := d.defaultHTTPClient().Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return networkProbe(dep.ProviderKey, dep.TypeKey, "AI provider request failed", latency)
	}
	defer resp.Body.Close()
	provider := readProbeResponse(resp)
	return httpProbeResult(dep.ProviderKey, dep.TypeKey, "GET "+spec.ModelsPath, provider, latency)
}

type aiProbeSpec struct {
	BaseURL    string
	ModelsPath string
	Auth       func(*http.Request, string)
}

var aiProbeSpecs = map[string]aiProbeSpec{
	"ai.apinet":        {BaseURL: "https://apinet.cloud/v1", ModelsPath: "/models", Auth: bearerProbeAuth},
	"ai.openai":        {BaseURL: "https://api.openai.com/v1", ModelsPath: "/models", Auth: bearerProbeAuth},
	"ai.openrouter":    {BaseURL: "https://openrouter.ai/api/v1", ModelsPath: "/models", Auth: bearerProbeAuth},
	"ai.mistral":       {BaseURL: "https://api.mistral.ai/v1", ModelsPath: "/models", Auth: bearerProbeAuth},
	"ai.groq":          {BaseURL: "https://api.groq.com/openai/v1", ModelsPath: "/models", Auth: bearerProbeAuth},
	"ai.xai":           {BaseURL: "https://api.x.ai/v1", ModelsPath: "/models", Auth: bearerProbeAuth},
	"ai.anthropic":     {BaseURL: "https://api.anthropic.com/v1", ModelsPath: "/models", Auth: anthropicProbeAuth},
	"ai.google_gemini": {BaseURL: "https://generativelanguage.googleapis.com/v1beta", ModelsPath: "/models", Auth: googleProbeAuth},
}

func bearerProbeAuth(req *http.Request, apiKey string) {
	req.Header.Set("Authorization", "Bearer "+apiKey)
}

func anthropicProbeAuth(req *http.Request, apiKey string) {
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

func googleProbeAuth(req *http.Request, apiKey string) {
	req.Header.Set("x-goog-api-key", apiKey)
}

type probeHTTPResponse struct {
	StatusCode int
	JSON       map[string]any
	Detail     string
}

func readProbeResponse(resp *http.Response) probeHTTPResponse {
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	out := probeHTTPResponse{StatusCode: resp.StatusCode}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err == nil {
		out.JSON = root
	}
	out.Detail = probeResponseDetail(root, raw)
	return out
}

func httpProbeResult(provider, typeKey, endpoint string, resp probeHTTPResponse, latency int64) probeResult {
	status := statusFromHTTP(resp.StatusCode)
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	detail := "provider returned HTTP " + fmt.Sprint(resp.StatusCode)
	if ok {
		status = probeStatusReachable
		detail = "provider probe succeeded"
	} else if resp.Detail != "" {
		detail = resp.Detail
	}
	return probeResult{
		OK:         ok,
		Checked:    true,
		Status:     status,
		Provider:   provider,
		TypeKey:    typeKey,
		Detail:     detail,
		Endpoint:   endpoint,
		StatusCode: resp.StatusCode,
		Retryable:  status == probeStatusRateLimited || status == probeStatusUnreachable,
		LatencyMS:  latency,
	}
}

func statusFromHTTP(code int) probeStatus {
	switch {
	case code >= 200 && code < 300:
		return probeStatusReachable
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		return probeStatusAuthFailed
	case code == http.StatusTooManyRequests:
		return probeStatusRateLimited
	case code == 0 || code >= 500:
		return probeStatusUnreachable
	default:
		return probeStatusMisconfigured
	}
}

func probeResponseDetail(root map[string]any, raw []byte) string {
	if root != nil {
		for _, key := range []string{"description", "error_description", "error", "message", "reason", "code"} {
			if v, exists := root[key]; exists {
				return truncateProbeDetail(fmt.Sprint(v), 240)
			}
		}
	}
	return truncateProbeDetail(string(raw), 240)
}

func smtpProbeConfigFromMetadata(metadata map[string]any, password string) (smtpProbeConfig, error) {
	host := strings.TrimSpace(probeString(metadata["host"]))
	if host == "" {
		return smtpProbeConfig{}, fmt.Errorf("SMTP host is required")
	}
	port := probeInt(metadata["port"], 587)
	if port <= 0 || port > 65535 {
		return smtpProbeConfig{}, fmt.Errorf("SMTP port must be between 1 and 65535")
	}
	username := strings.TrimSpace(probeString(metadata["username"]))
	password = strings.TrimSpace(password)
	if username != "" && password == "" {
		return smtpProbeConfig{}, fmt.Errorf("SMTP password is required when username is configured")
	}
	mode := strings.ToLower(strings.TrimSpace(probeString(metadata["encryption"])))
	if mode == "" {
		if port == 465 {
			mode = "tls"
		} else {
			mode = "starttls"
		}
	}
	if mode != "tls" && mode != "starttls" && mode != "none" {
		return smtpProbeConfig{}, fmt.Errorf("SMTP encryption must be none, starttls, or tls")
	}
	return smtpProbeConfig{
		Host:       host,
		Port:       port,
		Encryption: mode,
		Username:   username,
		Password:   password,
		Timeout:    dependencyProbeTimeout,
	}, nil
}

func defaultSMTPProbe(ctx context.Context, cfg smtpProbeConfig) probeResult {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = dependencyProbeTimeout
	}
	stages := []string{}
	fail := func(status probeStatus, stage, detail string, retryable bool) probeResult {
		if stage != "" {
			stages = append(stages, stage)
		}
		return probeResult{
			OK:        false,
			Checked:   true,
			Status:    status,
			Detail:    detail,
			Retryable: retryable,
			Stages:    append([]string{}, stages...),
		}
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port))
	dialer := &net.Dialer{Timeout: timeout}
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var conn net.Conn
	var err error
	start := time.Now()
	if cfg.Encryption == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.DialContext(dialCtx, "tcp", addr)
	}
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return fail(probeStatusUnreachable, "tcp_dial", "SMTP TCP/TLS connection failed", true)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	stages = append(stages, "tcp_connected")

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fail(probeStatusUnreachable, "smtp_client", "SMTP banner was not accepted", true)
	}
	defer client.Close()
	stages = append(stages, "smtp_banner")

	if cfg.Encryption == "starttls" {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return fail(probeStatusMisconfigured, "starttls", "SMTP server does not advertise STARTTLS", false)
		}
		if err := client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fail(probeStatusMisconfigured, "starttls", "SMTP STARTTLS failed", false)
		}
		stages = append(stages, "starttls")
	}
	if cfg.Username != "" {
		ok, _ := client.Extension("AUTH")
		if !ok {
			return fail(probeStatusMisconfigured, "auth", "SMTP server does not advertise AUTH", false)
		}
		if err := client.Auth(emailprovider.NewSMTPAuth(cfg.Username, cfg.Password, cfg.Host)); err != nil {
			code, _ := emailprovider.ExtractSMTPStatus(err.Error())
			status := probeStatusAuthFailed
			retryable := false
			if strings.HasPrefix(code, "4") {
				status = probeStatusUnreachable
				retryable = true
			}
			return fail(status, "auth", "SMTP authentication failed", retryable)
		}
		stages = append(stages, "auth")
	}
	if err := client.Noop(); err != nil {
		return fail(probeStatusUnreachable, "noop", "SMTP NOOP failed", true)
	}
	stages = append(stages, "noop")
	_ = client.Quit()
	stages = append(stages, "quit")
	return probeResult{
		OK:        true,
		Checked:   true,
		Status:    probeStatusReachable,
		Detail:    "SMTP handshake succeeded",
		LatencyMS: latency,
		Stages:    append([]string{}, stages...),
	}
}

func (d *Prober) defaultHTTPClient() *http.Client {
	if d != nil && d.HTTPClient != nil {
		return d.HTTPClient
	}
	return httptransport.NewDefaultClient(dependencyProbeTimeout)
}

func (d *Prober) maxProbeHTTPClient() (*http.Client, error) {
	if d != nil && d.MaxHTTPClient != nil {
		return d.MaxHTTPClient, nil
	}
	return maxprovider.NewHTTPClientFromEnv(dependencyProbeTimeout)
}

func probeBaseURL(metadata map[string]any, key, fallback string) (string, error) {
	base := strings.TrimSpace(probeString(metadata[key]))
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return "", fmt.Errorf("base URL must be https and must not include credentials")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func misconfiguredProbe(provider, typeKey, detail string) probeResult {
	return probeResult{OK: false, Checked: true, Status: probeStatusMisconfigured, Provider: provider, TypeKey: typeKey, Detail: detail}
}

func networkProbe(provider, typeKey, detail string, latency int64) probeResult {
	return probeResult{OK: false, Checked: true, Status: probeStatusUnreachable, Provider: provider, TypeKey: typeKey, Detail: detail, Retryable: true, LatencyMS: latency}
}

func probeString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func probeInt(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		i, err := typed.Int64()
		if err == nil {
			return int(i)
		}
	case string:
		var out int
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &out); err == nil {
			return out
		}
	}
	return fallback
}

func probeMaxAuthorizationHeader(secret string) string {
	secret = strings.TrimSpace(secret)
	if strings.HasPrefix(strings.ToLower(secret), "bearer ") {
		return strings.TrimSpace(secret[7:])
	}
	return secret
}

func truncateProbeDetail(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "..."
}

func mergeProbeExtra(left, right map[string]any) map[string]any {
	if len(left) == 0 {
		left = map[string]any{}
	}
	for key, value := range right {
		left[key] = value
	}
	return left
}

type Prober struct {
	HTTPClient    *http.Client
	MaxHTTPClient *http.Client
	smtpProbe     func(context.Context, smtpProbeConfig) probeResult
}

func New() *Prober { return &Prober{} }
func (d *Prober) Check(ctx context.Context, dep contracts.ProbeConfig) map[string]any {
	return d.Run(ctx, dep).ResponseMap()
}
