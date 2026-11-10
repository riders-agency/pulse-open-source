package repository

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
)

func (s *SiteChat) RecordPresenceTx(ctx context.Context, in PresenceInput) (*PresenceResponse, error) {
	in = normalizePresence(in)
	widget, err := s.lookupWidget(ctx, in.PublicKey)
	if err != nil {
		return nil, err
	}
	if in.AnonymousID == "" {
		in.AnonymousID = "anon_" + randomToken(18)
	}
	tx, err := s.transaction()
	if err != nil {
		return nil, err
	}

	visitorID := uuid.New()
	inserted := false
	row := tx.QueryRow(ctx, `
INSERT INTO site_visitors (
  visitor_id, company_id, project_id, chat_widget_id, anonymous_id, current_url, current_title,
  referrer, utm, browser, browser_version, os, device_type, screen_width, screen_height,
  language, timezone, metrika_client_id, metrika_user_id, metrika_session_id,
  user_agent_hash, ip_hash, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23::jsonb)
ON CONFLICT (chat_widget_id, anonymous_id) DO UPDATE SET
  last_seen_at = now(),
  current_url = EXCLUDED.current_url,
  current_title = EXCLUDED.current_title,
  referrer = COALESCE(NULLIF(EXCLUDED.referrer, ''), site_visitors.referrer),
  utm = CASE WHEN EXCLUDED.utm <> '{}'::jsonb THEN EXCLUDED.utm ELSE site_visitors.utm END,
  browser = COALESCE(NULLIF(EXCLUDED.browser, ''), site_visitors.browser),
  browser_version = COALESCE(NULLIF(EXCLUDED.browser_version, ''), site_visitors.browser_version),
  os = COALESCE(NULLIF(EXCLUDED.os, ''), site_visitors.os),
  device_type = COALESCE(NULLIF(EXCLUDED.device_type, ''), site_visitors.device_type),
  screen_width = COALESCE(EXCLUDED.screen_width, site_visitors.screen_width),
  screen_height = COALESCE(EXCLUDED.screen_height, site_visitors.screen_height),
  language = COALESCE(NULLIF(EXCLUDED.language, ''), site_visitors.language),
  timezone = COALESCE(NULLIF(EXCLUDED.timezone, ''), site_visitors.timezone),
  metrika_client_id = COALESCE(NULLIF(EXCLUDED.metrika_client_id, ''), site_visitors.metrika_client_id),
  metrika_user_id = COALESCE(NULLIF(EXCLUDED.metrika_user_id, ''), site_visitors.metrika_user_id),
  metrika_session_id = COALESCE(NULLIF(EXCLUDED.metrika_session_id, ''), site_visitors.metrika_session_id),
  metadata = site_visitors.metadata || EXCLUDED.metadata
RETURNING visitor_id::text, (xmax = 0) AS inserted`,
		visitorID.String(), widget.CompanyID, widget.ProjectID, widget.ChatWidgetID, in.AnonymousID,
		in.URL, in.Title, in.Referrer, jsonBytes(in.UTM), in.Browser, in.BrowserVersion, in.OS,
		in.DeviceType, nullableInt(in.ScreenWidth), nullableInt(in.ScreenHeight), in.Language, in.Timezone,
		in.MetrikaClientID, in.MetrikaUserID, in.MetrikaSessionID, hashString(in.UserAgent), hashIP(in.IP), jsonBytes(in.Metadata))
	var visitorIDStr string
	if err := row.Scan(&visitorIDStr, &inserted); err != nil {
		return nil, err
	}
	pageSessionID := presenceMetadataValue(in.Metadata, "page_session_id")
	sessionID := uuid.New()
	incomingSessionID, hasIncomingSession := uuid.Nil, false
	if sid, err := uuid.Parse(strings.TrimSpace(in.VisitSessionID)); err == nil {
		incomingSessionID = sid
		hasIncomingSession = true
	}
	if hasIncomingSession && pageSessionID != "" {
		var existingPageSessionID string
		err := tx.QueryRow(ctx, `
SELECT COALESCE(analytics_payload #>> '{metadata,page_session_id}', '')
  FROM site_visit_sessions
 WHERE visit_session_id = $1
   AND visitor_id = $2
   AND chat_widget_id = $3
   AND status = 'active'
   AND last_activity_at > now() - interval '30 minutes'
 FOR UPDATE`, incomingSessionID.String(), visitorIDStr, widget.ChatWidgetID).Scan(&existingPageSessionID)
		if err == nil && existingPageSessionID == pageSessionID {
			sessionID = incomingSessionID
		} else if err != nil && err != pgx.ErrNoRows {
			return nil, err
		} else {
			hasIncomingSession = false
		}
	} else if hasIncomingSession {
		sessionID = incomingSessionID
	}
	if !hasIncomingSession && pageSessionID != "" {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1)::bigint)`, "sitechat-session:"+widget.ChatWidgetID+":"+visitorIDStr+":"+pageSessionID); err != nil {
			return nil, err
		}
		var existingSessionID string
		err := tx.QueryRow(ctx, `
SELECT visit_session_id::text
  FROM site_visit_sessions
 WHERE visitor_id = $1
   AND chat_widget_id = $2
   AND status = 'active'
   AND COALESCE(analytics_payload #>> '{metadata,page_session_id}', '') = $3
   AND last_activity_at > now() - interval '30 minutes'
 ORDER BY last_activity_at DESC
 LIMIT 1
 FOR UPDATE`, visitorIDStr, widget.ChatWidgetID, pageSessionID).Scan(&existingSessionID)
		if err == nil {
			if sid, parseErr := uuid.Parse(existingSessionID); parseErr == nil {
				sessionID = sid
			}
		} else if err != pgx.ErrNoRows {
			return nil, err
		}
	}
	snapshot := buildSnapshot(visitorIDStr, sessionID.String(), widget.ChatWidgetID, in, inserted, false, false, "idle")
	_, err = tx.Exec(ctx, `
INSERT INTO site_visit_sessions (
  visit_session_id, visitor_id, company_id, project_id, chat_widget_id, current_url, current_title,
  viewport_width, viewport_height, scroll_x, scroll_y, device_type, browser, os, referrer, utm,
  analytics_payload, duration_seconds, last_duration_flush_at, visitor_state_snapshot
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb,$17::jsonb,$18,now(),$19::jsonb)
ON CONFLICT (visit_session_id) DO UPDATE SET
  last_activity_at = now(),
  status = 'active',
  ended_at = NULL,
  current_url = EXCLUDED.current_url,
  current_title = EXCLUDED.current_title,
  viewport_width = COALESCE(EXCLUDED.viewport_width, site_visit_sessions.viewport_width),
  viewport_height = COALESCE(EXCLUDED.viewport_height, site_visit_sessions.viewport_height),
  scroll_x = EXCLUDED.scroll_x,
  scroll_y = EXCLUDED.scroll_y,
  duration_seconds = GREATEST(site_visit_sessions.duration_seconds, EXCLUDED.duration_seconds),
  last_duration_flush_at = CASE
    WHEN EXCLUDED.duration_seconds > site_visit_sessions.duration_seconds THEN now()
    ELSE site_visit_sessions.last_duration_flush_at
  END,
  analytics_payload = site_visit_sessions.analytics_payload || EXCLUDED.analytics_payload`,
		sessionID.String(), visitorIDStr, widget.CompanyID, widget.ProjectID, widget.ChatWidgetID,
		in.URL, in.Title, nullableInt(in.ViewportWidth), nullableInt(in.ViewportHeight), in.ScrollX, in.ScrollY,
		in.DeviceType, in.Browser, in.OS, in.Referrer, jsonBytes(in.UTM), jsonBytes(analyticsPayload(in)),
		in.DurationSeconds, jsonBytes(snapshot))
	if err != nil {
		return nil, err
	}
	if pageSessionID != "" {
		_, _ = tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET status = 'ended',
       ended_at = COALESCE(ended_at, now()),
       visitor_state_snapshot =
         jsonb_set(
           jsonb_set(
             jsonb_set(visitor_state_snapshot, '{activity,is_active}', 'false'::jsonb, true),
             '{activity,connection_status}', '"offline"'::jsonb, true
           ),
           '{activity,offline_reason}', '"duplicate_page_session"'::jsonb, true
         )
 WHERE visitor_id = $1
   AND chat_widget_id = $2
   AND visit_session_id <> $3
   AND status = 'active'
   AND COALESCE(analytics_payload #>> '{metadata,page_session_id}', '') = $4`,
			visitorIDStr, widget.ChatWidgetID, sessionID.String(), pageSessionID)
	}
	var monitoringStatus string
	var operatorInvited, chatOpen bool
	var existingSnapshot map[string]any
	if err := tx.QueryRow(ctx, `
SELECT monitoring_status::text, operator_invited, chat_open, visitor_state_snapshot
  FROM site_visit_sessions
 WHERE visit_session_id = $1
 FOR UPDATE`, sessionID.String()).Scan(&monitoringStatus, &operatorInvited, &chatOpen, (*jsonMap)(&existingSnapshot)); err != nil {
		return nil, err
	}
	snapshot = mergePresenceSnapshot(existingSnapshot, buildSnapshot(visitorIDStr, sessionID.String(), widget.ChatWidgetID, in, inserted, operatorInvited, chatOpen, monitoringStatus))
	conversationID, err := ensureConversationForSessionTx(ctx, tx, widget, visitorIDStr, sessionID.String(), "", "")
	if err != nil {
		return nil, err
	}
	links := siteChatLinks(widget.ProjectID, visitorIDStr, sessionID.String(), conversationID)
	enrichSnapshotWithConversation(snapshot, conversationID, links)
	if _, err := tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET visitor_state_snapshot = $2::jsonb
 WHERE visit_session_id = $1`, sessionID.String(), jsonBytes(snapshot)); err != nil {
		return nil, err
	}
	if inserted {
		if err := enqueueFirstSeen(ctx, tx, widget, visitorIDStr, sessionID.String(), snapshot); err != nil {
			slog.Warn("site chat first_seen workflow enqueue failed",
				slog.String("project_id", widget.ProjectID),
				slog.String("chat_widget_id", widget.ChatWidgetID),
				slog.String("visitor_id", visitorIDStr),
				slog.String("visit_session_id", sessionID.String()),
				slog.String("error", err.Error()),
			)
		}
	}
	return &PresenceResponse{
		VisitorID:            visitorIDStr,
		VisitSessionID:       sessionID.String(),
		ChatWidgetID:         widget.ChatWidgetID,
		ProjectID:            widget.ProjectID,
		ConversationID:       conversationID,
		ConversationURL:      links.ConversationURL,
		PanelURL:             links.PanelURL,
		CallURL:              links.CallURL,
		IsFirstVisit:         inserted,
		MonitoringStatus:     monitoringStatus,
		Recorder:             map[string]any{"enabled": monitoringStatus == "active", "mode": "snapshot"},
		OperatorInvited:      operatorInvited,
		VisitorStateSnapshot: snapshot,
	}, nil
}

func (s *SiteChat) RecordEventsTx(ctx context.Context, in EventBatchInput) error {
	widget, err := s.lookupWidget(ctx, in.PublicKey)
	if err != nil {
		return err
	}
	visitorID, err := uuid.Parse(in.VisitorID)
	if err != nil {
		return ErrBadInput
	}
	sessionID, err := uuid.Parse(in.VisitSessionID)
	if err != nil {
		return ErrBadInput
	}
	tx, err := s.transaction()
	if err != nil {
		return err
	}
	for _, ev := range in.Events {
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
			widget.CompanyID, widget.ProjectID, widget.ChatWidgetID, visitorID.String(), sessionID.String(),
			ev.EventType, occurred, jsonBytes(ev.Payload), priority); err != nil {
			return err
		}
		if kind := cobrowseKind(ev.EventType); kind != "" {
			sequence := sequenceFromPayload(ev.Payload)
			if _, err := tx.Exec(ctx, `
INSERT INTO site_cobrowse_frames (
  company_id, project_id, chat_widget_id, visitor_id, visit_session_id, sequence, kind, occurred_at, payload_json
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb)
ON CONFLICT (visit_session_id, sequence) DO UPDATE SET payload_json = EXCLUDED.payload_json`,
				widget.CompanyID, widget.ProjectID, widget.ChatWidgetID, visitorID.String(), sessionID.String(),
				sequence, kind, occurred, jsonBytes(ev.Payload)); err != nil {
				return err
			}
		}
	}
	snapshot := map[string]any{}
	_ = tx.QueryRow(ctx, `
SELECT visitor_state_snapshot
  FROM site_visit_sessions
 WHERE visit_session_id = $1 AND visitor_id = $2`, sessionID.String(), visitorID.String()).Scan((*jsonMap)(&snapshot))
	applyEvents(snapshot, in.Events)
	_, err = tx.Exec(ctx, `
UPDATE site_visit_sessions
   SET last_activity_at = now(),
       event_counts = event_counts || $3::jsonb,
       visitor_state_snapshot = $4::jsonb
 WHERE visit_session_id = $1 AND visitor_id = $2`,
		sessionID.String(), visitorID.String(), jsonBytes(eventCounts(in.Events)), jsonBytes(snapshot))
	if err != nil {
		return err
	}
	_, _ = tx.Exec(ctx, `UPDATE site_visitors SET last_seen_at = now() WHERE visitor_id = $1`, visitorID.String())
	return nil
}

func (s *SiteChat) ActiveVisitors(ctx context.Context, projectID uuid.UUID) ([]VisitorSessionRow, error) {
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	if err := s.expireStaleVisitorSessions(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.P.Query(ctx, `
	WITH session_rows AS (
	  SELECT s.*,
	         NULLIF(s.analytics_payload #>> '{metadata,page_session_id}', '') AS page_session_key,
	         CASE WHEN EXISTS (
	           SELECT 1
	             FROM site_visit_connections vc
	            WHERE vc.visit_session_id = s.visit_session_id
	              AND vc.role = 'visitor'
	              AND vc.status = 'connected'
	              AND vc.last_heartbeat_at > now() - interval '45 seconds'
	         ) THEN 1 ELSE 0 END AS connection_rank,
	         COALESCE(conv.conversation_id::text, '') AS conversation_id,
	         COALESCE(conv.visitor_name::text, '') AS visitor_name,
	         COALESCE(conv.visitor_contact::text, '') AS visitor_contact,
	         COALESCE(conv.visitor_message_count, 0)::int AS visitor_message_count,
	         conv.last_message_at
    FROM site_visit_sessions s
    LEFT JOIN LATERAL (
      SELECT c.conversation_id,
             COALESCE(NULLIF(c.visitor_name, ''), latest_visitor.author_name, '') AS visitor_name,
             c.visitor_contact,
             COUNT(m.message_id) FILTER (WHERE m.author_kind = 'visitor' AND btrim(lower(m.body)) NOT IN ('запись звонка', 'запись разговора')) AS visitor_message_count,
             MAX(m.created_at) AS last_message_at,
             c.created_at
        FROM site_chat_conversations c
        LEFT JOIN site_chat_messages m ON m.conversation_id = c.conversation_id
        LEFT JOIN LATERAL (
          SELECT NULLIF(author_name, '') AS author_name
             FROM site_chat_messages lm
           WHERE lm.conversation_id = c.conversation_id
             AND lm.author_kind = 'visitor'
             AND btrim(lower(lm.body)) NOT IN ('запись звонка', 'запись разговора')
           ORDER BY lm.created_at DESC
           LIMIT 1
        ) latest_visitor ON true
       WHERE c.visit_session_id = s.visit_session_id
       GROUP BY c.conversation_id, c.visitor_name, c.visitor_contact, latest_visitor.author_name, c.created_at
       ORDER BY COUNT(m.message_id) FILTER (WHERE m.author_kind = 'visitor' AND btrim(lower(m.body)) NOT IN ('запись звонка', 'запись разговора')) DESC,
                MAX(m.created_at) DESC NULLS LAST,
                c.created_at ASC
       LIMIT 1
    ) conv ON true
   WHERE s.project_id = $1
     AND (
	       (
	         s.status = 'active'
	         AND COALESCE(s.visitor_state_snapshot #>> '{activity,connection_status}', 'online') <> 'offline'
	         AND (
           EXISTS (
             SELECT 1
               FROM site_visit_connections vc
              WHERE vc.visit_session_id = s.visit_session_id
                AND vc.role = 'visitor'
                AND vc.status = 'connected'
                AND vc.last_heartbeat_at > now() - interval '45 seconds'
           )
	           OR s.last_activity_at > now() - interval '75 seconds'
	         )
	       )
	       OR COALESCE(conv.visitor_message_count, 0) > 0
	     )
	),
filtered_sessions AS (
  SELECT s.*
    FROM session_rows s
   WHERE NOT (
     s.visitor_message_count = 0
     AND s.page_session_key IS NULL
     AND EXISTS (
       SELECT 1
         FROM session_rows newer
        WHERE newer.visitor_id = s.visitor_id
          AND newer.chat_widget_id = s.chat_widget_id
          AND newer.page_session_key IS NOT NULL
          AND newer.current_url IS NOT DISTINCT FROM s.current_url
          AND newer.last_activity_at >= s.last_activity_at - interval '2 minutes'
     )
   )
),
ranked_sessions AS (
  SELECT s.*,
         row_number() OVER (
           PARTITION BY s.visitor_id, s.chat_widget_id,
             CASE
               WHEN s.visitor_message_count > 0 THEN COALESCE(NULLIF(s.conversation_id, ''), s.visit_session_id::text)
               ELSE COALESCE(s.page_session_key, NULLIF(s.current_url, ''), s.visit_session_id::text)
             END
           ORDER BY s.connection_rank DESC, s.last_activity_at DESC, s.visit_session_id DESC
         ) AS rn
    FROM filtered_sessions s
)
SELECT v.visitor_id::text, s.visit_session_id::text, s.chat_widget_id::text, s.project_id::text,
       s.conversation_id,
       s.visitor_name,
       s.visitor_contact,
       s.current_url::text, s.current_title::text, s.status::text, s.started_at, s.ended_at,
       s.last_activity_at, s.duration_seconds, s.visitor_message_count,
       s.monitoring_status::text, s.operator_invited, s.chat_open, s.visitor_state_snapshot
  FROM ranked_sessions s
  JOIN site_visitors v ON v.visitor_id = s.visitor_id
 WHERE s.rn = 1
 ORDER BY s.connection_rank DESC, COALESCE(s.last_message_at, s.last_activity_at) DESC, s.last_activity_at DESC
 LIMIT 100`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []VisitorSessionRow{}
	for rows.Next() {
		var item VisitorSessionRow
		if err := rows.Scan(&item.VisitorID, &item.VisitSessionID, &item.ChatWidgetID, &item.ProjectID,
			&item.ConversationID, &item.VisitorName, &item.VisitorContact, &item.CurrentURL, &item.CurrentTitle, &item.Status, &item.StartedAt, &item.EndedAt,
			&item.LastActivityAt, &item.DurationSeconds, &item.VisitorMessageCount, &item.MonitoringStatus,
			&item.OperatorInvited, &item.ChatOpen, (*jsonMap)(&item.VisitorStateSnapshot)); err != nil {
			return nil, err
		}
		links := siteChatLinks(item.ProjectID, item.VisitorID, item.VisitSessionID, item.ConversationID)
		item.ConversationURL = links.ConversationURL
		item.PanelURL = links.PanelURL
		item.CallURL = links.CallURL
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *SiteChat) expireStaleVisitorSessions(ctx context.Context, projectID uuid.UUID) error {
	_, err := s.P.Exec(ctx, `
WITH stale AS (
  SELECT s.visit_session_id
    FROM site_visit_sessions s
   WHERE s.project_id = $1
     AND s.status = 'active'
     AND s.last_activity_at < now() - interval '90 seconds'
     AND NOT EXISTS (
       SELECT 1
         FROM site_visit_connections vc
        WHERE vc.visit_session_id = s.visit_session_id
          AND vc.role = 'visitor'
          AND vc.status = 'connected'
          AND vc.last_heartbeat_at > now() - interval '45 seconds'
     )
),
expired_connections AS (
  UPDATE site_visit_connections vc
     SET status = 'timed_out',
         closed_at = COALESCE(closed_at, now()),
         close_reason = COALESCE(NULLIF(close_reason, ''), 'stale_session_ttl')
    FROM stale
   WHERE vc.visit_session_id = stale.visit_session_id
     AND vc.role = 'visitor'
     AND vc.status = 'connected'
   RETURNING vc.visit_session_id
)
UPDATE site_visit_sessions s
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
  FROM stale
 WHERE s.visit_session_id = stale.visit_session_id`, projectID.String())
	return err
}

func (s *SiteChat) SetMonitoring(ctx context.Context, sessionID uuid.UUID, status string) error {
	projectID, err := s.sessionProject(ctx, sessionID)
	if err != nil {
		return err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return err
	}
	if status != "active" && status != "stopped" && status != "requested" {
		status = "active"
	}
	_, err = s.P.Exec(ctx, `
UPDATE site_visit_sessions
   SET monitoring_status = $2,
       visitor_state_snapshot = jsonb_set(
         jsonb_set(visitor_state_snapshot, '{signals,monitoring_active}', to_jsonb($2 = 'active'), true),
         '{monitoring_status}', to_jsonb($2::text), true
       )
 WHERE visit_session_id = $1`, sessionID.String(), status)
	return err
}
