package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (p *Projects) WithTx(tx pgx.Tx) *Projects { bound := *p; bound.P = tx; return &bound }
func (p *Projects) transaction() (pgx.Tx, error) {
	tx, ok := p.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("projects command requires an application transaction")
	}
	return tx, nil
}
