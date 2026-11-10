package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

type ProjectPaymentScheduleRow struct {
	ClientID           string  `json:"client_id"`
	ClientName         string  `json:"client_name"`
	ProjectID          string  `json:"project_id"`
	DueDate            string  `json:"due_date"`
	TotalAmount        float64 `json:"total_amount"`
	Currency           string  `json:"currency"`
	ServiceCount       int     `json:"service_count"`
	PaymentActionID    string  `json:"payment_action_id,omitempty"`
	PaymentStatus      string  `json:"payment_status,omitempty"`
	InvoiceScenarioKey string  `json:"invoice_scenario_key,omitempty"`
}

type CreateManualPaymentActionInput struct {
	ClientID        string `json:"client_id"`
	DueDate         string `json:"due_date"`
	InvoicePrepared bool   `json:"invoice_prepared"`
}

func (c *Clients) ListProjectPaymentSchedule(ctx context.Context, projectID, uid uuid.UUID) ([]ProjectPaymentScheduleRow, error) {
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT cs.client_id::text, cl.name, cs.project_id::text, cs.next_due_date::text,
       COALESCE(SUM(cs.amount),0)::float8, COALESCE(MIN(cs.currency),'RUB'), COUNT(*)::int,
       COALESCE(open_pa.payment_action_id::text,''), COALESCE(open_pa.status,''),
       CASE
         WHEN COUNT(DISTINCT COALESCE(NULLIF(COALESCE(cs.metadata_json #>> '{invoice_scenario,key}', cs.metadata_json ->> 'invoice_scenario_key'), ''), 'bank_invoice_responsible_email')) > 1
         THEN 'mixed_invoice_scenarios'
         ELSE MIN(COALESCE(NULLIF(COALESCE(cs.metadata_json #>> '{invoice_scenario,key}', cs.metadata_json ->> 'invoice_scenario_key'), ''), 'bank_invoice_responsible_email'))
       END AS invoice_scenario_key
  FROM client_services cs
  JOIN clients cl ON cl.client_id = cs.client_id
  LEFT JOIN LATERAL (
    SELECT pa.payment_action_id, pa.status
      FROM payment_actions pa
     WHERE pa.client_id = cs.client_id
       AND pa.project_id = cs.project_id
       AND pa.status IN ('pending','sent','snoozed')
     ORDER BY pa.due_date, pa.created_at
     LIMIT 1
  ) open_pa ON true
 WHERE cs.project_id = $1
   AND cs.status = 'active'
   AND cs.next_due_date IS NOT NULL
   AND cs.next_due_date <= (CURRENT_DATE + interval '30 days')::date
 GROUP BY cs.client_id, cl.name, cs.project_id, cs.next_due_date, open_pa.payment_action_id, open_pa.status
 ORDER BY cs.next_due_date, cl.name
 LIMIT 200`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProjectPaymentScheduleRow{}
	for rows.Next() {
		var row ProjectPaymentScheduleRow
		if err := rows.Scan(&row.ClientID, &row.ClientName, &row.ProjectID, &row.DueDate,
			&row.TotalAmount, &row.Currency, &row.ServiceCount, &row.PaymentActionID, &row.PaymentStatus, &row.InvoiceScenarioKey); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (c *Clients) CreateManualPaymentActionTx(ctx context.Context, projectID, uid uuid.UUID, in CreateManualPaymentActionInput) (string, error) {
	if err := authz.RequireProjectFinance(ctx, c.P, projectID); err != nil {
		return "", err
	}
	companyID, err := c.assertProject(ctx, projectID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	clientID, err := uuid.Parse(strings.TrimSpace(in.ClientID))
	if err != nil {
		return "", ErrInvalidIdentifiers
	}
	dueDate, err := parsePaymentDate(in.DueDate)
	if err != nil {
		return "", err
	}
	if existing, err := c.openPaymentActionForClientProject(ctx, clientID, projectID); existing != "" || err != nil {
		if err == nil && in.InvoicePrepared {
			_, err = c.P.Exec(ctx, `
UPDATE payment_actions
   SET status = 'sent',
       updated_at = clock_timestamp()
 WHERE payment_action_id = $1
   AND status = 'pending'`, existing)
		}
		return existing, err
	}
	var total float64
	var serviceCount int
	err = c.P.QueryRow(ctx, `
SELECT COALESCE(SUM(amount),0)::float8, COUNT(*)::int
  FROM client_services
 WHERE client_id = $1
   AND project_id = $2
   AND status = 'active'
   AND next_due_date = $3::date`, clientID.String(), projectID.String(), dueDate).Scan(&total, &serviceCount)
	if err != nil {
		return "", err
	}
	if serviceCount == 0 {
		return "", pgx.ErrNoRows
	}
	actionID := uuid.New()
	tx, err := c.transaction()
	if err != nil {
		return "", err
	}
	tag, err := tx.Exec(ctx, `
INSERT INTO payment_actions (
  payment_action_id, company_id, client_id, project_id, status, source,
  period_start, period_end, due_date, total_amount, currency)
VALUES ($1,$2,$3,$4,$7,'manual',$5::date,$5::date,$5::date,$6,'RUB')
ON CONFLICT DO NOTHING`, actionID.String(), companyID.String(), clientID.String(), projectID.String(), dueDate, total, paymentActionInitialStatus(in.InvoicePrepared))
	if err != nil {
		return "", err
	}
	if tag.RowsAffected() == 0 {
		return c.openPaymentActionForClientProject(ctx, clientID, projectID)
	}
	_, err = tx.Exec(ctx, `
INSERT INTO payment_action_items (
  payment_action_id, client_service_id, client_service_type_id, title, due_date, amount, currency)
SELECT $1, cs.client_service_id, cs.client_service_type_id, cs.name, cs.next_due_date,
       COALESCE(cs.amount,0), cs.currency
  FROM client_services cs
 WHERE cs.client_id = $2
   AND cs.project_id = $3
   AND cs.status = 'active'
   AND cs.next_due_date = $4::date
ON CONFLICT DO NOTHING`, actionID.String(), clientID.String(), projectID.String(), dueDate)
	if err != nil {
		return "", err
	}
	return actionID.String(), nil
}

func paymentActionInitialStatus(invoicePrepared bool) string {
	if invoicePrepared {
		return "sent"
	}
	return "pending"
}

func (c *Clients) openPaymentActionForClientProject(ctx context.Context, clientID, projectID uuid.UUID) (string, error) {
	var id string
	err := c.P.QueryRow(ctx, `
SELECT payment_action_id::text
  FROM payment_actions
 WHERE client_id = $1
   AND project_id = $2
   AND status IN ('pending','sent','snoozed')
 ORDER BY due_date, created_at
 LIMIT 1`, clientID.String(), projectID.String()).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return "", err
}
