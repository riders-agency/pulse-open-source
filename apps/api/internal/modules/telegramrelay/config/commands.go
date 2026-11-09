package config

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgconn"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

type Polling struct {
	Enabled                           bool
	TimeoutSeconds, Limit             int
	DeleteWebhook, DropPendingUpdates bool
}
type Settings struct {
	DependencyID, RelayKey, PublicURL, SecretHash, DefaultChatID string
	AllowedUpdates                                               []string
	CountryRelays                                                json.RawMessage
	Polling                                                      Polling
}

func Save(ctx context.Context, db dbpkg.Executor, in Settings) (pgconn.CommandTag, error) {
	var chat any
	if in.DefaultChatID != "" {
		chat = in.DefaultChatID
	}
	if in.SecretHash == "" {
		return db.Exec(ctx, `UPDATE telegram_relay_configs SET relay_key=$2,public_url=$3,allowed_updates=$4,default_chat_id=$5,country_relays_json=$6::jsonb,polling_enabled=$7,polling_timeout_seconds=$8,polling_limit=$9,polling_delete_webhook=$10,polling_drop_pending_updates=$11,polling_webhook_deleted_at=CASE WHEN $7 THEN polling_webhook_deleted_at ELSE NULL END,status='active',updated_at=now() WHERE action_dependency_id=$1`, in.DependencyID, in.RelayKey, in.PublicURL, in.AllowedUpdates, chat, in.CountryRelays, in.Polling.Enabled, in.Polling.TimeoutSeconds, in.Polling.Limit, in.Polling.DeleteWebhook, in.Polling.DropPendingUpdates)
	}
	return db.Exec(ctx, `INSERT INTO telegram_relay_configs(action_dependency_id,relay_key,public_url,webhook_secret_hash,allowed_updates,default_chat_id,country_relays_json,status,polling_enabled,polling_timeout_seconds,polling_limit,polling_delete_webhook,polling_drop_pending_updates)
 VALUES($1,$2,$3,$4,$5,$6,$7::jsonb,'active',$8,$9,$10,$11,$12)
 ON CONFLICT(action_dependency_id) DO UPDATE SET relay_key=EXCLUDED.relay_key,public_url=EXCLUDED.public_url,webhook_secret_hash=EXCLUDED.webhook_secret_hash,allowed_updates=EXCLUDED.allowed_updates,default_chat_id=EXCLUDED.default_chat_id,country_relays_json=EXCLUDED.country_relays_json,polling_enabled=EXCLUDED.polling_enabled,polling_timeout_seconds=EXCLUDED.polling_timeout_seconds,polling_limit=EXCLUDED.polling_limit,polling_delete_webhook=EXCLUDED.polling_delete_webhook,polling_drop_pending_updates=EXCLUDED.polling_drop_pending_updates,polling_webhook_deleted_at=CASE WHEN EXCLUDED.polling_enabled THEN telegram_relay_configs.polling_webhook_deleted_at ELSE NULL END,status='active',updated_at=now()`, in.DependencyID, in.RelayKey, in.PublicURL, in.SecretHash, in.AllowedUpdates, chat, in.CountryRelays, in.Polling.Enabled, in.Polling.TimeoutSeconds, in.Polling.Limit, in.Polling.DeleteWebhook, in.Polling.DropPendingUpdates)
}
func RotateSecret(ctx context.Context, db dbpkg.Executor, dependencyID, hash string) (pgconn.CommandTag, error) {
	return db.Exec(ctx, `UPDATE telegram_relay_configs SET webhook_secret_hash=$2,updated_at=now() WHERE action_dependency_id=$1`, dependencyID, hash)
}

type PollingState struct {
	RelayKey, PublicURL                                                            string
	AllowedUpdates                                                                 []string
	NextOffset, LastUpdateID, LastStartedAt, LastOKAt, LastError, WebhookDeletedAt any
}

func SetPolling(ctx context.Context, db dbpkg.Executor, dependencyID string, in Polling) (PollingState, error) {
	var out PollingState
	err := db.QueryRow(ctx, `UPDATE telegram_relay_configs SET polling_enabled=$2,polling_timeout_seconds=$3,polling_limit=$4,polling_delete_webhook=$5,polling_drop_pending_updates=$6,polling_webhook_deleted_at=CASE WHEN $2 AND $5 THEN polling_webhook_deleted_at ELSE NULL END,polling_locked_until=NULL,polling_last_error=NULL,updated_at=now() WHERE action_dependency_id=$1 RETURNING relay_key::text,public_url::text,allowed_updates,polling_next_offset,polling_last_update_id,polling_last_started_at,polling_last_ok_at,polling_last_error,polling_webhook_deleted_at`, dependencyID, in.Enabled, in.TimeoutSeconds, in.Limit, in.DeleteWebhook, in.DropPendingUpdates).Scan(&out.RelayKey, &out.PublicURL, &out.AllowedUpdates, &out.NextOffset, &out.LastUpdateID, &out.LastStartedAt, &out.LastOKAt, &out.LastError, &out.WebhookDeletedAt)
	return out, err
}
