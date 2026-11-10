package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

type AttachmentRow struct {
	AttachmentID     string         `json:"attachment_id"`
	ConversationID   string         `json:"conversation_id"`
	MessageID        string         `json:"message_id,omitempty"`
	CallSessionID    string         `json:"call_session_id,omitempty"`
	UploadedByKind   string         `json:"uploaded_by_kind"`
	UploadedByUserID string         `json:"uploaded_by_user_id,omitempty"`
	FileName         string         `json:"file_name"`
	ContentType      string         `json:"content_type"`
	SizeBytes        int64          `json:"size_bytes"`
	Kind             string         `json:"kind"`
	PublicURL        string         `json:"public_url,omitempty"`
	DownloadToken    string         `json:"download_token,omitempty"`
	StoragePath      string         `json:"-"`
	Metadata         map[string]any `json:"metadata_json"`
	CreatedAt        time.Time      `json:"created_at"`
}

type AttachmentInput struct {
	Scope            LiveAssistScope
	MessageID        string
	CallSessionID    string
	UploadedByKind   string
	UploadedByUserID string
	FileName         string
	ContentType      string
	SizeBytes        int64
	StoragePath      string
	DownloadToken    string
	Kind             string
	Metadata         map[string]any
}

func (s *SiteChat) OperatorAttachmentScope(ctx context.Context, conversationID, uid uuid.UUID) (LiveAssistScope, error) {
	projectID, err := s.conversationProject(ctx, conversationID)
	if err != nil {
		return LiveAssistScope{}, err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return LiveAssistScope{}, err
	}
	return s.ResolveOperatorScope(ctx, projectID, "", conversationID.String())
}

func (s *SiteChat) CreateAttachment(ctx context.Context, in AttachmentInput) (*AttachmentRow, error) {
	if in.Scope.ConversationID == "" || in.Scope.CompanyID == "" || in.Scope.ProjectID == "" {
		return nil, ErrBadInput
	}
	kind := strings.TrimSpace(in.Kind)
	if kind != "image" && kind != "audio" && kind != "video" {
		kind = "file"
	}
	uploadedBy := strings.TrimSpace(in.UploadedByKind)
	if uploadedBy != "operator" && uploadedBy != "system" {
		uploadedBy = "visitor"
	}
	if strings.TrimSpace(in.FileName) == "" || strings.TrimSpace(in.StoragePath) == "" || in.SizeBytes < 0 || strings.TrimSpace(in.DownloadToken) == "" {
		return nil, ErrBadInput
	}
	if strings.TrimSpace(in.ContentType) == "" {
		in.ContentType = "application/octet-stream"
	}
	id := uuid.NewString()
	var userID any
	if uid, err := uuid.Parse(strings.TrimSpace(in.UploadedByUserID)); err == nil && uid != uuid.Nil {
		userID = uid.String()
	}
	_, err := s.P.Exec(ctx, `
INSERT INTO site_chat_attachments (
  attachment_id, company_id, project_id, chat_widget_id, conversation_id, message_id,
  call_session_id, uploaded_by_kind, uploaded_by_user_id, file_name, content_type,
  size_bytes, storage_path, download_token, kind, metadata_json
) VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,NULLIF($6,'')::uuid,NULLIF($7,'')::uuid,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb)`,
		id, in.Scope.CompanyID, in.Scope.ProjectID, in.Scope.ChatWidgetID, in.Scope.ConversationID, strings.TrimSpace(in.MessageID),
		strings.TrimSpace(in.CallSessionID), uploadedBy, userID, trimLimit(in.FileName, 240), trimLimit(in.ContentType, 160),
		in.SizeBytes, in.StoragePath, in.DownloadToken, kind, jsonBytes(in.Metadata))
	if err != nil {
		return nil, err
	}
	return s.GetAttachmentByToken(ctx, id, in.DownloadToken)
}

func (s *SiteChat) GetAttachmentByToken(ctx context.Context, attachmentID, token string) (*AttachmentRow, error) {
	id, err := uuid.Parse(strings.TrimSpace(attachmentID))
	if err != nil {
		return nil, ErrBadInput
	}
	row := s.P.QueryRow(ctx, `
SELECT attachment_id::text, conversation_id::text, COALESCE(message_id::text,''), COALESCE(call_session_id::text,''),
       uploaded_by_kind::text, COALESCE(uploaded_by_user_id::text,''), file_name::text, content_type::text,
       size_bytes, storage_path::text, download_token::text, kind::text, metadata_json, created_at
  FROM site_chat_attachments
 WHERE attachment_id = $1 AND download_token = $2`, id.String(), strings.TrimSpace(token))
	return scanAttachment(row)
}

func (s *SiteChat) attachmentsForMessages(ctx context.Context, messageIDs []string) (map[string][]AttachmentRow, error) {
	out := map[string][]AttachmentRow{}
	if len(messageIDs) == 0 {
		return out, nil
	}
	rows, err := s.P.Query(ctx, `
SELECT attachment_id::text, conversation_id::text, COALESCE(message_id::text,''), COALESCE(call_session_id::text,''),
       uploaded_by_kind::text, COALESCE(uploaded_by_user_id::text,''), file_name::text, content_type::text,
       size_bytes, storage_path::text, download_token::text, kind::text, metadata_json, created_at
  FROM site_chat_attachments
 WHERE message_id = ANY($1::uuid[])
 ORDER BY created_at ASC`, messageIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		out[item.MessageID] = append(out[item.MessageID], *item)
	}
	return out, rows.Err()
}

func (s *SiteChat) CallRecordings(ctx context.Context, callID uuid.UUID) ([]AttachmentRow, error) {
	call, err := s.GetCall(ctx, callID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireProject(ctx, s.P, uuid.MustParse(call.ProjectID), authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := s.P.Query(ctx, `
SELECT attachment_id::text, conversation_id::text, COALESCE(message_id::text,''), COALESCE(call_session_id::text,''),
       uploaded_by_kind::text, COALESCE(uploaded_by_user_id::text,''), file_name::text, content_type::text,
       size_bytes, storage_path::text, download_token::text, kind::text, metadata_json, created_at
  FROM site_chat_attachments
 WHERE call_session_id = $1
 ORDER BY created_at ASC`, callID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AttachmentRow{}
	for rows.Next() {
		item, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (s *SiteChat) attachMessageFiles(ctx context.Context, tx pgx.Tx, conversationID, messageID uuid.UUID, attachmentIDs []string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(attachmentIDs))
	for _, raw := range attachmentIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return ErrBadInput
		}
		ids = append(ids, id.String())
	}
	tag, err := tx.Exec(ctx, `
UPDATE site_chat_attachments
   SET message_id = $2
 WHERE conversation_id = $1
   AND attachment_id = ANY($3::uuid[])
   AND message_id IS NULL`, conversationID.String(), messageID.String(), ids)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(ids)) {
		return ErrBadInput
	}
	return nil
}

func scanAttachment(row pgx.Row) (*AttachmentRow, error) {
	var item AttachmentRow
	if err := row.Scan(&item.AttachmentID, &item.ConversationID, &item.MessageID, &item.CallSessionID,
		&item.UploadedByKind, &item.UploadedByUserID, &item.FileName, &item.ContentType, &item.SizeBytes,
		&item.StoragePath, &item.DownloadToken, &item.Kind, (*jsonMap)(&item.Metadata), &item.CreatedAt); err != nil {
		return nil, mapErr(err)
	}
	item.PublicURL = "/api/v1/site-chat/attachments/" + item.AttachmentID + "/" + item.DownloadToken
	item.DownloadToken = ""
	return &item, nil
}
