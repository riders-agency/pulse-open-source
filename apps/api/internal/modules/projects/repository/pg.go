package repository

import (
	"context"
	"crypto/rand"
	"strings"

	clientcommands "riders.agency/pulse/api/internal/modules/clients/commands"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

type Projects struct {
	P dbpkg.DB
}

type ProjectRow struct {
	ID      string `json:"project_id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	HookKey string `json:"hook_key,omitempty"`
}

// ProjectDetail includes linkage for UI (company scoped).
type ProjectDetail struct {
	ProjectRow
	CompanyID string `json:"company_id"`
}

func (p *Projects) assertMember(ctx context.Context, companyID, uid uuid.UUID) error {
	if err := authz.RequireCompany(ctx, p.P, companyID, authz.PermRead); err != nil {
		return err
	}
	row := p.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.company_memberships_v1 WHERE company_id = $1 AND user_id = $2`,
		companyID.String(), uid.String())
	var one int
	if err := row.Scan(&one); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (p *Projects) List(ctx context.Context, companyID, uid uuid.UUID) ([]ProjectRow, error) {
	if err := p.assertMember(ctx, companyID, uid); err != nil {
		return nil, err
	}
	all, allowedIDs, err := authz.AllowProjectIDs(ctx, p.P, companyID)
	if err != nil {
		return nil, err
	}
	if !all && len(allowedIDs) == 0 {
		return []ProjectRow{}, nil
	}
	rows, err := p.P.Query(ctx, `
SELECT project_id::text, name::text, slug::text, hook_key::text
  FROM projects
 WHERE company_id = $1
   AND ($2::uuid[] IS NULL OR project_id = ANY($2::uuid[]))
 ORDER BY name`, companyID.String(), nullableUUIDFilter(all, allowedIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProjectRow{}
	for rows.Next() {
		var r ProjectRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.HookKey); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (p *Projects) InsertTx(ctx context.Context, companyID, uid uuid.UUID, name, slug string) (*ProjectRow, string, error) {
	if err := authz.RequireCompany(ctx, p.P, companyID, authz.PermWrite); err != nil {
		return nil, "", err
	}
	if err := p.assertMember(ctx, companyID, uid); err != nil {
		return nil, "", err
	}
	pid := uuid.New()
	hookKey := strings.ReplaceAll(uuid.NewString(), "-", "")
	rawSecret := make([]byte, 24)
	if _, err := rand.Read(rawSecret); err != nil {
		return nil, "", err
	}
	hSecret, err := bcrypt.GenerateFromPassword(rawSecret, bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	tx, err := p.transaction()
	if err != nil {
		return nil, "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO projects (
  project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id
) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		pid.String(), companyID.String(), name, slug, hookKey, string(hSecret), uid.String())
	if err != nil {
		return nil, "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, role) VALUES ($1,$2,'owner')`,
		pid.String(), uid.String())
	if err != nil {
		return nil, "", err
	}
	if err := clientcommands.SeedProjectServiceTypes(ctx, tx, companyID, pid); err != nil {
		return nil, "", err
	}
	pl := string(rawSecret)
	return &ProjectRow{ID: pid.String(), Name: name, Slug: slug, HookKey: hookKey},
		pl, nil
}

func (p *Projects) assertProject(ctx context.Context, projectID, uid uuid.UUID) error {
	if err := authz.RequireProject(ctx, p.P, projectID, authz.PermRead); err != nil {
		return err
	}
	row := p.P.QueryRow(ctx, `
SELECT 1 FROM projects pr
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE pr.project_id = $1`, projectID.String(), uid.String())
	var one int
	if err := row.Scan(&one); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (p *Projects) Get(ctx context.Context, projectID, uid uuid.UUID) (*ProjectDetail, error) {
	if err := authz.RequireProject(ctx, p.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	row := p.P.QueryRow(ctx, `
SELECT pr.project_id::text, pr.company_id::text, pr.name::text, pr.slug::text, pr.hook_key::text
  FROM projects pr
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE pr.project_id = $1`, projectID.String(), uid.String())
	var d ProjectDetail
	if err := row.Scan(&d.ID, &d.CompanyID, &d.Name, &d.Slug, &d.HookKey); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrForbidden
		}
		return nil, err
	}
	return &d, nil
}

func (p *Projects) Patch(ctx context.Context, projectID, uid uuid.UUID, name, slug *string) error {
	if err := authz.RequireProject(ctx, p.P, projectID, authz.PermWrite); err != nil {
		return err
	}
	if err := p.assertProject(ctx, projectID, uid); err != nil {
		return err
	}
	tag, err := p.P.Exec(ctx, `
UPDATE projects SET
  name = COALESCE($2, name),
  slug = COALESCE($3, slug),
  updated_at = now()
 WHERE project_id = $1`,
		projectID.String(),
		nullOpt(name),
		nullOpt(slug))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func nullOpt(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func (p *Projects) RotateHookSecret(ctx context.Context, projectID, uid uuid.UUID) (string, error) {
	if err := authz.RequireProject(ctx, p.P, projectID, authz.PermWrite); err != nil {
		return "", err
	}
	if err := p.assertProject(ctx, projectID, uid); err != nil {
		return "", err
	}
	rawSecret := make([]byte, 24)
	if _, err := rand.Read(rawSecret); err != nil {
		return "", err
	}
	hSecret, err := bcrypt.GenerateFromPassword(rawSecret, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	_, err = p.P.Exec(ctx, `
UPDATE projects SET hook_secret_hash = $2, updated_at = now() WHERE project_id = $1`,
		projectID.String(), string(hSecret))
	if err != nil {
		return "", err
	}
	return string(rawSecret), nil
}

func nullableUUIDFilter(all bool, ids []uuid.UUID) any {
	if all {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
