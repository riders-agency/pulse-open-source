package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *Runs) WithTx(tx pgx.Tx) *Runs { bound := *r; bound.P = tx; return &bound }
func (r *Runs) transaction() (pgx.Tx, error) {
	tx, ok := r.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("runs command requires an application transaction")
	}
	return tx, nil
}
