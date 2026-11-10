package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func (c *Clients) assertCompany(ctx context.Context, companyID, uid uuid.UUID, permission string) error {
	if err := authz.RequireCompany(ctx, c.P, companyID, permission); err != nil {
		return err
	}
	var one int
	err := c.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.company_memberships_v1 WHERE company_id = $1 AND user_id = $2`,
		companyID.String(), uid.String()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	return err
}

func (c *Clients) assertProject(ctx context.Context, projectID, uid uuid.UUID, permission string) (uuid.UUID, error) {
	if err := authz.RequireProject(ctx, c.P, projectID, permission); err != nil {
		return uuid.Nil, err
	}
	var companyID uuid.UUID
	err := c.P.QueryRow(ctx, `
SELECT company_id
  FROM pulse_read.projects_v1
 WHERE project_id = $1`, projectID.String()).Scan(&companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, errs.ErrForbidden
	}
	return companyID, err
}

func (c *Clients) clientCompany(ctx context.Context, clientID, uid uuid.UUID, permission string) (uuid.UUID, error) {
	if err := authz.RequireClient(ctx, c.P, clientID, permission); err != nil {
		return uuid.Nil, err
	}
	var companyID uuid.UUID
	err := c.P.QueryRow(ctx, `
SELECT cl.company_id
  FROM clients cl
 WHERE cl.client_id = $1 AND cl.status <> 'deleted'`, clientID.String()).Scan(&companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, errs.ErrForbidden
	}
	return companyID, err
}

func (c *Clients) companyLegalEntityCompany(ctx context.Context, entityID, uid uuid.UUID, permission string) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := c.P.QueryRow(ctx, `
SELECT cle.company_id
  FROM company_legal_entities cle
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = cle.company_id AND m.user_id = $2
 WHERE cle.company_legal_entity_id = $1 AND cle.status <> 'deleted'`,
		entityID.String(), uid.String()).Scan(&companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, errs.ErrForbidden
	}
	if err != nil {
		return uuid.Nil, err
	}
	if err := authz.RequireCompany(ctx, c.P, companyID, permission); err != nil {
		return uuid.Nil, err
	}
	return companyID, nil
}

func (c *Clients) assertLegalEntityInCompany(ctx context.Context, companyID, entityID uuid.UUID) error {
	var one int
	err := c.P.QueryRow(ctx, `
SELECT 1 FROM company_legal_entities
 WHERE company_legal_entity_id = $1 AND company_id = $2 AND status <> 'deleted'`,
		entityID.String(), companyID.String()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	return err
}

func (c *Clients) assertCompanyUser(ctx context.Context, companyID, uid uuid.UUID) error {
	var one int
	err := c.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.company_memberships_v1 WHERE company_id = $1 AND user_id = $2`,
		companyID.String(), uid.String()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	return err
}

func (c *Clients) userDisplayName(ctx context.Context, uid uuid.UUID) string {
	var firstName, email string
	_ = c.P.QueryRow(ctx, `
SELECT COALESCE(first_name,''), COALESCE(email::text,'') FROM pulse_read.users_v1 WHERE user_id = $1`, uid.String()).Scan(&firstName, &email)
	return firstNonEmpty(firstName, email)
}
