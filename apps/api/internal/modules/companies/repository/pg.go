package repository

import (
	"context"
	"errors"
	"strings"

	authcommands "riders.agency/pulse/api/internal/modules/auth/commands"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

var ErrCompanySlugTaken = errors.New("company slug already exists")

type Companies struct {
	P dbpkg.DB
}

type CompanyRow struct {
	ID   string `json:"company_id"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
	Role string `json:"role"`
}

func (c *Companies) List(ctx context.Context, uid uuid.UUID) ([]CompanyRow, error) {
	all, allowedIDs, err := authz.AllowCompanyIDs(ctx, c.P)
	if err != nil {
		return nil, err
	}
	if !all && len(allowedIDs) == 0 {
		return []CompanyRow{}, nil
	}
	q := `
SELECT c.company_id::text, COALESCE(c.name,''), COALESCE(c.slug,''), m.role::text
  FROM companies c
  JOIN company_memberships m ON m.company_id = c.company_id AND m.user_id = $1
 WHERE ($2::uuid[] IS NULL OR c.company_id = ANY($2::uuid[]))
 ORDER BY c.name`
	var filter any
	if !all {
		filter = uuidStrings(allowedIDs)
	}
	rows, err := c.P.Query(ctx, q, uid.String(), filter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyRow{}
	for rows.Next() {
		var r CompanyRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.Role); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (c *Companies) InsertTx(ctx context.Context, cid, uid uuid.UUID, slug, name string) error {
	if authz.IsAPIKey(ctx) {
		return errs.ErrForbidden
	}
	slug = strings.TrimSpace(slug)
	name = strings.TrimSpace(name)
	tx, err := c.transaction()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO companies (company_id, owner_user_id, created_by_user_id, slug, name, status)
VALUES ($1, $2, $2, $3, $4, 'active')`, cid.String(), uid.String(), slug, name)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return ErrCompanySlugTaken
		}
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO company_memberships (company_id, user_id, role)
VALUES ($1, $2, 'owner')`, cid.String(), uid.String())
	if err != nil {
		return err
	}
	err = authcommands.SetCurrentCompany(ctx, tx, uid, cid)
	if err != nil {
		return err
	}
	return nil
}

func (c *Companies) GetByID(ctx context.Context, companyID, uid uuid.UUID) (*CompanyRow, error) {
	if err := authz.RequireCompany(ctx, c.P, companyID, authz.PermRead); err != nil {
		return nil, err
	}
	var r CompanyRow
	err := c.P.QueryRow(ctx, `
SELECT c.company_id::text, COALESCE(c.name,''), COALESCE(c.slug,''), m.role::text
  FROM companies c
  JOIN company_memberships m ON m.company_id = c.company_id AND m.user_id = $2
 WHERE c.company_id = $1`, companyID.String(), uid.String()).
		Scan(&r.ID, &r.Name, &r.Slug, &r.Role)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Companies) Patch(ctx context.Context, companyID, uid uuid.UUID, name *string, slug *string) error {
	if err := authz.RequireCompany(ctx, c.P, companyID, authz.PermWrite); err != nil {
		return err
	}
	role, err := c.MembershipRole(ctx, companyID, uid)
	if err != nil {
		return err
	}
	if role != "owner" && role != "admin" {
		return errs.ErrForbidden
	}
	tag, err := c.P.Exec(ctx, `
UPDATE companies SET
    name = COALESCE($2, name),
    slug = COALESCE($3, slug),
    updated_at = now()
 WHERE company_id = $1`,
		companyID.String(),
		nullStr(name),
		nullStr(slug))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func nullStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func (c *Companies) MembershipRole(ctx context.Context, companyID, uid uuid.UUID) (string, error) {
	var role string
	err := c.P.QueryRow(ctx, `
SELECT role::text FROM company_memberships
 WHERE company_id = $1 AND user_id = $2`, companyID.String(), uid.String()).Scan(&role)
	return role, err
}

func (c *Companies) Switch(ctx context.Context, uid, cid uuid.UUID) error {
	return authcommands.SetCurrentCompany(ctx, c.P, uid, cid)
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
