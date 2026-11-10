package queries

import (
	"context"

	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

// EncryptedCredential is an internal runtime port. The caller has already resolved its authorized dependency.
func EncryptedCredential(ctx context.Context, db dbpkg.Executor, dependencyID string) ([]byte, error) {
	var ciphertext []byte
	err := db.QueryRow(ctx, `SELECT COALESCE(s.encrypted_value,''::bytea) FROM action_dependencies ad LEFT JOIN secrets s ON s.secret_id=ad.secret_id WHERE ad.action_dependency_id=$1 AND ad.status='active'`, dependencyID).Scan(&ciphertext)
	return ciphertext, err
}
