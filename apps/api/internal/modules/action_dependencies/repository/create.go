package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	relayconfig "riders.agency/pulse/api/internal/modules/telegramrelay/config"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func (d *Deps) CreateTx(ctx context.Context, pid, uid uuid.UUID, in DepCreate) (string, error) {
	if err := d.assertProject(ctx, pid, uid); err != nil {
		return "", err
	}
	tx, err := d.transaction()
	if err != nil {
		return "", err
	}
	var typeID string
	if err := tx.QueryRow(ctx, `
SELECT action_dependency_type_id::text FROM action_dependency_types WHERE dependency_type_key = $1`,
		in.DependencyTypeKey).
		Scan(&typeID); err != nil {
		return "", err
	}
	adid := uuid.New()
	var secretID *string
	if in.PlaintextSecret != "" {
		enc, err := cryptoapp.Seal([]byte(in.PlaintextSecret), d.MasterHEX)
		if err != nil {
			return "", err
		}
		sid := uuid.New()
		_, err = tx.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, name, encrypted_value, created_by_user_id)
VALUES ($1,$2,$3,$4,$5)`,
			sid.String(), pid.String(), "dep:"+adid.String(), enc, uid.String())
		if err != nil {
			return "", err
		}
		ss := sid.String()
		secretID = &ss
	}
	meta := "{}"
	if len(in.MetadataJSON) > 0 {
		meta = string(in.MetadataJSON)
	}
	if in.DependencyTypeKey == "smtp.server" {
		if err := validateSMTPMetadata(json.RawMessage(meta)); err != nil {
			return "", err
		}
	}
	var relayKey string
	var relayPublicURL string
	var relayAllowed []string
	var relayDefaultChatID string
	var relayCountryRelays json.RawMessage
	var relaySecret string
	var relayPolling telegramRelayPolling
	if in.DependencyTypeKey == "telegram.relay" {
		var err error
		meta, relayKey, relayPublicURL, relayAllowed, relayDefaultChatID, relayCountryRelays, relaySecret, relayPolling, err = telegramRelayMetadataForStorage(json.RawMessage(meta), true)
		if err != nil {
			return "", err
		}
	}
	secArg := secretIDValue(secretID)
	if _, errExec := tx.Exec(ctx, `
INSERT INTO action_dependencies (
  action_dependency_id, project_id, action_dependency_type_id,
  name, status, secret_id, metadata_json, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8)`,
		adid.String(), pid.String(), typeID, in.Name, depStatusFromSecret(secretID), secArg, meta, uid.String()); errExec != nil {
		return "", errExec
	}
	if in.DependencyTypeKey == "telegram.relay" {
		hash, err := bcrypt.GenerateFromPassword([]byte(relaySecret), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		_, err = relayconfig.Save(ctx, tx, relayconfig.Settings{DependencyID: adid.String(), RelayKey: relayKey, PublicURL: relayPublicURL, SecretHash: string(hash), AllowedUpdates: relayAllowed, DefaultChatID: relayDefaultChatID, CountryRelays: relayCountryRelays, Polling: relayconfig.Polling{Enabled: relayPolling.Enabled, TimeoutSeconds: relayPolling.TimeoutSeconds, Limit: relayPolling.Limit, DeleteWebhook: relayPolling.DeleteWebhook, DropPendingUpdates: relayPolling.DropPendingUpdates}})
		if err != nil {
			return "", err
		}
	}
	return adid.String(), nil
}

func depStatusFromSecret(secretID *string) string {
	if secretID == nil {
		return "disabled"
	}
	return "active"
}

func secretIDValue(secretID *string) any {
	if secretID == nil {
		return nil
	}
	return *secretID
}
