package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) ListResponsibles(ctx context.Context, clientID, uid uuid.UUID) ([]ClientResponsibleRow, error) {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT cr.client_responsible_id::text, cr.company_id::text, cr.client_id::text,
       COALESCE(cr.user_id::text,''), COALESCE(cr.company_legal_entity_id::text,''),
       COALESCE(cle.name,''), cr.display_name, cr.role, cr.is_primary, cr.status, COALESCE(cr.comment,'')
  FROM client_responsibles cr
  LEFT JOIN company_legal_entities cle ON cle.company_legal_entity_id = cr.company_legal_entity_id
 WHERE cr.client_id = $1 AND cr.status <> 'deleted'
 ORDER BY cr.is_primary DESC, cr.display_name`, clientID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ClientResponsibleRow{}
	for rows.Next() {
		var item ClientResponsibleRow
		if err := rows.Scan(&item.ClientResponsibleID, &item.CompanyID, &item.ClientID, &item.UserID,
			&item.CompanyLegalEntityID, &item.LegalEntityName, &item.DisplayName, &item.Role,
			&item.IsPrimary, &item.Status, &item.Comment); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (c *Clients) CreateResponsibleTx(ctx context.Context, clientID, uid uuid.UUID, in ClientResponsibleInput) (string, error) {
	companyID, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.Role = normalizeResponsibleRole(in.Role)
	in.Status = normalizeDefault(in.Status, "active")
	in.Comment = strings.TrimSpace(in.Comment)
	if in.UserID != nil && *in.UserID != uuid.Nil {
		if err := c.assertCompanyUser(ctx, companyID, *in.UserID); err != nil {
			return "", err
		}
		if in.DisplayName == "" {
			in.DisplayName = c.userDisplayName(ctx, *in.UserID)
		}
	}
	if in.CompanyLegalEntityID != nil && *in.CompanyLegalEntityID != uuid.Nil {
		if err := c.assertLegalEntityInCompany(ctx, companyID, *in.CompanyLegalEntityID); err != nil {
			return "", err
		}
	}
	if in.DisplayName == "" {
		return "", fmt.Errorf("display_name required")
	}
	id := uuid.New()
	tx, err := c.transaction()
	if err != nil {
		return "", err
	}
	if in.IsPrimary {
		_, _ = tx.Exec(ctx, `UPDATE client_responsibles SET is_primary = false WHERE client_id = $1`, clientID.String())
	}
	_, err = tx.Exec(ctx, `
INSERT INTO client_responsibles (
  client_responsible_id, company_id, client_id, user_id, company_legal_entity_id,
  display_name, role, is_primary, status, comment, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id.String(), companyID.String(), clientID.String(), nullableUUID(in.UserID), nullableUUID(in.CompanyLegalEntityID),
		in.DisplayName, in.Role, in.IsPrimary, in.Status, nullStr(in.Comment), uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) PatchResponsibleTx(ctx context.Context, clientID, responsibleID, uid uuid.UUID, patch ClientResponsiblePatch) error {
	companyID, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite)
	if err != nil {
		return err
	}
	if patch.UserID != nil && *patch.UserID != uuid.Nil {
		if err := c.assertCompanyUser(ctx, companyID, *patch.UserID); err != nil {
			return err
		}
	}
	if patch.CompanyLegalEntityID != nil && *patch.CompanyLegalEntityID != uuid.Nil {
		if err := c.assertLegalEntityInCompany(ctx, companyID, *patch.CompanyLegalEntityID); err != nil {
			return err
		}
	}
	role := any(nil)
	if patch.Role != nil {
		role = normalizeResponsibleRole(*patch.Role)
	}
	tx, err := c.transaction()
	if err != nil {
		return err
	}
	if patch.IsPrimary != nil && *patch.IsPrimary {
		_, _ = tx.Exec(ctx, `UPDATE client_responsibles SET is_primary = false WHERE client_id = $1`, clientID.String())
	}
	tag, err := tx.Exec(ctx, `
UPDATE client_responsibles SET
  user_id = COALESCE($3, user_id),
  company_legal_entity_id = COALESCE($4, company_legal_entity_id),
  display_name = COALESCE($5, display_name),
  role = COALESCE($6, role),
  is_primary = COALESCE($7, is_primary),
  status = COALESCE($8, status),
  comment = COALESCE($9, comment),
  updated_at = now()
WHERE client_id = $1 AND client_responsible_id = $2`,
		clientID.String(), responsibleID.String(), nullableUUID(patch.UserID), nullableUUID(patch.CompanyLegalEntityID),
		optTrim(patch.DisplayName), role, patch.IsPrimary, optStr(patch.Status), optTrim(patch.Comment))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (c *Clients) DeleteResponsible(ctx context.Context, clientID, responsibleID, uid uuid.UUID) error {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return err
	}
	_, err := c.P.Exec(ctx, `
UPDATE client_responsibles SET status = 'deleted', updated_at = now()
 WHERE client_id = $1 AND client_responsible_id = $2`, clientID.String(), responsibleID.String())
	return err
}
