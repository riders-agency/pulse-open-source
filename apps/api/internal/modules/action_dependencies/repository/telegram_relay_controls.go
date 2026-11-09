package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	relayconfig "riders.agency/pulse/api/internal/modules/telegramrelay/config"
)

func (d *Deps) RotateTelegramRelayWebhookSecret(ctx context.Context, pid, depID, uid uuid.UUID) (map[string]any, error) {
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return nil, err
	}
	var depType, relayKey, relayURL string
	var allowed []string
	err := d.P.QueryRow(ctx, `
SELECT t.dependency_type_key::text,
       COALESCE(trc.relay_key::text, ''),
       COALESCE(trc.public_url::text, ''),
       COALESCE(trc.allowed_updates, ARRAY[]::text[])
  FROM action_dependencies ad
  JOIN action_dependency_types t ON t.action_dependency_type_id = ad.action_dependency_type_id
  LEFT JOIN pulse_read.telegram_relay_configs_v1 trc ON trc.action_dependency_id = ad.action_dependency_id
 WHERE ad.action_dependency_id = $1 AND ad.project_id = $2`,
		depID.String(), pid.String()).Scan(&depType, &relayKey, &relayURL, &allowed)
	if err != nil {
		return nil, err
	}
	if depType != "telegram.relay" {
		return nil, fmt.Errorf("%w: интеграция не telegram.relay", ErrTelegramRelayMetadata)
	}
	if relayKey == "" || relayURL == "" {
		return nil, fmt.Errorf("%w: сначала сохраните metadata relay_key/webhook_secret", ErrTelegramRelayMetadata)
	}
	token, err := generateTelegramSecretToken(48)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	tag, err := relayconfig.RotateSecret(ctx, d.P, depID.String(), string(hash))
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("%w: telegram_relay_configs не найден", ErrTelegramRelayMetadata)
	}
	return map[string]any{
		"webhook_secret":  token,
		"relay_key":       relayKey,
		"relay_url":       relayURL,
		"allowed_updates": allowed,
	}, nil
}

func (d *Deps) SetTelegramRelayPollingTx(ctx context.Context, pid, depID, uid uuid.UUID, patch TelegramRelayPollingPatch) (map[string]any, error) {
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return nil, err
	}
	polling := telegramRelayPolling{
		Enabled:            patch.Enabled,
		TimeoutSeconds:     clampInt(patch.TimeoutSeconds, 50, 1, 60),
		Limit:              clampInt(patch.Limit, 100, 1, 100),
		DeleteWebhook:      patch.DeleteWebhook,
		DropPendingUpdates: patch.DropPendingUpdates,
	}
	metaPatch, _ := json.Marshal(map[string]any{
		"polling_enabled":              polling.Enabled,
		"polling_timeout_seconds":      polling.TimeoutSeconds,
		"polling_limit":                polling.Limit,
		"polling_delete_webhook":       polling.DeleteWebhook,
		"polling_drop_pending_updates": polling.DropPendingUpdates,
	})
	tx, err := d.transaction()
	if err != nil {
		return nil, err
	}
	var depType string
	if err := tx.QueryRow(ctx, `
SELECT t.dependency_type_key::text
  FROM action_dependencies ad
  JOIN action_dependency_types t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.action_dependency_id = $1 AND ad.project_id = $2`,
		depID.String(), pid.String()).Scan(&depType); err != nil {
		return nil, err
	}
	if depType != "telegram.relay" {
		return nil, fmt.Errorf("%w: интеграция не telegram.relay", ErrTelegramRelayMetadata)
	}
	_, err = tx.Exec(ctx, `
UPDATE action_dependencies
   SET metadata_json = COALESCE(metadata_json, '{}'::jsonb) || $3::jsonb,
       updated_at = now()
 WHERE action_dependency_id = $1 AND project_id = $2`,
		depID.String(), pid.String(), string(metaPatch))
	if err != nil {
		return nil, err
	}
	state, err := relayconfig.SetPolling(ctx, tx, depID.String(), relayconfig.Polling{Enabled: polling.Enabled, TimeoutSeconds: polling.TimeoutSeconds, Limit: polling.Limit, DeleteWebhook: polling.DeleteWebhook, DropPendingUpdates: polling.DropPendingUpdates})
	relayKey, relayURL, allowed := state.RelayKey, state.PublicURL, state.AllowedUpdates
	nextOffset, lastUpdateID, lastStartedAt, lastOKAt, lastError, webhookDeletedAt := state.NextOffset, state.LastUpdateID, state.LastStartedAt, state.LastOKAt, state.LastError, state.WebhookDeletedAt

	if err != nil {
		return nil, err
	}
	return map[string]any{
		"polling_enabled":              polling.Enabled,
		"polling_timeout_seconds":      polling.TimeoutSeconds,
		"polling_limit":                polling.Limit,
		"polling_delete_webhook":       polling.DeleteWebhook,
		"polling_drop_pending_updates": polling.DropPendingUpdates,
		"polling_next_offset":          nextOffset,
		"polling_last_update_id":       lastUpdateID,
		"polling_last_started_at":      lastStartedAt,
		"polling_last_ok_at":           lastOKAt,
		"polling_last_error":           lastError,
		"polling_webhook_deleted_at":   webhookDeletedAt,
		"relay_key":                    relayKey,
		"relay_url":                    relayURL,
		"allowed_updates":              allowed,
	}, nil
}
