package repository

import (
	"context"
	"encoding/json"

	"riders.agency/pulse/api/internal/modules/action_dependencies/contracts"

	"github.com/google/uuid"

	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func (d *Deps) LoadConnectionProbe(ctx context.Context, pid, depID, uid uuid.UUID) (contracts.ProbeConfig, error) {
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return contracts.ProbeConfig{}, err
	}
	var dep contracts.ProbeConfig
	var metadataRaw []byte
	var encrypted []byte
	if err := d.P.QueryRow(ctx, `
SELECT ad.name::text,
       ad.status::text,
       t.dependency_type_key::text,
       COALESCE(t.provider_key::text, ''),
       COALESCE(ad.metadata_json, '{}'::jsonb),
       COALESCE(s.encrypted_value, ''::bytea),
       ad.secret_id IS NOT NULL
  FROM action_dependencies ad
  JOIN action_dependency_types t ON t.action_dependency_type_id = ad.action_dependency_type_id
  LEFT JOIN secrets s ON s.secret_id = ad.secret_id
 WHERE ad.action_dependency_id = $1
   AND ad.project_id = $2
   AND ad.status <> 'deleted'`,
		depID.String(), pid.String()).
		Scan(&dep.Name, &dep.Status, &dep.TypeKey, &dep.ProviderKey, &metadataRaw, &encrypted, &dep.HasSecret); err != nil {
		return contracts.ProbeConfig{}, err
	}
	dep.Metadata = map[string]any{}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &dep.Metadata); err != nil {
			return contracts.ProbeConfig{}, err
		}
	}
	if dep.HasSecret {
		plaintext, err := cryptoapp.Open(encrypted, d.MasterHEX)
		if err != nil {
			return contracts.ProbeConfig{}, err
		}
		dep.Plaintext = string(plaintext)
	}
	return dep, nil
}

func (d *Deps) MarkConnectionCheck(ctx context.Context, pid, depID uuid.UUID) error {
	_, err := d.P.Exec(ctx, `UPDATE action_dependencies SET last_checked_at=now(),updated_at=now() WHERE action_dependency_id=$1 AND project_id=$2`, depID, pid)
	return err
}
