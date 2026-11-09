package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	depqueries "riders.agency/pulse/api/internal/modules/action_dependencies/queries"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func (r *Runner) loadDependency(ctx context.Context, depID string) (*dependencyRecord, error) {
	if strings.TrimSpace(depID) == "" {
		return nil, errors.New("integration id is required")
	}
	var typeKey, status string
	var metadataRaw []byte
	err := r.P.QueryRow(ctx, `
SELECT t.dependency_type_key::text, ad.status::text, ad.metadata_json
  FROM pulse_read.action_dependencies_v1 ad
  JOIN pulse_read.action_dependency_types_v1 t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.action_dependency_id = $1`, depID).Scan(&typeKey, &status, &metadataRaw)
	if err != nil {
		return nil, errors.New("action dependency not found")
	}
	encrypted, err := depqueries.EncryptedCredential(ctx, r.P, depID)
	if err != nil {
		return nil, err
	}
	plaintext, err := cryptoapp.Open(encrypted, r.MasterHEX)
	if err != nil {
		return nil, errors.New("dependency secret decrypt failed")
	}
	metadata := map[string]any{}
	_ = json.Unmarshal(metadataRaw, &metadata)
	return &dependencyRecord{TypeKey: typeKey, Status: status, Metadata: metadata, Plaintext: string(plaintext)}, nil
}
