package repository

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
)

func (s *SiteChat) InviteTx(ctx context.Context, sessionID, uid uuid.UUID) (*ConversationRow, error) {
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
	convID, _ := uuid.Parse(conversationID)
	return s.GetConversation(ctx, convID)
}

func (s *SiteChat) GetConversation(ctx context.Context, conversationID uuid.UUID) (*ConversationRow, error) {
	projectID, err := s.conversationProject(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	return s.getConversationNoAuth(ctx, conversationID)
}

func (s *SiteChat) Messages(ctx context.Context, conversationID uuid.UUID) ([]MessageRow, error) {
	projectID, err := s.conversationProject(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	return s.messagesNoAuth(ctx, conversationID)
}

func (s *SiteChat) OperatorMessageTx(ctx context.Context, conversationID, uid uuid.UUID, body string, attachmentIDs []string) (string, error) {
	projectID, err := s.conversationProject(ctx, conversationID)
	if err != nil {
		return "", err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return "", err
	}
	return s.insertMessageWithAttachments(ctx, conversationID, "operator", uid, "Оператор", body, attachmentIDs)
}

func (s *SiteChat) PublicConversationTx(ctx context.Context, in MessageInput) (*ConversationRow, error) {
	widget, err := s.lookupWidget(ctx, in.PublicKey)
	if err != nil {
		return nil, err
	}
	visitorID, err := uuid.Parse(in.VisitorID)
	if err != nil {
		return nil, ErrBadInput
	}
	sessionID, _ := uuid.Parse(strings.TrimSpace(in.VisitSessionID))
	tx, err := s.transaction()
	if err != nil {
		return nil, err
	}
	conversationID, err := ensureConversationForSessionTx(ctx, tx, widget, visitorID.String(), uuidOrEmpty(sessionID), strings.TrimSpace(in.AuthorName), strings.TrimSpace(in.VisitorContact))
	if err != nil {
		return nil, err
	}
	convUUID, _ := uuid.Parse(conversationID)
	if strings.TrimSpace(in.Body) != "" || len(in.AttachmentIDs) > 0 {
		_, _ = s.insertMessageWithAttachments(ctx, convUUID, "visitor", uuid.Nil, strings.TrimSpace(in.AuthorName), in.Body, in.AttachmentIDs)
	}
	return s.getConversationNoAuth(ctx, convUUID)
}

func (s *SiteChat) PublicMessageTx(ctx context.Context, in MessageInput) (string, error) {
	widget, err := s.lookupWidget(ctx, in.PublicKey)
	if err != nil {
		return "", err
	}
	conversationID, err := uuid.Parse(in.ConversationID)
	if err != nil {
		return "", ErrBadInput
	}
	if err := s.ensurePublicConversation(ctx, widget, conversationID, strings.TrimSpace(in.VisitorID)); err != nil {
		return "", err
	}
	if err := s.updateConversationVisitorIdentity(ctx, conversationID.String(), in.AuthorName, in.VisitorContact); err != nil {
		return "", err
	}
	return s.insertMessageWithAttachments(ctx, conversationID, "visitor", uuid.Nil, strings.TrimSpace(in.AuthorName), in.Body, in.AttachmentIDs)
}

func (s *SiteChat) PublicMessages(ctx context.Context, publicKey string, conversationID uuid.UUID, visitorID string) ([]MessageRow, error) {
	widget, err := s.lookupWidget(ctx, publicKey)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePublicConversation(ctx, widget, conversationID, strings.TrimSpace(visitorID)); err != nil {
		return nil, err
	}
	return s.messagesNoAuth(ctx, conversationID)
}

type widgetScope struct {
	ChatWidgetID string
	CompanyID    string
	ProjectID    string
}

type siteChatLinkSet struct {
	ConversationURL string `json:"conversation_url"`
	PanelURL        string `json:"panel_url"`
	CallURL         string `json:"call_url"`
}

func ensureConversationForSessionTx(ctx context.Context, tx pgx.Tx, widget widgetScope, visitorID, sessionID, visitorName, visitorContact string) (string, error) {
	visitorID = strings.TrimSpace(visitorID)
	sessionID = strings.TrimSpace(sessionID)
	visitorName = strings.TrimSpace(visitorName)
	visitorContact = strings.TrimSpace(visitorContact)
	if visitorID == "" {
		return "", ErrBadInput
	}
	var existing string
	if sessionID != "" {
		if _, err := tx.Exec(ctx, `
SELECT 1
  FROM site_visit_sessions
 WHERE visit_session_id = $1::uuid
 FOR UPDATE`, sessionID); err != nil {
			return "", err
		}
		err := tx.QueryRow(ctx, `
SELECT conversation_id::text
  FROM site_chat_conversations
 WHERE visit_session_id = $1::uuid
 ORDER BY created_at ASC
 LIMIT 1`, sessionID).Scan(&existing)
		if err == nil {
			if visitorName != "" || visitorContact != "" {
				_, _ = tx.Exec(ctx, `
UPDATE site_chat_conversations
   SET visitor_name = COALESCE(NULLIF($2,''), visitor_name),
       visitor_contact = COALESCE(NULLIF($3,''), visitor_contact)
 WHERE conversation_id = $1`, existing, visitorName, visitorContact)
			}
			return existing, nil
		}
		if err != pgx.ErrNoRows {
			return "", err
		}
	}
	conversationID := uuid.NewString()
	title := "Посетитель на сайте"
	_, err := tx.Exec(ctx, `
INSERT INTO site_chat_conversations (
  conversation_id, company_id, project_id, chat_widget_id, visitor_id, visit_session_id,
  title, visitor_name, visitor_contact
) VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::uuid,$7,$8,$9)`,
		conversationID, widget.CompanyID, widget.ProjectID, widget.ChatWidgetID, visitorID, sessionID, title, visitorName, visitorContact)
	if err != nil {
		return "", err
	}
	return conversationID, nil
}

func (s *SiteChat) ensurePublicConversation(ctx context.Context, widget widgetScope, conversationID uuid.UUID, visitorID string) error {
	var found int
	args := []any{conversationID.String(), widget.ChatWidgetID}
	filter := ""
	if visitorID != "" {
		if _, err := uuid.Parse(visitorID); err != nil {
			return ErrBadInput
		}
		filter = " AND visitor_id = $3::uuid"
		args = append(args, visitorID)
	}
	err := s.P.QueryRow(ctx, `
SELECT COUNT(*)::int
  FROM site_chat_conversations
 WHERE conversation_id = $1
   AND chat_widget_id = $2`+filter, args...).Scan(&found)
	if err != nil {
		return err
	}
	if found == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SiteChat) updateConversationVisitorIdentity(ctx context.Context, conversationID, visitorName, visitorContact string) error {
	visitorName = strings.TrimSpace(visitorName)
	visitorContact = strings.TrimSpace(visitorContact)
	if visitorName == "" && visitorContact == "" {
		return nil
	}
	_, err := s.P.Exec(ctx, `
UPDATE site_chat_conversations
   SET visitor_name = COALESCE(NULLIF($2,''), visitor_name),
       visitor_contact = COALESCE(NULLIF($3,''), visitor_contact)
 WHERE conversation_id = $1`, strings.TrimSpace(conversationID), visitorName, visitorContact)
	return err
}

func siteChatLinks(projectID, visitorID, sessionID, conversationID string) siteChatLinkSet {
	base := "https://pulse.riders.agency/site-chat"
	q := make([]string, 0, 4)
	if projectID != "" {
		q = append(q, "project_id="+neturlEscape(projectID))
	}
	if visitorID != "" {
		q = append(q, "visitor_id="+neturlEscape(visitorID))
	}
	if sessionID != "" {
		q = append(q, "visit_session_id="+neturlEscape(sessionID))
	}
	if conversationID != "" {
		q = append(q, "conversation_id="+neturlEscape(conversationID))
	}
	return siteChatLinkSet{
		ConversationURL: siteChatURL(base, q, ""),
		PanelURL:        siteChatURL(base, q, "panel"),
		CallURL:         siteChatURL(base, q, "call"),
	}
}

func siteChatURL(base string, params []string, mode string) string {
	q := append([]string{}, params...)
	if mode != "" {
		q = append(q, "mode="+neturlEscape(mode))
	}
	if len(q) == 0 {
		return base
	}
	return base + "?" + strings.Join(q, "&")
}

func enrichSnapshotWithConversation(snapshot map[string]any, conversationID string, links siteChatLinkSet) {
	if snapshot == nil {
		return
	}
	snapshot["conversation_id"] = conversationID
	snapshot["conversation_url"] = links.ConversationURL
	snapshot["panel_url"] = links.PanelURL
	snapshot["call_url"] = links.CallURL
	snapshot["links"] = map[string]any{
		"conversation_link": map[string]any{
			"label": "Открыть чат с посетителем",
			"href":  links.ConversationURL,
		},
		"visitor_panel_link": map[string]any{
			"label": "Панель посетителя",
			"href":  links.PanelURL,
		},
		"call_link": map[string]any{
			"label": "Звонок",
			"href":  links.CallURL,
		},
	}
}

func neturlEscape(v string) string {
	return url.QueryEscape(v)
}

func (s *SiteChat) projectCompany(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := s.P.QueryRow(ctx, `SELECT company_id FROM pulse_read.projects_v1 WHERE project_id = $1`, projectID.String()).Scan(&companyID)
	return companyID, mapErr(err)
}

func (s *SiteChat) widgetProject(ctx context.Context, widgetID uuid.UUID) (uuid.UUID, error) {
	var projectID uuid.UUID
	err := s.P.QueryRow(ctx, `SELECT project_id FROM site_chat_widgets WHERE chat_widget_id = $1`, widgetID.String()).Scan(&projectID)
	return projectID, mapErr(err)
}

func (s *SiteChat) sessionProject(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	var projectID uuid.UUID
	err := s.P.QueryRow(ctx, `SELECT project_id FROM site_visit_sessions WHERE visit_session_id = $1`, sessionID.String()).Scan(&projectID)
	return projectID, mapErr(err)
}

func (s *SiteChat) conversationProject(ctx context.Context, conversationID uuid.UUID) (uuid.UUID, error) {
	var projectID uuid.UUID
	err := s.P.QueryRow(ctx, `SELECT project_id FROM site_chat_conversations WHERE conversation_id = $1`, conversationID.String()).Scan(&projectID)
	return projectID, mapErr(err)
}

func (s *SiteChat) lookupWidget(ctx context.Context, publicKey string) (widgetScope, error) {
	var out widgetScope
	err := s.P.QueryRow(ctx, `
SELECT chat_widget_id::text, company_id::text, project_id::text
  FROM site_chat_widgets
 WHERE public_key = $1 AND status = 'active'`, strings.TrimSpace(publicKey)).
		Scan(&out.ChatWidgetID, &out.CompanyID, &out.ProjectID)
	return out, mapErr(err)
}

func (s *SiteChat) getConversationNoAuth(ctx context.Context, conversationID uuid.UUID) (*ConversationRow, error) {
	row := s.P.QueryRow(ctx, `
SELECT conversation_id::text, project_id::text, chat_widget_id::text, visitor_id::text,
       COALESCE(visit_session_id::text,''), status::text, title::text, visitor_name::text,
       visitor_contact::text, created_at, updated_at
  FROM site_chat_conversations
 WHERE conversation_id = $1`, conversationID.String())
	var out ConversationRow
	if err := row.Scan(&out.ConversationID, &out.ProjectID, &out.ChatWidgetID, &out.VisitorID,
		&out.VisitSessionID, &out.Status, &out.Title, &out.VisitorName, &out.VisitorContact,
		&out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	links := siteChatLinks(out.ProjectID, out.VisitorID, out.VisitSessionID, out.ConversationID)
	out.OperatorURL = links.ConversationURL
	out.CallURL = links.CallURL
	return &out, nil
}

func (s *SiteChat) messagesNoAuth(ctx context.Context, conversationID uuid.UUID) ([]MessageRow, error) {
	rows, err := s.P.Query(ctx, `
SELECT message_id::text, conversation_id::text, author_kind::text, COALESCE(author_user_id::text,''),
       author_name::text, body::text, created_at
  FROM site_chat_messages
 WHERE conversation_id = $1
 ORDER BY created_at ASC`, conversationID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MessageRow{}
	messageIDs := []string{}
	for rows.Next() {
		var item MessageRow
		if err := rows.Scan(&item.MessageID, &item.ConversationID, &item.AuthorKind, &item.AuthorUserID,
			&item.AuthorName, &item.Body, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
		messageIDs = append(messageIDs, item.MessageID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	attachments, err := s.attachmentsForMessages(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Attachments = attachments[out[i].MessageID]
	}
	return out, nil
}

func (s *SiteChat) insertMessage(ctx context.Context, conversationID uuid.UUID, authorKind string, uid uuid.UUID, authorName, body string) (string, error) {
	return s.insertMessageWithAttachments(ctx, conversationID, authorKind, uid, authorName, body, nil)
}

func (s *SiteChat) insertMessageWithAttachments(ctx context.Context, conversationID uuid.UUID, authorKind string, uid uuid.UUID, authorName, body string, attachmentIDs []string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" && len(attachmentIDs) == 0 {
		return "", ErrBadInput
	}
	var companyID, projectID string
	if err := s.P.QueryRow(ctx, `
SELECT company_id::text, project_id::text
  FROM site_chat_conversations
 WHERE conversation_id = $1`, conversationID.String()).Scan(&companyID, &projectID); err != nil {
		return "", mapErr(err)
	}
	id := uuid.New()
	var uidAny any
	if uid != uuid.Nil {
		uidAny = uid.String()
	}
	if authorName == "" {
		authorName = authorKind
	}
	tx, err := s.transaction()
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO site_chat_messages (
  message_id, conversation_id, company_id, project_id, author_kind, author_user_id, author_name, body
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id.String(), conversationID.String(), companyID, projectID, authorKind, uidAny, authorName, body)
	if err != nil {
		return "", err
	}
	if err := s.attachMessageFiles(ctx, tx, conversationID, id, attachmentIDs); err != nil {
		return "", err
	}
	_, _ = tx.Exec(ctx, `UPDATE site_chat_conversations SET updated_at = now() WHERE conversation_id = $1`, conversationID.String())
	if authorKind == "visitor" {
		_, _ = tx.Exec(ctx, `
UPDATE site_visit_sessions s
   SET chat_open = true,
       last_activity_at = now(),
       visitor_state_snapshot =
         jsonb_set(visitor_state_snapshot, '{signals,chat_open}', 'true'::jsonb, true)
  FROM site_chat_conversations c
 WHERE c.conversation_id = $1
   AND c.visit_session_id = s.visit_session_id`, conversationID.String())
	}
	return id.String(), nil
}
