package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func SeedProjectServiceTypes(ctx context.Context, tx pgx.Tx, companyID, projectID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
INSERT INTO client_service_types (
  company_id, project_id, scope_kind, service_type_key,
  display_name, description, status, sort_order)
SELECT $1::uuid, $2::uuid, 'project', service_type_key,
       display_name, description, status, sort_order
  FROM client_service_types
 WHERE project_id IS NULL
   AND status = 'active'
ON CONFLICT (project_id, service_type_key) WHERE project_id IS NOT NULL DO NOTHING`,
		companyID.String(), projectID.String())
	return err
}
