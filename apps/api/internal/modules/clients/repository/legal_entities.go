package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) ListCompanyLegalEntities(ctx context.Context, companyID, uid uuid.UUID) ([]CompanyLegalEntityRow, error) {
	if err := c.assertCompany(ctx, companyID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT company_legal_entity_id::text, company_id::text, entity_type, status, name,
       COALESCE(short_name,''), COALESCE(full_name,''), COALESCE(inn,''), COALESCE(kpp,''),
       COALESCE(ogrn,''), COALESCE(ogrnip,''), COALESCE(legal_address,''), COALESCE(actual_address,''),
       COALESCE(okved,''), COALESCE(comment,''), dadata_raw_json, created_at, updated_at
  FROM company_legal_entities
 WHERE company_id = $1 AND status <> 'deleted'
 ORDER BY name`, companyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCompanyLegalEntityRows(rows)
}

func (c *Clients) CreateCompanyLegalEntity(ctx context.Context, companyID, uid uuid.UUID, in CompanyLegalEntityInput) (string, error) {
	if err := c.assertCompany(ctx, companyID, uid, authz.PermWrite); err != nil {
		return "", err
	}
	in = normalizeCompanyLegalEntityInput(in)
	if err := ValidateClientIdentifiers(in.EntityType, in.INN, in.KPP, in.OGRN, in.OGRNIP); err != nil {
		return "", err
	}
	id := uuid.New()
	raw := "{}"
	if len(in.DaDataRawJSON) > 0 {
		raw = string(in.DaDataRawJSON)
	}
	_, err := c.P.Exec(ctx, `
INSERT INTO company_legal_entities (
  company_legal_entity_id, company_id, entity_type, status, name, short_name, full_name,
  inn, kpp, ogrn, ogrnip, legal_address, actual_address, okved, dadata_raw_json,
  comment, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb,$16,$17)`,
		id.String(), companyID.String(), in.EntityType, in.Status, in.Name, nullStr(in.ShortName),
		nullStr(in.FullName), nullStr(in.INN), nullStr(in.KPP), nullStr(in.OGRN), nullStr(in.OGRNIP),
		nullStr(in.LegalAddress), nullStr(in.ActualAddress), nullStr(in.OKVED), raw, nullStr(in.Comment), uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) GetCompanyLegalEntity(ctx context.Context, entityID, uid uuid.UUID) (*CompanyLegalEntityRow, error) {
	if _, err := c.companyLegalEntityCompany(ctx, entityID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT company_legal_entity_id::text, company_id::text, entity_type, status, name,
       COALESCE(short_name,''), COALESCE(full_name,''), COALESCE(inn,''), COALESCE(kpp,''),
       COALESCE(ogrn,''), COALESCE(ogrnip,''), COALESCE(legal_address,''), COALESCE(actual_address,''),
       COALESCE(okved,''), COALESCE(comment,''), dadata_raw_json, created_at, updated_at
  FROM company_legal_entities
 WHERE company_legal_entity_id = $1 AND status <> 'deleted'`, entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanCompanyLegalEntityRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return &items[0], nil
}

func (c *Clients) PatchCompanyLegalEntity(ctx context.Context, entityID, uid uuid.UUID, patch CompanyLegalEntityPatch) error {
	if _, err := c.companyLegalEntityCompany(ctx, entityID, uid, authz.PermWrite); err != nil {
		return err
	}
	var entityType, inn, kpp, ogrn, ogrnip string
	err := c.P.QueryRow(ctx, `
SELECT entity_type, COALESCE(inn,''), COALESCE(kpp,''), COALESCE(ogrn,''), COALESCE(ogrnip,'')
  FROM company_legal_entities WHERE company_legal_entity_id = $1`, entityID.String()).Scan(&entityType, &inn, &kpp, &ogrn, &ogrnip)
	if err != nil {
		return err
	}
	if patch.EntityType != nil {
		entityType = normalizeLegalEntityType(*patch.EntityType)
	}
	if patch.INN != nil {
		inn = normalizeDigits(*patch.INN)
	}
	if patch.KPP != nil {
		kpp = normalizeDigits(*patch.KPP)
	}
	if patch.OGRN != nil {
		ogrn = normalizeDigits(*patch.OGRN)
	}
	if patch.OGRNIP != nil {
		ogrnip = normalizeDigits(*patch.OGRNIP)
	}
	if err := ValidateClientIdentifiers(entityType, inn, kpp, ogrn, ogrnip); err != nil {
		return err
	}
	rawArg := any(nil)
	if len(patch.DaDataRawJSON) > 0 {
		rawArg = string(patch.DaDataRawJSON)
	}
	tag, err := c.P.Exec(ctx, `
UPDATE company_legal_entities SET
  entity_type = COALESCE($2, entity_type),
  status = COALESCE($3, status),
  name = COALESCE($4, name),
  short_name = COALESCE($5, short_name),
  full_name = COALESCE($6, full_name),
  inn = COALESCE($7, inn),
  kpp = COALESCE($8, kpp),
  ogrn = COALESCE($9, ogrn),
  ogrnip = COALESCE($10, ogrnip),
  legal_address = COALESCE($11, legal_address),
  actual_address = COALESCE($12, actual_address),
  okved = COALESCE($13, okved),
  comment = COALESCE($14, comment),
  dadata_raw_json = COALESCE($15::jsonb, dadata_raw_json),
  updated_at = now()
WHERE company_legal_entity_id = $1`,
		entityID.String(), optNormLegalEntityType(patch.EntityType), optStr(patch.Status), optTrim(patch.Name),
		optTrim(patch.ShortName), optTrim(patch.FullName), optDigits(patch.INN), optDigits(patch.KPP),
		optDigits(patch.OGRN), optDigits(patch.OGRNIP), optTrim(patch.LegalAddress), optTrim(patch.ActualAddress),
		optTrim(patch.OKVED), optTrim(patch.Comment), rawArg)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (c *Clients) DeleteCompanyLegalEntity(ctx context.Context, entityID, uid uuid.UUID) error {
	if _, err := c.companyLegalEntityCompany(ctx, entityID, uid, authz.PermWrite); err != nil {
		return err
	}
	_, err := c.P.Exec(ctx, `
UPDATE company_legal_entities SET status = 'deleted', updated_at = now()
 WHERE company_legal_entity_id = $1`, entityID.String())
	return err
}

func scanCompanyLegalEntityRows(rows pgx.Rows) ([]CompanyLegalEntityRow, error) {
	out := []CompanyLegalEntityRow{}
	for rows.Next() {
		var item CompanyLegalEntityRow
		var raw []byte
		if err := rows.Scan(&item.CompanyLegalEntityID, &item.CompanyID, &item.EntityType, &item.Status, &item.Name,
			&item.ShortName, &item.FullName, &item.INN, &item.KPP, &item.OGRN, &item.OGRNIP,
			&item.LegalAddress, &item.ActualAddress, &item.OKVED, &item.Comment, &raw,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.DaDataRawJSON = raw
		out = append(out, item)
	}
	return out, rows.Err()
}
