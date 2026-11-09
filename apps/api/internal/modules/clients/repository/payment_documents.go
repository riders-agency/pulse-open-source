package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) ListPaymentActionDocuments(ctx context.Context, actionID, uid uuid.UUID) ([]PaymentDocumentRow, error) {
	if err := c.assertPaymentAction(ctx, actionID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT payment_action_document_id::text, payment_action_id::text, file_name, content_type,
       size_bytes, kind, created_at
  FROM payment_action_documents
 WHERE payment_action_id = $1
 ORDER BY created_at DESC`, actionID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaymentDocumentRow{}
	for rows.Next() {
		var row PaymentDocumentRow
		if err := rows.Scan(&row.PaymentActionDocumentID, &row.PaymentActionID, &row.FileName,
			&row.ContentType, &row.SizeBytes, &row.Kind, &row.CreatedAt); err != nil {
			return nil, err
		}
		row.DownloadURL = "/api/v1/payment-action-documents/" + row.PaymentActionDocumentID
		out = append(out, row)
	}
	return out, rows.Err()
}

func (c *Clients) CreatePaymentActionDocument(ctx context.Context, actionID, uid uuid.UUID, in PaymentDocumentInput) (*PaymentDocumentRow, error) {
	if err := c.assertPaymentAction(ctx, actionID, uid, authz.PermWrite); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.FileName)
	if name == "" {
		return nil, fmt.Errorf("file_name is required")
	}
	path := strings.TrimSpace(in.StoragePath)
	if path == "" {
		return nil, fmt.Errorf("storage_path is required")
	}
	contentType := normalizeDefault(in.ContentType, "application/octet-stream")
	kind := normalizePaymentDocumentKind(in.Kind)
	var companyID, clientID, projectID string
	if err := c.P.QueryRow(ctx, `
SELECT company_id::text, client_id::text, project_id::text
  FROM payment_actions
 WHERE payment_action_id = $1`, actionID.String()).Scan(&companyID, &clientID, &projectID); err != nil {
		return nil, err
	}
	docID := uuid.New()
	var row PaymentDocumentRow
	err := c.P.QueryRow(ctx, `
INSERT INTO payment_action_documents (
  payment_action_document_id, payment_action_id, company_id, client_id, project_id,
  uploaded_by_user_id, file_name, content_type, size_bytes, kind, storage_path)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
RETURNING payment_action_document_id::text, payment_action_id::text, file_name, content_type,
          size_bytes, kind, created_at`,
		docID.String(), actionID.String(), companyID, clientID, projectID, uid.String(),
		name, contentType, in.SizeBytes, kind, path).
		Scan(&row.PaymentActionDocumentID, &row.PaymentActionID, &row.FileName, &row.ContentType,
			&row.SizeBytes, &row.Kind, &row.CreatedAt)
	if err != nil {
		return nil, err
	}
	if kind == "invoice" {
		_, _ = c.P.Exec(ctx, `
UPDATE payment_actions
   SET status = 'sent',
       last_reminded_at = COALESCE(last_reminded_at, clock_timestamp()),
       updated_at = clock_timestamp()
 WHERE payment_action_id = $1
   AND status = 'pending'`, actionID.String())
	}
	row.DownloadURL = "/api/v1/payment-action-documents/" + row.PaymentActionDocumentID
	return &row, nil
}

func (c *Clients) GetPaymentActionDocument(ctx context.Context, documentID, uid uuid.UUID) (*PaymentDocumentRow, string, error) {
	var row PaymentDocumentRow
	var storagePath string
	err := c.P.QueryRow(ctx, `
SELECT payment_action_document_id::text, payment_action_id::text, file_name, content_type,
       size_bytes, kind, storage_path, created_at
  FROM payment_action_documents
 WHERE payment_action_document_id = $1`, documentID.String()).
		Scan(&row.PaymentActionDocumentID, &row.PaymentActionID, &row.FileName, &row.ContentType,
			&row.SizeBytes, &row.Kind, &storagePath, &row.CreatedAt)
	if err != nil {
		return nil, "", err
	}
	actionID, err := uuid.Parse(row.PaymentActionID)
	if err != nil {
		return nil, "", err
	}
	if err := c.assertPaymentAction(ctx, actionID, uid, authz.PermRead); err != nil {
		return nil, "", err
	}
	row.DownloadURL = "/api/v1/payment-action-documents/" + row.PaymentActionDocumentID
	return &row, storagePath, nil
}
