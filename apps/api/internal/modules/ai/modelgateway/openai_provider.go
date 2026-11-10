package modelgateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAICompatibleProvider struct {
	Client *http.Client
	Config Config
}

func (p OpenAICompatibleProvider) Chat(ctx context.Context, call ChatCall) (ChatResult, error) {
	payload := map[string]any{
		"model":       call.ModelID,
		"messages":    openAIMessages(call.Messages),
		"temperature": call.Temperature,
		"max_tokens":  call.MaxTokens,
	}
	if len(call.ResponseFormat) > 0 {
		payload["response_format"] = call.ResponseFormat
	}
	if len(call.ToolSchemas) > 0 {
		payload["tools"] = call.ToolSchemas
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Config.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return ChatResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return ChatResult{}, typedProviderError("provider_unavailable", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ChatResult{}, typedProviderError("provider_http", fmt.Errorf("model provider status %d", resp.StatusCode))
	}
	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ChatResult{}, typedProviderError("provider_bad_json", err)
	}
	if len(parsed.Choices) == 0 {
		return ChatResult{}, typedProviderError("provider_empty_response", fmt.Errorf("chat returned no choices"))
	}
	content := parsed.Choices[0].Message.Content
	responseJSON, err := parseStructuredResponse(content, call.ResponseFormat)
	if err != nil {
		return ChatResult{}, err
	}
	return ChatResult{
		Content:      content,
		ResponseJSON: responseJSON,
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
		Metadata: map[string]any{
			"provider": "openai_compatible",
			"model":    call.ModelID,
		},
	}, nil
}

func (p OpenAICompatibleProvider) Embed(ctx context.Context, call EmbedCall) (EmbedResult, error) {
	payload := map[string]any{"model": call.ModelID, "input": call.Input}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Config.BaseURL+"/embeddings", bytes.NewReader(raw))
	if err != nil {
		return EmbedResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return EmbedResult{}, typedProviderError("provider_unavailable", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return EmbedResult{}, typedProviderError("provider_http", fmt.Errorf("embedding provider status %d", resp.StatusCode))
	}
	var parsed embeddingResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return EmbedResult{}, typedProviderError("provider_bad_json", err)
	}
	embeddings := make([][]float64, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		embeddings = append(embeddings, item.Embedding)
	}
	return EmbedResult{
		Embeddings:  embeddings,
		InputTokens: parsed.Usage.PromptTokens,
		Metadata: map[string]any{
			"provider": "openai_compatible",
			"model":    call.ModelID,
		},
	}, nil
}

func (p OpenAICompatibleProvider) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

func openAIMessages(messages []Message) []map[string]any {
	out := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		if role == "" {
			role = "user"
		}
		out = append(out, map[string]any{"role": role, "content": message.Content})
	}
	return out
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
	} `json:"usage"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens int64 `json:"prompt_tokens"`
	} `json:"usage"`
}
