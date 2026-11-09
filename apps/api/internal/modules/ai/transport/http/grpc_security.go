package http

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type aiAgentGRPCSecurity struct {
	Target       string
	Mode         string
	InsecureDev  bool
	ServiceToken bool
}

func aiAgentGRPCTarget() string {
	target := strings.TrimSpace(os.Getenv("AI_AGENT_GRPC_ADDR"))
	if target == "" {
		target = strings.TrimSpace(os.Getenv("PULSE_AI_AGENT_GRPC_ADDR"))
	}
	if target == "" {
		target = strings.TrimSpace(os.Getenv("LANGGRAPH_AGENT_GRPC_ADDR"))
	}
	if target == "" {
		target = "127.0.0.1:9097"
	}
	if strings.HasPrefix(target, ":") {
		target = "127.0.0.1" + target
	}
	return target
}

func aiAgentGRPCDialOptions(target string) ([]grpc.DialOption, aiAgentGRPCSecurity, error) {
	security := aiAgentGRPCSecurity{Target: target, Mode: "internal_token"}
	token := aiAgentServiceToken()
	security.ServiceToken = token != ""
	insecureDev := boolEnv("AI_AGENT_GRPC_INSECURE_DEV") || boolEnv("AI_AGENT_INSECURE_DEV")
	security.InsecureDev = insecureDev
	if insecureDev {
		if !isLoopbackTarget(target) && !boolEnv("AI_AGENT_GRPC_INSECURE_ALLOW_NONLOCAL") {
			return nil, security, fmt.Errorf("AI agent insecure dev gRPC target must be loopback: %s", target)
		}
		security.Mode = "insecure_dev"
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, security, nil
	}
	if token == "" {
		return nil, security, fmt.Errorf("AI agent gRPC requires AI_AGENT_SERVICE_TOKEN or explicit AI_AGENT_GRPC_INSECURE_DEV=true")
	}
	if !isLoopbackTarget(target) && !boolEnv("AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN") {
		return nil, security, fmt.Errorf("AI agent token gRPC target must be loopback or unix socket unless AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN=true: %s", target)
	}
	return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, security, nil
}

func aiAgentOutgoingContext(ctx context.Context) context.Context {
	token := aiAgentServiceToken()
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx,
		"x-pulse-service", "pulse-api",
		"x-pulse-service-token", token,
	)
}

func aiAgentServiceToken() string {
	if token := strings.TrimSpace(os.Getenv("AI_AGENT_SERVICE_TOKEN")); token != "" {
		return token
	}
	return strings.TrimSpace(os.Getenv("AI_AGENT_GRPC_SERVICE_TOKEN"))
}

func boolEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func isLoopbackTarget(target string) bool {
	host := target
	if strings.HasPrefix(host, "unix:") {
		return true
	}
	if strings.Contains(host, "://") {
		return false
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
