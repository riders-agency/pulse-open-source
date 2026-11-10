package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskAttachmentInput struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	StoragePath string
}

func (w *WorkTracker) Attachments(ctx context.Context, taskID, userID uuid.UUID) ([]TaskAttachment, error) {
	access, _, err := w.AccessForTask(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	task, err := w.TaskRow(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !access.CanSeeTask(*task) || access.IsClientGuest {
		return []TaskAttachment{}, nil
	}
	rows, err := w.pool().Query(ctx, `
SELECT ta.task_attachment_id::text, ta.task_id::text, ta.file_name, ta.content_type,
       ta.size_bytes, COALESCE(u.display_name, u.email::text, 'System'), ta.created_at
  FROM task_attachments ta
  LEFT JOIN pulse_read.users_v1 u ON u.user_id = ta.uploaded_by_user_id
 WHERE ta.task_id = $1
 ORDER BY ta.created_at DESC`, taskID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TaskAttachment{}
	for rows.Next() {
		var row TaskAttachment
		if err := rows.Scan(&row.TaskAttachmentID, &row.TaskID, &row.FileName, &row.ContentType, &row.SizeBytes, &row.UploadedByName, &row.CreatedAt); err != nil {
			return nil, err
		}
		row.DownloadURL = "/api/v1/work/task-attachments/" + row.TaskAttachmentID
		out = append(out, row)
	}
	return out, rows.Err()
}

func (w *WorkTracker) CreateAttachment(ctx context.Context, taskID, userID uuid.UUID, in TaskAttachmentInput) (*TaskAttachment, error) {
	access, _, err := w.AccessForTask(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	task, err := w.TaskRow(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !access.CanEditTaskInWorkflow(task.WorkflowID) || !access.CanSeeTask(*task) || access.IsClientGuest {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(in.FileName)
	if name == "" {
		return nil, fmt.Errorf("file_name is required")
	}
	path := strings.TrimSpace(in.StoragePath)
	if path == "" {
		return nil, fmt.Errorf("storage_path is required")
	}
	contentType := strings.TrimSpace(in.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	attachmentID := uuid.New()
	var row TaskAttachment
	err = w.pool().QueryRow(ctx, `
INSERT INTO task_attachments (
  task_attachment_id, company_id, project_id, client_id, task_id, uploaded_by_user_id,
  file_name, content_type, size_bytes, storage_path
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING task_attachment_id::text, task_id::text, file_name, content_type, size_bytes, created_at`,
		attachmentID.String(), access.CompanyID.String(), access.ProjectID.String(), nullableTextUUID(task.ClientID),
		taskID.String(), userID.String(), name, contentType, in.SizeBytes, path).
		Scan(&row.TaskAttachmentID, &row.TaskID, &row.FileName, &row.ContentType, &row.SizeBytes, &row.CreatedAt)
	if err != nil {
		return nil, err
	}
	if _, err := w.pool().Exec(ctx, `UPDATE tasks SET attachments_count = attachments_count + 1, revision = revision + 1, updated_at = now() WHERE task_id = $1`, taskID.String()); err != nil {
		return nil, err
	}
	if _, err := w.pool().Exec(ctx, `
INSERT INTO task_activity_events (company_id, task_id, actor_user_id, event_type, payload_json)
VALUES ($1,$2,$3,'attachment.added', jsonb_build_object('file_name', $4::text))`,
		access.CompanyID.String(), taskID.String(), userID.String(), name); err != nil {
		return nil, err
	}
	row.DownloadURL = "/api/v1/work/task-attachments/" + row.TaskAttachmentID
	return &row, nil
}

func (w *WorkTracker) Attachment(ctx context.Context, attachmentID, userID uuid.UUID) (*TaskAttachment, string, error) {
	var row TaskAttachment
	var storagePath string
	err := w.pool().QueryRow(ctx, `
SELECT ta.task_attachment_id::text, ta.task_id::text, ta.file_name, ta.content_type,
       ta.size_bytes, COALESCE(u.display_name, u.email::text, 'System'), ta.storage_path, ta.created_at
  FROM task_attachments ta
  LEFT JOIN pulse_read.users_v1 u ON u.user_id = ta.uploaded_by_user_id
 WHERE ta.task_attachment_id = $1`, attachmentID.String()).
		Scan(&row.TaskAttachmentID, &row.TaskID, &row.FileName, &row.ContentType, &row.SizeBytes, &row.UploadedByName, &storagePath, &row.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	taskID, err := uuid.Parse(row.TaskID)
	if err != nil {
		return nil, "", err
	}
	access, _, err := w.AccessForTask(ctx, taskID, userID)
	if err != nil {
		return nil, "", err
	}
	task, err := w.TaskRow(ctx, taskID)
	if err != nil {
		return nil, "", err
	}
	if !access.CanSeeTask(*task) || access.IsClientGuest {
		return nil, "", ErrForbidden
	}
	row.DownloadURL = "/api/v1/work/task-attachments/" + row.TaskAttachmentID
	return &row, storagePath, nil
}
