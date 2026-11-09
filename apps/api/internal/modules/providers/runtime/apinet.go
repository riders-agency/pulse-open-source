package runtime

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

const defaultAPINETNormalizeInstructions = "Ты редактор срочных сообщений Демо. Исправь орфографию, пунктуацию и формулировки; сохрани факты, числа, имена, телефоны, адреса и смысл; не добавляй неподтвержденные данные; добавь 1-3 уместных эмодзи максимум; сделай текст понятным для публикации в Telegram-канале. Верни только готовый текст без пояснений и markdown."

func executeAPINETNormalizeTextAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeAPINETNormalizeText)
}

func (r *Runner) executeAPINETNormalizeText(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadDependency(ctx, node.Config.ActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "ai.apinet", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	if dep.TypeKey != "ai.apinet" {
		err := errors.New("ai.apinet.normalize_text requires ai.apinet dependency")
		return failedAction(input, "{}", map[string]any{"provider": "ai.apinet", "ok": false, "error": err.Error()}, false, "dependency_type", err)
	}
	if dep.Status != "active" {
		err := fmt.Errorf("ai.apinet dependency is %s", dep.Status)
		return failedAction(input, "{}", map[string]any{"provider": "ai.apinet", "ok": false, "status": dep.Status}, false, "dependency_status", err)
	}
	if strings.TrimSpace(dep.Plaintext) == "" {
		err := errors.New("ai.apinet api key is empty")
		return failedAction(input, "{}", map[string]any{"provider": "ai.apinet", "ok": false, "error": err.Error()}, false, "dependency", err)
	}

	source := strings.TrimSpace(renderTemplate(node.Config.SourceText, runInput, steps, tpl))
	if source == "" {
		err := errors.New("source_text is required")
		return failedAction(input, "{}", map[string]any{"provider": "ai.apinet", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	instructions := strings.TrimSpace(renderTemplate(node.Config.Instructions, runInput, steps, tpl))
	if instructions == "" {
		instructions = defaultAPINETNormalizeInstructions
	}
	model := apinetModel(node.Config, dep.Metadata)
	baseURL := apinetBaseURL(dep.Metadata)
	timeout := durationOrDefault(node.Config.RetryPolicy.Timeout, 45*time.Second)
	temperature := 0.2
	if node.Config.Temperature != nil {
		temperature = *node.Config.Temperature
	}
	maxTokens := node.Config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 800
	}

	requestInput, _ := json.Marshal(map[string]any{
		"provider":                "ai.apinet",
		"model":                   model,
		"source_text_length":      len([]rune(source)),
		"instructions_sha256":     sha256Hex(instructions),
		"temperature":             temperature,
		"max_tokens":              maxTokens,
		"success_criteria":        node.Config.SuccessCriteria,
		"redacted_dependency":     node.Config.ActionDependencyID != "",
		"provider_input_redacted": true,
	})
	payload, _ := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": instructions},
			{"role": "user", "content": source},
		},
		"temperature": temperature,
		"max_tokens":  maxTokens,
	})

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		cancel()
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "ai.apinet", "ok": false, "error": "request build failed"}, false, "request", errors.New("ai.apinet request build failed"))
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(dep.Plaintext))
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient().Do(req)
	cancel()
	if err != nil {
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "ai.apinet", "ok": false, "error": "ai.apinet request failed"}, true, "network", errors.New("ai.apinet request failed"))
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	resp.Body.Close()

	provider := sanitizeAPINETResponse(resp.StatusCode, raw, model)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		providerJSON, _ := json.Marshal(provider)
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		err := fmt.Errorf("ai.apinet returned HTTP %d", resp.StatusCode)
		return actionExecution{InputJSON: string(requestInput), OutputJSON: "{}", ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: "provider_http", Err: err}
	}

	normalized := strings.TrimSpace(apinetResponseText(raw))
	if normalized == "" {
		providerJSON, _ := json.Marshal(provider)
		err := errors.New("ai.apinet returned empty normalized text")
		return actionExecution{InputJSON: string(requestInput), OutputJSON: "{}", ProviderResponseJSON: string(providerJSON), Retryable: false, ErrorType: "provider", Err: err}
	}
	output := map[string]any{
		"provider":        "ai.apinet",
		"model":           provider["model"],
		"original_text":   source,
		"normalized_text": normalized,
		"changed":         normalized != source,
		"expose": map[string]any{
			"original_text":   source,
			"normalized_text": normalized,
			"model":           provider["model"],
		},
	}
	if usage, ok := provider["usage"]; ok {
		output["usage"] = usage
	}
	outputJSON, _ := json.Marshal(output)
	providerJSON, _ := json.Marshal(provider)
	return actionExecution{InputJSON: string(requestInput), OutputJSON: string(outputJSON), ProviderResponseJSON: string(providerJSON)}
}

func apinetModel(cfg actionConfig, metadata map[string]any) string {
	return firstNonEmpty(
		cfg.Model,
		stringMetadata(metadata, "default_chat_model_id"),
		stringMetadata(metadata, "default_model_id"),
		stringMetadata(metadata, "default_tool_model_id"),
		"gpt-5-chat-latest",
	)
}

func apinetBaseURL(metadata map[string]any) string {
	baseURL := firstNonEmpty(stringMetadata(metadata, "base_url"), "https://apinet.cloud/v1")
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "https://apinet.cloud/v1"
	}
	return baseURL
}

func sanitizeAPINETResponse(statusCode int, raw []byte, fallbackModel string) map[string]any {
	provider := map[string]any{
		"provider":                   "ai.apinet",
		"ok":                         statusCode >= 200 && statusCode < 300,
		"status_code":                statusCode,
		"model":                      fallbackModel,
		"provider_response_redacted": true,
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err == nil {
		if id, ok := parsed["id"].(string); ok {
			provider["id"] = id
		}
		if model, ok := parsed["model"].(string); ok && strings.TrimSpace(model) != "" {
			provider["model"] = model
		}
		if usage, ok := parsed["usage"].(map[string]any); ok {
			provider["usage"] = usage
		}
		if errObj, ok := parsed["error"].(map[string]any); ok {
			provider["error_type"] = fmt.Sprint(errObj["type"])
			provider["error_code"] = fmt.Sprint(errObj["code"])
		}
	}
	return provider
}

func apinetResponseText(raw []byte) string {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return ""
	}
	if strings.TrimSpace(parsed.Choices[0].Message.Content) != "" {
		return parsed.Choices[0].Message.Content
	}
	return parsed.Choices[0].Text
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
