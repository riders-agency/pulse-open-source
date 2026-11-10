package http

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"riders.agency/pulse/api/internal/platform/problem"
)

func RuntimeReadiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runtime := firstNonEmptyEnv("AI_EMPLOYEE_RUNTIME", "langgraph")
		target := firstNonEmptyEnv("AI_AGENT_GRPC_ADDR", firstNonEmptyEnv("PULSE_AI_AGENT_GRPC_ADDR", firstNonEmptyEnv("LANGGRAPH_AGENT_GRPC_ADDR", "")))
		if target == "" && runtime == "langgraph" {
			target = aiAgentGRPCTarget()
		}
		if strings.HasPrefix(target, ":") {
			target = "127.0.0.1" + target
		}
		python := map[string]any{
			"configured": false,
			"ok":         false,
			"target":     "",
			"status":     "not_configured",
		}
		if target != "" {
			python["configured"] = true
			python["target"] = target
			python["security"] = aiAgentGRPCSecurityReadiness(target)
			ok, status, errText := checkPythonGRPC(r.Context(), target)
			python["ok"] = ok
			python["status"] = status
			if errText != "" {
				python["error"] = errText
			}
		}
		capability := evaluateAICapability(r.Context(), capabilityOptions{CheckPythonGRPC: true})
		ok := runtime != "langgraph" || boolValue(python["ok"])
		statusCode := http.StatusOK
		if !ok {
			statusCode = http.StatusServiceUnavailable
		}
		problem.JSON(w, statusCode, map[string]any{
			"ok":            ok,
			"runtime":       runtime,
			"go_api":        map[string]any{"ok": true, "status": "serving"},
			"python_grpc":   python,
			"capability":    capability,
			"rollout":       aiRolloutRuntimeSettings(),
			"checked_at":    time.Now().UTC().Format(time.RFC3339),
			"ready_policy":  "python_grpc_required_when_runtime_is_langgraph",
			"endpoint_note": "AI_AGENT_GRPC_ADDR overrides the default local Python gRPC target.",
		})
	}
}

func checkPythonGRPC(parent context.Context, target string) (bool, string, string) {
	ctx, cancel := context.WithTimeout(parent, 1500*time.Millisecond)
	defer cancel()
	opts, _, err := aiAgentGRPCDialOptions(target)
	if err != nil {
		return false, "security_not_configured", sanitizeReadinessError(err)
	}
	conn, err := connectAgent(ctx, target, opts)
	if err != nil {
		return false, "unavailable", sanitizeReadinessError(err)
	}
	defer conn.Close()
	resp, err := healthpb.NewHealthClient(conn).Check(aiAgentOutgoingContext(ctx), &healthpb.HealthCheckRequest{})
	if err != nil {
		return false, "health_check_failed", sanitizeReadinessError(err)
	}
	if resp.GetStatus() == healthpb.HealthCheckResponse_SERVING {
		return true, "serving", ""
	}
	return false, strings.ToLower(resp.GetStatus().String()), ""
}

func aiAgentGRPCSecurityReadiness(target string) map[string]any {
	_, security, err := aiAgentGRPCDialOptions(target)
	status := "secure"
	if security.InsecureDev {
		status = "degraded"
	}
	if err != nil {
		status = "unavailable"
	}
	out := map[string]any{
		"status":        status,
		"mode":          security.Mode,
		"insecure_dev":  security.InsecureDev,
		"service_token": security.ServiceToken,
	}
	if err != nil {
		out["error"] = sanitizeReadinessError(err)
	}
	return out
}

func firstNonEmptyEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}
	return fallback
}

func sanitizeReadinessError(err error) string {
	if err == nil {
		return ""
	}
	text := err.Error()
	if len(text) > 240 {
		text = text[:240]
	}
	return text
}

func boolValue(value any) bool {
	v, ok := value.(bool)
	return ok && v
}
