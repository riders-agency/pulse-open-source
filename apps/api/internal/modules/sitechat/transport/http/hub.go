package http

import (
	"context"
	"encoding/json"
	"log/slog"
	stdhttp "net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	sitechat "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/problem"
)

const (
	wsWriteWait        = 8 * time.Second
	wsPongWait         = 35 * time.Second
	wsPingPeriod       = 10 * time.Second
	wsVisitorGoneAfter = 30 * time.Second
)

type WSEnvelope struct {
	ID             string         `json:"id"`
	Type           string         `json:"type"`
	SentAt         time.Time      `json:"sent_at"`
	ProjectID      string         `json:"project_id,omitempty"`
	ChatWidgetID   string         `json:"chat_widget_id,omitempty"`
	VisitorID      string         `json:"visitor_id,omitempty"`
	VisitSessionID string         `json:"visit_session_id,omitempty"`
	ConversationID string         `json:"conversation_id,omitempty"`
	CallSessionID  string         `json:"call_session_id,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
}

type Hub struct {
	S       *sitechatsvc.Commands
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
}

type wsClient struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan WSEnvelope
	role         string
	connectionID string
	scope        sitechat.LiveAssistScope
	lastSeen     time.Time
	closed       chan struct{}
	leftOnce     sync.Once
}

func NewHub(s *sitechatsvc.Commands) *Hub {
	return &Hub{S: s, clients: map[*wsClient]struct{}{}}
}

func PublicWS(hub *Hub) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if hub == nil || hub.S == nil {
			problem.WriteModule(w, stdhttp.StatusServiceUnavailable, "site_chat", "ws_unavailable", "Live Assist недоступен", "WebSocket hub не инициализирован.", "Повторите позже.")
			return
		}
		q := r.URL.Query()
		scope, err := hub.S.ResolvePublicScope(r.Context(), q.Get("public_key"), q.Get("visitor_id"), q.Get("visit_session_id"), q.Get("conversation_id"))
		if err != nil {
			writeSiteChatErr(w, err)
			return
		}
		hub.serve(w, r, "visitor", scope)
	}
}

func OperatorWS(hub *Hub) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if hub == nil || hub.S == nil {
			problem.WriteModule(w, stdhttp.StatusServiceUnavailable, "site_chat", "ws_unavailable", "Live Assist недоступен", "WebSocket hub не инициализирован.", "Повторите позже.")
			return
		}
		projectID, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("project_id")))
		if err != nil {
			problem.WriteModule(w, stdhttp.StatusBadRequest, "site_chat", "bad_project_id", "Некорректный project_id", "Операторский WebSocket требует UUID проекта.", "Проверьте URL подключения.")
			return
		}
		scope, err := hub.S.ResolveOperatorScope(r.Context(), projectID, r.URL.Query().Get("visit_session_id"), r.URL.Query().Get("conversation_id"))
		if err != nil {
			writeSiteChatErr(w, err)
			return
		}
		hub.serve(w, r, "operator", scope)
	}
}

func (h *Hub) serve(w stdhttp.ResponseWriter, r *stdhttp.Request, role string, scope sitechat.LiveAssistScope) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 8 * time.Second,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		CheckOrigin: func(*stdhttp.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	connectionID, err := h.S.RegisterConnection(r.Context(), scope, sitechat.ConnectionInput{
		Role:         role,
		ConnectionID: r.URL.Query().Get("connection_id"),
		UserAgent:    r.UserAgent(),
		Metadata: map[string]any{
			"remote_addr": r.RemoteAddr,
			"ws_path":     r.URL.Path,
		},
	})
	if err != nil {
		_ = conn.Close()
		return
	}
	client := &wsClient{
		hub:          h,
		conn:         conn,
		send:         make(chan WSEnvelope, 64),
		role:         role,
		connectionID: connectionID,
		scope:        scope,
		lastSeen:     time.Now(),
		closed:       make(chan struct{}),
	}
	h.register(client)
	client.enqueue(envelope("ack", scope, map[string]any{
		"connection_id": connectionID,
		"role":          role,
		"heartbeat_ms":  10000,
		"timeout_ms":    30000,
	}), false)
	if role == "visitor" {
		h.broadcastOperators(scope, envelope("presence.update", scope, map[string]any{
			"connection_status": "online",
			"visit_session_id":  scope.VisitSessionID,
		}))
	}
	go client.writePump()
	go client.readPump()
	go client.timeoutWatch()
}

func (h *Hub) register(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *wsClient, status, reason string) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
	if c.role == "visitor" {
		c.closeAndMaybeMarkVisitorLeft(status, reason)
		return
	}
	_ = h.S.CloseConnection(context.Background(), c.connectionID, status, reason)
}

func (c *wsClient) closeAndMaybeMarkVisitorLeft(status, reason string) {
	if c.role != "visitor" {
		return
	}
	c.leftOnce.Do(func() {
		ctx := context.Background()
		_ = c.hub.S.CloseConnection(ctx, c.connectionID, status, reason)
		active, err := c.hub.S.HasActiveVisitorConnections(ctx, c.scope, c.connectionID)
		if err == nil && active {
			return
		}
		_ = c.hub.S.MarkVisitorOffline(ctx, c.scope, reason)
		c.hub.broadcastOperators(c.scope, visitorLeftEnvelope(c.scope, reason))
	})
}

func (h *Hub) broadcastOperators(scope sitechat.LiveAssistScope, msg WSEnvelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.role != "operator" || client.scope.ProjectID != scope.ProjectID {
			continue
		}
		client.enqueue(msg, isLowPriority(msg.Type))
	}
}

func (h *Hub) broadcastVisitors(scope sitechat.LiveAssistScope, msg WSEnvelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.role != "visitor" {
			continue
		}
		if client.scope.VisitSessionID == scope.VisitSessionID || (scope.ConversationID != "" && client.scope.ConversationID == scope.ConversationID) {
			client.enqueue(msg, isLowPriority(msg.Type))
		}
	}
}

func (h *Hub) broadcastConversation(scope sitechat.LiveAssistScope, msg WSEnvelope) {
	h.broadcastOperators(scope, msg)
	h.broadcastVisitors(scope, msg)
}

func (c *wsClient) enqueue(msg WSEnvelope, lowPriority bool) {
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	if lowPriority {
		select {
		case c.send <- msg:
		default:
		}
		return
	}
	select {
	case c.send <- msg:
	case <-time.After(750 * time.Millisecond):
	}
}

func (c *wsClient) readPump() {
	defer func() {
		_ = c.conn.Close()
		closeOnce(c.closed)
		c.hub.unregister(c, "closed", "disconnect")
	}()
	c.conn.SetReadLimit(512 * 1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	c.conn.SetPongHandler(func(string) error {
		c.lastSeen = time.Now()
		return c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})
	for {
		var msg WSEnvelope
		if err := c.conn.ReadJSON(&msg); err != nil {
			return
		}
		c.lastSeen = time.Now()
		c.handle(msg)
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.closed:
			return
		}
	}
}

func (c *wsClient) timeoutWatch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if c.role != "visitor" {
				continue
			}
			if time.Since(c.lastSeen) <= wsVisitorGoneAfter {
				continue
			}
			c.hub.unregister(c, "timed_out", "heartbeat_timeout")
			_ = c.conn.Close()
			return
		case <-c.closed:
			return
		}
	}
}

func (c *wsClient) handle(msg WSEnvelope) {
	msg.Type = strings.TrimSpace(msg.Type)
	if msg.Type == "" {
		return
	}
	switch msg.Type {
	case "visitor.heartbeat":
		_ = c.hub.S.TouchConnection(context.Background(), c.connectionID)
		durationSeconds := intPayload(msg.Payload, "duration_seconds")
		_ = c.hub.S.MarkVisitorHeartbeat(context.Background(), c.scope, durationSeconds)
		c.hub.broadcastOperators(c.scope, envelope("presence.update", c.scope, map[string]any{
			"connection_status": "online",
			"duration_seconds":  durationSeconds,
		}))
	case "visitor.left":
		reason := stringPayload(msg.Payload, "reason")
		if reason == "" {
			reason = "visitor_left"
		}
		_ = c.hub.S.MarkVisitorHeartbeat(context.Background(), c.scope, intPayload(msg.Payload, "duration_seconds"))
		c.closeAndMaybeMarkVisitorLeft("closed", reason)
	case "event.batch":
		events := eventsFromPayload(msg.Payload)
		if len(events) > 0 {
			_ = c.hub.S.RecordLiveEvents(context.Background(), c.scope, events)
			c.hub.broadcastOperators(c.scope, withEnvelopeScope(msg, c.scope))
		}
	case "cursor.move", "touch.move", "click", "scroll", "viewport", "page.navigate", "monitoring.snapshot", "monitoring.incremental", "widget_state":
		_ = c.hub.S.RecordLiveEvents(context.Background(), c.scope, []sitechat.EventInput{{
			EventType:  msg.Type,
			OccurredAt: time.Now().UTC().Format(time.RFC3339),
			Payload:    msg.Payload,
			Priority:   priorityForType(msg.Type),
		}})
		c.hub.broadcastOperators(c.scope, withEnvelopeScope(msg, c.scope))
	case "chat.message":
		c.handleChatMessage(msg)
	case "invite.seen", "invite.accepted", "invite.declined":
		_ = c.hub.S.RecordInviteReceipt(context.Background(), sitechat.InviteReceiptInput{
			PublicKey:      stringPayload(msg.Payload, "public_key"),
			VisitorID:      c.scope.VisitorID,
			VisitSessionID: c.scope.VisitSessionID,
			ConversationID: c.scope.ConversationID,
			Status:         strings.TrimPrefix(msg.Type, "invite."),
		})
		c.hub.broadcastOperators(c.scope, withEnvelopeScope(msg, c.scope))
	case "monitoring.start":
		if c.role != "operator" {
			return
		}
		if id, err := uuid.Parse(c.scope.VisitSessionID); err == nil {
			_ = c.hub.S.SetMonitoring(context.Background(), id, "active")
			c.hub.broadcastVisitors(c.scope, envelope("monitoring.start", c.scope, msg.Payload))
		}
	case "monitoring.stop":
		if c.role != "operator" {
			return
		}
		if id, err := uuid.Parse(c.scope.VisitSessionID); err == nil {
			_ = c.hub.S.SetMonitoring(context.Background(), id, "stopped")
			c.hub.broadcastVisitors(c.scope, envelope("monitoring.stop", c.scope, msg.Payload))
		}
	case "call.request", "call.offer", "call.accept", "call.reject", "call.cancel", "call.end", "call.recording", "call.audio_chunk", "webrtc.offer", "webrtc.answer", "webrtc.ice", "consent.accepted", "consent.declined":
		scope := c.scope
		if msg.CallSessionID != "" {
			_ = c.hub.S.MarkCallSignal(context.Background(), msg.CallSessionID)
			if callID, err := uuid.Parse(strings.TrimSpace(msg.CallSessionID)); err == nil {
				if call, err := c.hub.S.GetCall(context.Background(), callID); err == nil {
					scope = callScope(call)
				}
			}
		}
		c.hub.broadcastConversation(scope, withEnvelopeScope(msg, scope))
	default:
		slog.Debug("site chat websocket ignored message", slog.String("type", msg.Type))
	}
}

func (c *wsClient) handleChatMessage(msg WSEnvelope) {
	body := strings.TrimSpace(stringPayload(msg.Payload, "body"))
	attachmentIDs := stringSlicePayload(msg.Payload, "attachment_ids")
	if (body == "" && len(attachmentIDs) == 0) || c.scope.ConversationID == "" {
		return
	}
	var id string
	var err error
	if c.role == "operator" {
		convID, parseErr := uuid.Parse(c.scope.ConversationID)
		if parseErr != nil {
			return
		}
		id, err = c.hub.S.LiveOperatorMessage(context.Background(), convID, body)
	} else {
		id, err = c.hub.S.PublicMessage(context.Background(), sitechat.MessageInput{
			PublicKey:      stringPayload(msg.Payload, "public_key"),
			ConversationID: c.scope.ConversationID,
			VisitSessionID: c.scope.VisitSessionID,
			VisitorID:      c.scope.VisitorID,
			AuthorName:     stringPayload(msg.Payload, "author_name"),
			VisitorContact: stringPayload(msg.Payload, "visitor_contact"),
			Body:           body,
			AttachmentIDs:  attachmentIDs,
		})
	}
	if err != nil {
		c.enqueue(envelope("error", c.scope, map[string]any{"code": "chat_message_failed", "detail": err.Error()}), false)
		return
	}
	if msg.Payload == nil {
		msg.Payload = map[string]any{}
	}
	msg.Payload["message_id"] = id
	c.hub.broadcastConversation(c.scope, withEnvelopeScope(msg, c.scope))
}

func envelope(kind string, scope sitechat.LiveAssistScope, payload map[string]any) WSEnvelope {
	return WSEnvelope{
		ID:             uuid.NewString(),
		Type:           kind,
		SentAt:         time.Now().UTC(),
		ProjectID:      scope.ProjectID,
		ChatWidgetID:   scope.ChatWidgetID,
		VisitorID:      scope.VisitorID,
		VisitSessionID: scope.VisitSessionID,
		ConversationID: scope.ConversationID,
		Payload:        payload,
	}
}

func visitorLeftEnvelope(scope sitechat.LiveAssistScope, reason string) WSEnvelope {
	if strings.TrimSpace(reason) == "" {
		reason = "disconnect"
	}
	return envelope("visitor.left", scope, map[string]any{
		"reason":            reason,
		"connection_status": "offline",
		"visit_session_id":  scope.VisitSessionID,
		"conversation_id":   scope.ConversationID,
		"last_seen_at":      time.Now().UTC().Format(time.RFC3339),
	})
}

func withEnvelopeScope(msg WSEnvelope, scope sitechat.LiveAssistScope) WSEnvelope {
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	msg.SentAt = time.Now().UTC()
	msg.ProjectID = scope.ProjectID
	msg.ChatWidgetID = scope.ChatWidgetID
	msg.VisitorID = scope.VisitorID
	msg.VisitSessionID = scope.VisitSessionID
	if msg.ConversationID == "" {
		msg.ConversationID = scope.ConversationID
	}
	return msg
}

func eventsFromPayload(payload map[string]any) []sitechat.EventInput {
	raw, ok := payload["events"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var events []sitechat.EventInput
	if json.Unmarshal(b, &events) != nil {
		return nil
	}
	if len(events) > 50 {
		return events[:50]
	}
	return events
}

func stringPayload(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, _ := payload[key].(string)
	return strings.TrimSpace(value)
}

func intPayload(payload map[string]any, key string) int {
	if payload == nil {
		return 0
	}
	switch value := payload[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	}
	return 0
}

func stringSlicePayload(payload map[string]any, key string) []string {
	if payload == nil {
		return nil
	}
	raw, ok := payload[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, strings.TrimSpace(text))
			}
		}
		return out
	default:
		return nil
	}
}

func priorityForType(kind string) string {
	if isLowPriority(kind) {
		return "low"
	}
	return "normal"
}

func isLowPriority(kind string) bool {
	switch kind {
	case "cursor.move", "touch.move", "scroll", "viewport", "monitoring.incremental", "event.batch":
		return true
	default:
		return false
	}
}

func closeOnce(ch chan struct{}) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}
