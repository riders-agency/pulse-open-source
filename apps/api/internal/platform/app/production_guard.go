package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func isProductionEnvironment(value string) bool {
	return strings.EqualFold(value, "production")
}

func guardNoProductionDemoTenant(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	err := pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM users
   WHERE email = 'demo@riders-pulse.local'::citext
     AND deleted_at IS NULL
)`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check production demo tenant: %w", err)
	}
	if exists {
		return fmt.Errorf("production database contains demo tenant; apply demo cleanup before starting")
	}
	return nil
}
