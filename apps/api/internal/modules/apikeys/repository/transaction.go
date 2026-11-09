package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (a *APIKeys) WithTx(tx pgx.Tx) *APIKeys { bound := *a; bound.P = tx; return &bound }
func (a *APIKeys) transaction() (pgx.Tx, error) {
	tx, ok := a.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("apikeys command requires an application transaction")
	}
	return tx, nil
}
