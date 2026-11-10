package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	depcommands "riders.agency/pulse/api/internal/modules/action_dependencies/commands"
	depqueries "riders.agency/pulse/api/internal/modules/action_dependencies/queries"
	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func (a *AI) listAPINETModels(ctx context.Context, projectID, dependencyID uuid.UUID, refresh bool) ([]ModelDescriptor, string, error) {
	if !refresh {
		var raw []byte
		err := a.P.QueryRow(ctx, `
SELECT models_json FROM ai_model_catalog_cache
 WHERE action_dependency_id = $1
   AND expires_at > now()
   AND status = 'fresh'`, dependencyID).Scan(&raw)
		if err == nil && len(raw) > 0 {
			var cached []ModelDescriptor
			if json.Unmarshal(raw, &cached) == nil && len(cached) > 0 {
				return cached, "cache", nil
			}
		}
	}
	dep, err := a.loadDependency(ctx, projectID, dependencyID)
	if err != nil {
		return nil, "", err
	}
	if dep.DependencyType != "ai.apinet" {
		return nil, "", fmt.Errorf("unsupported live provider %s", dep.DependencyType)
	}
	models, err := a.fetchAPINETModels(ctx, dep)
	modelsJSON, _ := json.Marshal(models)
	status := "fresh"
	errText := ""
	if err != nil {
		status = "error"
		errText = err.Error()
	}
	_, _ = a.P.Exec(ctx, `
INSERT INTO ai_model_catalog_cache (action_dependency_id, provider_key, models_json, fetched_at, expires_at, status, error_text)
VALUES ($1, $2, $3::jsonb, now(), now() + interval '24 hours', $4, $5)
ON CONFLICT (action_dependency_id) DO UPDATE SET
  provider_key = EXCLUDED.provider_key,
  models_json = EXCLUDED.models_json,
  fetched_at = now(),
  expires_at = now() + interval '24 hours',
  status = EXCLUDED.status,
  error_text = EXCLUDED.error_text`, dependencyID, dep.ProviderKey, modelsJSON, status, errText)
	if err != nil {
		return nil, "provider_error", err
	}
	if len(models) == 0 {
		models = aiproviders.FallbackAPINETModels(dep.Metadata, dep.ProviderKey)
	}
	return models, "apinet_live", nil
}

func (a *AI) EvaluateModels(ctx context.Context, projectID, dependencyID, userID uuid.UUID) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermWrite); err != nil {
		return nil, err
	}
	dep, err := a.loadDependency(ctx, projectID, dependencyID)
	if err != nil {
		return nil, err
	}
	if dep.DependencyType != "ai.apinet" {
		return nil, fmt.Errorf("model evaluation currently supports ai.apinet")
	}
	models, _, err := a.listAPINETModels(ctx, projectID, dependencyID, true)
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		models = aiproviders.FallbackAPINETModels(dep.Metadata, dep.ProviderKey)
	}
	candidates := selectEvaluationCandidates(models, dep.Metadata)
	results := make([]map[string]any, 0, len(candidates))
	bestModel := ""
	bestScore := -1.0
	for _, model := range candidates {
		start := time.Now()
		prompt := `Верни только JSON без markdown: {"intent":"task.create","entities":{"title":"Проверить оплату Орион","client_query":"Орион"},"patch":{"format":"riders-pulse-workflow-patch-v1","operations":[]}}`
		content, usage, err := a.apinetChat(ctx, projectID, uuid.Nil, dep, model.ID, []apinetChatMessage{
			{Role: "system", Content: "Ты проверяешь пригодность модели для строгого JSON и typed tools Riders Pulse."},
			{Role: "user", Content: prompt},
		}, 320)
		score := scoreEvaluation(content, err, time.Since(start))
		capability := map[string]any{
			"chat_completions": true,
			"json_schema_soft": score >= 55,
			"tools_internal":   true,
			"error":            "",
		}
		if err != nil {
			capability["error"] = "provider_error"
		}
		benchmark := map[string]any{
			"latency_ms":      usage.LatencyMS,
			"sample_redacted": true,
		}
		_, _ = a.P.Exec(ctx, `
INSERT INTO ai_model_evaluations (
  project_id, action_dependency_id, model_id, provider_key, score,
  capability_json, benchmark_json, selected_for
) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb,$8)`,
			projectID, dependencyID, model.ID, dep.ProviderKey, score,
			mustJSON(capability), mustJSON(benchmark), "")
		results = append(results, map[string]any{
			"model_id":        model.ID,
			"score":           score,
			"capability_json": capability,
			"benchmark_json":  benchmark,
		})
		if score > bestScore {
			bestScore = score
			bestModel = model.ID
		}
	}
	if bestModel == "" && len(models) > 0 {
		bestModel = models[0].ID
		bestScore = 1
	}
	if bestModel != "" {
		dep.Metadata["default_model_id"] = bestModel
		dep.Metadata["default_chat_model_id"] = bestModel
		dep.Metadata["default_tool_model_id"] = bestModel
		dep.Metadata["default_patch_model_id"] = bestModel
		if err := depcommands.SetDefaultModels(ctx, a.P, projectID, dependencyID, json.RawMessage(mustJSON(dep.Metadata))); err != nil {
			return nil, err
		}

		_, _ = a.P.Exec(ctx, `
UPDATE project_ai_employees
   SET action_dependency_id = $2, model_id = $3, updated_at = now()
 WHERE project_id = $1`, projectID, dependencyID, bestModel)
		_, _ = a.P.Exec(ctx, `
UPDATE ai_model_evaluations
   SET selected_for = 'default_chat,default_tool,default_patch'
 WHERE project_id = $1 AND action_dependency_id = $2 AND model_id = $3`,
			projectID, dependencyID, bestModel)
	}
	return map[string]any{
		"action_dependency_id": dependencyID.String(),
		"provider_key":         dep.ProviderKey,
		"selected_model_id":    bestModel,
		"score":                bestScore,
		"items":                results,
	}, nil
}

func (a *AI) fetchAPINETModels(ctx context.Context, dep aiDependency) ([]ModelDescriptor, error) {
	if a.ModelClient == nil {
		return nil, fmt.Errorf("model gateway not configured")
	}
	return a.ModelClient.FetchModels(ctx, dep)
}
func (a *AI) apinetChat(ctx context.Context, projectID, sessionID uuid.UUID, dep aiDependency, modelID string, messages []apinetChatMessage, maxTokens int) (string, aiTurnUsage, error) {
	if a.ModelClient == nil {
		return "", aiTurnUsage{}, fmt.Errorf("model gateway not configured")
	}
	text, usage, err := a.ModelClient.Chat(ctx, dep, modelID, messages, maxTokens)
	a.recordUsage(ctx, projectID, sessionID, usage)
	return text, usage, err
}

func (a *AI) sessionProvider(ctx context.Context, projectID, sessionID uuid.UUID) (aiDependency, string, json.RawMessage, error) {
	var depText, modelID string
	var pageContext []byte
	err := a.P.QueryRow(ctx, `
SELECT COALESCE(s.action_dependency_id::text, e.action_dependency_id::text, ''),
       COALESCE(NULLIF(s.model_id,''), NULLIF(e.model_id,''), ''),
       COALESCE(s.page_context_json, '{}'::jsonb)
  FROM ai_chat_sessions s
  JOIN project_ai_employees e ON e.project_ai_employee_id = s.project_ai_employee_id
 WHERE s.ai_chat_session_id = $1
   AND s.project_id = $2`, sessionID, projectID).Scan(&depText, &modelID, &pageContext)
	if err != nil {
		return aiDependency{}, "", nil, err
	}
	depID := uuid.Nil
	if depText != "" {
		depID, _ = uuid.Parse(depText)
	}
	if depID == uuid.Nil {
		found, err := a.defaultAIProvider(ctx, projectID)
		if err != nil {
			return aiDependency{}, "", pageContext, err
		}
		depID = found
	}
	dep, err := a.loadDependency(ctx, projectID, depID)
	if err != nil {
		return aiDependency{}, "", pageContext, err
	}
	if modelID == "" {
		modelID = aiproviders.ModelFromMetadata(dep.Metadata)
	}
	if modelID == "" {
		if models, _, err := a.listAPINETModels(ctx, projectID, dep.ID, false); err == nil && len(models) > 0 {
			modelID = models[0].ID
		}
	}
	if modelID != "" {
		_, _ = a.P.Exec(ctx, `
UPDATE ai_chat_sessions SET action_dependency_id = $2, model_id = $3 WHERE ai_chat_session_id = $1`,
			sessionID, dep.ID, modelID)
	}
	return dep, modelID, json.RawMessage(pageContext), nil
}

func (a *AI) defaultAIProvider(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := a.P.QueryRow(ctx, `
SELECT ad.action_dependency_id
  FROM pulse_read.action_dependencies_v1 ad
  JOIN pulse_read.action_dependency_types_v1 t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.project_id = $1
   AND ad.status = 'active'
   AND t.dependency_type_key LIKE 'ai.%'
 ORDER BY CASE WHEN t.dependency_type_key = 'ai.apinet' THEN 0 ELSE 1 END, ad.updated_at DESC
 LIMIT 1`, projectID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, fmt.Errorf("active AI integration required")
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (a *AI) loadDependency(ctx context.Context, projectID, dependencyID uuid.UUID) (aiDependency, error) {
	var dep aiDependency
	var metaBytes []byte
	err := a.P.QueryRow(ctx, `
SELECT ad.action_dependency_id, ad.project_id, ad.name,
       t.provider_key::text, t.dependency_type_key::text,
       COALESCE(ad.metadata_json, '{}'::jsonb)
  FROM pulse_read.action_dependencies_v1 ad
  JOIN pulse_read.action_dependency_types_v1 t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.project_id = $1
   AND ad.action_dependency_id = $2
   AND ad.status <> 'deleted'`, projectID, dependencyID).Scan(
		&dep.ID, &dep.ProjectID, &dep.Name, &dep.ProviderKey, &dep.DependencyType, &metaBytes)
	if err != nil {
		if err == pgx.ErrNoRows {
			return dep, ErrNotFound
		}
		return dep, err
	}
	_ = json.Unmarshal(metaBytes, &dep.Metadata)
	if dep.Metadata == nil {
		dep.Metadata = map[string]any{}
	}
	encrypted, err := depqueries.EncryptedCredential(ctx, a.P, dependencyID.String())
	if err != nil {
		return dep, err
	}
	if len(encrypted) > 0 {
		plain, err := cryptoapp.Open(encrypted, a.MasterHEX)
		if err != nil {
			return dep, err
		}
		dep.APIKey = strings.TrimSpace(string(plain))
	}
	return dep, nil
}

func (a *AI) recordUsage(ctx context.Context, projectID, sessionID uuid.UUID, usage aiTurnUsage) {
	if usage.ProviderKey == "" && usage.ModelID == "" && usage.InputTokens == 0 && usage.OutputTokens == 0 {
		return
	}
	var sid any
	if sessionID != uuid.Nil {
		sid = sessionID
	}
	_, _ = a.P.Exec(ctx, `
INSERT INTO ai_usage_events (
  project_id, ai_chat_session_id, provider_key, model_id,
  input_tokens, output_tokens, latency_ms, provider_response_json
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)`,
		projectID, sid, usage.ProviderKey, usage.ModelID, usage.InputTokens, usage.OutputTokens, usage.LatencyMS, mustJSON(usage.ProviderResponse))
}
