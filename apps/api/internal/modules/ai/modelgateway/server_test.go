package modelgateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
)

func TestFakeProviderChatReturnsStructuredResponse(t *testing.T) {
	srv := newTestServer(t, Config{ProviderKey: "fake", DefaultModelID: "fake-model"})
	resp, err := srv.Chat(context.Background(), &aiagentv1.ModelChatRequest{
		RequestId: "req-1",
		Purpose:   aiagentv1.ModelPurpose_MODEL_PURPOSE_INTENT_ROUTER,
		Messages: []*aiagentv1.ModelMessage{
			{Role: "user", Content: "classify this"},
		},
		ResponseFormatJson: mustTestStruct(t, map[string]any{"type": "json_object"}),
		MaxTokens:          100,
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if resp.GetUsage().GetProviderKey() != "fake" || resp.GetUsage().GetModelId() != "fake-model" {
		t.Fatalf("unexpected usage: %+v", resp.GetUsage())
	}
	if got := resp.GetResponseJson().AsMap()["provider"]; got != "fake" {
		t.Fatalf("provider = %v, want fake", got)
	}
	if resp.GetContent() == "" {
		t.Fatalf("expected content")
	}
}

func TestOpenAICompatibleProviderChatRedactsStructuredResponse(t *testing.T) {
	var sawAuthorization bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		sawAuthorization = r.Header.Get("Authorization") == "Bearer test-secret"
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"{\"answer\":\"ok\",\"token\":\"leak\"}"}}],
			"usage":{"prompt_tokens":11,"completion_tokens":7}
		}`))
	}))
	defer upstream.Close()

	srv := newTestServer(t, Config{
		ProviderKey:    "openai-compatible",
		BaseURL:        upstream.URL,
		APIKey:         "test-secret",
		DefaultModelID: "model-a",
		AllowedModels:  map[string]struct{}{"model-a": {}},
	})
	resp, err := srv.Chat(context.Background(), &aiagentv1.ModelChatRequest{
		Purpose: aiagentv1.ModelPurpose_MODEL_PURPOSE_FINAL_RESPONSE,
		Messages: []*aiagentv1.ModelMessage{
			{Role: "user", Content: "answer"},
		},
		ResponseFormatJson: mustTestStruct(t, map[string]any{"type": "json_object"}),
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if !sawAuthorization {
		t.Fatalf("upstream did not receive provider authorization")
	}
	response := resp.GetResponseJson().AsMap()
	if response["token"] != "[redacted]" {
		t.Fatalf("token was not redacted: %#v", response)
	}
	if usage := resp.GetUsage(); usage.GetInputTokens() != 11 || usage.GetOutputTokens() != 7 {
		t.Fatalf("usage = %+v", usage)
	}
	metadata := resp.GetUsage().GetMetadataJson().AsMap()
	if raw, _ := json.Marshal(metadata); string(raw) == "" || string(raw) == "null" {
		t.Fatalf("expected metadata")
	}
}

func TestChatUsesRequestScopedResolverWhenEnvGatewayDisabled(t *testing.T) {
	var sawAuthorization bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuthorization = r.Header.Get("Authorization") == "Bearer resolver-secret"
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"{\"intent_key\":\"client.search\"}"}}],
			"usage":{"prompt_tokens":5,"completion_tokens":3}
		}`))
	}))
	defer upstream.Close()

	srv, err := NewServer(
		Config{},
		http.DefaultClient,
		WithConfigResolver(staticConfigResolver{cfg: Config{
			ProviderKey:    "openai-compatible",
			BaseURL:        upstream.URL,
			APIKey:         "resolver-secret",
			DefaultModelID: "resolver-model",
		}}),
	)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	resp, err := srv.Chat(context.Background(), &aiagentv1.ModelChatRequest{
		ProjectId: "11111111-1111-1111-1111-111111111111",
		SessionId: "22222222-2222-2222-2222-222222222222",
		Purpose:   aiagentv1.ModelPurpose_MODEL_PURPOSE_INTENT_ROUTER,
		Messages: []*aiagentv1.ModelMessage{
			{Role: "user", Content: "выведи клиентов"},
		},
		ResponseFormatJson: mustTestStruct(t, map[string]any{"type": "json_object"}),
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if !sawAuthorization {
		t.Fatalf("resolver provider authorization was not used")
	}
	if resp.GetUsage().GetProviderKey() != "openai-compatible" || resp.GetUsage().GetModelId() != "resolver-model" {
		t.Fatalf("usage = %+v", resp.GetUsage())
	}
}

func TestChatRejectsDisallowedModel(t *testing.T) {
	srv := newTestServer(t, Config{
		ProviderKey:    "fake",
		DefaultModelID: "model-a",
		AllowedModels:  map[string]struct{}{"model-b": {}},
	})
	_, err := srv.Chat(context.Background(), &aiagentv1.ModelChatRequest{
		Purpose: aiagentv1.ModelPurpose_MODEL_PURPOSE_PLANNER,
		Messages: []*aiagentv1.ModelMessage{
			{Role: "user", Content: "plan"},
		},
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("status = %v, want PermissionDenied, err=%v", status.Code(err), err)
	}
}

func TestStreamChatSendsDeltaThenDone(t *testing.T) {
	srv := newTestServer(t, Config{ProviderKey: "fake", DefaultModelID: "fake-model"})
	stream := &captureModelStream{ctx: context.Background()}
	err := srv.StreamChat(&aiagentv1.ModelChatRequest{
		Purpose: aiagentv1.ModelPurpose_MODEL_PURPOSE_FINAL_RESPONSE,
		Messages: []*aiagentv1.ModelMessage{
			{Role: "user", Content: "hello"},
		},
	}, stream)
	if err != nil {
		t.Fatalf("StreamChat returned error: %v", err)
	}
	if len(stream.events) != 2 || stream.events[0].GetDelta() == "" || !stream.events[1].GetDone() {
		t.Fatalf("unexpected stream events: %+v", stream.events)
	}
}

func TestFakeProviderEmbed(t *testing.T) {
	srv := newTestServer(t, Config{ProviderKey: "fake", DefaultModelID: "fake-model"})
	resp, err := srv.Embed(context.Background(), &aiagentv1.EmbedRequest{
		RequestId: "embed-1",
		Input:     []string{"one", "two"},
	})
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	if len(resp.GetEmbeddings()) != 2 || len(resp.GetEmbeddings()[0].GetValues()) != 3 {
		t.Fatalf("unexpected embeddings: %+v", resp.GetEmbeddings())
	}
}

func newTestServer(t *testing.T, cfg Config) *Server {
	t.Helper()
	if cfg.Timeout == 0 {
		cfg.Timeout = time.Second
	}
	srv, err := NewServer(cfg, http.DefaultClient)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return srv
}

type captureModelStream struct {
	grpc.ServerStream
	ctx    context.Context
	events []*aiagentv1.ModelChatEvent
}

func (s *captureModelStream) Context() context.Context {
	return s.ctx
}

func (s *captureModelStream) Send(event *aiagentv1.ModelChatEvent) error {
	s.events = append(s.events, event)
	return nil
}

type staticConfigResolver struct {
	cfg Config
}

func (r staticConfigResolver) ResolveModelGatewayConfig(context.Context, uuid.UUID, uuid.UUID) (Config, error) {
	return r.cfg, nil
}

func mustTestStruct(t *testing.T, value map[string]any) *structpb.Struct {
	t.Helper()
	out, err := structpb.NewStruct(value)
	if err != nil {
		t.Fatalf("NewStruct: %v", err)
	}
	return out
}
