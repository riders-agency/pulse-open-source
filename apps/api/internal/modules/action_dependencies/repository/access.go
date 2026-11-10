package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func (d *Deps) assertProject(ctx context.Context, pid, uid uuid.UUID) error {
	if err := authz.RequireProject(ctx, d.P, pid, authz.PermRead); err != nil {
		return err
	}
	row := d.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.projects_v1 pr
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE pr.project_id = $1`, pid.String(), uid.String())
	var x int
	if err := row.Scan(&x); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (d *Deps) dependencyMembershipStatus(ctx context.Context, pid, depID, uid uuid.UUID) (status string, err error) {
	err = d.P.QueryRow(ctx, `
SELECT ad.status::text FROM action_dependencies ad
 JOIN pulse_read.projects_v1 pr ON pr.project_id = ad.project_id
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $3
 WHERE ad.action_dependency_id = $1 AND ad.project_id = $2`,
		depID.String(), pid.String(), uid.String()).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrForbidden
		}
		return "", err
	}
	return status, nil
}

func (d *Deps) assertDepInProject(ctx context.Context, pid, depID, uid uuid.UUID) error {
	row := d.P.QueryRow(ctx, `
SELECT 1 FROM action_dependencies ad
 JOIN pulse_read.projects_v1 pr ON pr.project_id = ad.project_id
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $3
 WHERE ad.action_dependency_id = $1 AND ad.project_id = $2 AND ad.status <> 'deleted'`,
		depID.String(), pid.String(), uid.String())
	var x int
	if err := row.Scan(&x); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}
