package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	depqueries "riders.agency/pulse/api/internal/modules/action_dependencies/queries"

	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
	"riders.agency/pulse/api/internal/platform/httptransport"
)

const botAPIProxyMaxBody = 50 << 20

// BotAPIProxy forwards Bot API calls through Riders Pulse by relay_key.
//
// Clients use a Telegram-compatible endpoint template:
// https://pulse.riders.agency/api/v1/telegram/bot-api/bot%s/%s
// where %s is the relay_key, not the real BotFather token. The request must
// include the relay secret in X-Relay-Secret or Authorization: Bearer.

func (h *Relay) LoadBotAPIConfig(ctx context.Context, relayKey string) (*relayBotAPIConfig, error) {
	row := h.P.QueryRow(ctx, `
SELECT trc.action_dependency_id::text,
       ad.project_id::text,
       trc.relay_key::text,
       trc.public_url::text,
       trc.webhook_secret_hash::text,
       trc.country_relays_json
  FROM telegram_relay_configs trc
  JOIN pulse_read.action_dependencies_v1 ad ON ad.action_dependency_id = trc.action_dependency_id
 WHERE trc.relay_key = $1
   AND trc.status = 'active'
   AND ad.status = 'active'`, relayKey)
	var cfg relayBotAPIConfig
	if err := row.Scan(
		&cfg.ActionDependencyID,
		&cfg.ProjectID,
		&cfg.RelayKey,
		&cfg.PublicURL,
		&cfg.WebhookSecretHash,
		&cfg.CountryRelaysJSON,
	); err != nil {
		return nil, err
	}
	encrypted, err := depqueries.EncryptedCredential(ctx, h.P, cfg.ActionDependencyID)
	if err != nil {
		return nil, err
	}
	token, err := cryptoapp.Open(encrypted, h.MasterHEX)
	if err != nil {
		return nil, fmt.Errorf("bot token decrypt failed")
	}
	cfg.BotToken = string(token)
	return &cfg, nil
}

func (h *Relay) ForwardBotAPI(ctx context.Context, token, method string, body []byte, contentType string) ([]byte, int, string, error) {
	client := h.HTTPClient
	if client == nil {
		client = httptransport.NewDefaultClient(30 * time.Second)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, telegramBotAPIURL(token, method), bytes.NewReader(body))
	if err != nil {
		return nil, 0, "", err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, botAPIProxyMaxBody))
	return raw, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func ValidBotAPIMethod(method string) bool {
	if method == "" || len(method) > 80 {
		return false
	}
	for _, r := range method {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}
