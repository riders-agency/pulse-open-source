package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (c *Companies) WithTx(tx pgx.Tx) *Companies { bound := *c; bound.P = tx; return &bound }
func (c *Companies) transaction() (pgx.Tx, error) {
	tx, ok := c.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("companies command requires an application transaction")
	}
	return tx, nil
}
