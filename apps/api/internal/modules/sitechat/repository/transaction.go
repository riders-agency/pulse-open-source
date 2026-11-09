package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *SiteChat) WithTx(tx pgx.Tx) *SiteChat { bound := *s; bound.P = tx; return &bound }
func (s *SiteChat) transaction() (pgx.Tx, error) {
	tx, ok := s.P.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("sitechat command requires an application transaction")
	}
	return tx, nil
}
