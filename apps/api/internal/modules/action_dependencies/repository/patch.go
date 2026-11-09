package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	relayconfig "riders.agency/pulse/api/internal/modules/telegramrelay/config"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func (d *Deps) PatchTx(ctx context.Context, pid, depID, uid uuid.UUID, patch DepPatch) error {
	st, err := d.dependencyMembershipStatus(ctx, pid, depID, uid)
	if err != nil {
		return err
	}
	if st == "deleted" {
		if patch.Name != nil || len(patch.MetadataJSON) > 0 {
			return errors.New("archived dependency: only restore (status active)")
		}
		if patch.Status == nil || *patch.Status != "active" {
			return errors.New("archived dependency can only be restored to active")
		}
		_, errExec := d.P.Exec(ctx, `
UPDATE action_dependencies SET
  status = CASE WHEN secret_id IS NULL THEN 'disabled'::text ELSE 'active'::text END,
  updated_at = now()
 WHERE action_dependency_id = $1 AND project_id = $2 AND status = 'deleted'`,
			depID.String(), pid.String())
		return errExec
	}
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return err
	}
	var depTypeKey string
	if err := d.P.QueryRow(ctx, `
SELECT t.dependency_type_key::text
  FROM action_dependencies ad
  JOIN action_dependency_types t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.action_dependency_id = $1`, depID.String()).Scan(&depTypeKey); err != nil {
		return err
	}
	if len(patch.MetadataJSON) > 0 && depTypeKey == "smtp.server" {
		if err := validateSMTPMetadata(patch.MetadataJSON); err != nil {
			return err
		}
	}
	var relayKey string
	var relayPublicURL string
	var relayAllowed []string
	var relayDefaultChatID string
	var relayCountryRelays json.RawMessage
	var relaySecret string
	var relayPolling telegramRelayPolling
	if len(patch.MetadataJSON) > 0 && depTypeKey == "telegram.relay" {
		var err error
		var sanitized string
		sanitized, relayKey, relayPublicURL, relayAllowed, relayDefaultChatID, relayCountryRelays, relaySecret, relayPolling, err = telegramRelayMetadataForStorage(patch.MetadataJSON, false)
		if err != nil {
			return err
		}
		patch.MetadataJSON = json.RawMessage(sanitized)
	}
	if patch.Status != nil && *patch.Status == "deleted" {
		return errors.New("use DELETE to archive integration")
	}
	var sets []string
	var args []any
	i := 1
	if patch.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, *patch.Name)
		i++
	}
	if patch.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", i))
		args = append(args, *patch.Status)
		i++
	}
	if len(patch.MetadataJSON) > 0 {
		sets = append(sets, fmt.Sprintf("metadata_json = $%d::jsonb", i))
		args = append(args, string(patch.MetadataJSON))
		i++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, depID.String())
	q := fmt.Sprintf(`UPDATE action_dependencies SET %s, updated_at = now() WHERE action_dependency_id = $%d`,
		strings.Join(sets, ", "), i)
	tx, err := d.transaction()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, q, args...); err != nil {
		return err
	}
	if depTypeKey == "telegram.relay" && len(patch.MetadataJSON) > 0 {
		if relaySecret != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(relaySecret), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			_, err = relayconfig.Save(ctx, tx, relayconfig.Settings{DependencyID: depID.String(), RelayKey: relayKey, PublicURL: relayPublicURL, SecretHash: string(hash), AllowedUpdates: relayAllowed, DefaultChatID: relayDefaultChatID, CountryRelays: relayCountryRelays, Polling: relayconfig.Polling{Enabled: relayPolling.Enabled, TimeoutSeconds: relayPolling.TimeoutSeconds, Limit: relayPolling.Limit, DeleteWebhook: relayPolling.DeleteWebhook, DropPendingUpdates: relayPolling.DropPendingUpdates}})
			if err != nil {
				return err
			}
		} else {
			tag, err := relayconfig.Save(ctx, tx, relayconfig.Settings{DependencyID: depID.String(), RelayKey: relayKey, PublicURL: relayPublicURL, SecretHash: "", AllowedUpdates: relayAllowed, DefaultChatID: relayDefaultChatID, CountryRelays: relayCountryRelays, Polling: relayconfig.Polling{Enabled: relayPolling.Enabled, TimeoutSeconds: relayPolling.TimeoutSeconds, Limit: relayPolling.Limit, DeleteWebhook: relayPolling.DeleteWebhook, DropPendingUpdates: relayPolling.DropPendingUpdates}})
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return fmt.Errorf("%w: webhook_secret нужен для первой настройки relay", ErrTelegramRelayMetadata)
			}
		}
	}
	return nil
}

func (d *Deps) Delete(ctx context.Context, pid, depID, uid uuid.UUID) error {
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return err
	}
	_, err := d.P.Exec(ctx, `
UPDATE action_dependencies SET status = 'deleted', updated_at = now()
 WHERE action_dependency_id = $1 AND project_id = $2`,
		depID.String(), pid.String())
	return err
}

func (d *Deps) RotateSecret(ctx context.Context, pid, depID, uid uuid.UUID, plaintext string) error {
	if plaintext == "" {
		return ErrPlaintextRequired
	}
	if err := d.assertDepInProject(ctx, pid, depID, uid); err != nil {
		return err
	}
	enc, err := cryptoapp.Seal([]byte(plaintext), d.MasterHEX)
	if err != nil {
		return err
	}
	var secretID string
	if err := d.P.QueryRow(ctx, `
SELECT COALESCE(ad.secret_id::text, '')
  FROM action_dependencies ad
 WHERE ad.action_dependency_id = $1 AND ad.project_id = $2`,
		depID.String(), pid.String()).Scan(&secretID); err != nil {
		return err
	}
	if secretID != "" {
		_, err = d.P.Exec(ctx, `UPDATE secrets SET encrypted_value = $2 WHERE secret_id = $1`,
			secretID, enc)
		return err
	}
	sid := uuid.New()
	_, err = d.P.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, name, encrypted_value, created_by_user_id)
VALUES ($1,$2,$3,$4,$5)`,
		sid.String(), pid.String(), "dep:"+depID.String(), enc, uid.String())
	if err != nil {
		return err
	}
	_, err = d.P.Exec(ctx, `
UPDATE action_dependencies SET secret_id = $2, status = 'active', updated_at = now()
 WHERE action_dependency_id = $1`,
		depID.String(), sid.String())
	return err
}
