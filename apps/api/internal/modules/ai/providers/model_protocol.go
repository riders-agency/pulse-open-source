package providers

import "github.com/google/uuid"

type ModelUsage struct {
	ProviderKey      string         `json:"provider_key,omitempty"`
	ModelID          string         `json:"model_id,omitempty"`
	InputTokens      int            `json:"input_tokens"`
	OutputTokens     int            `json:"output_tokens"`
	LatencyMS        int            `json:"latency_ms"`
	ProviderResponse map[string]any `json:"provider_response_json,omitempty"`
}
type ModelDependency struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	ProviderKey    string
	DependencyType string
	Metadata       map[string]any
	APIKey         string
}
type apinetChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apinetChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
