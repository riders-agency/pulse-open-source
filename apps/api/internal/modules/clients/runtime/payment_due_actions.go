package runtime

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (h *PaymentDueTrigger) ensureDuePaymentActions(ctx context.Context, tx pgx.Tx, projectID string, daysBeforeDue int) error {
	_, err := tx.Exec(ctx, `
WITH due AS (
  SELECT cs.company_id, cs.client_id, cs.project_id,
         MIN(cs.next_due_date)::date AS due_date,
         COALESCE(SUM(cs.amount), 0)::numeric AS total
    FROM client_services cs
   WHERE cs.project_id = $1::uuid
     AND cs.status = 'active'
     AND cs.next_due_date IS NOT NULL
     AND cs.next_due_date <= (CURRENT_DATE + ($2 * interval '1 day'))::date
     AND NOT EXISTS (
       SELECT 1
         FROM payment_actions pa
        WHERE pa.client_id = cs.client_id
          AND pa.project_id = cs.project_id
          AND pa.status IN ('pending','sent','snoozed')
     )
   GROUP BY cs.company_id, cs.client_id, cs.project_id
),
inserted AS (
  INSERT INTO payment_actions (
    payment_action_id, company_id, client_id, project_id, status, source,
    period_start, period_end, due_date, total_amount, currency,
    last_reminded_at, remind_after_at)
  SELECT gen_random_uuid(), d.company_id, d.client_id, d.project_id, 'pending', 'workflow',
         d.due_date, d.due_date, d.due_date, d.total, 'RUB', NULL, NULL
    FROM due d
  RETURNING payment_action_id, client_id, project_id, due_date
)
INSERT INTO payment_action_items (
  payment_action_id, client_service_id, client_service_type_id, title, due_date, amount, currency)
SELECT i.payment_action_id, cs.client_service_id, cs.client_service_type_id, cs.name, cs.next_due_date,
       COALESCE(cs.amount, 0), cs.currency
  FROM inserted i
  JOIN client_services cs
    ON cs.client_id = i.client_id
   AND cs.project_id = i.project_id
   AND cs.status = 'active'
   AND cs.next_due_date IS NOT NULL
   AND cs.next_due_date <= (CURRENT_DATE + ($2 * interval '1 day'))::date
ON CONFLICT DO NOTHING`, projectID, daysBeforeDue)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE payment_actions pa
   SET total_amount = sub.total
  FROM (
    SELECT i.payment_action_id, COALESCE(SUM(i.amount), 0)::numeric AS total
      FROM payment_action_items i
      JOIN payment_actions pa2 ON pa2.payment_action_id = i.payment_action_id
     WHERE pa2.project_id = $1::uuid
       AND pa2.status = 'pending'
     GROUP BY i.payment_action_id
  ) sub
 WHERE pa.payment_action_id = sub.payment_action_id`, projectID)
	return err
}
