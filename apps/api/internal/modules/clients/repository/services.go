package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func (c *Clients) ListServiceTypes(ctx context.Context) ([]ServiceTypeRow, error) {
	rows, err := c.P.Query(ctx, `
SELECT client_service_type_id::text,
       COALESCE(company_id::text, ''),
       COALESCE(project_id::text, ''),
       COALESCE(scope_kind, 'global'),
       service_type_key,
       display_name,
       COALESCE(description,''),
       status,
       sort_order
  FROM client_service_types
 WHERE status = 'active'
   AND project_id IS NULL
 ORDER BY sort_order, display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ServiceTypeRow{}
	for rows.Next() {
		var item ServiceTypeRow
		if err := rows.Scan(&item.ClientServiceTypeID, &item.CompanyID, &item.ProjectID, &item.ScopeKind, &item.ServiceTypeKey, &item.DisplayName, &item.Description, &item.Status, &item.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (c *Clients) ListProjectServiceTypes(ctx context.Context, projectID, uid uuid.UUID) ([]ServiceTypeRow, error) {
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
WITH project_types AS (
  SELECT client_service_type_id, company_id, project_id, scope_kind, service_type_key,
         display_name, description, status, sort_order
    FROM client_service_types
   WHERE project_id = $1
     AND status = 'active'
),
fallback_types AS (
  SELECT client_service_type_id, company_id, project_id, COALESCE(scope_kind, 'global') AS scope_kind,
         service_type_key, display_name, description, status, sort_order
    FROM client_service_types
   WHERE project_id IS NULL
     AND status = 'active'
     AND NOT EXISTS (SELECT 1 FROM project_types)
)
SELECT client_service_type_id::text,
       COALESCE(company_id::text, ''),
       COALESCE(project_id::text, ''),
       COALESCE(scope_kind, 'global'),
       service_type_key,
       display_name,
       COALESCE(description,''),
       status,
       sort_order
  FROM project_types
UNION ALL
SELECT client_service_type_id::text,
       COALESCE(company_id::text, ''),
       COALESCE(project_id::text, ''),
       COALESCE(scope_kind, 'global'),
       service_type_key,
       display_name,
       COALESCE(description,''),
       status,
       sort_order
  FROM fallback_types
 ORDER BY sort_order, display_name`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ServiceTypeRow{}
	for rows.Next() {
		var item ServiceTypeRow
		if err := rows.Scan(&item.ClientServiceTypeID, &item.CompanyID, &item.ProjectID, &item.ScopeKind, &item.ServiceTypeKey, &item.DisplayName, &item.Description, &item.Status, &item.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (c *Clients) CreateProjectServiceType(ctx context.Context, projectID, uid uuid.UUID, in ServiceTypeInput) (string, error) {
	companyID, err := c.assertProject(ctx, projectID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	in = normalizeServiceTypeInput(in)
	var existingID string
	_ = c.P.QueryRow(ctx, `
SELECT client_service_type_id::text
  FROM client_service_types
 WHERE project_id = $1
   AND status = 'active'
   AND (service_type_key = $2 OR lower(btrim(display_name)) = lower(btrim($3)))
 ORDER BY sort_order, created_at
 LIMIT 1`, projectID.String(), in.ServiceTypeKey, in.DisplayName).Scan(&existingID)
	if existingID != "" {
		return existingID, nil
	}
	id := uuid.New()
	_, err = c.P.Exec(ctx, `
INSERT INTO client_service_types (
  client_service_type_id, company_id, project_id, scope_kind,
  service_type_key, display_name, description, status, sort_order)
VALUES ($1,$2,$3,'project',$4,$5,$6,$7,$8)`,
		id.String(), companyID.String(), projectID.String(), in.ServiceTypeKey,
		in.DisplayName, nullStr(in.Description), in.Status, in.SortOrder)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) PatchProjectServiceType(ctx context.Context, projectID, serviceTypeID, uid uuid.UUID, patch ServiceTypePatch) error {
	if _, err := c.assertProject(ctx, projectID, uid, authz.PermWrite); err != nil {
		return err
	}
	tag, err := c.P.Exec(ctx, `
UPDATE client_service_types SET
  service_type_key = COALESCE($3, service_type_key),
  display_name = COALESCE($4, display_name),
  description = COALESCE($5, description),
  sort_order = COALESCE($6, sort_order),
  status = COALESCE($7, status),
  updated_at = now()
WHERE project_id = $1
  AND client_service_type_id = $2`,
		projectID.String(), serviceTypeID.String(), optServiceTypeKey(patch.ServiceTypeKey),
		optTrim(patch.DisplayName), optStr(patch.Description), patch.SortOrder, optServiceTypeStatus(patch.Status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (c *Clients) ArchiveProjectServiceType(ctx context.Context, projectID, serviceTypeID, uid uuid.UUID) error {
	status := "archived"
	return c.PatchProjectServiceType(ctx, projectID, serviceTypeID, uid, ServiceTypePatch{Status: &status})
}

func (c *Clients) ListServices(ctx context.Context, clientID, uid uuid.UUID, projectID *uuid.UUID) ([]ServiceRow, error) {
	companyID, err := c.clientCompany(ctx, clientID, uid, authz.PermRead)
	if err != nil {
		return nil, err
	}
	if projectID != nil {
		projectCompanyID, err := c.assertProject(ctx, *projectID, uid, authz.PermRead)
		if err != nil {
			return nil, err
		}
		if projectCompanyID != companyID {
			return nil, errs.ErrForbidden
		}
	}
	rows, err := c.P.Query(ctx, `
SELECT cs.client_service_id::text, cs.company_id::text, cs.client_id::text, cs.project_id::text,
       cs.client_service_type_id::text, st.service_type_key, st.display_name, cs.name, cs.status,
       COALESCE(cs.amount,0)::float8, COALESCE(cs.raw_price_text,''), cs.currency, cs.billing_period,
       COALESCE(cs.next_due_date::text,''), cs.remind_days_before, cs.remind_every_days,
       COALESCE(cs.metadata_json, '{}'::jsonb)
  FROM client_services cs
  JOIN client_service_types st ON st.client_service_type_id = cs.client_service_type_id
  JOIN pulse_read.project_members_v1 pm ON pm.project_id = cs.project_id AND pm.user_id = $3
 WHERE cs.client_id = $1
   AND cs.status <> 'deleted'
   AND ($2::uuid IS NULL OR cs.project_id = $2)
 ORDER BY cs.next_due_date NULLS LAST, cs.name`, clientID.String(), nullableUUID(projectID), uid.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServiceRows(rows)
}

func (c *Clients) CreateServiceTx(ctx context.Context, clientID, uid uuid.UUID, in ServiceInput) (string, error) {
	companyID, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	projectCompanyID, err := c.assertProject(ctx, in.ProjectID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	if companyID != projectCompanyID {
		return "", errs.ErrForbidden
	}
	typeID, err := c.resolveServiceType(ctx, in.ProjectID, in.ClientServiceTypeID, in.ServiceTypeKey)
	if err != nil {
		return "", err
	}
	in = normalizeServiceInput(in)
	if in.NextDueDate != "" {
		nextDueDate, err := parsePaymentDate(in.NextDueDate)
		if err != nil {
			return "", fmt.Errorf("next_due_date: %w", err)
		}
		in.NextDueDate = nextDueDate
	}
	id := uuid.New()
	meta := "{}"
	if len(in.MetadataJSON) > 0 {
		meta = string(in.MetadataJSON)
	}
	tx, err := c.transaction()
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO client_projects (client_id, project_id, company_id)
VALUES ($1,$2,$3)
ON CONFLICT (client_id, project_id) DO NOTHING`, clientID.String(), in.ProjectID.String(), companyID.String())
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO client_services (
  client_service_id, company_id, client_id, project_id, client_service_type_id, name, status,
  amount, raw_price_text, currency, billing_period, next_due_date, remind_days_before,
  remind_every_days, metadata_json, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::date,$13,$14,$15::jsonb,$16)`,
		id.String(), companyID.String(), clientID.String(), in.ProjectID.String(), typeID.String(), in.Name,
		in.Status, in.Amount, nullStr(in.RawPriceText), in.Currency, in.BillingPeriod, nullStr(in.NextDueDate),
		in.RemindDaysBefore, in.RemindEveryDays, meta, uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) GetService(ctx context.Context, serviceID, uid uuid.UUID) (*ServiceRow, error) {
	if err := c.assertService(ctx, serviceID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT cs.client_service_id::text, cs.company_id::text, cs.client_id::text, cs.project_id::text,
       cs.client_service_type_id::text, st.service_type_key, st.display_name, cs.name, cs.status,
       COALESCE(cs.amount,0)::float8, COALESCE(cs.raw_price_text,''), cs.currency, cs.billing_period,
       COALESCE(cs.next_due_date::text,''), cs.remind_days_before, cs.remind_every_days,
       COALESCE(cs.metadata_json, '{}'::jsonb)
  FROM client_services cs
  JOIN client_service_types st ON st.client_service_type_id = cs.client_service_type_id
 WHERE cs.client_service_id = $1 AND cs.status <> 'deleted'`, serviceID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanServiceRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return &items[0], nil
}

func (c *Clients) PatchService(ctx context.Context, serviceID, uid uuid.UUID, patch ServicePatch) error {
	if err := c.assertService(ctx, serviceID, uid, authz.PermWrite); err != nil {
		return err
	}
	if patch.NextDueDate != nil && strings.TrimSpace(*patch.NextDueDate) != "" {
		nextDueDate, err := parsePaymentDate(*patch.NextDueDate)
		if err != nil {
			return fmt.Errorf("next_due_date: %w", err)
		}
		patch.NextDueDate = &nextDueDate
	}
	var serviceProjectID uuid.UUID
	if err := c.P.QueryRow(ctx, `
SELECT project_id FROM client_services WHERE client_service_id = $1 AND status <> 'deleted'`, serviceID.String()).Scan(&serviceProjectID); err != nil {
		return err
	}
	var typeID any
	if patch.ClientServiceTypeID != nil || patch.ServiceTypeKey != nil {
		var requestedID uuid.UUID
		if patch.ClientServiceTypeID != nil {
			requestedID = *patch.ClientServiceTypeID
		}
		key := ""
		if patch.ServiceTypeKey != nil {
			key = *patch.ServiceTypeKey
		}
		resolved, err := c.resolveServiceType(ctx, serviceProjectID, requestedID, key)
		if err != nil {
			return err
		}
		typeID = resolved.String()
	}
	metaArg := any(nil)
	if len(patch.MetadataJSON) > 0 {
		metaArg = string(patch.MetadataJSON)
	}
	tag, err := c.P.Exec(ctx, `
UPDATE client_services SET
  client_service_type_id = COALESCE($2, client_service_type_id),
  name = COALESCE($3, name),
  status = COALESCE($4, status),
  amount = COALESCE($5, amount),
  raw_price_text = COALESCE($6, raw_price_text),
  currency = COALESCE($7, currency),
  billing_period = COALESCE($8, billing_period),
  next_due_date = COALESCE($9::date, next_due_date),
  remind_days_before = COALESCE($10, remind_days_before),
  remind_every_days = COALESCE($11, remind_every_days),
  metadata_json = COALESCE($12::jsonb, metadata_json),
  updated_at = now()
WHERE client_service_id = $1`, serviceID.String(), typeID, optTrim(patch.Name), optStr(patch.Status),
		patch.Amount, optTrim(patch.RawPriceText), optStr(patch.Currency), optStr(patch.BillingPeriod),
		optDate(patch.NextDueDate), patch.RemindDaysBefore, patch.RemindEveryDays, metaArg)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (c *Clients) DeleteService(ctx context.Context, serviceID, uid uuid.UUID) error {
	if err := c.assertService(ctx, serviceID, uid, authz.PermWrite); err != nil {
		return err
	}
	_, err := c.P.Exec(ctx, `UPDATE client_services SET status = 'deleted', updated_at = now() WHERE client_service_id = $1`, serviceID.String())
	return err
}

func (c *Clients) assertService(ctx context.Context, serviceID, uid uuid.UUID, permission string) error {
	if err := authz.RequireClientService(ctx, c.P, serviceID, permission); err != nil {
		return err
	}
	var one int
	err := c.P.QueryRow(ctx, `
SELECT 1
  FROM client_services cs
 WHERE cs.client_service_id = $1 AND cs.status <> 'deleted'`, serviceID.String()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	return err
}

func (c *Clients) resolveServiceType(ctx context.Context, projectID, id uuid.UUID, key string) (uuid.UUID, error) {
	if id != uuid.Nil {
		var exists uuid.UUID
		err := c.P.QueryRow(ctx, `
SELECT client_service_type_id FROM client_service_types
 WHERE client_service_type_id = $1
   AND status = 'active'
   AND (project_id = $2 OR project_id IS NULL)`, id.String(), projectID.String()).Scan(&exists)
		return exists, err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return uuid.Nil, fmt.Errorf("service type required")
	}
	var out uuid.UUID
	err := c.P.QueryRow(ctx, `
SELECT client_service_type_id FROM client_service_types
 WHERE service_type_key = $1
   AND status = 'active'
   AND (project_id = $2 OR project_id IS NULL)
 ORDER BY CASE WHEN project_id = $2 THEN 0 ELSE 1 END
 LIMIT 1`, key, projectID.String()).Scan(&out)
	return out, err
}

func scanServiceRows(rows pgx.Rows) ([]ServiceRow, error) {
	out := []ServiceRow{}
	for rows.Next() {
		var item ServiceRow
		if err := rows.Scan(&item.ClientServiceID, &item.CompanyID, &item.ClientID, &item.ProjectID,
			&item.ClientServiceTypeID, &item.ServiceTypeKey, &item.ServiceTypeName, &item.Name,
			&item.Status, &item.Amount, &item.RawPriceText, &item.Currency, &item.BillingPeriod,
			&item.NextDueDate, &item.RemindDaysBefore, &item.RemindEveryDays, &item.MetadataJSON); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
