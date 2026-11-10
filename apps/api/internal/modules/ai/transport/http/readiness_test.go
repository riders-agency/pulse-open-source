package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckPythonGRPCServing(t *testing.T) {
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	ok, status, errText := checkPythonGRPC(context.Background(), listener.Addr().String())
	if !ok {
		t.Fatalf("expected serving grpc health, status=%s err=%s", status, errText)
	}
	if status != "serving" {
		t.Fatalf("unexpected status %q", status)
	}
}

func TestRuntimeReadinessDefaultsToLangGraph(t *testing.T) {
	t.Setenv("AI_EMPLOYEE_RUNTIME", "")
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	t.Setenv("AI_AGENT_GRPC_SERVICE_TOKEN", "")

	recorder := httptest.NewRecorder()
	RuntimeReadiness().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/status/ai-runtime", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(recorder.Body.String(), `"runtime":"langgraph"`) {
		t.Fatalf("response does not show langgraph runtime: %s", recorder.Body.String())
	}
}

func TestCheckPythonGRPCRequiresSecurityMode(t *testing.T) {
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	t.Setenv("AI_AGENT_GRPC_SERVICE_TOKEN", "")
	ok, status, errText := checkPythonGRPC(context.Background(), "127.0.0.1:1")
	if ok {
		t.Fatal("expected readiness failure without service token or insecure dev flag")
	}
	if status != "security_not_configured" {
		t.Fatalf("status = %q, want security_not_configured", status)
	}
	if errText == "" {
		t.Fatal("expected explanatory readiness error")
	}
}

func TestAIAgentGRPCSecurityReadinessReportsBoundaryStatus(t *testing.T) {
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	if got := aiAgentGRPCSecurityReadiness("127.0.0.1:9097"); got["status"] != "degraded" {
		t.Fatalf("insecure local status = %#v, want degraded", got)
	}

	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "service-token")
	if got := aiAgentGRPCSecurityReadiness("127.0.0.1:9097"); got["status"] != "secure" {
		t.Fatalf("token local status = %#v, want secure", got)
	}

	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	if got := aiAgentGRPCSecurityReadiness("127.0.0.1:9097"); got["status"] != "unavailable" {
		t.Fatalf("missing security status = %#v, want unavailable", got)
	}
}
