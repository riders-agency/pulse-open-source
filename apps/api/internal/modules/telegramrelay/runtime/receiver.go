package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/executor"
	"riders.agency/pulse/api/internal/platform/httptransport"
	"riders.agency/pulse/api/internal/platform/outbox"
)

type Relay struct {
	P          *pgxpool.Pool
	E          *executor.Engine
	HTTPClient *http.Client
	MasterHEX  string
}

type relayConfig struct {
	ActionDependencyID string
	ProjectID          string
	RelayKey           string
	PublicURL          string
	WebhookSecretHash  string
	CountryRelaysJSON  []byte
}

type relayBotAPIConfig struct {
	relayConfig
	BotToken string
}

type relayWorkflow struct {
	WorkflowID string
	VersionID  string
	Nodes      []relayNode
}

type relayNode struct {
	ActionKey     string          `json:"action_key"`
	ActionTypeKey string          `json:"action_type_key"`
	Config        relayNodeConfig `json:"config"`
	RawConfig     map[string]any  `json:"-"`
	Raw           map[string]any  `json:"-"`
}

type relayNodeConfig struct {
	ActionDependencyID string `json:"action_dependency_id"`
	RelayKey           string `json:"relay_key"`
	Enabled            *bool  `json:"enabled"`
}

type relayDefinition struct {
	Nodes []relayNode `json:"nodes"`
}

type normalizedUpdate struct {
	UpdateID   *int64
	UpdateType string
	Command    string
	ChatID     string
	UserID     string
	MessageID  *int64
	Text       string
	Normalized map[string]any
}

type botSession struct {
	ID      string
	State   string
	Context map[string]any
	ChatID  string
	UserID  string
}

type Origin struct {
	Ingress   string
	Headers   map[string]any
	UserAgent string
}

func (h *Relay) IngestPayload(ctx context.Context, cfg *relayConfig, raw []byte, payload map[string]any, origin Origin) (map[string]any, error) {
	if origin.Ingress == "" {
		origin.Ingress = "webhook"
	}
	if origin.Headers == nil {
		origin.Headers = map[string]any{}
	}
	normalized := normalizeTelegramUpdate(cfg.RelayKey, payload)
	normalized.Normalized["ingress"] = origin.Ingress

	eventID := uuid.New()
	tx, err := h.P.Begin(ctx)
	if err != nil {
		return nil, errors.New("tx")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
INSERT INTO telegram_relay_events (
  telegram_relay_event_id, action_dependency_id, project_id, relay_key,
  update_id, update_type, command, chat_id, user_id, message_id, text,
  payload_json, normalized_json, raw_body, headers_json, user_agent, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13::jsonb,$14,$15::jsonb,$16,'received')`,
		eventID.String(), cfg.ActionDependencyID, cfg.ProjectID, cfg.RelayKey,
		int64OrNil(normalized.UpdateID), normalized.UpdateType, normalized.Command, normalized.ChatID,
		normalized.UserID, int64OrNil(normalized.MessageID), normalized.Text,
		jsonString(payload), jsonString(normalized.Normalized), raw, jsonString(origin.Headers),
		origin.UserAgent)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" && normalized.UpdateID != nil {
			return h.loadDuplicate(ctx, cfg.ActionDependencyID, *normalized.UpdateID)
		}
		return nil, errors.New("event log")
	}
	session, err := h.upsertBotSession(ctx, tx, cfg, eventID.String(), normalized)
	if err != nil {
		return nil, errors.New("bot session")
	}

	workflows, err := h.workflowsForRelay(ctx, tx, cfg.ProjectID, cfg.ActionDependencyID, cfg.RelayKey)
	if err != nil {
		return nil, errors.New("workflow lookup")
	}
	runIDs := make([]string, 0, len(workflows))
	for _, wf := range workflows {
		runID := uuid.New()
		input := relayRunInput(cfg, eventID.String(), normalized.Normalized, payload, session)
		eventActionKey := "telegram_relay"
		if len(wf.Nodes) > 0 && strings.TrimSpace(wf.Nodes[0].ActionKey) != "" {
			eventActionKey = strings.TrimSpace(wf.Nodes[0].ActionKey)
		}
		outboxPayload := map[string]any{
			"workflow_run_id":     runID.String(),
			"project_id":          cfg.ProjectID,
			"workflow_id":         wf.WorkflowID,
			"workflow_version_id": wf.VersionID,
			"source_type":         "telegram_relay",
			"input_json":          input,
			"event_action_key":    eventActionKey,
			"event_message":       "telegram relay update queued",
			"event_payload_json":  input,
		}
		if _, err := outbox.EnqueueTx(ctx, tx, outbox.EnqueueRequest{
			TypeKey:        "workflow.run.requested",
			AggregateType:  "workflow_run",
			AggregateID:    runID.String(),
			IdempotencyKey: "workflow.run.requested:" + runID.String(),
			PayloadJSON:    []byte(jsonString(outboxPayload)),
		}); err != nil {
			return nil, errors.New("run")
		}
		runIDs = append(runIDs, runID.String())
	}
	// The outbox worker creates workflow_runs after this transaction commits.
	// Writing last_workflow_run_id here would point at a row that does not
	// exist yet and abort the whole relay ingest transaction via the FK.
	status := "ignored"
	queuedAt := "NULL"
	if len(runIDs) > 0 {
		status = "queued"
		queuedAt = "now()"
	}
	updateSQL := fmt.Sprintf(`
UPDATE telegram_relay_events
   SET workflow_run_ids_json = $2::jsonb,
       status = $3,
       queued_at = %s
 WHERE telegram_relay_event_id = $1`, queuedAt)
	if _, err := tx.Exec(ctx, updateSQL, eventID.String(), jsonString(runIDs), status); err != nil {
		return nil, errors.New("event update")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, errors.New("commit")
	}

	if h.E != nil && len(runIDs) > 0 {
		h.E.Notify()
	}
	forwardCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	h.forwardRelays(forwardCtx, eventID.String(), cfg.CountryRelaysJSON, raw, cfg.RelayKey)
	cancel()
	slog.Info("telegram relay update received",
		"telegram_relay_event_id", eventID.String(),
		"relay_key", cfg.RelayKey,
		"ingress", origin.Ingress,
		"workflow_run_count", len(runIDs),
	)
	return map[string]any{
		"ok":                      true,
		"telegram_relay_event_id": eventID.String(),
		"workflow_run_ids":        runIDs,
		"status":                  status,
		"ingress":                 origin.Ingress,
	}, nil
}

func (h *Relay) LoadConfig(ctx context.Context, relayKey string) (*relayConfig, error) {
	row := h.P.QueryRow(ctx, `
SELECT trc.action_dependency_id::text,
       ad.project_id::text,
       trc.relay_key::text,
       trc.public_url::text,
       trc.webhook_secret_hash::text,
       trc.country_relays_json
  FROM telegram_relay_configs trc
  JOIN pulse_read.action_dependencies_v1 ad ON ad.action_dependency_id = trc.action_dependency_id
 WHERE trc.relay_key = $1
   AND trc.status = 'active'
   AND ad.status = 'active'`, relayKey)
	var cfg relayConfig
	if err := row.Scan(&cfg.ActionDependencyID, &cfg.ProjectID, &cfg.RelayKey, &cfg.PublicURL, &cfg.WebhookSecretHash, &cfg.CountryRelaysJSON); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (h *Relay) loadDuplicate(ctx context.Context, depID string, updateID int64) (map[string]any, error) {
	var eventID string
	var runIDs []byte
	err := h.P.QueryRow(ctx, `
SELECT telegram_relay_event_id::text, workflow_run_ids_json
  FROM telegram_relay_events
 WHERE action_dependency_id = $1 AND update_id = $2`, depID, updateID).Scan(&eventID, &runIDs)
	if err != nil {
		return nil, err
	}
	_, _ = h.P.Exec(ctx, `
UPDATE telegram_relay_events
   SET status = CASE WHEN status = 'queued' THEN status ELSE 'duplicate' END,
       processed_at = COALESCE(processed_at, now())
 WHERE telegram_relay_event_id = $1`, eventID)
	var ids []string
	_ = json.Unmarshal(runIDs, &ids)
	return map[string]any{
		"ok":                      true,
		"duplicate":               true,
		"telegram_relay_event_id": eventID,
		"workflow_run_ids":        ids,
	}, nil
}

func normalizeTelegramUpdate(relayKey string, payload map[string]any) normalizedUpdate {
	updateType, updateValue := telegramUpdateType(payload)
	message := messageFromUpdate(updateType, updateValue)
	chat := mapAt(message, "chat")
	from := mapAt(message, "from")
	if updateType == "callback_query" {
		if callbackFrom := mapAt(updateValue, "from"); len(callbackFrom) > 0 {
			from = callbackFrom
		}
	}
	text := firstNonEmpty(stringAt(message, "text"), stringAt(message, "caption"))
	updateID := int64At(payload, "update_id")
	messageID := int64At(message, "message_id")
	chatID := numberOrStringAt(chat, "id")
	userID := numberOrStringAt(from, "id")
	callbackData := ""
	if updateType == "callback_query" {
		callbackData = stringAt(updateValue, "data")
	}
	command := telegramCommand(message, text)
	if command == "" {
		fields := strings.Fields(callbackData)
		if len(fields) > 0 && strings.HasPrefix(fields[0], "/") {
			command = commandName(fields[0])
		}
	}
	normalized := map[string]any{
		"relay_key":     relayKey,
		"update_id":     int64Value(updateID),
		"update_type":   updateType,
		"command":       command,
		"text":          text,
		"caption":       stringAt(message, "caption"),
		"chat_id":       chatID,
		"user_id":       userID,
		"message_id":    int64Value(messageID),
		"callback_data": callbackData,
		"chat":          chat,
		"from":          from,
		"message":       message,
		"raw_update":    payload,
	}
	return normalizedUpdate{
		UpdateID:   updateID,
		UpdateType: updateType,
		Command:    command,
		ChatID:     chatID,
		UserID:     userID,
		MessageID:  messageID,
		Text:       text,
		Normalized: normalized,
	}
}

func telegramUpdateType(payload map[string]any) (string, map[string]any) {
	keys := []string{
		"message", "edited_message", "channel_post", "edited_channel_post",
		"business_connection", "business_message", "edited_business_message",
		"deleted_business_messages", "guest_message", "callback_query",
		"inline_query", "chosen_inline_result", "shipping_query", "pre_checkout_query",
		"purchased_paid_media", "poll", "poll_answer", "my_chat_member", "chat_member", "chat_join_request",
		"message_reaction", "message_reaction_count", "chat_boost", "removed_chat_boost",
		"managed_bot",
	}
	for _, key := range keys {
		if value, ok := payload[key].(map[string]any); ok {
			return key, value
		}
	}
	return "unknown", map[string]any{}
}

func messageFromUpdate(updateType string, value map[string]any) map[string]any {
	switch updateType {
	case "callback_query":
		return mapAt(value, "message")
	case "my_chat_member", "chat_member":
		return map[string]any{"chat": mapAt(value, "chat"), "from": mapAt(value, "from")}
	case "business_connection":
		return map[string]any{"from": mapAt(value, "user"), "chat": map[string]any{"id": value["user_chat_id"], "type": "private"}}
	case "deleted_business_messages", "message_reaction", "message_reaction_count", "chat_boost", "removed_chat_boost", "chat_join_request":
		return map[string]any{"chat": mapAt(value, "chat"), "from": firstMapAt(value, "from", "user")}
	case "inline_query", "chosen_inline_result", "shipping_query", "pre_checkout_query", "purchased_paid_media", "poll_answer":
		return map[string]any{"from": firstMapAt(value, "from", "user")}
	default:
		return value
	}
}

func telegramCommand(message map[string]any, text string) string {
	for _, entityKey := range []string{"entities", "caption_entities"} {
		for _, raw := range arrayAt(message, entityKey) {
			entity, ok := raw.(map[string]any)
			if !ok || stringAt(entity, "type") != "bot_command" || intAt(entity, "offset") != 0 {
				continue
			}
			length := intAt(entity, "length")
			if length > 0 && len([]rune(text)) >= length {
				return commandName(string([]rune(text)[:length]))
			}
		}
	}
	if strings.HasPrefix(text, "/") {
		first := strings.Fields(text)
		if len(first) > 0 {
			return commandName(first[0])
		}
	}
	return ""
}

func firstMapAt(m map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if out := mapAt(m, key); len(out) > 0 {
			return out
		}
	}
	return map[string]any{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func commandName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "/")
	value = strings.SplitN(value, "@", 2)[0]
	return strings.TrimSpace(value)
}

func (h *Relay) workflowsForRelay(ctx context.Context, tx pgx.Tx, projectID, depID, relayKey string) ([]relayWorkflow, error) {
	rows, err := tx.Query(ctx, `
SELECT w.workflow_id::text, w.current_workflow_version_id::text, wv.definition_json
  FROM pulse_read.workflows_v1 w
  JOIN pulse_read.workflow_versions_v1 wv ON wv.workflow_version_id = w.current_workflow_version_id
 WHERE w.project_id = $1
   AND w.status = 'active'
   AND w.current_workflow_version_id IS NOT NULL`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]relayWorkflow, 0)
	for rows.Next() {
		var wf relayWorkflow
		var def []byte
		if err := rows.Scan(&wf.WorkflowID, &wf.VersionID, &def); err != nil {
			return nil, err
		}
		wf.Nodes = relayNodesForDefinition(def, depID, relayKey)
		if len(wf.Nodes) > 0 {
			out = append(out, wf)
		}
	}
	return out, rows.Err()
}

func relayNodesForDefinition(def []byte, depID, relayKey string) []relayNode {
	var parsed relayDefinition
	_ = json.Unmarshal(def, &parsed)
	out := make([]relayNode, 0)
	for _, node := range parsed.Nodes {
		if strings.TrimSpace(node.ActionTypeKey) != "telegram.relay.trigger" {
			continue
		}
		if node.Config.Enabled != nil && !*node.Config.Enabled {
			continue
		}
		cfgDep := strings.TrimSpace(node.Config.ActionDependencyID)
		cfgRelayKey := strings.TrimSpace(node.Config.RelayKey)
		if cfgDep != "" && cfgDep == depID {
			out = append(out, normalizedRelayNode(node))
			continue
		}
		if cfgDep == "" && cfgRelayKey != "" && cfgRelayKey == relayKey {
			out = append(out, normalizedRelayNode(node))
		}
	}
	return out
}

func normalizedRelayNode(node relayNode) relayNode {
	node.ActionKey = strings.TrimSpace(node.ActionKey)
	if node.ActionKey == "" {
		node.ActionKey = "telegram_relay"
	}
	return node
}

func (h *Relay) upsertBotSession(ctx context.Context, tx pgx.Tx, cfg *relayConfig, eventID string, normalized normalizedUpdate) (*botSession, error) {
	if strings.TrimSpace(normalized.ChatID) == "" {
		return nil, nil
	}
	userID := strings.TrimSpace(normalized.UserID)
	var id, state string
	var rawContext []byte
	err := tx.QueryRow(ctx, `
INSERT INTO bot_sessions (
  project_id, action_dependency_id, relay_key, chat_id, user_id,
  last_update_id, last_telegram_relay_event_id, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,now())
ON CONFLICT (project_id, action_dependency_id, relay_key, chat_id, user_id) DO UPDATE SET
  last_update_id = EXCLUDED.last_update_id,
  last_telegram_relay_event_id = EXCLUDED.last_telegram_relay_event_id,
  updated_at = now()
RETURNING bot_session_id::text, state_key::text, context_json`,
		cfg.ProjectID, cfg.ActionDependencyID, cfg.RelayKey, normalized.ChatID, userID,
		int64OrNil(normalized.UpdateID), eventID).Scan(&id, &state, &rawContext)
	if err != nil {
		return nil, err
	}
	contextJSON := map[string]any{}
	_ = json.Unmarshal(rawContext, &contextJSON)
	return &botSession{
		ID:      id,
		State:   state,
		Context: contextJSON,
		ChatID:  normalized.ChatID,
		UserID:  userID,
	}, nil
}

func relayRunInput(cfg *relayConfig, eventID string, normalized map[string]any, payload map[string]any, session *botSession) map[string]any {
	pulse := map[string]any{
		"source_type":               "telegram_relay",
		"telegram_relay_key":        cfg.RelayKey,
		"telegram_relay_event_id":   eventID,
		"action_dependency_id":      cfg.ActionDependencyID,
		"telegram_update_id":        normalized["update_id"],
		"telegram_update_type":      normalized["update_type"],
		"telegram_command":          normalized["command"],
		"telegram_relay_public_url": cfg.PublicURL,
	}
	out := map[string]any{
		"_pulse":         pulse,
		"telegram":       payload,
		"telegram_relay": normalized,
	}
	if session != nil {
		out["bot_session"] = map[string]any{
			"bot_session_id": session.ID,
			"state_key":      session.State,
			"context_json":   session.Context,
			"chat_id":        session.ChatID,
			"user_id":        session.UserID,
			"relay_key":      cfg.RelayKey,
		}
	}
	if cfg.RelayKey != "" {
		out[cfg.RelayKey] = normalized
	}
	return out
}

type countryRelay struct {
	Name                string `json:"name"`
	Region              string `json:"region"`
	URL                 string `json:"url"`
	Enabled             *bool  `json:"enabled"`
	AuthorizationHeader string `json:"authorization_header"`
	SecretToken         string `json:"secret_token"`
}

func (h *Relay) forwardRelays(ctx context.Context, eventID string, relaysJSON []byte, raw []byte, relayKey string) {
	var relays []countryRelay
	if err := json.Unmarshal(relaysJSON, &relays); err != nil || len(relays) == 0 {
		return
	}
	client := h.HTTPClient
	if client == nil {
		client = httptransport.NewDefaultClient(5 * time.Second)
	}
	for _, relay := range relays {
		enabled := relay.Enabled == nil || *relay.Enabled
		if !enabled {
			h.recordForward(ctx, eventID, relay, "skipped", 0, map[string]any{}, "relay disabled")
			continue
		}
		u, err := neturl.Parse(strings.TrimSpace(relay.URL))
		if err != nil || u.Scheme != "https" || u.Host == "" {
			h.recordForward(ctx, eventID, relay, "skipped", 0, map[string]any{}, "relay url must be absolute https")
			continue
		}
		reqCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, relay.URL, bytes.NewReader(raw))
		if err != nil {
			cancel()
			h.recordForward(ctx, eventID, relay, "failed", 0, map[string]any{}, err.Error())
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Riders-Relay-Source", relayKey)
		if relay.AuthorizationHeader != "" {
			req.Header.Set("Authorization", relay.AuthorizationHeader)
		}
		if relay.SecretToken != "" {
			req.Header.Set("X-Relay-Secret", relay.SecretToken)
		}
		resp, err := client.Do(req)
		cancel()
		if err != nil {
			h.recordForward(ctx, eventID, relay, "failed", 0, requestHeadersFrom(req.Header), err.Error())
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
		_ = resp.Body.Close()
		response := map[string]any{"body": strings.TrimSpace(string(body))}
		var parsed any
		if json.Unmarshal(body, &parsed) == nil {
			response["json"] = parsed
		}
		status := "succeeded"
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			status = "failed"
		}
		h.recordForward(ctx, eventID, relay, status, resp.StatusCode, requestHeadersFrom(req.Header), "")
		_, _ = h.P.Exec(ctx, `
UPDATE telegram_relay_forwards
   SET response_json = $3::jsonb
 WHERE telegram_relay_event_id = $1 AND relay_url = $2`,
			eventID, relay.URL, jsonString(response))
	}
}

func (h *Relay) recordForward(ctx context.Context, eventID string, relay countryRelay, status string, statusCode int, headers map[string]any, errText string) {
	if h.P == nil {
		return
	}
	_, _ = h.P.Exec(ctx, `
INSERT INTO telegram_relay_forwards (
  telegram_relay_event_id, relay_url, relay_region, status, status_code,
  request_headers_json, response_json, error, finished_at)
VALUES ($1,$2,$3,$4,$5,$6::jsonb,'{}'::jsonb,$7,now())`,
		eventID, relay.URL, relay.Region, status, nilIfZero(statusCode), jsonString(headers), nullIfEmpty(errText))
}

func requestHeadersFrom(headers http.Header) map[string]any {
	out := map[string]any{}
	for key, values := range headers {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "authorization") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") {
			out[key] = "***"
			continue
		}
		if len(values) == 1 {
			out[key] = values[0]
		} else {
			out[key] = values
		}
	}
	return out
}

func jsonString(value any) string {
	b, _ := json.Marshal(value)
	return string(b)
}

func int64OrNil(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func int64Value(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nilIfZero(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func mapAt(root map[string]any, key string) map[string]any {
	value, _ := root[key].(map[string]any)
	if value == nil {
		return map[string]any{}
	}
	return value
}

func arrayAt(root map[string]any, key string) []any {
	value, _ := root[key].([]any)
	return value
}

func stringAt(root map[string]any, key string) string {
	value, _ := root[key].(string)
	return strings.TrimSpace(value)
}

func intAt(root map[string]any, key string) int {
	switch value := root[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	case int64:
		return int(value)
	case json.Number:
		out, _ := value.Int64()
		return int(out)
	default:
		return 0
	}
}

func int64At(root map[string]any, key string) *int64 {
	switch value := root[key].(type) {
	case float64:
		out := int64(value)
		return &out
	case int:
		out := int64(value)
		return &out
	case int64:
		out := value
		return &out
	case json.Number:
		out, err := value.Int64()
		if err == nil {
			return &out
		}
	}
	return nil
}

func numberOrStringAt(root map[string]any, key string) string {
	switch value := root[key].(type) {
	case string:
		return strings.TrimSpace(value)
	case float64:
		return fmt.Sprintf("%.0f", value)
	case int:
		return fmt.Sprintf("%d", value)
	case int64:
		return fmt.Sprintf("%d", value)
	case json.Number:
		return value.String()
	default:
		return ""
	}
}
