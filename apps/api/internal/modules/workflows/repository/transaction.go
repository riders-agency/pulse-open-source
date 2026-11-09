package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (w *Workflows) WithTx(tx pgx.Tx) *Workflows { bound := *w; bound.P = tx; return &bound }
func (w *Workflows) transaction() (pgx.Tx, error) {
	tx, ok := w.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("workflows command requires an application transaction")
	}
	return tx, nil
}
