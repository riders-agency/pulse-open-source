package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (a *AI) WithTx(tx pgx.Tx) *AI { bound := *a; bound.P = tx; return &bound }
func (a *AI) transaction() (pgx.Tx, error) {
	tx, ok := a.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("ai command requires an application transaction")
	}
	return tx, nil
}
