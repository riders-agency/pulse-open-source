package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

type LiveAssistScope struct {
	CompanyID       string
	ProjectID       string
	ChatWidgetID    string
	VisitorID       string
	VisitSessionID  string
	ConversationID  string
	CurrentURL      string
	CurrentTitle    string
	VisitorSnapshot map[string]any
}

type ConnectionInput struct {
	Role         string
	ConnectionID string
	UserAgent    string
	Metadata     map[string]any
}

type InviteInput struct {
	Message  string `json:"message"`
	Template string `json:"template"`
	Force    bool   `json:"force"`
}

type InviteReceiptInput struct {
	PublicKey      string `json:"public_key"`
	VisitorID      string `json:"visitor_id"`
	VisitSessionID string `json:"visit_session_id"`
	ConversationID string `json:"conversation_id"`
	Status         string `json:"status"`
}

type CallInput struct {
	PublicKey                string `json:"public_key"`
	VisitorID                string `json:"visitor_id"`
	VisitSessionID           string `json:"visit_session_id"`
	ConversationID           string `json:"conversation_id"`
	MediaKind                string `json:"media_kind"`
	RecordingEnabled         *bool  `json:"recording_enabled"`
	PrivacyConsentAccepted   bool   `json:"privacy_consent_accepted"`
	RecordingConsentAccepted bool   `json:"recording_consent_accepted"`
	EndReason                string `json:"end_reason"`
}

type CallSessionRow struct {
	CallSessionID              string         `json:"call_session_id"`
	CompanyID                  string         `json:"company_id"`
	ProjectID                  string         `json:"project_id"`
	ChatWidgetID               string         `json:"chat_widget_id"`
	VisitorID                  string         `json:"visitor_id"`
	VisitSessionID             string         `json:"visit_session_id,omitempty"`
	ConversationID             string         `json:"conversation_id,omitempty"`
	Direction                  string         `json:"direction"`
	RequestedBy                string         `json:"requested_by"`
	RequestedByUserID          string         `json:"requested_by_user_id,omitempty"`
	Status                     string         `json:"status"`
	RecordingEnabled           bool           `json:"recording_enabled"`
	MediaKind                  string         `json:"media_kind"`
	PrivacyConsentAcceptedAt   *time.Time     `json:"privacy_consent_accepted_at,omitempty"`
	RecordingConsentAcceptedAt *time.Time     `json:"recording_consent_accepted_at,omitempty"`
	StartedAt                  *time.Time     `json:"started_at,omitempty"`
	EndedAt                    *time.Time     `json:"ended_at,omitempty"`
	LastSignalAt               *time.Time     `json:"last_signal_at,omitempty"`
	EndReason                  string         `json:"end_reason"`
	Metadata                   map[string]any `json:"metadata_json"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
}

type CobrowseFrameRow struct {
	FrameID        string         `json:"frame_id"`
	VisitSessionID string         `json:"visit_session_id"`
	Sequence       int64          `json:"sequence"`
	Kind           string         `json:"kind"`
	OccurredAt     time.Time      `json:"occurred_at"`
	Payload        map[string]any `json:"payload_json"`
	CreatedAt      time.Time      `json:"created_at"`
}

func (s *SiteChat) ProjectForSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	return s.sessionProject(ctx, sessionID)
}

func (s *SiteChat) ResolvePublicScope(ctx context.Context, publicKey, visitorID, sessionID, conversationID string) (LiveAssistScope, error) {
	widget, err := s.lookupWidget(ctx, publicKey)
	if err != nil {
		return LiveAssistScope{}, err
	}
	if _, err := uuid.Parse(strings.TrimSpace(visitorID)); err != nil {
		return LiveAssistScope{}, ErrBadInput
	}
	if _, err := uuid.Parse(strings.TrimSpace(sessionID)); err != nil {
		return LiveAssistScope{}, ErrBadInput
	}
	var scope LiveAssistScope
	err = s.P.QueryRow(ctx, `
SELECT s.company_id::text, s.project_id::text, s.chat_widget_id::text, s.visitor_id::text,
       s.visit_session_id::text, COALESCE(c.conversation_id::text, ''), s.current_url::text,
       s.current_title::text, s.visitor_state_snapshot
  FROM site_visit_sessions s
  LEFT JOIN LATERAL (
    SELECT conversation_id
      FROM site_chat_conversations c
     WHERE c.visit_session_id = s.visit_session_id
     ORDER BY c.created_at ASC
     LIMIT 1
  ) c ON true
 WHERE s.visit_session_id = $1
   AND s.visitor_id = $2
   AND s.chat_widget_id = $3`, strings.TrimSpace(sessionID), strings.TrimSpace(visitorID), widget.ChatWidgetID).
		Scan(&scope.CompanyID, &scope.ProjectID, &scope.ChatWidgetID, &scope.VisitorID, &scope.VisitSessionID,
			&scope.ConversationID, &scope.CurrentURL, &scope.CurrentTitle, (*jsonMap)(&scope.VisitorSnapshot))
	if err != nil {
		return LiveAssistScope{}, mapErr(err)
	}
	if strings.TrimSpace(conversationID) != "" {
		convUUID, err := uuid.Parse(strings.TrimSpace(conversationID))
		if err != nil {
			return LiveAssistScope{}, ErrBadInput
		}
		if err := s.ensurePublicConversation(ctx, widget, convUUID, visitorID); err != nil {
			return LiveAssistScope{}, err
		}
		scope.ConversationID = strings.TrimSpace(conversationID)
	}
	return scope, nil
}

func (s *SiteChat) ResolveOperatorScope(ctx context.Context, projectID uuid.UUID, sessionID, conversationID string) (LiveAssistScope, error) {
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return LiveAssistScope{}, err
	}
	scope := LiveAssistScope{ProjectID: projectID.String()}
	if strings.TrimSpace(sessionID) != "" {
		err := s.P.QueryRow(ctx, `
SELECT company_id::text, project_id::text, chat_widget_id::text, visitor_id::text,
       visit_session_id::text, current_url::text, current_title::text, visitor_state_snapshot
  FROM site_visit_sessions
 WHERE visit_session_id = $1 AND project_id = $2`, strings.TrimSpace(sessionID), projectID.String()).
			Scan(&scope.CompanyID, &scope.ProjectID, &scope.ChatWidgetID, &scope.VisitorID, &scope.VisitSessionID,
				&scope.CurrentURL, &scope.CurrentTitle, (*jsonMap)(&scope.VisitorSnapshot))
		if err != nil {
			return LiveAssistScope{}, mapErr(err)
		}
	}
	if strings.TrimSpace(conversationID) != "" {
		var visitSessionID string
		err := s.P.QueryRow(ctx, `
SELECT company_id::text, project_id::text, chat_widget_id::text, visitor_id::text,
       COALESCE(visit_session_id::text, '')
  FROM site_chat_conversations
 WHERE conversation_id = $1 AND project_id = $2`, strings.TrimSpace(conversationID), projectID.String()).
			Scan(&scope.CompanyID, &scope.ProjectID, &scope.ChatWidgetID, &scope.VisitorID, &visitSessionID)
		if err != nil {
			return LiveAssistScope{}, mapErr(err)
		}
		scope.ConversationID = strings.TrimSpace(conversationID)
		if scope.VisitSessionID == "" {
			scope.VisitSessionID = visitSessionID
		}
	}
	if scope.CompanyID == "" {
		var companyID string
		if err := s.P.QueryRow(ctx, `SELECT company_id::text FROM pulse_read.projects_v1 WHERE project_id = $1`, projectID.String()).Scan(&companyID); err != nil {
			return LiveAssistScope{}, mapErr(err)
		}
		scope.CompanyID = companyID
	}
	return scope, nil
}

func (s *SiteChat) RegisterConnection(ctx context.Context, scope LiveAssistScope, in ConnectionInput) (string, error) {
	role := strings.TrimSpace(in.Role)
	if role != "operator" {
		role = "visitor"
	}
	connectionID := strings.TrimSpace(in.ConnectionID)
	if _, err := uuid.Parse(connectionID); err != nil {
		connectionID = uuid.NewString()
	}
	_, err := s.P.Exec(ctx, `
INSERT INTO site_visit_connections (
  connection_id, company_id, project_id, chat_widget_id, visitor_id, visit_session_id,
  role, status, last_heartbeat_at, user_agent_hash, metadata_json
) VALUES ($1,$2,$3,NULLIF($4,'')::uuid,NULLIF($5,'')::uuid,NULLIF($6,'')::uuid,$7,'connected',now(),$8,$9::jsonb)
ON CONFLICT (connection_id) DO UPDATE SET
  status = 'connected',
  last_heartbeat_at = now(),
  closed_at = NULL,
  close_reason = '',
  metadata_json = site_visit_connections.metadata_json || EXCLUDED.metadata_json`,
		connectionID, scope.CompanyID, scope.ProjectID, scope.ChatWidgetID, scope.VisitorID,
		scope.VisitSessionID, role, hashString(in.UserAgent), jsonBytes(in.Metadata))
	if err != nil {
		return "", err
	}
	if role == "visitor" && scope.VisitSessionID != "" {
		_ = s.MarkVisitorOnline(ctx, scope)
	}
	return connectionID, nil
}

func (s *SiteChat) TouchConnection(ctx context.Context, connectionID string) error {
	_, err := s.P.Exec(ctx, `
UPDATE site_visit_connections
   SET last_heartbeat_at = now(), status = 'connected'
 WHERE connection_id = $1`, strings.TrimSpace(connectionID))
	return err
}

func (s *SiteChat) CloseConnection(ctx context.Context, connectionID, status, reason string) error {
	status = strings.TrimSpace(status)
	if status != "timed_out" {
		status = "closed"
	}
	_, err := s.P.Exec(ctx, `
UPDATE site_visit_connections
   SET status = $2, closed_at = now(), close_reason = $3
 WHERE connection_id = $1`, strings.TrimSpace(connectionID), status, trimLimit(reason, 200))
	return err
}

func (s *SiteChat) HasActiveVisitorConnections(ctx context.Context, scope LiveAssistScope, excludeConnectionID string) (bool, error) {
	if strings.TrimSpace(scope.VisitSessionID) == "" {
		return false, nil
	}
	var active bool
	err := s.P.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM site_visit_connections
   WHERE visit_session_id = $1
     AND role = 'visitor'
     AND status = 'connected'
     AND (NULLIF($2, '') IS NULL OR connection_id <> NULLIF($2, '')::uuid)
     AND last_heartbeat_at > now() - interval '45 seconds'
)`, strings.TrimSpace(scope.VisitSessionID), strings.TrimSpace(excludeConnectionID)).Scan(&active)
	return active, err
}

func (s *SiteChat) MarkVisitorOnline(ctx context.Context, scope LiveAssistScope) error {
	return s.MarkVisitorHeartbeat(ctx, scope, 0)
}

func (s *SiteChat) MarkVisitorHeartbeat(ctx context.Context, scope LiveAssistScope, durationSeconds int) error {
	if scope.VisitSessionID == "" {
		return nil
	}
	if durationSeconds < 0 {
		durationSeconds = 0
	}
	if durationSeconds > 24*60*60 {
		durationSeconds = 24 * 60 * 60
	}
	_, err := s.P.Exec(ctx, `
UPDATE site_visit_sessions
   SET status = 'active',
       ended_at = NULL,
       last_activity_at = now(),
       duration_seconds = GREATEST(duration_seconds, $2),
       last_duration_flush_at = CASE WHEN $2 > duration_seconds THEN now() ELSE last_duration_flush_at END,
       visitor_state_snapshot =
         jsonb_set(
           jsonb_set(
             jsonb_set(
               jsonb_set(visitor_state_snapshot, '{activity,is_active}', 'true'::jsonb, true),
               '{activity,connection_status}', '"online"'::jsonb, true
             ),
             '{activity,dwell_seconds}', to_jsonb(GREATEST(duration_seconds, $2)), true
           ),
           '{activity,last_seen_at}', to_jsonb(now()), true
         )
 WHERE visit_session_id = $1`, scope.VisitSessionID, durationSeconds)
	return err
}

func (s *SiteChat) MarkVisitorOffline(ctx context.Context, scope LiveAssistScope, reason string) error {
	if scope.VisitSessionID == "" {
		return nil
	}
	_, err := s.P.Exec(ctx, `
UPDATE site_visit_sessions
   SET status = 'ended',
       ended_at = COALESCE(ended_at, now()),
       duration_seconds = GREATEST(duration_seconds, EXTRACT(EPOCH FROM (COALESCE(ended_at, now()) - started_at))::integer),
       visitor_state_snapshot =
         jsonb_set(
           jsonb_set(
             jsonb_set(
               jsonb_set(visitor_state_snapshot, '{activity,is_active}', 'false'::jsonb, true),
               '{activity,connection_status}', '"offline"'::jsonb, true
             ),
             '{activity,dwell_seconds}', to_jsonb(GREATEST(duration_seconds, EXTRACT(EPOCH FROM (COALESCE(ended_at, now()) - started_at))::integer)), true
           ),
           '{activity,last_seen_at}', to_jsonb(now()), true
         )
 WHERE visit_session_id = $1`, scope.VisitSessionID)
	if err != nil {
		return err
	}
	return s.endOpenCallsForSession(ctx, scope.VisitSessionID, "visitor_offline")
}

func (s *SiteChat) InviteWithInputTx(ctx context.Context, sessionID, uid uuid.UUID, in InviteInput) (*ConversationRow, error) {
	projectID, err := s.sessionProject(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return nil, err
	}
	tx, err := s.transaction()
	if err != nil {
		return nil, err
	}
	var companyID, widgetID, visitorID string
	if err := tx.QueryRow(ctx, `
SELECT company_id::text, chat_widget_id::text, visitor_id::text
  FROM site_visit_sessions
 WHERE visit_session_id = $1`, sessionID.String()).Scan(&companyID, &widgetID, &visitorID); err != nil {
		return nil, mapErr(err)
	}
	conversationID, err := ensureConversationForSessionTx(ctx, tx, widgetScope{
		ChatWidgetID: widgetID,
		CompanyID:    companyID,
		ProjectID:    projectID.String(),
	}, visitorID, sessionID.String(), "", "")
	if err != nil {
		return nil, err
	}
	_, _ = tx.Exec(ctx, `
UPDATE site_chat_conversations
   SET created_by_user_id = COALESCE(created_by_user_id, $2)
 WHERE conversation_id = $1`, conversationID, uid.String())
	_, err = tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET operator_invited = true,
       chat_open = true,
       visitor_state_snapshot = jsonb_set(jsonb_set(visitor_state_snapshot, '{signals,operator_invited}', 'true'::jsonb, true), '{signals,chat_open}', 'true'::jsonb, true)
 WHERE visit_session_id = $1`, sessionID.String())
	if err != nil {
		return nil, err
	}
	message := strings.TrimSpace(in.Message)
	if message == "" {
		message = "Оператор на связи. Помочь на этой странице?"
	}
	if err := insertSystemMessageTx(ctx, tx, conversationID, companyID, projectID.String(), "Оператор пригласил посетителя: "+message, map[string]any{
		"event_type": "operator.invite",
		"template":   strings.TrimSpace(in.Template),
		"force":      in.Force,
	}); err != nil {
		return nil, err
	}
	convID, _ := uuid.Parse(conversationID)
	return s.GetConversation(ctx, convID)
}

func (s *SiteChat) RecordInviteReceiptTx(ctx context.Context, in InviteReceiptInput) error {
	scope, err := s.ResolvePublicScope(ctx, in.PublicKey, in.VisitorID, in.VisitSessionID, in.ConversationID)
	if err != nil {
		return err
	}
	status := strings.TrimSpace(in.Status)
	if status != "seen" && status != "accepted" && status != "declined" {
		return ErrBadInput
	}
	tx, err := s.transaction()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET chat_open = CASE WHEN $2 = 'accepted' THEN true ELSE chat_open END,
       visitor_state_snapshot = jsonb_set(
         jsonb_set(visitor_state_snapshot, '{signals,operator_invited}', to_jsonb($2 <> 'declined'), true),
         '{signals,chat_open}', to_jsonb($2 = 'accepted'), true
       )
 WHERE visit_session_id = $1`, scope.VisitSessionID, status)
	if err != nil {
		return err
	}
	if scope.ConversationID != "" {
		body := map[string]string{
			"seen":     "Посетитель увидел приглашение",
			"accepted": "Посетитель принял приглашение",
			"declined": "Посетитель отклонил приглашение",
		}[status]
		if err := insertSystemMessageTx(ctx, tx, scope.ConversationID, scope.CompanyID, scope.ProjectID, body, map[string]any{
			"event_type": "invite." + status,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SiteChat) LiveOperatorMessageTx(ctx context.Context, conversationID uuid.UUID, body string) (string, error) {
	return s.insertMessage(ctx, conversationID, "operator", uuid.Nil, "Оператор", body)
}

func (s *SiteChat) OperatorCallOfferTx(ctx context.Context, sessionID, uid uuid.UUID, in CallInput) (*CallSessionRow, error) {
	projectID, err := s.sessionProject(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return nil, err
	}
	scope, err := s.ResolveOperatorScope(ctx, projectID, sessionID.String(), strings.TrimSpace(in.ConversationID))
	if err != nil {
		return nil, err
	}
	if err := s.requireCallableVisitor(ctx, scope); err != nil {
		return nil, err
	}
	return s.createCall(ctx, scope, uid, "operator_to_visitor", "operator", in, "Оператор предложил звонок")
}

func (s *SiteChat) PublicCallRequestTx(ctx context.Context, in CallInput) (*CallSessionRow, error) {
	scope, err := s.ResolvePublicScope(ctx, in.PublicKey, in.VisitorID, in.VisitSessionID, in.ConversationID)
	if err != nil {
		return nil, err
	}
	if err := s.requireCallableVisitor(ctx, scope); err != nil {
		return nil, err
	}
	return s.createCall(ctx, scope, uuid.Nil, "visitor_to_operator", "visitor", in, "Посетитель запросил звонок")
}

func (s *SiteChat) createCall(ctx context.Context, scope LiveAssistScope, uid uuid.UUID, direction, requestedBy string, in CallInput, message string) (*CallSessionRow, error) {
	mediaKind := strings.TrimSpace(in.MediaKind)
	if mediaKind != "video" {
		mediaKind = "audio"
	}
	recordingEnabled := false
	if in.RecordingEnabled != nil {
		recordingEnabled = *in.RecordingEnabled
	}
	if scope.ConversationID == "" && scope.VisitSessionID != "" {
		convID, err := s.ensureConversationForScope(ctx, scope)
		if err != nil {
			return nil, err
		}
		scope.ConversationID = convID
	}
	tx, err := s.transaction()
	if err != nil {
		return nil, err
	}
	callID := uuid.NewString()
	status := "ringing"
	if direction == "operator_to_visitor" {
		status = "offered"
	}
	var uidAny any
	if uid != uuid.Nil {
		uidAny = uid.String()
	}
	_, err = tx.Exec(ctx, `
INSERT INTO site_call_sessions (
  call_session_id, company_id, project_id, chat_widget_id, visitor_id, visit_session_id,
  conversation_id, direction, requested_by, requested_by_user_id, status, recording_enabled,
  media_kind, last_signal_at, metadata_json
) VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::uuid,NULLIF($7,'')::uuid,$8,$9,$10,$11,$12,$13,now(),$14::jsonb)`,
		callID, scope.CompanyID, scope.ProjectID, scope.ChatWidgetID, scope.VisitorID, scope.VisitSessionID,
		scope.ConversationID, direction, requestedBy, uidAny, status, recordingEnabled, mediaKind,
		jsonBytes(map[string]any{"source": "live_assist"}))
	if err != nil {
		return nil, err
	}
	_, _ = tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET visitor_state_snapshot =
     jsonb_set(
       jsonb_set(visitor_state_snapshot, '{activity,call_status}', to_jsonb($2::text), true),
       '{activity,call_session_id}', to_jsonb($3::text), true
     )
 WHERE visit_session_id = $1`, scope.VisitSessionID, status, callID)
	if scope.ConversationID != "" {
		if err := insertSystemMessageTx(ctx, tx, scope.ConversationID, scope.CompanyID, scope.ProjectID, message, map[string]any{
			"event_type":        "call.request",
			"call_session_id":   callID,
			"direction":         direction,
			"requested_by":      requestedBy,
			"recording_enabled": recordingEnabled,
			"media_kind":        mediaKind,
		}); err != nil {
			return nil, err
		}
	}
	return s.GetCall(ctx, uuid.MustParse(callID))
}

func (s *SiteChat) AcceptCallTx(ctx context.Context, callID uuid.UUID, actor string, in CallInput) (*CallSessionRow, error) {
	call, err := s.GetCall(ctx, callID)
	if err != nil {
		return nil, err
	}
	if actor == "visitor" {
		if err := s.ensurePublicCallInput(ctx, call, in); err != nil {
			return nil, err
		}
		if !in.PrivacyConsentAccepted {
			return nil, ErrBadInput
		}
		if call.RecordingEnabled && !in.RecordingConsentAccepted {
			return nil, ErrBadInput
		}
	} else if err := authz.RequireProject(ctx, s.P, uuid.MustParse(call.ProjectID), authz.PermWrite); err != nil {
		return nil, err
	}
	if err := s.requireCallableVisitor(ctx, callLiveScope(call)); err != nil {
		return nil, err
	}
	_, err = s.P.Exec(ctx, `
UPDATE site_call_sessions
   SET status = 'accepted',
       started_at = COALESCE(started_at, now()),
       privacy_consent_accepted_at = CASE WHEN $2 THEN COALESCE(privacy_consent_accepted_at, now()) ELSE privacy_consent_accepted_at END,
       recording_consent_accepted_at = CASE WHEN $3 THEN COALESCE(recording_consent_accepted_at, now()) ELSE recording_consent_accepted_at END,
       last_signal_at = now()
 WHERE call_session_id = $1`, callID.String(), in.PrivacyConsentAccepted, in.RecordingConsentAccepted)
	if err != nil {
		return nil, err
	}
	_ = s.updateCallActivity(ctx, call, "accepted")
	_ = s.writeCallSystemMessage(ctx, callID, "Звонок принят", "call.accept")
	return s.GetCall(ctx, callID)
}

func (s *SiteChat) RejectCallTx(ctx context.Context, callID uuid.UUID, actor string, in CallInput) (*CallSessionRow, error) {
	call, err := s.GetCall(ctx, callID)
	if err != nil {
		return nil, err
	}
	if actor == "visitor" {
		if err := s.ensurePublicCallInput(ctx, call, in); err != nil {
			return nil, err
		}
	} else if err := authz.RequireProject(ctx, s.P, uuid.MustParse(call.ProjectID), authz.PermWrite); err != nil {
		return nil, err
	}
	_, err = s.P.Exec(ctx, `
UPDATE site_call_sessions
   SET status = 'rejected', ended_at = COALESCE(ended_at, now()), end_reason = COALESCE(NULLIF($2,''), 'rejected'), last_signal_at = now()
 WHERE call_session_id = $1`, callID.String(), trimLimit(in.EndReason, 200))
	if err != nil {
		return nil, err
	}
	_ = s.updateCallActivity(ctx, call, "rejected")
	_ = s.writeCallSystemMessage(ctx, callID, "Звонок отклонен", "call.reject")
	return s.GetCall(ctx, callID)
}

func (s *SiteChat) EndCallTx(ctx context.Context, callID uuid.UUID, in CallInput) (*CallSessionRow, error) {
	call, err := s.GetCall(ctx, callID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, uuid.MustParse(call.ProjectID), authz.PermWrite); err != nil {
		return nil, err
	}
	_, err = s.P.Exec(ctx, `
UPDATE site_call_sessions
   SET status = 'ended', ended_at = COALESCE(ended_at, now()), end_reason = COALESCE(NULLIF($2,''), 'ended'), last_signal_at = now()
 WHERE call_session_id = $1`, callID.String(), trimLimit(in.EndReason, 200))
	if err != nil {
		return nil, err
	}
	_ = s.updateCallActivity(ctx, call, "ended")
	_ = s.writeCallSystemMessage(ctx, callID, "Звонок завершен", "call.end")
	return s.GetCall(ctx, callID)
}

func (s *SiteChat) PublicEndCallTx(ctx context.Context, callID uuid.UUID, in CallInput) (*CallSessionRow, error) {
	call, err := s.GetCall(ctx, callID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePublicCallInput(ctx, call, in); err != nil {
		return nil, err
	}
	_, err = s.P.Exec(ctx, `
UPDATE site_call_sessions
   SET status = 'ended', ended_at = COALESCE(ended_at, now()), end_reason = COALESCE(NULLIF($2,''), 'visitor_ended'), last_signal_at = now()
 WHERE call_session_id = $1`, callID.String(), trimLimit(in.EndReason, 200))
	if err != nil {
		return nil, err
	}
	_ = s.updateCallActivity(ctx, call, "ended")
	_ = s.writeCallSystemMessage(ctx, callID, "Звонок завершен посетителем", "call.end")
	return s.GetCall(ctx, callID)
}

func (s *SiteChat) MarkCallSignal(ctx context.Context, callID string) error {
	if _, err := uuid.Parse(strings.TrimSpace(callID)); err != nil {
		return ErrBadInput
	}
	_, err := s.P.Exec(ctx, `UPDATE site_call_sessions SET last_signal_at = now() WHERE call_session_id = $1`, strings.TrimSpace(callID))
	return err
}

func (s *SiteChat) requireCallableVisitor(ctx context.Context, scope LiveAssistScope) error {
	if strings.TrimSpace(scope.VisitSessionID) == "" {
		return ErrVisitorOffline
	}
	var status, connectionStatus string
	err := s.P.QueryRow(ctx, `
SELECT status::text, COALESCE(visitor_state_snapshot #>> '{activity,connection_status}', '')
  FROM site_visit_sessions
 WHERE visit_session_id = $1
   AND project_id = $2`, strings.TrimSpace(scope.VisitSessionID), strings.TrimSpace(scope.ProjectID)).
		Scan(&status, &connectionStatus)
	if err != nil {
		return mapErr(err)
	}
	if status != "active" || connectionStatus == "offline" {
		return ErrVisitorOffline
	}
	active, err := s.HasActiveVisitorConnections(ctx, scope, "")
	if err != nil {
		return err
	}
	if !active {
		return ErrVisitorOffline
	}
	return nil
}

func callLiveScope(call *CallSessionRow) LiveAssistScope {
	if call == nil {
		return LiveAssistScope{}
	}
	return LiveAssistScope{
		CompanyID:      call.CompanyID,
		ProjectID:      call.ProjectID,
		ChatWidgetID:   call.ChatWidgetID,
		VisitorID:      call.VisitorID,
		VisitSessionID: call.VisitSessionID,
		ConversationID: call.ConversationID,
	}
}

func (s *SiteChat) endOpenCallsForSession(ctx context.Context, visitSessionID, reason string) error {
	_, err := s.P.Exec(ctx, `
UPDATE site_call_sessions
   SET status = 'ended',
       ended_at = COALESCE(ended_at, now()),
       end_reason = COALESCE(NULLIF(end_reason, ''), $2),
       last_signal_at = now()
 WHERE visit_session_id = $1
   AND status IN ('ringing', 'offered', 'accepted')`, strings.TrimSpace(visitSessionID), trimLimit(reason, 200))
	return err
}

func (s *SiteChat) updateCallActivity(ctx context.Context, call *CallSessionRow, status string) error {
	if call == nil || strings.TrimSpace(call.VisitSessionID) == "" {
		return nil
	}
	_, err := s.P.Exec(ctx, `
UPDATE site_visit_sessions
   SET visitor_state_snapshot =
     jsonb_set(
       jsonb_set(visitor_state_snapshot, '{activity,call_status}', to_jsonb($2::text), true),
       '{activity,call_session_id}', to_jsonb($3::text), true
     )
 WHERE visit_session_id = $1`, call.VisitSessionID, status, call.CallSessionID)
	return err
}

func (s *SiteChat) GetCall(ctx context.Context, callID uuid.UUID) (*CallSessionRow, error) {
	row := s.P.QueryRow(ctx, `
SELECT call_session_id::text, company_id::text, project_id::text, chat_widget_id::text,
       visitor_id::text, COALESCE(visit_session_id::text,''), COALESCE(conversation_id::text,''),
       direction::text, requested_by::text, COALESCE(requested_by_user_id::text,''), status::text,
       recording_enabled, media_kind::text, privacy_consent_accepted_at, recording_consent_accepted_at,
       started_at, ended_at, last_signal_at, end_reason::text, metadata_json, created_at, updated_at
  FROM site_call_sessions
 WHERE call_session_id = $1`, callID.String())
	var out CallSessionRow
	if err := row.Scan(&out.CallSessionID, &out.CompanyID, &out.ProjectID, &out.ChatWidgetID, &out.VisitorID,
		&out.VisitSessionID, &out.ConversationID, &out.Direction, &out.RequestedBy, &out.RequestedByUserID,
		&out.Status, &out.RecordingEnabled, &out.MediaKind, &out.PrivacyConsentAcceptedAt,
		&out.RecordingConsentAcceptedAt, &out.StartedAt, &out.EndedAt, &out.LastSignalAt,
		&out.EndReason, (*jsonMap)(&out.Metadata), &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &out, nil
}

func (s *SiteChat) Snapshot(ctx context.Context, sessionID uuid.UUID) (map[string]any, error) {
	projectID, err := s.sessionProject(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	var out map[string]any
	err = s.P.QueryRow(ctx, `
SELECT visitor_state_snapshot
  FROM site_visit_sessions
 WHERE visit_session_id = $1`, sessionID.String()).Scan((*jsonMap)(&out))
	return out, mapErr(err)
}

func (s *SiteChat) CobrowseFrames(ctx context.Context, sessionID uuid.UUID, sinceSequence int64) ([]CobrowseFrameRow, error) {
	projectID, err := s.sessionProject(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := s.P.Query(ctx, `
SELECT frame_id::text, visit_session_id::text, sequence, kind::text, occurred_at, payload_json, created_at
  FROM site_cobrowse_frames
 WHERE visit_session_id = $1
   AND sequence > $2
 ORDER BY sequence ASC
 LIMIT 500`, sessionID.String(), sinceSequence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CobrowseFrameRow{}
	for rows.Next() {
		var item CobrowseFrameRow
		if err := rows.Scan(&item.FrameID, &item.VisitSessionID, &item.Sequence, &item.Kind, &item.OccurredAt, (*jsonMap)(&item.Payload), &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *SiteChat) RecordLiveEventsTx(ctx context.Context, scope LiveAssistScope, events []EventInput) error {
	if scope.VisitSessionID == "" || scope.VisitorID == "" {
		return ErrBadInput
	}
	tx, err := s.transaction()
	if err != nil {
		return err
	}
	for _, ev := range events {
		ev.EventType = strings.TrimSpace(ev.EventType)
		if ev.EventType == "" {
			continue
		}
		occurred := time.Now()
		if t, err := time.Parse(time.RFC3339, ev.OccurredAt); err == nil {
			occurred = t
		}
		priority := normalizePriority(ev.Priority)
		if _, err := tx.Exec(ctx, `
INSERT INTO site_visitor_events (
  company_id, project_id, chat_widget_id, visitor_id, visit_session_id, event_type, occurred_at, payload, priority
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)`,
			scope.CompanyID, scope.ProjectID, scope.ChatWidgetID, scope.VisitorID, scope.VisitSessionID,
			ev.EventType, occurred, jsonBytes(ev.Payload), priority); err != nil {
			return err
		}
		if cobrowseKind(ev.EventType) != "" {
			sequence := sequenceFromPayload(ev.Payload)
			if _, err := tx.Exec(ctx, `
INSERT INTO site_cobrowse_frames (
  company_id, project_id, chat_widget_id, visitor_id, visit_session_id, sequence, kind, occurred_at, payload_json
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb)
ON CONFLICT (visit_session_id, sequence) DO UPDATE SET payload_json = EXCLUDED.payload_json`,
				scope.CompanyID, scope.ProjectID, scope.ChatWidgetID, scope.VisitorID, scope.VisitSessionID,
				sequence, cobrowseKind(ev.EventType), occurred, jsonBytes(ev.Payload)); err != nil {
				return err
			}
		}
	}
	snapshot := map[string]any{}
	_ = tx.QueryRow(ctx, `
SELECT visitor_state_snapshot
  FROM site_visit_sessions
 WHERE visit_session_id = $1 AND visitor_id = $2`, scope.VisitSessionID, scope.VisitorID).Scan((*jsonMap)(&snapshot))
	applyEvents(snapshot, events)
	_, err = tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET last_activity_at = now(),
       event_counts = event_counts || $3::jsonb,
       visitor_state_snapshot = $4::jsonb
 WHERE visit_session_id = $1 AND visitor_id = $2`,
		scope.VisitSessionID, scope.VisitorID, jsonBytes(eventCounts(events)), jsonBytes(snapshot))
	if err != nil {
		return err
	}
	_, _ = tx.Exec(ctx, `UPDATE site_visitors SET last_seen_at = now() WHERE visitor_id = $1`, scope.VisitorID)
	return nil
}

func (s *SiteChat) ensureConversationForScope(ctx context.Context, scope LiveAssistScope) (string, error) {
	tx, err := s.transaction()
	if err != nil {
		return "", err
	}
	conversationID, err := ensureConversationForSessionTx(ctx, tx, widgetScope{
		ChatWidgetID: scope.ChatWidgetID,
		CompanyID:    scope.CompanyID,
		ProjectID:    scope.ProjectID,
	}, scope.VisitorID, scope.VisitSessionID, "", "")
	if err != nil {
		return "", err
	}
	return conversationID, nil
}

func (s *SiteChat) ensurePublicCallInput(ctx context.Context, call *CallSessionRow, in CallInput) error {
	if strings.TrimSpace(in.PublicKey) == "" {
		return ErrBadInput
	}
	scope, err := s.ResolvePublicScope(ctx, in.PublicKey, in.VisitorID, in.VisitSessionID, in.ConversationID)
	if err != nil {
		return err
	}
	if scope.VisitorID != call.VisitorID || (call.VisitSessionID != "" && scope.VisitSessionID != call.VisitSessionID) {
		return ErrNotFound
	}
	return nil
}

func (s *SiteChat) writeCallSystemMessage(ctx context.Context, callID uuid.UUID, body, eventType string) error {
	call, err := s.GetCall(ctx, callID)
	if err != nil || call.ConversationID == "" {
		return err
	}
	convID, _ := uuid.Parse(call.ConversationID)
	return s.insertSystemMessage(ctx, convID, body, map[string]any{
		"event_type":      eventType,
		"call_session_id": call.CallSessionID,
		"status":          call.Status,
	})
}

func (s *SiteChat) insertSystemMessage(ctx context.Context, conversationID uuid.UUID, body string, metadata map[string]any) error {
	var companyID, projectID string
	if err := s.P.QueryRow(ctx, `
SELECT company_id::text, project_id::text
  FROM site_chat_conversations
 WHERE conversation_id = $1`, conversationID.String()).Scan(&companyID, &projectID); err != nil {
		return mapErr(err)
	}
	tx, err := s.transaction()
	if err != nil {
		return err
	}
	if err := insertSystemMessageTx(ctx, tx, conversationID.String(), companyID, projectID, body, metadata); err != nil {
		return err
	}
	return nil
}

func insertSystemMessageTx(ctx context.Context, tx pgx.Tx, conversationID, companyID, projectID, body string, metadata map[string]any) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	_, err := tx.Exec(ctx, `
INSERT INTO site_chat_messages (
  message_id, conversation_id, company_id, project_id, author_kind, author_name, body, metadata
) VALUES ($1,$2,$3,$4,'system','Pulse',$5,$6::jsonb)`,
		uuid.NewString(), conversationID, companyID, projectID, body, jsonBytes(metadata))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE site_chat_conversations SET updated_at = now() WHERE conversation_id = $1`, conversationID)
	return err
}

func cobrowseKind(eventType string) string {
	switch strings.TrimSpace(eventType) {
	case "monitoring.snapshot":
		return "snapshot"
	case "monitoring.incremental":
		return "mutation"
	case "cursor.move", "cursor":
		return "cursor"
	case "touch.move", "touch":
		return "touch"
	case "scroll":
		return "scroll"
	case "viewport":
		return "viewport"
	case "page.navigate", "page":
		return "navigation"
	case "click":
		return "click"
	case "diagnostic":
		return "diagnostic"
	default:
		return ""
	}
}

func sequenceFromPayload(payload map[string]any) int64 {
	if payload != nil {
		switch v := payload["sequence"].(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		case json.Number:
			if n, err := v.Int64(); err == nil {
				return n
			}
		}
	}
	return time.Now().UnixNano()
}
