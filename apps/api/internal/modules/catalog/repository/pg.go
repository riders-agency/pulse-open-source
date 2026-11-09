package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Catalog struct {
	P *pgxpool.Pool
}

func (c *Catalog) ListActionTypes(ctx context.Context) ([]map[string]any, error) {
	rows, err := c.P.Query(ctx, `
SELECT action_key::text, display_name::text, category::text,
       requires_dependency, COALESCE(provider_key,'')::text,
       config_schema_json::text, result_schema_json::text
  FROM action_types ORDER BY category, display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJSON(rows)
}

func (c *Catalog) GetActionType(ctx context.Context, key string) (map[string]any, error) {
	row := c.P.QueryRow(ctx, `
SELECT action_key::text, display_name::text, category::text,
       requires_dependency, COALESCE(provider_key,'')::text,
       config_schema_json::text, result_schema_json::text
  FROM action_types WHERE action_key = $1`, key)
	var ak, dn, cat, pk string
	var rd bool
	var cfg, res []byte
	if err := row.Scan(&ak, &dn, &cat, &rd, &pk, &cfg, &res); err != nil {
		return nil, err
	}
	return decodeRow(ak, dn, cat, rd, pk, cfg, res), nil
}

func (c *Catalog) ListDepTypes(ctx context.Context) ([]map[string]any, error) {
	rows, err := c.P.Query(ctx, `
SELECT dependency_type_key::text, display_name::text, provider_key::text,
       auth_schema_json::text, metadata_schema_json::text
  FROM pulse_read.action_dependency_types_v1 ORDER BY provider_key, display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var k, dn, pk string
		var auth, meta []byte
		if err := rows.Scan(&k, &dn, &pk, &auth, &meta); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"dependency_type_key":  k,
			"display_name":         dn,
			"provider_key":         pk,
			"auth_schema_json":     json.RawMessage(auth),
			"metadata_schema_json": json.RawMessage(meta),
		})
	}
	return out, rows.Err()
}

func (c *Catalog) GetDepType(ctx context.Context, key string) (map[string]any, error) {
	row := c.P.QueryRow(ctx, `
SELECT dependency_type_key::text, display_name::text, provider_key::text,
       auth_schema_json::text, metadata_schema_json::text
  FROM pulse_read.action_dependency_types_v1 WHERE dependency_type_key = $1`, key)
	var k, dn, pk string
	var auth, meta []byte
	if err := row.Scan(&k, &dn, &pk, &auth, &meta); err != nil {
		return nil, err
	}
	return map[string]any{
		"dependency_type_key":  k,
		"display_name":         dn,
		"provider_key":         pk,
		"auth_schema_json":     json.RawMessage(auth),
		"metadata_schema_json": json.RawMessage(meta),
	}, nil
}

func scanJSON(rows pgx.Rows) ([]map[string]any, error) {
	out := make([]map[string]any, 0)
	for rows.Next() {
		var ak, dn, cat, pk string
		var rd bool
		var cfg, res []byte
		if err := rows.Scan(&ak, &dn, &cat, &rd, &pk, &cfg, &res); err != nil {
			return nil, err
		}
		out = append(out, decodeRow(ak, dn, cat, rd, pk, cfg, res))
	}
	return out, rows.Err()
}

func decodeRow(ak, dn, cat string, rd bool, pk string, cfg, res []byte) map[string]any {
	return map[string]any{
		"action_key":          ak,
		"display_name":        dn,
		"category":            cat,
		"requires_dependency": rd,
		"provider_key":        pk,
		"config_schema_json":  json.RawMessage(cfg),
		"result_schema_json":  json.RawMessage(res),
	}
}
