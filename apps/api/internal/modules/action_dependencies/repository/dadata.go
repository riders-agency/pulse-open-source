package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"riders.agency/pulse/api/internal/modules/action_dependencies/contracts"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func (d *Deps) GetDaDataIntegration(ctx context.Context, companyID, uid uuid.UUID) (*contracts.DaDataIntegrationRow, error) {
	if err := authz.RequireCompany(ctx, d.P, companyID, authz.PermRead); err != nil {
		return nil, err
	}
	var row contracts.DaDataIntegrationRow
	err := d.P.QueryRow(ctx, `
SELECT ad.action_dependency_id::text, ad.name, COALESCE(ad.metadata_json->>'base_url',''), ad.updated_at
  FROM action_dependencies ad
  JOIN action_dependency_types dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.company_id = $1
   AND ad.scope_kind = 'company'
   AND ad.status <> 'deleted'
   AND dt.dependency_type_key = 'dadata.party'
 ORDER BY ad.updated_at DESC
 LIMIT 1`, companyID.String()).Scan(&row.ActionDependencyID, &row.Name, &row.BaseURL, &row.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		if strings.TrimSpace(os.Getenv("DADATA_API_KEY")) != "" {
			return &contracts.DaDataIntegrationRow{
				Configured: true,
				Name:       "DaData",
				BaseURL:    strings.TrimSpace(os.Getenv("DADATA_BASE_URL")),
				Source:     "env",
			}, nil
		}
		return &contracts.DaDataIntegrationRow{Configured: false, Source: "none"}, nil
	}
	if err != nil {
		return nil, err
	}
	row.Configured = true
	row.Source = "company"
	if row.BaseURL == "" {
		row.BaseURL = "https://suggestions.dadata.ru/suggestions/api/4_1/rs"
	}
	return &row, nil
}

func (d *Deps) SaveDaDataIntegrationTx(ctx context.Context, tx pgx.Tx, companyID, uid uuid.UUID, in contracts.DaDataIntegrationInput) error {
	if err := authz.RequireCompany(ctx, d.P, companyID, authz.PermWrite); err != nil {
		return err
	}
	apiKey := strings.TrimSpace(in.APIKey)
	if apiKey == "" {
		return fmt.Errorf("api_key required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(in.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://suggestions.dadata.ru/suggestions/api/4_1/rs"
	}
	enc, err := cryptoapp.Seal([]byte(apiKey), d.MasterHEX)
	if err != nil {
		return err
	}
	var typeID string
	if err := tx.QueryRow(ctx, `
SELECT action_dependency_type_id::text FROM action_dependency_types WHERE dependency_type_key = 'dadata.party'`).Scan(&typeID); err != nil {
		return err
	}
	var depID, secretID string
	err = tx.QueryRow(ctx, `
SELECT ad.action_dependency_id::text, COALESCE(ad.secret_id::text,'')
  FROM action_dependencies ad
  JOIN action_dependency_types dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.company_id = $1
   AND ad.scope_kind = 'company'
   AND dt.dependency_type_key = 'dadata.party'
 ORDER BY ad.updated_at DESC
 LIMIT 1`, companyID.String()).Scan(&depID, &secretID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	meta := map[string]string{"base_url": baseURL}
	metaBytes, _ := json.Marshal(meta)
	if depID == "" {
		newDepID := uuid.New()
		newSecretID := uuid.New()
		_, err = tx.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, company_id, scope_kind, name, encrypted_value, created_by_user_id)
VALUES ($1,NULL,$2,'company',$3,$4,$5)`,
			newSecretID.String(), companyID.String(), "dep:"+newDepID.String(), enc, uid.String())
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO action_dependencies (
  action_dependency_id, project_id, company_id, scope_kind, action_dependency_type_id,
  name, status, secret_id, metadata_json, created_by_user_id)
VALUES ($1,NULL,$2,'company',$3,$4,'active',$5,$6::jsonb,$7)`,
			newDepID.String(), companyID.String(), typeID, "DaData: ЮЛ и ИП", newSecretID.String(), string(metaBytes), uid.String())
		if err != nil {
			return err
		}
		depID = newDepID.String()
	} else {
		if secretID == "" {
			newSecretID := uuid.New()
			_, err = tx.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, company_id, scope_kind, name, encrypted_value, created_by_user_id)
VALUES ($1,NULL,$2,'company',$3,$4,$5)`,
				newSecretID.String(), companyID.String(), "dep:"+depID, enc, uid.String())
			if err != nil {
				return err
			}
			secretID = newSecretID.String()
		} else {
			_, err = tx.Exec(ctx, `
UPDATE secrets
   SET encrypted_value = $2, company_id = $3, project_id = NULL, scope_kind = 'company', updated_at = now()
 WHERE secret_id = $1`, secretID, enc, companyID.String())
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `
UPDATE action_dependencies
   SET company_id = $2,
       project_id = NULL,
       scope_kind = 'company',
       action_dependency_type_id = $3,
       name = 'DaData: ЮЛ и ИП',
       status = 'active',
       secret_id = $4,
       metadata_json = $5::jsonb,
       updated_at = now()
 WHERE action_dependency_id = $1`, depID, companyID.String(), typeID, secretID, string(metaBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) GetDaDataProjectIntegration(ctx context.Context, projectID, uid uuid.UUID) (*contracts.DaDataIntegrationRow, error) {
	if _, err := d.dadataProjectScope(ctx, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	var row contracts.DaDataIntegrationRow
	err := d.P.QueryRow(ctx, `
SELECT ad.action_dependency_id::text, ad.name, COALESCE(ad.metadata_json->>'base_url',''), ad.updated_at
  FROM action_dependencies ad
  JOIN action_dependency_types dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.project_id = $1
   AND ad.scope_kind = 'project'
   AND ad.status <> 'deleted'
   AND dt.dependency_type_key = 'dadata.party'
 ORDER BY ad.updated_at DESC
 LIMIT 1`, projectID.String()).Scan(&row.ActionDependencyID, &row.Name, &row.BaseURL, &row.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		if strings.TrimSpace(os.Getenv("DADATA_API_KEY")) != "" {
			return &contracts.DaDataIntegrationRow{
				Configured: true,
				Name:       "DaData",
				BaseURL:    strings.TrimSpace(os.Getenv("DADATA_BASE_URL")),
				Source:     "env",
			}, nil
		}
		return &contracts.DaDataIntegrationRow{Configured: false, Source: "none"}, nil
	}
	if err != nil {
		return nil, err
	}
	row.Configured = true
	row.Source = "project"
	if row.BaseURL == "" {
		row.BaseURL = "https://suggestions.dadata.ru/suggestions/api/4_1/rs"
	}
	return &row, nil
}

func (d *Deps) SaveDaDataProjectIntegrationTx(ctx context.Context, tx pgx.Tx, projectID, uid uuid.UUID, in contracts.DaDataIntegrationInput) error {
	companyID, err := d.dadataProjectScope(ctx, projectID, authz.PermWrite)
	if err != nil {
		return err
	}
	apiKey := strings.TrimSpace(in.APIKey)
	if apiKey == "" {
		return fmt.Errorf("api_key required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(in.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://suggestions.dadata.ru/suggestions/api/4_1/rs"
	}
	enc, err := cryptoapp.Seal([]byte(apiKey), d.MasterHEX)
	if err != nil {
		return err
	}
	var typeID string
	if err := tx.QueryRow(ctx, `
SELECT action_dependency_type_id::text FROM action_dependency_types WHERE dependency_type_key = 'dadata.party'`).Scan(&typeID); err != nil {
		return err
	}
	var depID, secretID string
	err = tx.QueryRow(ctx, `
SELECT ad.action_dependency_id::text, COALESCE(ad.secret_id::text,'')
  FROM action_dependencies ad
  JOIN action_dependency_types dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.project_id = $1
   AND ad.scope_kind = 'project'
   AND dt.dependency_type_key = 'dadata.party'
 ORDER BY ad.updated_at DESC
 LIMIT 1`, projectID.String()).Scan(&depID, &secretID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	meta := map[string]string{"base_url": baseURL}
	metaBytes, _ := json.Marshal(meta)
	if depID == "" {
		newDepID := uuid.New()
		newSecretID := uuid.New()
		_, err = tx.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, company_id, scope_kind, name, encrypted_value, created_by_user_id)
VALUES ($1,$2,$3,'project',$4,$5,$6)`,
			newSecretID.String(), projectID.String(), companyID.String(), "dep:"+newDepID.String(), enc, uid.String())
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO action_dependencies (
  action_dependency_id, project_id, company_id, scope_kind, action_dependency_type_id,
  name, status, secret_id, metadata_json, created_by_user_id)
VALUES ($1,$2,$3,'project',$4,$5,'active',$6,$7::jsonb,$8)`,
			newDepID.String(), projectID.String(), companyID.String(), typeID, "DaData: ЮЛ и ИП", newSecretID.String(), string(metaBytes), uid.String())
		if err != nil {
			return err
		}
		depID = newDepID.String()
	} else {
		if secretID == "" {
			newSecretID := uuid.New()
			_, err = tx.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, company_id, scope_kind, name, encrypted_value, created_by_user_id)
VALUES ($1,$2,$3,'project',$4,$5,$6)`,
				newSecretID.String(), projectID.String(), companyID.String(), "dep:"+depID, enc, uid.String())
			if err != nil {
				return err
			}
			secretID = newSecretID.String()
		} else {
			_, err = tx.Exec(ctx, `
UPDATE secrets
   SET encrypted_value = $2, project_id = $3, company_id = $4, scope_kind = 'project', updated_at = now()
 WHERE secret_id = $1`, secretID, enc, projectID.String(), companyID.String())
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `
UPDATE action_dependencies
   SET project_id = $2,
       company_id = $3,
       scope_kind = 'project',
       action_dependency_type_id = $4,
       name = 'DaData: ЮЛ и ИП',
       status = 'active',
       secret_id = $5,
       metadata_json = $6::jsonb,
       updated_at = now()
 WHERE action_dependency_id = $1`, depID, projectID.String(), companyID.String(), typeID, secretID, string(metaBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) DaDataCompanyLookupConfig(ctx context.Context, companyID, uid uuid.UUID) (contracts.DaDataLookupConfig, error) {
	if err := authz.RequireCompany(ctx, d.P, companyID, authz.PermRead); err != nil {
		return contracts.DaDataLookupConfig{}, err
	}
	apiKey, baseURL, err := d.dadataConfig(ctx, companyID)
	if err != nil {
		return contracts.DaDataLookupConfig{}, err
	}
	return contracts.DaDataLookupConfig{APIKey: apiKey, BaseURL: baseURL}, nil
}

func (d *Deps) DaDataProjectLookupConfig(ctx context.Context, projectID, uid uuid.UUID) (contracts.DaDataLookupConfig, error) {
	if _, err := d.dadataProjectScope(ctx, projectID, authz.PermRead); err != nil {
		return contracts.DaDataLookupConfig{}, err
	}
	apiKey, baseURL, err := d.dadataProjectConfig(ctx, projectID)
	if err != nil {
		return contracts.DaDataLookupConfig{}, err
	}
	return contracts.DaDataLookupConfig{APIKey: apiKey, BaseURL: baseURL}, nil
}

func (d *Deps) dadataConfig(ctx context.Context, companyID uuid.UUID) (string, string, error) {
	var encrypted []byte
	var metadata []byte
	err := d.P.QueryRow(ctx, `
SELECT s.encrypted_value, COALESCE(ad.metadata_json, '{}'::jsonb)
  FROM action_dependencies ad
  JOIN action_dependency_types dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
  JOIN secrets s ON s.secret_id = ad.secret_id
 WHERE ad.company_id = $1
   AND ad.scope_kind = 'company'
   AND ad.status = 'active'
   AND dt.dependency_type_key = 'dadata.party'
 ORDER BY ad.created_at DESC
 LIMIT 1`, companyID.String()).Scan(&encrypted, &metadata)
	if errors.Is(err, pgx.ErrNoRows) {
		if key := strings.TrimSpace(os.Getenv("DADATA_API_KEY")); key != "" {
			return key, strings.TrimSpace(os.Getenv("DADATA_BASE_URL")), nil
		}
		return "", "", contracts.ErrDaDataNotConfigured
	}
	if err != nil {
		return "", "", err
	}
	plain, err := cryptoapp.Open(encrypted, d.MasterHEX)
	if err != nil {
		return "", "", err
	}
	var meta struct {
		BaseURL string `json:"base_url"`
	}
	_ = json.Unmarshal(metadata, &meta)
	return strings.TrimSpace(string(plain)), strings.TrimSpace(meta.BaseURL), nil
}

func (d *Deps) dadataProjectConfig(ctx context.Context, projectID uuid.UUID) (string, string, error) {
	var encrypted []byte
	var metadata []byte
	err := d.P.QueryRow(ctx, `
SELECT s.encrypted_value, COALESCE(ad.metadata_json, '{}'::jsonb)
  FROM action_dependencies ad
  JOIN action_dependency_types dt ON dt.action_dependency_type_id = ad.action_dependency_type_id
  JOIN secrets s ON s.secret_id = ad.secret_id
 WHERE ad.project_id = $1
   AND ad.scope_kind = 'project'
   AND ad.status = 'active'
   AND dt.dependency_type_key = 'dadata.party'
 ORDER BY ad.created_at DESC
 LIMIT 1`, projectID.String()).Scan(&encrypted, &metadata)
	if errors.Is(err, pgx.ErrNoRows) {
		if key := strings.TrimSpace(os.Getenv("DADATA_API_KEY")); key != "" {
			return key, strings.TrimSpace(os.Getenv("DADATA_BASE_URL")), nil
		}
		return "", "", contracts.ErrDaDataNotConfigured
	}
	if err != nil {
		return "", "", err
	}
	plain, err := cryptoapp.Open(encrypted, d.MasterHEX)
	if err != nil {
		return "", "", err
	}
	var meta struct {
		BaseURL string `json:"base_url"`
	}
	_ = json.Unmarshal(metadata, &meta)
	return strings.TrimSpace(string(plain)), strings.TrimSpace(meta.BaseURL), nil
}

func (d *Deps) dadataProjectScope(ctx context.Context, projectID uuid.UUID, permission string) (uuid.UUID, error) {
	if err := authz.RequireProject(ctx, d.P, projectID, permission); err != nil {
		return uuid.Nil, err
	}
	var companyID uuid.UUID
	err := d.P.QueryRow(ctx, `SELECT company_id FROM pulse_read.projects_v1 WHERE project_id=$1`, projectID).Scan(&companyID)
	return companyID, err
}
