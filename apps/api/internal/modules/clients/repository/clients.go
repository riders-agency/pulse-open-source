package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) List(ctx context.Context, companyID, uid uuid.UUID, search string) ([]ClientRow, error) {
	if err := c.assertCompany(ctx, companyID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	search = strings.TrimSpace(search)
	rows, err := c.P.Query(ctx, `
SELECT cl.client_id::text, cl.company_id::text, cl.client_type, cl.status, cl.name,
       COALESCE(cl.short_name,''), COALESCE(cl.full_name,''), COALESCE(cl.inn,''), COALESCE(cl.kpp,''),
       COALESCE(cl.ogrn,''), COALESCE(cl.ogrnip,''), COALESCE(cl.legal_address,''), COALESCE(cl.actual_address,''),
       COALESCE(cl.okved,''), COALESCE(cl.comment,''), cl.dadata_raw_json,
       COUNT(DISTINCT cp.project_id)::int, COUNT(DISTINCT cs.client_service_id)::int,
       cl.created_at, cl.updated_at
  FROM clients cl
  LEFT JOIN client_projects cp ON cp.client_id = cl.client_id
  LEFT JOIN client_services cs ON cs.client_id = cl.client_id AND cs.status <> 'deleted'
 WHERE cl.company_id = $1
   AND cl.status <> 'deleted'
   AND (
     $2::text = ''
     OR cl.search_vector @@ websearch_to_tsquery('russian'::regconfig, $2::text)
     OR cl.name ILIKE '%' || $2::text || '%'
     OR COALESCE(cl.short_name, '') ILIKE '%' || $2::text || '%'
     OR COALESCE(cl.full_name, '') ILIKE '%' || $2::text || '%'
     OR COALESCE(cl.inn, '') ILIKE $2::text || '%'
     OR COALESCE(cl.ogrn, '') ILIKE $2::text || '%'
     OR COALESCE(cl.ogrnip, '') ILIKE $2::text || '%'
   )
 GROUP BY cl.client_id
 ORDER BY
   CASE
     WHEN $2::text = '' THEN 0
     WHEN lower(cl.name) = lower($2::text) THEN 0
     WHEN lower(cl.name) LIKE lower($2::text) || '%' THEN 1
     WHEN COALESCE(cl.inn, '') LIKE $2::text || '%' THEN 2
     ELSE 3
   END,
   CASE
     WHEN $2::text = '' THEN COUNT(DISTINCT cp.project_id)::int + COUNT(DISTINCT cs.client_service_id)::int
     ELSE ts_rank_cd(cl.search_vector, websearch_to_tsquery('russian'::regconfig, $2::text))
   END DESC,
   cl.name
 LIMIT 40`, companyID.String(), search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClientRows(rows)
}

func (c *Clients) ListByProject(ctx context.Context, projectID, uid uuid.UUID) ([]ClientRow, error) {
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT cl.client_id::text, cl.company_id::text, cl.client_type, cl.status, cl.name,
       COALESCE(cl.short_name,''), COALESCE(cl.full_name,''), COALESCE(cl.inn,''), COALESCE(cl.kpp,''),
       COALESCE(cl.ogrn,''), COALESCE(cl.ogrnip,''), COALESCE(cl.legal_address,''), COALESCE(cl.actual_address,''),
       COALESCE(cl.okved,''), COALESCE(cl.comment,''), cl.dadata_raw_json,
       COUNT(DISTINCT cp_all.project_id)::int, COUNT(DISTINCT cs.client_service_id)::int,
       cl.created_at, cl.updated_at
  FROM client_projects cp
  JOIN clients cl ON cl.client_id = cp.client_id
  LEFT JOIN client_projects cp_all ON cp_all.client_id = cl.client_id
  LEFT JOIN client_services cs ON cs.client_id = cl.client_id AND cs.project_id = cp.project_id AND cs.status <> 'deleted'
 WHERE cp.project_id = $1
   AND cl.status <> 'deleted'
 GROUP BY cl.client_id
 ORDER BY cl.name`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClientRows(rows)
}

func (c *Clients) Create(ctx context.Context, companyID, uid uuid.UUID, in ClientInput) (string, error) {
	if err := c.assertCompany(ctx, companyID, uid, authz.PermWrite); err != nil {
		return "", err
	}
	in = normalizeClientInput(in)
	if err := ValidateClientIdentifiers(in.ClientType, in.INN, in.KPP, in.OGRN, in.OGRNIP); err != nil {
		return "", err
	}
	id := uuid.New()
	raw := "{}"
	if len(in.DaDataRawJSON) > 0 {
		raw = string(in.DaDataRawJSON)
	}
	_, err := c.P.Exec(ctx, `
INSERT INTO clients (
  client_id, company_id, client_type, status, name, short_name, full_name,
  inn, kpp, ogrn, ogrnip, legal_address, actual_address, okved, dadata_raw_json,
  comment, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb,$16,$17)`,
		id.String(), companyID.String(), in.ClientType, in.Status, in.Name, nullStr(in.ShortName),
		nullStr(in.FullName), nullStr(in.INN), nullStr(in.KPP), nullStr(in.OGRN), nullStr(in.OGRNIP),
		nullStr(in.LegalAddress), nullStr(in.ActualAddress), nullStr(in.OKVED), raw, nullStr(in.Comment), uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) CreateForProjectTx(ctx context.Context, projectID, uid uuid.UUID, in ClientInput) (string, error) {
	companyID, err := c.assertProject(ctx, projectID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	in = normalizeClientInput(in)
	if err := ValidateClientIdentifiers(in.ClientType, in.INN, in.KPP, in.OGRN, in.OGRNIP); err != nil {
		return "", err
	}
	id := uuid.New()
	raw := "{}"
	if len(in.DaDataRawJSON) > 0 {
		raw = string(in.DaDataRawJSON)
	}
	tx, err := c.transaction()
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO clients (
  client_id, company_id, client_type, status, name, short_name, full_name,
  inn, kpp, ogrn, ogrnip, legal_address, actual_address, okved, dadata_raw_json,
  comment, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb,$16,$17)`,
		id.String(), companyID.String(), in.ClientType, in.Status, in.Name, nullStr(in.ShortName),
		nullStr(in.FullName), nullStr(in.INN), nullStr(in.KPP), nullStr(in.OGRN), nullStr(in.OGRNIP),
		nullStr(in.LegalAddress), nullStr(in.ActualAddress), nullStr(in.OKVED), raw, nullStr(in.Comment), uid.String())
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO client_projects (client_id, project_id, company_id)
VALUES ($1,$2,$3)
ON CONFLICT (client_id, project_id) DO NOTHING`, id.String(), projectID.String(), companyID.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) Get(ctx context.Context, clientID, uid uuid.UUID) (*ClientRow, error) {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT cl.client_id::text, cl.company_id::text, cl.client_type, cl.status, cl.name,
       COALESCE(cl.short_name,''), COALESCE(cl.full_name,''), COALESCE(cl.inn,''), COALESCE(cl.kpp,''),
       COALESCE(cl.ogrn,''), COALESCE(cl.ogrnip,''), COALESCE(cl.legal_address,''), COALESCE(cl.actual_address,''),
       COALESCE(cl.okved,''), COALESCE(cl.comment,''), cl.dadata_raw_json,
       COUNT(DISTINCT cp.project_id)::int, COUNT(DISTINCT cs.client_service_id)::int,
       cl.created_at, cl.updated_at
  FROM clients cl
  LEFT JOIN client_projects cp ON cp.client_id = cl.client_id
  LEFT JOIN client_services cs ON cs.client_id = cl.client_id AND cs.status <> 'deleted'
 WHERE cl.client_id = $1 AND cl.status <> 'deleted'
 GROUP BY cl.client_id`, clientID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanClientRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return &items[0], nil
}

func (c *Clients) Patch(ctx context.Context, clientID, uid uuid.UUID, patch ClientPatch) error {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return err
	}
	var clientType, inn, kpp, ogrn, ogrnip string
	err := c.P.QueryRow(ctx, `
SELECT client_type, COALESCE(inn,''), COALESCE(kpp,''), COALESCE(ogrn,''), COALESCE(ogrnip,'')
  FROM clients WHERE client_id = $1`, clientID.String()).Scan(&clientType, &inn, &kpp, &ogrn, &ogrnip)
	if err != nil {
		return err
	}
	if patch.ClientType != nil {
		clientType = normalizeClientType(*patch.ClientType)
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
	if err := ValidateClientIdentifiers(clientType, inn, kpp, ogrn, ogrnip); err != nil {
		return err
	}
	rawArg := any(nil)
	if len(patch.DaDataRawJSON) > 0 {
		rawArg = string(patch.DaDataRawJSON)
	}
	tag, err := c.P.Exec(ctx, `
UPDATE clients SET
  client_type = COALESCE($2, client_type),
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
WHERE client_id = $1`,
		clientID.String(), optNormClientType(patch.ClientType), optStr(patch.Status), optTrim(patch.Name),
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

func (c *Clients) Delete(ctx context.Context, clientID, uid uuid.UUID) error {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return err
	}
	_, err := c.P.Exec(ctx, `UPDATE clients SET status = 'deleted', updated_at = now() WHERE client_id = $1`, clientID.String())
	return err
}

func scanClientRows(rows pgx.Rows) ([]ClientRow, error) {
	out := []ClientRow{}
	for rows.Next() {
		var item ClientRow
		var raw []byte
		if err := rows.Scan(&item.ClientID, &item.CompanyID, &item.ClientType, &item.Status, &item.Name,
			&item.ShortName, &item.FullName, &item.INN, &item.KPP, &item.OGRN, &item.OGRNIP,
			&item.LegalAddress, &item.ActualAddress, &item.OKVED, &item.Comment, &raw,
			&item.ProjectsCount, &item.ServicesCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.DaDataRawJSON = raw
		out = append(out, item)
	}
	return out, rows.Err()
}
