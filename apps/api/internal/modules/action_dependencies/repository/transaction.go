package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (d *Deps) WithTx(tx pgx.Tx) *Deps { bound := *d; bound.P = tx; return &bound }
func (d *Deps) transaction() (pgx.Tx, error) {
	tx, ok := d.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("action_dependencies command requires an application transaction")
	}
	return tx, nil
}
