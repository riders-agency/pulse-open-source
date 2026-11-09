package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func (d *Deps) List(ctx context.Context, pid, uid uuid.UUID, includeArchived bool) ([]map[string]any, error) {
	if err := d.assertProject(ctx, pid, uid); err != nil {
		return nil, err
	}
	rows, err := d.P.Query(ctx, `
SELECT ad.action_dependency_id::text, ad.name::text, ad.status::text,
       t.dependency_type_key::text,
       CASE WHEN t.dependency_type_key = 'telegram.relay' THEN
         COALESCE(ad.metadata_json, '{}'::jsonb) ||
         jsonb_strip_nulls(jsonb_build_object(
           'relay_key', trc.relay_key,
           'relay_url', trc.public_url,
           'allowed_updates', trc.allowed_updates,
           'polling_enabled', trc.polling_enabled,
           'polling_timeout_seconds', trc.polling_timeout_seconds,
           'polling_limit', trc.polling_limit,
           'polling_next_offset', trc.polling_next_offset,
           'polling_last_update_id', trc.polling_last_update_id,
           'polling_last_started_at', trc.polling_last_started_at,
           'polling_last_ok_at', trc.polling_last_ok_at,
           'polling_last_error', trc.polling_last_error,
           'polling_delete_webhook', trc.polling_delete_webhook,
           'polling_drop_pending_updates', trc.polling_drop_pending_updates,
           'polling_webhook_deleted_at', trc.polling_webhook_deleted_at
         ))
       ELSE COALESCE(ad.metadata_json, '{}'::jsonb) END,
       COALESCE(t.provider_key::text, ''), COALESCE(t.display_name::text, '')
  FROM action_dependencies ad
  JOIN action_dependency_types t ON t.action_dependency_type_id = ad.action_dependency_type_id
  LEFT JOIN pulse_read.telegram_relay_configs_v1 trc ON trc.action_dependency_id = ad.action_dependency_id
 WHERE ad.project_id = $1
   AND ($2::bool OR ad.status <> 'deleted')
 ORDER BY ad.status <> 'deleted' DESC, t.provider_key, ad.name`, pid.String(), includeArchived)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, nm, st, dk, prov, disp string
		var meta []byte
		if err := rows.Scan(&id, &nm, &st, &dk, &meta, &prov, &disp); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"action_dependency_id":         id,
			"name":                         nm,
			"status":                       st,
			"dependency_type_key":          dk,
			"metadata_json":                json.RawMessage(meta),
			"provider_key":                 prov,
			"dependency_type_display_name": disp,
		})
	}
	return out, rows.Err()
}

func (d *Deps) Get(ctx context.Context, pid, depID, uid uuid.UUID) (map[string]any, error) {
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return nil, err
	}
	row := d.P.QueryRow(ctx, `
SELECT ad.action_dependency_id::text, ad.name::text, ad.status::text,
       t.dependency_type_key::text,
       CASE WHEN t.dependency_type_key = 'telegram.relay' THEN
         COALESCE(ad.metadata_json, '{}'::jsonb) ||
         jsonb_strip_nulls(jsonb_build_object(
           'relay_key', trc.relay_key,
           'relay_url', trc.public_url,
           'allowed_updates', trc.allowed_updates,
           'polling_enabled', trc.polling_enabled,
           'polling_timeout_seconds', trc.polling_timeout_seconds,
           'polling_limit', trc.polling_limit,
           'polling_next_offset', trc.polling_next_offset,
           'polling_last_update_id', trc.polling_last_update_id,
           'polling_last_started_at', trc.polling_last_started_at,
           'polling_last_ok_at', trc.polling_last_ok_at,
           'polling_last_error', trc.polling_last_error,
           'polling_delete_webhook', trc.polling_delete_webhook,
           'polling_drop_pending_updates', trc.polling_drop_pending_updates,
           'polling_webhook_deleted_at', trc.polling_webhook_deleted_at
         ))
       ELSE COALESCE(ad.metadata_json, '{}'::jsonb) END,
       ad.last_checked_at
  FROM action_dependencies ad
  JOIN action_dependency_types t ON t.action_dependency_type_id = ad.action_dependency_type_id
  LEFT JOIN pulse_read.telegram_relay_configs_v1 trc ON trc.action_dependency_id = ad.action_dependency_id
 WHERE ad.action_dependency_id = $1`, depID.String())
	var id, nm, st, dk string
	var meta []byte
	var last any
	if err := row.Scan(&id, &nm, &st, &dk, &meta, &last); err != nil {
		return nil, err
	}
	return map[string]any{
		"action_dependency_id": id,
		"name":                 nm,
		"status":               st,
		"dependency_type_key":  dk,
		"metadata_json":        json.RawMessage(meta),
		"last_checked_at":      last,
	}, nil
}
