package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

type APIKeys struct {
	P dbpkg.DB
}

type GrantInput struct {
	ScopeKind   string    `json:"scope_kind"`
	ScopeID     uuid.UUID `json:"scope_id,omitempty"`
	Permissions []string  `json:"permissions"`
}

type APIKeyRow struct {
	APIKeyID   string     `json:"api_key_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Grants     []GrantRow `json:"grants"`
}

type GrantRow struct {
	ScopeKind   string   `json:"scope_kind"`
	ScopeID     string   `json:"scope_id,omitempty"`
	Permissions []string `json:"permissions"`
}

func (a *APIKeys) List(ctx context.Context, userID uuid.UUID) ([]APIKeyRow, error) {
	rows, err := a.P.Query(ctx, `
SELECT api_key_id::text,
       name::text,
       key_prefix::text,
       status::text,
       created_at,
       last_used_at,
       expires_at
  FROM api_keys
 WHERE user_id = $1
   AND revoked_at IS NULL
 ORDER BY created_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]APIKeyRow, 0)
	for rows.Next() {
		var row APIKeyRow
		if err := rows.Scan(&row.APIKeyID, &row.Name, &row.KeyPrefix, &row.Status, &row.CreatedAt, &row.LastUsedAt, &row.ExpiresAt); err != nil {
			return nil, err
		}
		grants, err := a.grants(ctx, row.APIKeyID)
		if err != nil {
			return nil, err
		}
		row.Grants = grants
		out = append(out, row)
	}
	return out, rows.Err()
}

func (a *APIKeys) CreateTx(ctx context.Context, userID uuid.UUID, name string, grants []GrantInput, expiresAt *time.Time) (*APIKeyRow, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "API key"
	}
	prefix, secret, plaintext, err := newAPIKeyToken()
	if err != nil {
		return nil, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	keyID := uuid.New()
	tx, err := a.transaction()
	if err != nil {
		return nil, "", err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO api_keys (api_key_id, user_id, name, key_prefix, key_hash, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		keyID.String(), userID.String(), name, prefix, string(hash), expiresAt)
	if err != nil {
		return nil, "", err
	}
	for _, grant := range normalizeGrants(grants) {
		var scopeID any
		if grant.ScopeID != uuid.Nil {
			scopeID = grant.ScopeID.String()
		}
		_, err = tx.Exec(ctx, `
INSERT INTO api_key_grants (api_key_id, scope_kind, scope_id, permissions)
VALUES ($1, $2, $3, $4)`,
			keyID.String(), grant.ScopeKind, scopeID, grant.Permissions)
		if err != nil {
			return nil, "", err
		}
	}
	row := &APIKeyRow{
		APIKeyID:  keyID.String(),
		Name:      name,
		KeyPrefix: prefix,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}
	row.Grants, err = a.grants(ctx, row.APIKeyID)
	if err != nil {
		return nil, "", err
	}
	_ = secret
	return row, plaintext, nil
}

func (a *APIKeys) Revoke(ctx context.Context, userID, keyID uuid.UUID) error {
	tag, err := a.P.Exec(ctx, `
UPDATE api_keys
   SET status = 'revoked',
       revoked_at = now(),
       revoked_reason = 'user_revoked'
 WHERE api_key_id = $1
   AND user_id = $2
   AND revoked_at IS NULL`, keyID.String(), userID.String())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (a *APIKeys) grants(ctx context.Context, keyID string) ([]GrantRow, error) {
	rows, err := a.P.Query(ctx, `
SELECT scope_kind::text, COALESCE(scope_id::text, ''), permissions
  FROM api_key_grants
 WHERE api_key_id = $1
 ORDER BY scope_kind, scope_id`, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]GrantRow, 0)
	for rows.Next() {
		var row GrantRow
		if err := rows.Scan(&row.ScopeKind, &row.ScopeID, &row.Permissions); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func newAPIKeyToken() (prefix string, secret string, plaintext string, err error) {
	prefixBytes := make([]byte, 4)
	secretBytes := make([]byte, 24)
	if _, err = rand.Read(prefixBytes); err != nil {
		return "", "", "", err
	}
	if _, err = rand.Read(secretBytes); err != nil {
		return "", "", "", err
	}
	prefix = hex.EncodeToString(prefixBytes)
	secret = hex.EncodeToString(secretBytes)
	return prefix, secret, "rp_live_" + prefix + "_" + secret, nil
}

func normalizeGrants(grants []GrantInput) []GrantInput {
	out := make([]GrantInput, 0, len(grants))
	for _, grant := range grants {
		grant.ScopeKind = strings.TrimSpace(grant.ScopeKind)
		if grant.ScopeKind == "" {
			continue
		}
		if len(grant.Permissions) == 0 {
			grant.Permissions = []string{"read"}
		}
		out = append(out, grant)
	}
	return out
}

var _ = pgx.ErrNoRows
