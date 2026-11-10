package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) ListPaymentActions(ctx context.Context, projectID, uid uuid.UUID) ([]PaymentActionRow, error) {
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, paymentActionSelect()+`
 WHERE pa.project_id = $1
 ORDER BY pa.due_date, pa.created_at DESC
 LIMIT 200`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out, err := scanPaymentRows(rows, false)
	if err != nil {
		return nil, err
	}
	if err := c.hydratePaymentActionItems(ctx, out); err != nil {
		return nil, err
	}
	if err := c.hydratePaymentActionDocuments(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Clients) ListClientPaymentActions(ctx context.Context, clientID, uid uuid.UUID, projectID *uuid.UUID) ([]PaymentActionRow, error) {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	args := []any{clientID.String()}
	where := ` WHERE pa.client_id = $1`
	if projectID != nil {
		if _, err := c.assertProject(ctx, *projectID, uid, authz.PermRead); err != nil {
			return nil, err
		}
		args = append(args, projectID.String())
		where += ` AND pa.project_id = $2`
	}
	rows, err := c.P.Query(ctx, paymentActionSelect()+where+`
 ORDER BY COALESCE(pa.confirmed_at, pa.due_date::timestamptz) DESC, pa.created_at DESC
 LIMIT 200`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out, err := scanPaymentRows(rows, false)
	if err != nil {
		return nil, err
	}
	if err := c.hydratePaymentActionItems(ctx, out); err != nil {
		return nil, err
	}
	if err := c.hydratePaymentActionDocuments(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Clients) GetPaymentAction(ctx context.Context, actionID, uid uuid.UUID) (*PaymentActionRow, error) {
	if err := c.assertPaymentAction(ctx, actionID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, paymentActionSelect()+` WHERE pa.payment_action_id = $1`, actionID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanPaymentRows(rows, false)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	itemRows, err := c.P.Query(ctx, `
SELECT i.payment_action_item_id::text, i.client_service_id::text, COALESCE(st.service_type_key,''),
       i.title, i.due_date::text, i.amount::float8, i.currency
  FROM payment_action_items i
  LEFT JOIN client_service_types st ON st.client_service_type_id = i.client_service_type_id
 WHERE i.payment_action_id = $1
 ORDER BY i.due_date, i.title`, actionID.String())
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var item PaymentActionItem
		if err := itemRows.Scan(&item.PaymentActionItemID, &item.ClientServiceID, &item.ServiceTypeKey,
			&item.Title, &item.DueDate, &item.Amount, &item.Currency); err != nil {
			return nil, err
		}
		items[0].Items = append(items[0].Items, item)
	}
	if err := itemRows.Err(); err != nil {
		return nil, err
	}
	factRows, err := c.P.Query(ctx, `
SELECT payment_fact_id::text, source, COALESCE(payment_method, 'legal_invoice'), amount::float8, currency, paid_at, created_at
  FROM payment_facts
 WHERE payment_action_id = $1
 ORDER BY paid_at DESC`, actionID.String())
	if err != nil {
		return nil, err
	}
	defer factRows.Close()
	for factRows.Next() {
		var fact PaymentFactRow
		if err := factRows.Scan(&fact.PaymentFactID, &fact.Source, &fact.PaymentMethod, &fact.Amount, &fact.Currency, &fact.PaidAt, &fact.CreatedAt); err != nil {
			return nil, err
		}
		items[0].Facts = append(items[0].Facts, fact)
	}
	if err := factRows.Err(); err != nil {
		return nil, err
	}
	docs, err := c.ListPaymentActionDocuments(ctx, actionID, uid)
	if err != nil {
		return nil, err
	}
	items[0].Documents = docs
	return &items[0], nil
}
