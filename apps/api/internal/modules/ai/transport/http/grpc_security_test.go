package http

import "testing"

func TestAIAgentGRPCTokenModeRejectsRemotePlaintextTarget(t *testing.T) {
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN", "")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "service-token")

	if _, security, err := aiAgentGRPCDialOptions("10.0.0.10:9097"); err == nil {
		t.Fatalf("expected remote token target to be rejected; security=%#v", security)
	}
}

func TestAIAgentGRPCTokenModeAllowsLoopbackAndUnixTargets(t *testing.T) {
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_INSECURE_DEV", "")
	t.Setenv("AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN", "")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "service-token")

	for _, target := range []string{"127.0.0.1:9097", "localhost:9097", "unix:/run/riders-pulse/ai-agent.sock"} {
		if _, _, err := aiAgentGRPCDialOptions(target); err != nil {
			t.Fatalf("expected %s to be allowed: %v", target, err)
		}
	}
}
