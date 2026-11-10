package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/modules/clients/invoicing"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

const bankInvoiceColumns = `payment_action_id::text, revision, kind, payer_name, tax_id, purpose,
COALESCE(due_date::text,''), amount_minor::text, currency, status, updated_at`

func scanBankInvoice(row pgx.Row) (*invoicing.Draft, error) {
	var draft invoicing.Draft
	err := row.Scan(&draft.PaymentActionID, &draft.Revision, &draft.Kind, &draft.PayerName,
		&draft.TaxID, &draft.Purpose, &draft.DueDate, &draft.AmountMinor, &draft.Currency, &draft.Status, &draft.UpdatedAt)
	return &draft, err
}

func (c *Clients) ReadBankInvoice(ctx context.Context, actionID, userID uuid.UUID) (*invoicing.Workspace, error) {
	if err := c.assertPaymentAction(ctx, actionID, userID, authz.PermRead); err != nil {
		return nil, err
	}
	if err := authz.RequirePaymentFinance(ctx, c.P, actionID, authz.PermRead); err != nil {
		return nil, err
	}
	var result invoicing.Workspace
	permissionErr := authz.RequirePaymentFinance(ctx, c.P, actionID, authz.PermWrite)
	if permissionErr != nil && !errors.Is(permissionErr, errs.ErrForbidden) {
		return nil, permissionErr
	}
	result.CanEdit = permissionErr == nil
	err := c.P.QueryRow(ctx, `SELECT (pa.total_amount*100)::bigint::text, pa.currency, pa.status, cl.name
FROM payment_actions pa JOIN clients cl ON cl.client_id=pa.client_id WHERE pa.payment_action_id=$1`, actionID).
		Scan(&result.AmountMinor, &result.Currency, &result.PaymentStatus, &result.PayerName)
	if err != nil {
		return nil, err
	}
	result.Draft, err = scanBankInvoice(c.P.QueryRow(ctx, `SELECT `+bankInvoiceColumns+` FROM bank_invoice_drafts WHERE payment_action_id=$1`, actionID))
	if errors.Is(err, pgx.ErrNoRows) {
		result.Draft, err = nil, nil
	}
	return &result, err
}

func (c *Clients) SaveBankInvoice(ctx context.Context, actionID, userID, key uuid.UUID, hash string, in invoicing.Input) (*invoicing.Draft, error) {
	if err := c.assertPaymentAction(ctx, actionID, userID, authz.PermWrite); err != nil {
		return nil, err
	}
	draft, err := scanBankInvoice(c.P.QueryRow(ctx, saveBankInvoiceSQL, actionID, in.Revision,
		in.Kind, in.PayerName, in.TaxID, in.Purpose, in.DueDate, key, hash))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, invoicing.ErrConflict
	}
	return draft, err
}

// The payment lock serializes draft changes with payment confirmation. A retry of
// the latest identical command returns the same revision, without changing money.
const saveBankInvoiceSQL = `WITH payment AS (
 SELECT payment_action_id, total_amount, currency FROM payment_actions
 WHERE payment_action_id=$1 AND status IN ('pending','sent','snoozed') AND total_amount>0 FOR UPDATE
)
INSERT INTO bank_invoice_drafts AS d
 (payment_action_id,revision,kind,payer_name,tax_id,purpose,due_date,amount_minor,currency,request_key,request_hash)
SELECT payment_action_id,1,$3,$4,$5,$6,NULLIF($7,'')::date,(total_amount*100)::bigint,currency,$8,$9
FROM payment WHERE $2=0 OR EXISTS (SELECT 1 FROM bank_invoice_drafts WHERE payment_action_id=$1)
ON CONFLICT (payment_action_id) DO UPDATE SET
 revision=CASE WHEN d.request_key=$8 THEN d.revision ELSE d.revision+1 END,
 kind=EXCLUDED.kind,payer_name=EXCLUDED.payer_name,tax_id=EXCLUDED.tax_id,purpose=EXCLUDED.purpose,
 due_date=EXCLUDED.due_date,
 amount_minor=CASE WHEN d.request_key=$8 THEN d.amount_minor ELSE EXCLUDED.amount_minor END,
 currency=EXCLUDED.currency,request_key=$8,request_hash=$9,
 updated_at=CASE WHEN d.request_key=$8 THEN d.updated_at ELSE now() END
WHERE d.status='draft' AND ((d.revision=$2 AND d.request_key<>$8) OR (d.request_key=$8 AND d.request_hash=$9))
RETURNING ` + bankInvoiceColumns
