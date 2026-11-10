package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func (r *Runner) loadTelegramDependency(ctx context.Context, depID, actionName string) (*dependencyRecord, error) {
	dep, err := r.loadDependency(ctx, depID)
	if err != nil {
		return nil, err
	}
	if dep.TypeKey != "telegram.bot" && dep.TypeKey != "telegram.relay" {
		return nil, fmt.Errorf("%s requires telegram.bot or telegram.relay dependency", actionName)
	}
	if dep.Status != "active" {
		return nil, fmt.Errorf("telegram dependency is %s", dep.Status)
	}
	return dep, nil
}

func telegramDependencyErrorType(err error) string {
	msg := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(msg, "requires telegram.bot"):
		return "dependency_type"
	case strings.HasPrefix(msg, "telegram dependency is "):
		return "dependency_status"
	default:
		return "dependency"
	}
}

func (r *Runner) callTelegramJSON(ctx context.Context, dep *dependencyRecord, actionType, method string, body map[string]any, timeout time.Duration) (map[string]any, actionExecution) {
	payload, _ := json.Marshal(body)
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, "https://api.telegram.org/bot"+dep.Plaintext+"/"+method, bytes.NewReader(payload))
	if err != nil {
		cancel()
		ex := failedAction("{}", "{}", map[string]any{"provider": "telegram", "ok": false, "error": "request build failed"}, false, "request", errors.New("telegram request build failed"))
		return map[string]any{"provider": "telegram", "ok": false}, ex
	}
	defer cancel()
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient().Do(req)
	if err != nil {
		ex := failedAction("{}", "{}", map[string]any{"provider": "telegram", "ok": false, "error": "telegram request failed"}, true, "network", errors.New("telegram request failed"))
		return map[string]any{"provider": "telegram", "ok": false}, ex
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	provider := sanitizeTelegramResponse(resp.StatusCode, raw)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		providerJSON, _ := json.Marshal(provider)
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		err := fmt.Errorf("telegram returned HTTP %d", resp.StatusCode)
		report := diagnosticsForFailure(actionType, provider, retryable, "provider_http", err)
		output := providerdiagnostics.OutputWithReport("{}", report)
		return provider, actionExecution{InputJSON: "{}", OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: "provider_http", Err: err}
	}
	return provider, actionExecution{}
}

func sanitizeTelegramResponse(statusCode int, raw []byte) map[string]any {
	out := map[string]any{"provider": "telegram", "status_code": statusCode, "provider_response_redacted": true}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		out["ok"] = false
		out["description"] = strings.TrimSpace(string(raw))
		return out
	}
	out["raw_response_json"] = payload
	if ok, exists := payload["ok"]; exists {
		out["ok"] = ok
	}
	if desc, exists := payload["description"]; exists {
		out["description"] = desc
	}
	if params, ok := payload["parameters"].(map[string]any); ok {
		out["parameters"] = params
	}
	if result, ok := payload["result"].(map[string]any); ok {
		out["message_id"] = result["message_id"]
		out["date"] = result["date"]
		if chat, ok := result["chat"].(map[string]any); ok {
			out["chat"] = map[string]any{"id": chat["id"], "type": chat["type"], "username": chat["username"]}
		}
	}
	return out
}
