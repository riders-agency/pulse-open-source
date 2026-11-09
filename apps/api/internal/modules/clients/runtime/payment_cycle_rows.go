package runtime

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
)

type paymentCycleRow struct {
	CycleState      string
	StateLabel      string
	PaymentActionID string
	PaymentStatus   string
	CompanyID       string
	ClientID        string
	ClientName      string
	ProjectID       string
	ProjectName     string
	ProjectSlug     string
	DueDate         string
	PeriodStart     string
	PeriodEnd       string
	TotalAmount     float64
	Currency        string
	ServiceCount    int
	ServiceSummary  string
	Source          string
}

func (row paymentCycleRow) Map() map[string]any {
	paymentURL := ""
	if strings.TrimSpace(row.PaymentActionID) != "" {
		paymentURL = paymentActionURL(row.PaymentActionID)
	}
	return map[string]any{
		"cycle_state":       row.CycleState,
		"state_label":       row.StateLabel,
		"payment_action_id": row.PaymentActionID,
		"payment_status":    row.PaymentStatus,
		"payment_url":       paymentURL,
		"client": map[string]any{
			"client_id":  row.ClientID,
			"company_id": row.CompanyID,
			"name":       row.ClientName,
			"url":        clientURL(row.ClientID),
		},
		"project": map[string]any{
			"project_id": row.ProjectID,
			"name":       row.ProjectName,
			"slug":       row.ProjectSlug,
			"url":        projectURL(row.ProjectID),
		},
		"totals": map[string]any{
			"amount":       row.TotalAmount,
			"currency":     row.Currency,
			"due_date":     row.DueDate,
			"period_start": row.PeriodStart,
			"period_end":   row.PeriodEnd,
			"status":       row.PaymentStatus,
			"status_label": row.StateLabel,
		},
		"service_count":   row.ServiceCount,
		"service_summary": row.ServiceSummary,
		"source":          row.Source,
	}
}

const paymentCycleCollectSQL = `
WITH open_actions AS (
  SELECT
    CASE WHEN pa.status IN ('sent','snoozed') THEN 'sent' ELSE 'scheduled' END AS cycle_state,
    CASE WHEN pa.status IN ('sent','snoozed') THEN 'Отправлен' ELSE 'Запланирован' END AS state_label,
    pa.payment_action_id::text AS payment_action_id,
    pa.status AS payment_status,
    pa.company_id::text AS company_id,
    pa.client_id::text AS client_id,
    cl.name AS client_name,
    pa.project_id::text AS project_id,
    p.name AS project_name,
    p.slug AS project_slug,
    pa.due_date::text AS due_date,
    pa.period_start::text AS period_start,
    pa.period_end::text AS period_end,
    pa.total_amount::float8 AS total_amount,
    pa.currency AS currency,
    COALESCE(COUNT(i.payment_action_item_id), 0)::int AS service_count,
    COALESCE(string_agg(DISTINCT NULLIF(i.title, ''), ', ' ORDER BY NULLIF(i.title, '')), '') AS service_summary,
    pa.source AS source
  FROM payment_actions pa
  JOIN clients cl ON cl.client_id = pa.client_id
  JOIN pulse_read.projects_v1 p ON p.project_id = pa.project_id
  LEFT JOIN payment_action_items i ON i.payment_action_id = pa.payment_action_id
 WHERE pa.project_id = $1
   AND pa.status IN ('pending','sent','snoozed')
 GROUP BY pa.payment_action_id, cl.name, p.name, p.slug
),
scheduled_services AS (
  SELECT
    'scheduled' AS cycle_state,
    'Запланирован' AS state_label,
    '' AS payment_action_id,
    'scheduled' AS payment_status,
    cs.company_id::text AS company_id,
    cs.client_id::text AS client_id,
    cl.name AS client_name,
    cs.project_id::text AS project_id,
    p.name AS project_name,
    p.slug AS project_slug,
    cs.next_due_date::text AS due_date,
    cs.next_due_date::text AS period_start,
    cs.next_due_date::text AS period_end,
    COALESCE(SUM(cs.amount), 0)::float8 AS total_amount,
    COALESCE(MIN(cs.currency), 'RUB') AS currency,
    COUNT(*)::int AS service_count,
    COALESCE(string_agg(DISTINCT NULLIF(cs.name, ''), ', ' ORDER BY NULLIF(cs.name, '')), '') AS service_summary,
    'schedule' AS source
  FROM client_services cs
  JOIN clients cl ON cl.client_id = cs.client_id
  JOIN pulse_read.projects_v1 p ON p.project_id = cs.project_id
 WHERE cs.project_id = $1
   AND cs.status = 'active'
   AND cs.next_due_date IS NOT NULL
   AND cs.next_due_date <= (CURRENT_DATE + ($2::int * interval '1 day'))::date
   AND NOT EXISTS (
     SELECT 1
       FROM payment_actions pa
      WHERE pa.client_id = cs.client_id
        AND pa.project_id = cs.project_id
        AND pa.period_start = cs.next_due_date
        AND pa.status IN ('pending','sent','snoozed')
   )
 GROUP BY cs.company_id, cs.client_id, cl.name, cs.project_id, p.name, p.slug, cs.next_due_date
)
SELECT *
  FROM open_actions
UNION ALL
SELECT *
  FROM scheduled_services
 ORDER BY due_date, client_name`

func scanPaymentCycleRows(rows pgx.Rows) ([]paymentCycleRow, error) {
	out := []paymentCycleRow{}
	for rows.Next() {
		var row paymentCycleRow
		if err := rows.Scan(&row.CycleState, &row.StateLabel, &row.PaymentActionID, &row.PaymentStatus,
			&row.CompanyID, &row.ClientID, &row.ClientName, &row.ProjectID, &row.ProjectName, &row.ProjectSlug,
			&row.DueDate, &row.PeriodStart, &row.PeriodEnd, &row.TotalAmount, &row.Currency, &row.ServiceCount, &row.ServiceSummary, &row.Source); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *Runner) collectPaymentCycleRows(ctx context.Context, projectID string, daysBeforeDue int) ([]paymentCycleRow, error) {
	rows, err := r.P.Query(ctx, paymentCycleCollectSQL, projectID, daysBeforeDue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPaymentCycleRows(rows)
}
