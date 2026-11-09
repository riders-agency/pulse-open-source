package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"
	"riders.agency/pulse/api/internal/platform/authz"
)

func ModelProviders() []ModelProvider {
	return aiproviders.ModelProviders()
}

func (a *AI) ListModels(ctx context.Context, projectID, dependencyID, userID uuid.UUID, refresh bool) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	var providerKey, depTypeKey, depName string
	err := a.P.QueryRow(ctx, `
SELECT t.provider_key::text, t.dependency_type_key::text, ad.name::text
  FROM pulse_read.action_dependencies_v1 ad
  JOIN pulse_read.action_dependency_types_v1 t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.project_id = $1
   AND ad.action_dependency_id = $2
   AND (t.dependency_type_key LIKE 'ai.%' OR t.dependency_type_key = 'network.vless_proxy')
   AND ad.status <> 'deleted'`,
		projectID, dependencyID).Scan(&providerKey, &depTypeKey, &depName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	models := modelsForDependency(depTypeKey)
	source := "static_mvp_catalog"
	if depTypeKey == "ai.apinet" {
		liveModels, liveSource, err := a.listAPINETModels(ctx, projectID, dependencyID, refresh)
		if err != nil && len(models) == 0 {
			return nil, err
		}
		if err == nil {
			models = liveModels
			source = liveSource
		}
	}
	modelsJSON, _ := json.Marshal(models)
	if refresh {
		_, _ = a.P.Exec(ctx, `
INSERT INTO ai_model_catalog_cache (action_dependency_id, provider_key, models_json, fetched_at, expires_at, status, error_text)
VALUES ($1, $2, $3::jsonb, now(), now() + interval '15 minutes', 'fresh', '')
ON CONFLICT (action_dependency_id) DO UPDATE SET
  provider_key = EXCLUDED.provider_key,
  models_json = EXCLUDED.models_json,
  fetched_at = now(),
  expires_at = now() + interval '15 minutes',
  status = 'fresh',
  error_text = ''`, dependencyID, providerKey, modelsJSON)
	}
	return map[string]any{
		"action_dependency_id": dependencyID.String(),
		"name":                 depName,
		"provider_key":         providerKey,
		"dependency_type_key":  depTypeKey,
		"models":               models,
		"source":               source,
		"refreshed":            refresh,
		"fetched_at":           time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func modelsForDependency(depTypeKey string) []ModelDescriptor {
	switch depTypeKey {
	case "ai.apinet":
		return []ModelDescriptor{
			{ID: "gpt-5.5", Name: "GPT-5.5 via APINET", ProviderKey: "ai.apinet", SupportedParameters: []string{"chat_completions", "json_object"}, SupportsTools: false, SupportsJSONSchema: true, Notes: "Fallback only; refresh /models to get the live APINET catalog."},
			{ID: "gpt-5", Name: "GPT-5 via APINET", ProviderKey: "ai.apinet", SupportedParameters: []string{"chat_completions", "json_object"}, SupportsTools: false, SupportsJSONSchema: true, Notes: "Fallback only; refresh /models to get the live APINET catalog."},
		}
	case "ai.openai":
		return []ModelDescriptor{
			{ID: "gpt-5", Name: "GPT-5", ProviderKey: "ai.openai", SupportedParameters: []string{"tools", "json_schema", "max_output_tokens"}, SupportsTools: true, SupportsJSONSchema: true},
			{ID: "gpt-5-mini", Name: "GPT-5 mini", ProviderKey: "ai.openai", SupportedParameters: []string{"tools", "json_schema", "max_output_tokens"}, SupportsTools: true, SupportsJSONSchema: true},
			{ID: "gpt-4.1", Name: "GPT-4.1", ProviderKey: "ai.openai", SupportedParameters: []string{"tools", "json_schema", "max_output_tokens"}, SupportsTools: true, SupportsJSONSchema: true},
		}
	case "ai.openrouter":
		return []ModelDescriptor{
			{ID: "openrouter/auto", Name: "OpenRouter Auto", ProviderKey: "ai.openrouter", SupportedParameters: []string{"route", "fallbacks"}, SupportsTools: false, SupportsJSONSchema: false, Notes: "Точный список приходит из OpenRouter model catalog после подключения adapter refresh."},
			{ID: "anthropic/claude-sonnet-4", Name: "Claude Sonnet via OpenRouter", ProviderKey: "ai.openrouter", SupportedParameters: []string{"route"}, SupportsTools: false, SupportsJSONSchema: false},
			{ID: "google/gemini-2.5-pro", Name: "Gemini Pro via OpenRouter", ProviderKey: "ai.openrouter", SupportedParameters: []string{"route"}, SupportsTools: false, SupportsJSONSchema: false},
		}
	case "ai.google_gemini":
		return []ModelDescriptor{
			{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", ProviderKey: "ai.google_gemini", SupportedParameters: []string{"json_schema"}, SupportsTools: false, SupportsJSONSchema: true},
			{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", ProviderKey: "ai.google_gemini", SupportedParameters: []string{"json_schema"}, SupportsTools: false, SupportsJSONSchema: true},
		}
	case "ai.anthropic":
		return []ModelDescriptor{
			{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", ProviderKey: "ai.anthropic", SupportedParameters: []string{"tools"}, SupportsTools: true, SupportsJSONSchema: false},
			{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", ProviderKey: "ai.anthropic", SupportedParameters: []string{"tools"}, SupportsTools: true, SupportsJSONSchema: false},
		}
	case "ai.mistral":
		return []ModelDescriptor{
			{ID: "mistral-large-latest", Name: "Mistral Large", ProviderKey: "ai.mistral", SupportedParameters: []string{"json_schema"}, SupportsTools: false, SupportsJSONSchema: true},
			{ID: "codestral-latest", Name: "Codestral", ProviderKey: "ai.mistral", SupportedParameters: []string{}, SupportsTools: false, SupportsJSONSchema: false},
		}
	case "ai.groq":
		return []ModelDescriptor{
			{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B Versatile", ProviderKey: "ai.groq", SupportedParameters: []string{"openai_compatible"}, SupportsTools: false, SupportsJSONSchema: false},
			{ID: "openai/gpt-oss-120b", Name: "GPT OSS 120B", ProviderKey: "ai.groq", SupportedParameters: []string{"openai_compatible"}, SupportsTools: false, SupportsJSONSchema: false},
		}
	case "ai.xai":
		return []ModelDescriptor{
			{ID: "grok-4", Name: "Grok 4", ProviderKey: "ai.xai", SupportedParameters: []string{"openai_compatible"}, SupportsTools: false, SupportsJSONSchema: false},
			{ID: "grok-3", Name: "Grok 3", ProviderKey: "ai.xai", SupportedParameters: []string{"openai_compatible"}, SupportsTools: false, SupportsJSONSchema: false},
		}
	default:
		return []ModelDescriptor{}
	}
}
