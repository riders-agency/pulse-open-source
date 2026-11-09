package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/httptransport"
)

func (a *ModelClient) FetchModels(ctx context.Context, dep ModelDependency) ([]ModelDescriptor, error) {
	base, err := SafeProviderBaseURL(dep.Metadata, "https://apinet.cloud/v1")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+dep.APIKey)
	req.Header.Set("Accept", "application/json")
	resp, err := a.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("apinet models status %d", resp.StatusCode)
	}
	models := ParseProviderModels(body, dep.ProviderKey)
	if len(models) == 0 {
		return nil, fmt.Errorf("apinet models response has no models")
	}
	return models, nil
}

func (a *ModelClient) Chat(ctx context.Context, dep ModelDependency, modelID string, messages []ChatMessage, maxTokens int) (string, ModelUsage, error) {
	if modelID == "" {
		modelID = ModelFromMetadata(dep.Metadata)
	}
	if modelID == "" {
		return "", ModelUsage{}, fmt.Errorf("model_id required")
	}
	base, err := SafeProviderBaseURL(dep.Metadata, "https://apinet.cloud/v1")
	if err != nil {
		return "", ModelUsage{}, err
	}
	payload := apinetChatRequest{
		Model:       modelID,
		Messages:    messages,
		Temperature: 0.1,
		MaxTokens:   maxTokens,
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", ModelUsage{}, err
	}
	req.Header.Set("Authorization", "Bearer "+dep.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	start := time.Now()
	resp, err := a.client().Do(req)
	latency := int(time.Since(start).Milliseconds())
	usage := ModelUsage{
		ProviderKey: dep.ProviderKey,
		ModelID:     modelID,
		LatencyMS:   latency,
		ProviderResponse: map[string]any{
			"provider":                   "ai",
			"provider_key":               dep.ProviderKey,
			"base_url_host":              ProviderBaseHost(base),
			"model":                      modelID,
			"ok":                         false,
			"request_redacted":           true,
			"provider_response_redacted": true,
		},
	}
	if err != nil {

		return "", usage, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		usage.ProviderResponse["status_code"] = resp.StatusCode

		return "", usage, fmt.Errorf("apinet chat status %d", resp.StatusCode)
	}
	var parsed apinetChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		usage.ProviderResponse["status_code"] = resp.StatusCode

		return "", usage, err
	}
	usage.InputTokens = parsed.Usage.PromptTokens
	usage.OutputTokens = parsed.Usage.CompletionTokens
	usage.ProviderResponse["ok"] = true
	usage.ProviderResponse["status_code"] = resp.StatusCode
	usage.ProviderResponse["response_id"] = parsed.ID
	if len(parsed.Choices) > 0 {
		usage.ProviderResponse["finish_reason"] = parsed.Choices[0].FinishReason
	}

	if len(parsed.Choices) == 0 {
		return "", usage, fmt.Errorf("apinet chat returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), usage, nil
}

func SafeProviderBaseURL(meta map[string]any, fallback string) (string, error) {
	base := strings.TrimSpace(stringValue(meta["base_url"]))
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("provider base_url must use https")
	}
	if parsed.Hostname() != "apinet.cloud" {
		return "", fmt.Errorf("provider host is not allowlisted")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func ProviderBaseHost(base string) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func ParseProviderModels(raw []byte, providerKey string) []ModelDescriptor {
	var anyValue any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&anyValue); err != nil {
		return nil
	}
	var entries []any
	switch root := anyValue.(type) {
	case map[string]any:
		if data, ok := root["data"].([]any); ok {
			entries = data
		} else if data, ok := root["models"].([]any); ok {
			entries = data
		}
	case []any:
		entries = root
	}
	out := make([]ModelDescriptor, 0, len(entries))
	seen := map[string]bool{}
	for _, entry := range entries {
		id := ""
		name := ""
		switch value := entry.(type) {
		case string:
			id = value
			name = value
		case map[string]any:
			id = stringValue(value["id"])
			if id == "" {
				id = stringValue(value["name"])
			}
			name = defaultString(stringValue(value["name"]), id)
		}
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, ModelDescriptor{
			ID:                  id,
			Name:                defaultString(name, id),
			ProviderKey:         providerKey,
			SupportedParameters: []string{"chat_completions", "json_object"},
			SupportsTools:       false,
			SupportsJSONSchema:  true,
			Notes:               "Получено через APINET /models",
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return ModelRank(out[i].ID) > ModelRank(out[j].ID)
	})
	return out
}

func FallbackAPINETModels(meta map[string]any, providerKey string) []ModelDescriptor {
	ids := make([]string, 0, 8)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		for _, existing := range ids {
			if existing == v {
				return
			}
		}
		ids = append(ids, v)
	}
	for _, key := range []string{"default_chat_model_id", "default_model_id", "default_tool_model_id", "default_patch_model_id"} {
		add(stringValue(meta[key]))
	}
	if raw, ok := meta["allowed_model_ids"].([]any); ok {
		for _, item := range raw {
			add(fmt.Sprint(item))
		}
	}
	for _, id := range []string{"gpt-5.5", "gpt-5.4-nano-low", "gpt-5"} {
		add(id)
	}
	out := make([]ModelDescriptor, 0, len(ids))
	for _, id := range ids {
		out = append(out, ModelDescriptor{
			ID:                  id,
			Name:                id,
			ProviderKey:         providerKey,
			SupportedParameters: []string{"chat_completions", "json_object"},
			SupportsTools:       false,
			SupportsJSONSchema:  true,
			Notes:               "Fallback из metadata APINET, используется если /models не вернул список",
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return ModelRank(out[i].ID) > ModelRank(out[j].ID) })
	return out
}

func ModelFromMetadata(meta map[string]any) string {
	for _, key := range []string{"default_chat_model_id", "default_model_id", "default_tool_model_id"} {
		if v := strings.TrimSpace(stringValue(meta[key])); v != "" {
			return v
		}
	}
	return ""
}

func TruncateForPrompt(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "..."
}

type ModelClient struct{ HTTP *http.Client }

func (a *ModelClient) client() *http.Client {
	if a != nil && a.HTTP != nil {
		return a.HTTP
	}
	return httptransport.NewDefaultClient(45 * time.Second)
}
func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func ModelRank(id string) int {
	lower := strings.ToLower(id)
	score := 0
	if strings.Contains(lower, "gpt-5") {
		score += 50
	}
	if strings.Contains(lower, "gpt-4.1") {
		score += 40
	}
	if strings.Contains(lower, "claude") {
		score += 30
	}
	if strings.Contains(lower, "gemini") {
		score += 25
	}
	if strings.Contains(lower, "mini") || strings.Contains(lower, "flash") {
		score -= 5
	}
	return score
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
