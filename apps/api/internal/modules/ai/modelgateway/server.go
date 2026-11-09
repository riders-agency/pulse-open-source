package modelgateway

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"riders.agency/pulse/api/internal/platform/demomode"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
)

type Server struct {
	aiagentv1.UnimplementedModelGatewayServiceServer
	config   Config
	provider Provider
	client   *http.Client
	resolver ConfigResolver
	clock    func() time.Time
}

type ConfigResolver interface {
	ResolveModelGatewayConfig(ctx context.Context, projectID, sessionID uuid.UUID) (Config, error)
}

type Option func(*Server)

func WithConfigResolver(resolver ConfigResolver) Option {
	return func(s *Server) {
		s.resolver = resolver
	}
}

func NewServerFromEnv(environment string, client *http.Client, opts ...Option) (*Server, error) {
	cfg := LoadConfigFromEnv(environment)
	return NewServer(cfg, client, opts...)
}

func NewServer(cfg Config, client *http.Client, opts ...Option) (*Server, error) {
	cfg = cfg.normalize()
	provider, err := providerForConfig(cfg, client)
	if err != nil {
		return nil, err
	}
	server := &Server{config: cfg, provider: provider, client: client, clock: time.Now}
	for _, opt := range opts {
		opt(server)
	}
	return server, nil
}

func (s *Server) Chat(ctx context.Context, req *aiagentv1.ModelChatRequest) (*aiagentv1.ModelChatResponse, error) {
	if err := demomode.ReserveModelCall(req.GetUserId()); err != nil {
		return nil, status.Error(codes.ResourceExhausted, "demo_model_budget_exhausted")
	}
	start := s.clock()
	cfg, provider, err := s.providerForRequest(ctx, req.GetProjectId(), req.GetSessionId())
	if err != nil {
		return nil, err
	}
	call, err := s.chatCall(req, cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, call.Timeout)
	defer cancel()
	result, err := provider.Chat(ctx, call)
	if err != nil {
		return nil, grpcError(err)
	}
	return &aiagentv1.ModelChatResponse{
		Content:      result.Content,
		ResponseJson: mustStruct(result.ResponseJSON),
		Usage:        usage(cfg.ProviderKey, call.ModelID, result, time.Since(start)),
	}, nil
}

func (s *Server) StreamChat(req *aiagentv1.ModelChatRequest, stream grpc.ServerStreamingServer[aiagentv1.ModelChatEvent]) error {
	resp, err := s.Chat(stream.Context(), req)
	if err != nil {
		return err
	}
	if resp.GetContent() != "" {
		if err := stream.Send(&aiagentv1.ModelChatEvent{Delta: resp.GetContent()}); err != nil {
			return err
		}
	}
	return stream.Send(&aiagentv1.ModelChatEvent{Done: true, Usage: resp.GetUsage()})
}

func (s *Server) Embed(ctx context.Context, req *aiagentv1.EmbedRequest) (*aiagentv1.EmbedResponse, error) {
	if err := demomode.ReserveModelCall(req.GetUserId()); err != nil {
		return nil, status.Error(codes.ResourceExhausted, "demo_model_budget_exhausted")
	}
	start := s.clock()
	cfg, provider, err := s.providerForRequest(ctx, req.GetProjectId(), "")
	if err != nil {
		return nil, err
	}
	call, err := s.embedCall(req, cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, call.Timeout)
	defer cancel()
	result, err := provider.Embed(ctx, call)
	if err != nil {
		return nil, grpcError(err)
	}
	out := make([]*structpb.ListValue, 0, len(result.Embeddings))
	for _, embedding := range result.Embeddings {
		values := make([]any, 0, len(embedding))
		for _, number := range embedding {
			values = append(values, number)
		}
		out = append(out, mustList(values))
	}
	return &aiagentv1.EmbedResponse{
		Embeddings: out,
		Usage: &aiagentv1.Usage{
			ProviderKey:  cfg.ProviderKey,
			ModelId:      call.ModelID,
			InputTokens:  result.InputTokens,
			OutputTokens: 0,
			LatencyMs:    time.Since(start).Milliseconds(),
			MetadataJson: mustStruct(result.Metadata),
		},
	}, nil
}

func (s *Server) chatCall(req *aiagentv1.ModelChatRequest, cfg Config) (ChatCall, error) {
	if req.GetPurpose() == aiagentv1.ModelPurpose_MODEL_PURPOSE_UNSPECIFIED {
		return ChatCall{}, invalidArgument("purpose is required")
	}
	if len(req.GetMessages()) == 0 {
		return ChatCall{}, invalidArgument("messages are required")
	}
	modelID := cfg.DefaultModelID
	if err := validateModel(cfg, modelID); err != nil {
		return ChatCall{}, err
	}
	return ChatCall{
		RequestID:      req.GetRequestId(),
		CompanyID:      req.GetCompanyId(),
		ProjectID:      req.GetProjectId(),
		UserID:         req.GetUserId(),
		SessionID:      req.GetSessionId(),
		ThreadID:       req.GetThreadId(),
		Purpose:        purposeName(req.GetPurpose()),
		ModelID:        modelID,
		Messages:       protoMessages(req.GetMessages()),
		ResponseFormat: structMap(req.GetResponseFormatJson()),
		ToolSchemas:    structMap(req.GetToolSchemasJson()),
		MaxTokens:      maxTokens(cfg, req.GetMaxTokens()),
		Temperature:    req.GetTemperature(),
		Timeout:        timeout(cfg, req.GetTimeoutMs()),
	}, nil
}

func (s *Server) embedCall(req *aiagentv1.EmbedRequest, cfg Config) (EmbedCall, error) {
	if len(req.GetInput()) == 0 {
		return EmbedCall{}, invalidArgument("input is required")
	}
	modelID := cfg.EmbeddingModelID
	if err := validateModel(cfg, modelID); err != nil {
		return EmbedCall{}, err
	}
	return EmbedCall{
		RequestID: req.GetRequestId(),
		CompanyID: req.GetCompanyId(),
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
		ModelID:   modelID,
		Input:     req.GetInput(),
		Timeout:   cfg.Timeout,
	}, nil
}

func (s *Server) providerForRequest(ctx context.Context, projectIDText, sessionIDText string) (Config, Provider, error) {
	if s.config.ProviderKey != "" {
		return s.config, s.provider, nil
	}
	if s.resolver == nil {
		return s.config, s.provider, nil
	}
	projectID, err := uuid.Parse(strings.TrimSpace(projectIDText))
	if err != nil || projectID == uuid.Nil {
		return s.config, s.provider, nil
	}
	sessionID := uuid.Nil
	if strings.TrimSpace(sessionIDText) != "" {
		sessionID, _ = uuid.Parse(strings.TrimSpace(sessionIDText))
	}
	cfg, err := s.resolver.ResolveModelGatewayConfig(ctx, projectID, sessionID)
	if err != nil {
		return Config{}, nil, err
	}
	cfg = cfg.normalize()
	provider, err := providerForConfig(cfg, s.client)
	if err != nil {
		return Config{}, nil, err
	}
	return cfg, provider, nil
}

func validateModel(cfg Config, modelID string) error {
	if strings.TrimSpace(modelID) == "" {
		return status.Error(codes.FailedPrecondition, "model_not_configured")
	}
	if len(cfg.AllowedModels) == 0 {
		return nil
	}
	if _, ok := cfg.AllowedModels[modelID]; !ok {
		return status.Error(codes.PermissionDenied, "model_not_allowed")
	}
	return nil
}

func maxTokens(cfg Config, requested int32) int {
	if requested <= 0 {
		return cfg.DefaultMaxTokens
	}
	if int(requested) > cfg.HardMaxTokens {
		return cfg.HardMaxTokens
	}
	return int(requested)
}

func timeout(cfg Config, requested int32) time.Duration {
	if requested <= 0 {
		return cfg.Timeout
	}
	timeout := time.Duration(requested) * time.Millisecond
	if timeout > cfg.Timeout {
		return cfg.Timeout
	}
	return timeout
}

func providerForConfig(cfg Config, client *http.Client) (Provider, error) {
	switch cfg.ProviderKey {
	case "fake":
		return FakeProvider{}, nil
	case "apinet", "openai", "openai-compatible", "openai_compatible":
		if cfg.BaseURL == "" || cfg.APIKey == "" {
			return nil, status.Error(codes.FailedPrecondition, "model_provider_not_configured")
		}
		return OpenAICompatibleProvider{Client: client, Config: cfg}, nil
	case "":
		return disabledProvider{}, nil
	default:
		return nil, status.Error(codes.FailedPrecondition, "model_provider_not_supported")
	}
}

type disabledProvider struct{}

func (disabledProvider) Chat(context.Context, ChatCall) (ChatResult, error) {
	return ChatResult{}, GatewayError{GRPCCode: codes.FailedPrecondition, PublicCode: "model_gateway_disabled"}
}

func (disabledProvider) Embed(context.Context, EmbedCall) (EmbedResult, error) {
	return EmbedResult{}, GatewayError{GRPCCode: codes.FailedPrecondition, PublicCode: "model_gateway_disabled"}
}

func protoMessages(messages []*aiagentv1.ModelMessage) []Message {
	out := make([]Message, 0, len(messages))
	for _, message := range messages {
		out = append(out, Message{
			Role:     strings.TrimSpace(message.GetRole()),
			Content:  message.GetContent(),
			Metadata: structMap(message.GetMetadataJson()),
		})
	}
	return out
}

func purposeName(purpose aiagentv1.ModelPurpose) string {
	return strings.TrimPrefix(purpose.String(), "MODEL_PURPOSE_")
}

func usage(providerKey, modelID string, result ChatResult, latency time.Duration) *aiagentv1.Usage {
	return &aiagentv1.Usage{
		ProviderKey:  providerKey,
		ModelId:      modelID,
		InputTokens:  result.InputTokens,
		OutputTokens: result.OutputTokens,
		LatencyMs:    latency.Milliseconds(),
		MetadataJson: mustStruct(result.Metadata),
	}
}

func structMap(value *structpb.Struct) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value.AsMap()
}

func mustStruct(value map[string]any) *structpb.Struct {
	out, err := structpb.NewStruct(value)
	if err != nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	return out
}

func mustList(value []any) *structpb.ListValue {
	out, err := structpb.NewList(value)
	if err != nil {
		return &structpb.ListValue{Values: []*structpb.Value{}}
	}
	return out
}
