package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (ev *WebhookEvents) WithTx(tx pgx.Tx) *WebhookEvents { bound := *ev; bound.P = tx; return &bound }
func (ev *WebhookEvents) transaction() (pgx.Tx, error) {
	tx, ok := ev.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("webhooks command requires an application transaction")
	}
	return tx, nil
}
