package office

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func isMissing(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
func (h Handler) apply(ctx context.Context, id identity, steps map[string]any) (map[string]any, error) {
	var approved bool
	err := h.P.QueryRow(ctx, `SELECT approved FROM workflow_office_decisions WHERE workflow_run_id=$1`, id.RunID).Scan(&approved)
	if err != nil {
		return nil, fmt.Errorf("нет решения человека")
	}
	if !approved {
		return map[string]any{"applied": false, "message": "Человек отклонил изменения. Сайт и счета не изменены."}, nil
	}
	page, invoice := step(steps, "website"), step(steps, "invoice")
	if page == nil || invoice == nil {
		return nil, fmt.Errorf("подготовка страницы или счёта отключена; применить неполный результат нельзя")
	}
	p, i := []byte(`{}`), []byte(`{}`)
	if page["skipped"] != true {
		p, _ = json.Marshal(page)
	}
	if invoice["skipped"] != true {
		i, _ = json.Marshal(invoice)
	}
	err = uow.Run(ctx, h.P, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO workflow_office_artifacts(workflow_run_id,project_id,page_json,invoice_json)
  VALUES($1,$2,$3,$4) ON CONFLICT(workflow_run_id) DO NOTHING`, id.RunID, id.ProjectID, p, i)
		return err
	})
	return map[string]any{"applied": err == nil, "page": page, "invoice": invoice, "message": "Страница учебного сайта опубликована, черновик счёта сохранён. Отправки в банк нет."}, err
}
