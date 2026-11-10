package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func (c *Clients) assertPaymentAction(ctx context.Context, actionID, uid uuid.UUID, permission string) error {
	guard := authz.RequirePaymentAction
	if permission == authz.PermWrite {
		guard = authz.RequirePaymentFinance
	}
	if err := guard(ctx, c.P, actionID, permission); err != nil {
		return err
	}
	var one int
	err := c.P.QueryRow(ctx, `
SELECT 1
  FROM payment_actions pa
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pa.company_id AND m.user_id = $2
 WHERE pa.payment_action_id = $1`, actionID.String(), uid.String()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	return err
}

func paymentActionSelect() string {
	return `
SELECT pa.payment_action_id::text, pa.company_id::text, pa.client_id::text, cl.name,
       pa.project_id::text, p.name, COALESCE(pa.workflow_run_id::text,''), pa.status, pa.source,
       COALESCE(pa.payment_method, 'legal_invoice'), pa.period_start::text, pa.period_end::text, pa.due_date::text, pa.total_amount::float8,
       pa.currency, pa.last_reminded_at, pa.snoozed_until, pa.confirmed_at, COALESCE(pa.comment,''), pa.created_at
  FROM payment_actions pa
  JOIN clients cl ON cl.client_id = pa.client_id
  JOIN pulse_read.projects_v1 p ON p.project_id = pa.project_id`
}

func scanPaymentRows(rows pgx.Rows, _ bool) ([]PaymentActionRow, error) {
	out := []PaymentActionRow{}
	for rows.Next() {
		var item PaymentActionRow
		if err := rows.Scan(&item.PaymentActionID, &item.CompanyID, &item.ClientID, &item.ClientName,
			&item.ProjectID, &item.ProjectName, &item.WorkflowRunID, &item.Status, &item.Source,
			&item.PaymentMethod, &item.PeriodStart, &item.PeriodEnd, &item.DueDate, &item.TotalAmount, &item.Currency,
			&item.LastRemindedAt, &item.SnoozedUntil, &item.ConfirmedAt, &item.Comment, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (c *Clients) hydratePaymentActionItems(ctx context.Context, rows []PaymentActionRow) error {
	for i := range rows {
		itemRows, err := c.P.Query(ctx, `
SELECT i.payment_action_item_id::text, i.client_service_id::text, COALESCE(st.service_type_key,''),
       i.title, i.due_date::text, i.amount::float8, i.currency
  FROM payment_action_items i
  LEFT JOIN client_service_types st ON st.client_service_type_id = i.client_service_type_id
 WHERE i.payment_action_id = $1
 ORDER BY i.due_date, i.title`, rows[i].PaymentActionID)
		if err != nil {
			return err
		}
		for itemRows.Next() {
			var item PaymentActionItem
			if err := itemRows.Scan(&item.PaymentActionItemID, &item.ClientServiceID, &item.ServiceTypeKey,
				&item.Title, &item.DueDate, &item.Amount, &item.Currency); err != nil {
				itemRows.Close()
				return err
			}
			rows[i].Items = append(rows[i].Items, item)
		}
		if err := itemRows.Err(); err != nil {
			itemRows.Close()
			return err
		}
		itemRows.Close()
	}
	return nil
}

func (c *Clients) hydratePaymentActionDocuments(ctx context.Context, rows []PaymentActionRow) error {
	for i := range rows {
		docRows, err := c.P.Query(ctx, `
SELECT payment_action_document_id::text, payment_action_id::text, file_name, content_type,
       size_bytes, kind, created_at
  FROM payment_action_documents
 WHERE payment_action_id = $1
 ORDER BY created_at DESC`, rows[i].PaymentActionID)
		if err != nil {
			return err
		}
		for docRows.Next() {
			var row PaymentDocumentRow
			if err := docRows.Scan(&row.PaymentActionDocumentID, &row.PaymentActionID, &row.FileName,
				&row.ContentType, &row.SizeBytes, &row.Kind, &row.CreatedAt); err != nil {
				docRows.Close()
				return err
			}
			row.DownloadURL = "/api/v1/payment-action-documents/" + row.PaymentActionDocumentID
			rows[i].Documents = append(rows[i].Documents, row)
		}
		if err := docRows.Err(); err != nil {
			docRows.Close()
			return err
		}
		docRows.Close()
	}
	return nil
}

func (c *Clients) paymentEventByIdempotency(ctx context.Context, q pgx.Tx, eventSource, idempotencyKey string) (*PaymentEventRow, error) {
	var row PaymentEventRow
	err := q.QueryRow(ctx, `
SELECT payment_event_id::text, event_type, event_source, event_subject, occurred_at,
       idempotency_key, client_id::text, COALESCE(client_service_id::text, ''),
       COALESCE(payment_action_id::text, ''), COALESCE(payment_fact_id::text, ''),
       COALESCE(amount,0)::float8, currency, COALESCE(period_start::text, ''),
       COALESCE(period_end::text, ''), paid_at, source, COALESCE(payment_method, 'legal_invoice'), trigger_type, status, processed_at
  FROM payment_events
 WHERE event_source = $1 AND idempotency_key = $2`,
		eventSource, idempotencyKey).
		Scan(&row.EventID, &row.EventType, &row.EventSource, &row.EventSubject, &row.OccurredAt,
			&row.IdempotencyKey, &row.ClientID, &row.ClientServiceID, &row.PaymentActionID,
			&row.PaymentFactID, &row.Amount, &row.Currency, &row.PeriodStart, &row.PeriodEnd,
			&row.PaidAt, &row.Source, &row.PaymentMethod, &row.TriggerType, &row.Status, &row.ProcessedAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}
