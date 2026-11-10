package repository

import (
	"context"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func (c *Clients) ListClientProjects(ctx context.Context, clientID, uid uuid.UUID) ([]ProjectLinkRow, error) {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT p.project_id::text, p.company_id::text, p.name, p.slug
  FROM client_projects cp
  JOIN pulse_read.projects_v1 p ON p.project_id = cp.project_id
 WHERE cp.client_id = $1
 ORDER BY p.name`, clientID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProjectLinkRow{}
	for rows.Next() {
		var item ProjectLinkRow
		if err := rows.Scan(&item.ProjectID, &item.CompanyID, &item.Name, &item.Slug); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (c *Clients) LinkProject(ctx context.Context, clientID, projectID, uid uuid.UUID) error {
	clientCompanyID, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite)
	if err != nil {
		return err
	}
	projectCompanyID, err := c.assertProject(ctx, projectID, uid, authz.PermWrite)
	if err != nil {
		return err
	}
	if clientCompanyID != projectCompanyID {
		return errs.ErrForbidden
	}
	_, err = c.P.Exec(ctx, `
INSERT INTO client_projects (client_id, project_id, company_id)
VALUES ($1,$2,$3)
ON CONFLICT (client_id, project_id) DO NOTHING`, clientID.String(), projectID.String(), clientCompanyID.String())
	return err
}

func (c *Clients) UnlinkProject(ctx context.Context, clientID, projectID, uid uuid.UUID) error {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return err
	}
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermWrite); err != nil {
		return err
	}
	_, err := c.P.Exec(ctx, `DELETE FROM client_projects WHERE client_id = $1 AND project_id = $2`, clientID.String(), projectID.String())
	return err
}
