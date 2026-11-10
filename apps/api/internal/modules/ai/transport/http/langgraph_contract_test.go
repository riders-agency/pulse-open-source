package http

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
)

func TestLangGraphGRPCToSSEContract(t *testing.T) {
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	runID := "contract-run"
	threadID := "pulse:" + projectID.String() + ":" + sessionID.String()
	server := &contractAgentRuntimeServer{events: []*aiagentv1.AgentEvent{
		contractEvent(1, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_RUN_STARTED, "started"),
		contractEvent(2, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_STATUS, "planning"),
		contractEventWithCard(3, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_CARD_CREATED, map[string]any{"card_type": "info", "title": "created"}),
		contractEventWithCard(4, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_CARD_UPDATED, map[string]any{"card_type": "info", "title": "updated"}),
		contractEventWithCard(5, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_CARD_PATCHED, map[string]any{"card_type": "info", "title": "patched"}),
		contractEventWithTool(6, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_PROPOSED, map[string]any{"tool_key": "contract.inspect"}),
		contractEventWithTool(7, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_STARTED, map[string]any{"tool_key": "contract.inspect"}),
		contractEventWithTool(8, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_TOOL_RESULT, map[string]any{"tool_key": "contract.inspect", "status": "succeeded"}),
		contractEventWithInterrupt(9, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_APPROVAL_REQUIRED, "approval", map[string]any{"interrupt_id": "i-1"}),
		contractEventWithInterrupt(10, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_INTERRUPT, "interrupt", map[string]any{"reason": "wait"}),
		contractDeltaEvent(11, runID, threadID, sessionID.String(), "delta"),
		contractEvent(12, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_MESSAGE_FINAL, "final answer"),
		contractEventWithError(13, runID, threadID, sessionID.String(), "runtime warning", map[string]any{"code": "contract_warning"}),
		contractEventWithDebug(14, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_MEMORY_UPDATED, "memory", map[string]any{"memory_class": "working"}),
		contractEvent(15, runID, threadID, sessionID.String(), aiagentv1.AgentEventType(99), "future event"),
		contractEvent(16, runID, threadID, sessionID.String(), aiagentv1.AgentEventType_AGENT_EVENT_TYPE_DONE, "done"),
	}}
	addr, stop := serveContractRuntime(t, server)
	defer stop()

	t.Setenv("AI_AGENT_GRPC_ADDR", addr)
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")

	recorder := httptest.NewRecorder()
	_, err := streamLangGraphTurn(context.Background(), recorder, recorder, nil, projectID, sessionID, userID, airepo.MessageInput{
		Body:              "  проверь контракт  ",
		LocalDate:         "2026-06-25",
		Timezone:          "Europe/Moscow",
		PageContextJSON:   json.RawMessage(`{"route":"/x","nested":{"items":[1,true,"тест"]}}`),
		ClientContextJSON: json.RawMessage(`{"surface":"dashboard"}`),
		UIStateJSON:       json.RawMessage(`{"panel":"ai","count":2}`),
	})
	if err == nil {
		t.Fatal("expected terminal gRPC error to prevent persistence in this contract test")
	}
	if !strings.Contains(err.Error(), "contract terminal") {
		t.Fatalf("unexpected terminal error: %v", err)
	}

	req := server.request()
	if req == nil {
		t.Fatal("fake Python runtime did not receive StartTurn")
	}
	if got := server.startCount(); got != 1 {
		t.Fatalf("StartTurn call count = %d, want 1", got)
	}
	if req.GetBody() != "проверь контракт" || req.GetLocalDate() != "2026-06-25" || req.GetTimezone() != "Europe/Moscow" {
		t.Fatalf("unexpected request fields: body=%q date=%q timezone=%q", req.GetBody(), req.GetLocalDate(), req.GetTimezone())
	}
	pageContext := req.GetPageContextJson().AsMap()
	if nested, ok := pageContext["nested"].(map[string]any); !ok || len(nested["items"].([]any)) != 3 {
		t.Fatalf("page_context_json struct edge case lost: %#v", pageContext)
	}

	records := parseSSERecords(t, recorder.Body.String())
	wantTypes := []string{
		"STATUS",
		"STATUS",
		"AGENT_EVENT_TYPE_CARD_CREATED",
		"AGENT_EVENT_TYPE_CARD_UPDATED",
		"AGENT_EVENT_TYPE_CARD_PATCHED",
		"AGENT_EVENT_TYPE_TOOL_PROPOSED",
		"AGENT_EVENT_TYPE_TOOL_STARTED",
		"AGENT_EVENT_TYPE_TOOL_RESULT",
		"AGENT_EVENT_TYPE_APPROVAL_REQUIRED",
		"AGENT_EVENT_TYPE_INTERRUPT",
		"MESSAGE_DELTA",
		"ERROR",
		"MEMORY_UPDATED",
		"UNKNOWN_AGENT_EVENT",
	}
	assertSSETypesInOrder(t, records, wantTypes)
	if got := countSSEEvent(records, "memory"); got != 1 {
		t.Fatalf("memory SSE count = %d, want 1", got)
	}
	if got := countSSEEvent(records, "done"); got != 0 {
		t.Fatalf("DONE from Python must not be forwarded directly, got %d done events", got)
	}
	assertMonotonicSSESeq(t, records)
}

func TestLangGraphGRPCCancellationPropagates(t *testing.T) {
	server := &blockingAgentRuntimeServer{}
	addr, stop := serveContractRuntime(t, server)
	defer stop()

	t.Setenv("AI_AGENT_GRPC_ADDR", addr)
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	recorder := httptest.NewRecorder()
	_, err := streamLangGraphTurn(ctx, recorder, recorder, nil, uuid.New(), uuid.New(), uuid.New(), airepo.MessageInput{
		Body: "cancel",
	})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestLangGraphStreamRequiresV3ToolGateway(t *testing.T) {
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	server := &contractAgentRuntimeServer{}
	addr, stop := serveContractRuntime(t, server)
	defer stop()

	t.Setenv("AI_AGENT_GRPC_ADDR", addr)
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")

	recorder := httptest.NewRecorder()
	_, err := streamLangGraphTurn(context.Background(), recorder, recorder, nil, projectID, sessionID, userID, airepo.MessageInput{
		Body: "покажи задачи",
	})
	if err == nil || !strings.Contains(err.Error(), "tool_gateway_graph_loop") {
		t.Fatalf("missing graph loop should stop before runtime start: %v", err)
	}
}

func TestAIStreamErrorMessageKeepsActionableCheckpointFailure(t *testing.T) {
	message := aiStreamErrorMessage(errors.New("langgraph_checkpoint_error: langgraph checkpoint metadata failed: CheckViolation"))
	if !strings.Contains(message, "checkpoint") {
		t.Fatalf("checkpoint failure should stay actionable, got %q", message)
	}
	if message == "Не удалось выполнить ход AI-сотрудника." {
		t.Fatal("checkpoint failure was reduced to generic stream error")
	}
}

func TestLangGraphThreadStateContractSanitizesAndScopes(t *testing.T) {
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	server := &contractAgentRuntimeServer{threadState: &aiagentv1.ThreadStateResponse{
		ThreadId: "pulse:" + projectID.String() + ":" + sessionID.String(),
		RunId:    "run-1",
		StateJson: mustStruct(map[string]any{
			"ready":                 true,
			"runtime":               "langgraph",
			"current_step":          "await_approval",
			"active_mode":           "approval",
			"checkpoint_found":      true,
			"dialogue_state":        map[string]any{"schema": "riders-pulse-dialogue-state-v1"},
			"active_tool_call_id":   "",
			"pending_interrupt":     true,
			"raw_prompt":            "must not leak",
			"provider_payload_json": map[string]any{"secret": "must not leak"},
		}),
		Cards: []*structpb.Struct{mustStruct(map[string]any{
			"card_id":    "card-1",
			"card_type":  "task_create_proposal",
			"revision":   2,
			"status":     "needs_approval",
			"title":      "Create task",
			"debug_json": map[string]any{"secret": "must not leak"},
		})},
	}}
	addr, stop := serveContractRuntime(t, server)
	defer stop()

	t.Setenv("AI_AGENT_GRPC_ADDR", addr)
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")

	state, err := loadLangGraphThreadState(context.Background(), map[string]any{
		"company_id":               "44444444-4444-4444-4444-444444444444",
		"status":                   "waiting_approval",
		"session_revision":         7,
		"active_tool_call_id":      "55555555-5555-5555-5555-555555555555",
		"active_patch_proposal_id": "66666666-6666-6666-6666-666666666666",
		"langgraph_thread_id":      "pulse:" + projectID.String() + ":" + sessionID.String(),
	}, projectID, sessionID, userID)
	if err != nil {
		t.Fatalf("load thread state: %v", err)
	}
	if state["status"] != "ok" || state["current_step"] != "await_approval" || state["thread_revision"] != 7 {
		t.Fatalf("unexpected sanitized state: %#v", state)
	}
	if dialogue, ok := state["dialogue_state"].(map[string]any); !ok || dialogue["schema"] != "riders-pulse-dialogue-state-v1" {
		t.Fatalf("dialogue_state missing from sanitized state: %#v", state)
	}
	if state["active_tool_call_id"] != "55555555-5555-5555-5555-555555555555" {
		t.Fatalf("active tool call fallback lost: %#v", state)
	}
	if _, ok := state["raw_prompt"]; ok {
		t.Fatalf("raw prompt leaked: %#v", state)
	}
	cards, ok := state["cards"].([]map[string]any)
	if !ok || len(cards) != 1 {
		t.Fatalf("cards not sanitized: %#v", state["cards"])
	}
	if _, ok := cards[0]["debug_json"]; ok {
		t.Fatalf("card debug_json leaked: %#v", cards[0])
	}
	req := server.threadStateRequest()
	if req == nil {
		t.Fatal("fake Python runtime did not receive GetThreadState")
	}
	if req.GetAuthContext().GetCompanyId() != "44444444-4444-4444-4444-444444444444" ||
		req.GetProjectId() != projectID.String() ||
		req.GetSessionId() != sessionID.String() {
		t.Fatalf("bad thread-state scope: %#v", req)
	}
}

func TestLangGraphCancelBridgeScopesRequest(t *testing.T) {
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	companyID := "44444444-4444-4444-4444-444444444444"
	server := &contractAgentRuntimeServer{cancelResponse: &aiagentv1.CancelRunResponse{
		Ok:     true,
		Status: "canceled",
	}}
	addr, stop := serveContractRuntime(t, server)
	defer stop()

	t.Setenv("AI_AGENT_GRPC_ADDR", addr)
	t.Setenv("AI_AGENT_GRPC_INSECURE_DEV", "true")
	t.Setenv("AI_AGENT_SERVICE_TOKEN", "")
	t.Setenv("AI_EMPLOYEE_RUNTIME", "langgraph")
	t.Setenv("PULSE_AI_RESUME_READY", "true")
	t.Setenv("PULSE_AI_CHECKPOINTER_READY", "true")

	handled, err := cancelLangGraphSession(context.Background(), airepo.SessionRuntimeContext{
		CompanyID:         companyID,
		ProjectID:         projectID,
		SessionID:         sessionID,
		AgentRuntime:      "langgraph",
		LangGraphThreadID: "thread-1",
		Status:            "running",
	}, userID)
	if err != nil {
		t.Fatalf("cancel langgraph session: %v", err)
	}
	if !handled {
		t.Fatal("expected LangGraph cancel to be handled")
	}
	req := server.cancelRequest()
	if req == nil {
		t.Fatal("fake Python runtime did not receive CancelRun")
	}
	if req.GetAuthContext().GetCompanyId() != companyID ||
		req.GetAuthContext().GetProjectId() != projectID.String() ||
		req.GetAuthContext().GetUserId() != userID.String() ||
		req.GetProjectId() != projectID.String() ||
		req.GetSessionId() != sessionID.String() ||
		req.GetThreadId() != "thread-1" {
		t.Fatalf("bad cancel scope: %#v", req)
	}
}

type contractAgentRuntimeServer struct {
	aiagentv1.UnimplementedAgentRuntimeServiceServer
	mu              sync.Mutex
	seen            *aiagentv1.TurnRequest
	stateSeen       *aiagentv1.GetThreadStateRequest
	cancelSeen      *aiagentv1.CancelRunRequest
	events          []*aiagentv1.AgentEvent
	threadState     *aiagentv1.ThreadStateResponse
	cancelResponse  *aiagentv1.CancelRunResponse
	started         int
	threadStateSeen int
}

func (s *contractAgentRuntimeServer) StartTurn(req *aiagentv1.TurnRequest, stream grpc.ServerStreamingServer[aiagentv1.AgentEvent]) error {
	s.mu.Lock()
	s.started++
	s.seen = req
	s.mu.Unlock()
	for _, event := range s.events {
		if err := stream.Send(event); err != nil {
			return err
		}
	}
	return status.Error(codes.Canceled, "contract terminal")
}

func (s *contractAgentRuntimeServer) request() *aiagentv1.TurnRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen
}

func (s *contractAgentRuntimeServer) startCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

func (s *contractAgentRuntimeServer) GetThreadState(ctx context.Context, req *aiagentv1.GetThreadStateRequest) (*aiagentv1.ThreadStateResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.threadStateSeen++
	s.stateSeen = req
	out := s.threadState
	s.mu.Unlock()
	if out == nil {
		return &aiagentv1.ThreadStateResponse{ThreadId: req.GetThreadId(), StateJson: mustStruct(map[string]any{"ready": true})}, nil
	}
	return out, nil
}

func (s *contractAgentRuntimeServer) CancelRun(ctx context.Context, req *aiagentv1.CancelRunRequest) (*aiagentv1.CancelRunResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cancelSeen = req
	out := s.cancelResponse
	s.mu.Unlock()
	if out == nil {
		return &aiagentv1.CancelRunResponse{Ok: true, Status: "canceled"}, nil
	}
	return out, nil
}

func (s *contractAgentRuntimeServer) threadStateRequest() *aiagentv1.GetThreadStateRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stateSeen
}

func (s *contractAgentRuntimeServer) cancelRequest() *aiagentv1.CancelRunRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancelSeen
}

type blockingAgentRuntimeServer struct {
	aiagentv1.UnimplementedAgentRuntimeServiceServer
}

func (s *blockingAgentRuntimeServer) StartTurn(_ *aiagentv1.TurnRequest, stream grpc.ServerStreamingServer[aiagentv1.AgentEvent]) error {
	<-stream.Context().Done()
	return stream.Context().Err()
}

func serveContractRuntime(t *testing.T, runtime aiagentv1.AgentRuntimeServiceServer) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	aiagentv1.RegisterAgentRuntimeServiceServer(server, runtime)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("contract gRPC server: %v", err)
		}
	}()
	return listener.Addr().String(), func() {
		server.Stop()
		_ = listener.Close()
	}
}

func contractEvent(seq int64, runID, threadID, sessionID string, eventType aiagentv1.AgentEventType, message string) *aiagentv1.AgentEvent {
	return &aiagentv1.AgentEvent{
		Seq:       seq,
		RunId:     runID,
		ThreadId:  threadID,
		SessionId: sessionID,
		EventType: eventType,
		Message:   message,
	}
}

func contractDeltaEvent(seq int64, runID, threadID, sessionID, delta string) *aiagentv1.AgentEvent {
	event := contractEvent(seq, runID, threadID, sessionID, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_MESSAGE_DELTA, "")
	event.BodyDelta = delta
	return event
}

func contractEventWithCard(seq int64, runID, threadID, sessionID string, eventType aiagentv1.AgentEventType, card map[string]any) *aiagentv1.AgentEvent {
	event := contractEvent(seq, runID, threadID, sessionID, eventType, "card")
	event.Card = mustStruct(card)
	return event
}

func contractEventWithTool(seq int64, runID, threadID, sessionID string, eventType aiagentv1.AgentEventType, tool map[string]any) *aiagentv1.AgentEvent {
	event := contractEvent(seq, runID, threadID, sessionID, eventType, "tool")
	event.ToolCall = mustStruct(tool)
	return event
}

func contractEventWithInterrupt(seq int64, runID, threadID, sessionID string, eventType aiagentv1.AgentEventType, message string, interrupt map[string]any) *aiagentv1.AgentEvent {
	event := contractEvent(seq, runID, threadID, sessionID, eventType, message)
	event.Interrupt = mustStruct(interrupt)
	return event
}

func contractEventWithError(seq int64, runID, threadID, sessionID, message string, payload map[string]any) *aiagentv1.AgentEvent {
	event := contractEvent(seq, runID, threadID, sessionID, aiagentv1.AgentEventType_AGENT_EVENT_TYPE_ERROR, message)
	event.Error = mustStruct(payload)
	return event
}

func contractEventWithDebug(seq int64, runID, threadID, sessionID string, eventType aiagentv1.AgentEventType, message string, debug map[string]any) *aiagentv1.AgentEvent {
	event := contractEvent(seq, runID, threadID, sessionID, eventType, message)
	event.Debug = mustStruct(debug)
	return event
}

func mustStruct(value map[string]any) *structpb.Struct {
	out, err := structpb.NewStruct(value)
	if err != nil {
		panic(err)
	}
	return out
}

type sseRecord struct {
	Event string
	Data  map[string]any
}

func parseSSERecords(t *testing.T, raw string) []sseRecord {
	t.Helper()
	blocks := strings.Split(strings.TrimSpace(raw), "\n\n")
	records := make([]sseRecord, 0, len(blocks))
	for _, block := range blocks {
		var record sseRecord
		for _, line := range strings.Split(block, "\n") {
			if after, ok := strings.CutPrefix(line, "event: "); ok {
				record.Event = after
			}
			if after, ok := strings.CutPrefix(line, "data: "); ok {
				if err := json.Unmarshal([]byte(after), &record.Data); err != nil {
					t.Fatalf("decode SSE data %q: %v", after, err)
				}
			}
		}
		if record.Event != "" {
			records = append(records, record)
		}
	}
	return records
}

func assertSSETypesInOrder(t *testing.T, records []sseRecord, want []string) {
	t.Helper()
	got := make([]string, 0, len(records))
	for _, record := range records {
		if value, ok := record.Data["type"].(string); ok {
			got = append(got, value)
		}
	}
	cursor := 0
	for _, wantType := range want {
		found := false
		for cursor < len(got) {
			if got[cursor] == wantType {
				found = true
				cursor++
				break
			}
			cursor++
		}
		if !found {
			t.Fatalf("SSE type %q not found in order; got %#v", wantType, got)
		}
	}
}

func countSSEEvent(records []sseRecord, event string) int {
	count := 0
	for _, record := range records {
		if record.Event == event {
			count++
		}
	}
	return count
}

func assertMonotonicSSESeq(t *testing.T, records []sseRecord) {
	t.Helper()
	var last float64
	for _, record := range records {
		seq, ok := record.Data["seq"].(float64)
		if !ok {
			continue
		}
		if seq < last {
			t.Fatalf("SSE seq moved backwards: prev=%v current=%v record=%#v", last, seq, record)
		}
		last = seq
	}
}
