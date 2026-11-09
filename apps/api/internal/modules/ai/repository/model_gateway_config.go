package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/modules/ai/modelgateway"
)

func (a *AI) ResolveModelGatewayConfig(ctx context.Context, projectID, sessionID uuid.UUID) (modelgateway.Config, error) {
	if a == nil || a.P == nil {
		return modelgateway.Config{}, fmt.Errorf("ai repository is not configured")
	}
	dep, modelID, err := a.modelGatewayDependency(ctx, projectID, sessionID)
	if err != nil {
		return modelgateway.Config{}, err
	}
	if dep.DependencyType != "ai.apinet" {
		return modelgateway.Config{}, fmt.Errorf("unsupported model gateway provider %s", dep.DependencyType)
	}
	if modelID == "" {
		modelID = aiproviders.ModelFromMetadata(dep.Metadata)
	}
	if modelID == "" {
		if models, _, listErr := a.listAPINETModels(ctx, projectID, dep.ID, false); listErr == nil && len(models) > 0 {
			modelID = models[0].ID
		}
	}
	base, err := aiproviders.SafeProviderBaseURL(dep.Metadata, "https://apinet.cloud/v1")
	if err != nil {
		return modelgateway.Config{}, err
	}
	return modelgateway.Config{
		ProviderKey:      "apinet",
		BaseURL:          base,
		APIKey:           dep.APIKey,
		DefaultModelID:   modelID,
		EmbeddingModelID: modelID,
		AllowedModels:    modelGatewayAllowedModels(dep.Metadata),
		DefaultMaxTokens: 700,
		HardMaxTokens:    1800,
		Timeout:          20 * time.Second,
	}, nil
}

func (a *AI) modelGatewayDependency(ctx context.Context, projectID, sessionID uuid.UUID) (aiDependency, string, error) {
	var dep aiDependency
	modelID := ""
	var err error
	if sessionID != uuid.Nil {
		dep, modelID, _, err = a.sessionProvider(ctx, projectID, sessionID)
	}
	if dep.ID != uuid.Nil {
		return dep, modelID, nil
	}
	depID, fallbackErr := a.defaultAIProvider(ctx, projectID)
	if fallbackErr != nil {
		if err != nil {
			return aiDependency{}, "", err
		}
		return aiDependency{}, "", fallbackErr
	}
	dep, err = a.loadDependency(ctx, projectID, depID)
	if err != nil {
		return aiDependency{}, "", err
	}
	return dep, aiproviders.ModelFromMetadata(dep.Metadata), nil
}

func modelGatewayAllowedModels(meta map[string]any) map[string]struct{} {
	allowed := map[string]struct{}{}
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			allowed[value] = struct{}{}
		}
	}
	for _, key := range []string{
		"default_chat_model_id",
		"default_model_id",
		"default_tool_model_id",
		"default_patch_model_id",
	} {
		add(stringValue(meta[key]))
	}
	if raw, ok := meta["allowed_model_ids"].([]any); ok {
		for _, item := range raw {
			add(fmt.Sprint(item))
		}
	}
	return allowed
}
