package office

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"net"
	"os"
	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	"strings"
	"time"
)

func prepareWithAgent(ctx context.Context, identity identity, action string, mail map[string]any) (map[string]any, map[string]any, error) {
	target := strings.TrimSpace(os.Getenv("AI_AGENT_GRPC_ADDR"))
	if target == "" {
		target = strings.TrimSpace(os.Getenv("PULSE_AI_AGENT_GRPC_ADDR"))
	}
	if target == "" {
		target = "127.0.0.1:9097"
	}
	if strings.HasPrefix(target, ":") {
		target = "127.0.0.1" + target
	}
	host, _, err := net.SplitHostPort(target)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		return nil, nil, fmt.Errorf("workflow agent requires loopback gRPC")
	}
	token := os.Getenv("AI_AGENT_SERVICE_TOKEN")
	if token == "" {
		token = os.Getenv("AI_AGENT_GRPC_SERVICE_TOKEN")
	}
	if token == "" {
		return nil, nil, fmt.Errorf("workflow agent service token unavailable")
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-pulse-service-token", token, "x-pulse-service", "pulse-worker")
	input, err := structpb.NewStruct(mail)
	if err != nil {
		return nil, nil, err
	}
	result, err := aiagentv1.NewWorkflowAgentServiceClient(conn).Prepare(ctx, &aiagentv1.WorkflowAgentRequest{RequestId: identity.RunID + ":" + action, CompanyId: identity.CompanyID, ProjectId: identity.ProjectID, UserId: identity.UserID, Action: action, Mail: input})
	if err != nil {
		return nil, nil, fmt.Errorf("ИИ-сотрудник временно недоступен: %w", err)
	}
	var usage map[string]any
	raw, _ := protojson.Marshal(result.Usage)
	_ = json.Unmarshal(raw, &usage)
	return result.Output.AsMap(), map[string]any{"provider": "langgraph", "usage": usage, "executed_nodes": result.ExecutedNodes}, nil
}
