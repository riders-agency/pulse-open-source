package providers

type ModelProvider struct {
	ProviderKey       string   `json:"provider_key"`
	DependencyTypeKey string   `json:"dependency_type_key"`
	DisplayName       string   `json:"display_name"`
	Status            string   `json:"status"`
	DefaultBaseURL    string   `json:"default_base_url"`
	Capabilities      []string `json:"capabilities"`
	ToolLoop          string   `json:"tool_loop"`
	DocsURL           string   `json:"docs_url"`
}

type ModelDescriptor struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	ProviderKey         string   `json:"provider_key"`
	ContextLength       int      `json:"context_length,omitempty"`
	SupportedParameters []string `json:"supported_parameters"`
	SupportsTools       bool     `json:"supports_tools"`
	SupportsJSONSchema  bool     `json:"supports_json_schema"`
	Notes               string   `json:"notes,omitempty"`
}

func ModelProviders() []ModelProvider {
	return []ModelProvider{
		{
			ProviderKey:       "ai.apinet",
			DependencyTypeKey: "ai.apinet",
			DisplayName:       "APINET / Riders Agency AI Gateway",
			Status:            "available",
			DefaultBaseURL:    "https://apinet.cloud/v1",
			Capabilities:      []string{"chat_completions", "models", "embeddings", "openai_compatible", "tool_argument_generation"},
			ToolLoop:          "riders_pulse_typed_tools",
			DocsURL:           "https://apinet.cloud/",
		},
		{
			ProviderKey:       "ai.openai",
			DependencyTypeKey: "ai.openai",
			DisplayName:       "OpenAI",
			Status:            "available",
			DefaultBaseURL:    "https://api.openai.com/v1",
			Capabilities:      []string{"chat", "responses", "tools", "json_schema", "patch_proposals"},
			ToolLoop:          "native",
			DocsURL:           "https://platform.openai.com/docs",
		},
		{
			ProviderKey:       "ai.openrouter",
			DependencyTypeKey: "ai.openrouter",
			DisplayName:       "OpenRouter",
			Status:            "planned",
			DefaultBaseURL:    "https://openrouter.ai/api/v1",
			Capabilities:      []string{"chat", "model_catalog", "routing", "fallbacks"},
			ToolLoop:          "adapter_pending",
			DocsURL:           "https://openrouter.ai/docs",
		},
		{
			ProviderKey:       "ai.google_gemini",
			DependencyTypeKey: "ai.google_gemini",
			DisplayName:       "Google Gemini",
			Status:            "planned",
			DefaultBaseURL:    "https://generativelanguage.googleapis.com/v1beta",
			Capabilities:      []string{"chat", "json_schema", "vision"},
			ToolLoop:          "adapter_pending",
			DocsURL:           "https://ai.google.dev/gemini-api/docs",
		},
		{
			ProviderKey:       "ai.anthropic",
			DependencyTypeKey: "ai.anthropic",
			DisplayName:       "Anthropic Claude",
			Status:            "planned",
			DefaultBaseURL:    "https://api.anthropic.com/v1",
			Capabilities:      []string{"chat", "long_context", "tools"},
			ToolLoop:          "adapter_pending",
			DocsURL:           "https://docs.anthropic.com",
		},
		{
			ProviderKey:       "ai.mistral",
			DependencyTypeKey: "ai.mistral",
			DisplayName:       "Mistral AI",
			Status:            "planned",
			DefaultBaseURL:    "https://api.mistral.ai/v1",
			Capabilities:      []string{"chat", "json_schema"},
			ToolLoop:          "adapter_pending",
			DocsURL:           "https://docs.mistral.ai",
		},
		{
			ProviderKey:       "ai.groq",
			DependencyTypeKey: "ai.groq",
			DisplayName:       "Groq",
			Status:            "planned",
			DefaultBaseURL:    "https://api.groq.com/openai/v1",
			Capabilities:      []string{"chat", "openai_compatible", "low_latency"},
			ToolLoop:          "adapter_pending",
			DocsURL:           "https://console.groq.com/docs",
		},
		{
			ProviderKey:       "ai.xai",
			DependencyTypeKey: "ai.xai",
			DisplayName:       "xAI",
			Status:            "planned",
			DefaultBaseURL:    "https://api.x.ai/v1",
			Capabilities:      []string{"chat", "openai_compatible"},
			ToolLoop:          "adapter_pending",
			DocsURL:           "https://docs.x.ai",
		},
	}
}
