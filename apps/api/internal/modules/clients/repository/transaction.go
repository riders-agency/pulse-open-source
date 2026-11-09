package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (c *Clients) WithTx(tx pgx.Tx) *Clients { bound := *c; bound.P = tx; return &bound }
func (c *Clients) transaction() (pgx.Tx, error) {
	tx, ok := c.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("clients command requires an application transaction")
	}
	return tx, nil
}
