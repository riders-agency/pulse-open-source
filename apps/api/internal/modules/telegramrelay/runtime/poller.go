package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	stdhttp "net/http"
	"strings"
	"time"

	depqueries "riders.agency/pulse/api/internal/modules/action_dependencies/queries"

	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/executor"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
	"riders.agency/pulse/api/internal/platform/httptransport"
)

type Poller struct {
	P          *pgxpool.Pool
	E          *executor.Engine
	MasterHEX  string
	HTTPClient *stdhttp.Client
	Interval   time.Duration
}

type pollingConfig struct {
	relayConfig
	BotToken           string
	AllowedUpdates     []string
	TimeoutSeconds     int
	Limit              int
	NextOffset         *int64
	DeleteWebhook      bool
	DropPendingUpdates bool
	WebhookDeletedAt   *time.Time
}

type botAPIEnvelope struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
	ErrorCode   int             `json:"error_code"`
}

func (p *Poller) Run(ctx context.Context) {
	interval := p.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := p.tick(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Warn("telegram polling tick failed", slog.Any("err", err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Poller) tick(ctx context.Context) error {
	configs, err := p.claimConfigs(ctx, 8)
	if err != nil {
		return err
	}
	for _, cfg := range configs {
		go p.pollOne(ctx, cfg)
	}
	return nil
}

func (p *Poller) claimConfigs(ctx context.Context, limit int) ([]pollingConfig, error) {
	rows, err := p.P.Query(ctx, `
WITH picked AS (
  SELECT trc.action_dependency_id
    FROM telegram_relay_configs trc
    JOIN pulse_read.action_dependencies_v1 ad ON ad.action_dependency_id = trc.action_dependency_id
   WHERE trc.status = 'active'
     AND ad.status = 'active'
     AND trc.polling_enabled = true
     AND (trc.polling_locked_until IS NULL OR trc.polling_locked_until < now())
   ORDER BY COALESCE(trc.polling_last_started_at, 'epoch'::timestamptz), trc.updated_at
   LIMIT $1
   FOR UPDATE OF trc SKIP LOCKED
)
UPDATE telegram_relay_configs trc
   SET polling_locked_until = now() + ((trc.polling_timeout_seconds + 20)::text || ' seconds')::interval,
       polling_last_started_at = now()
  FROM picked p
  JOIN pulse_read.action_dependencies_v1 ad ON ad.action_dependency_id = p.action_dependency_id
 WHERE trc.action_dependency_id = p.action_dependency_id
 RETURNING trc.action_dependency_id::text,
       ad.project_id::text,
       trc.relay_key::text,
       trc.public_url::text,
       trc.webhook_secret_hash::text,
       trc.country_relays_json,
       trc.allowed_updates,
       trc.polling_timeout_seconds,
       trc.polling_limit,
       trc.polling_next_offset,
       trc.polling_delete_webhook,
       trc.polling_drop_pending_updates,
       trc.polling_webhook_deleted_at`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]pollingConfig, 0)
	for rows.Next() {
		var cfg pollingConfig
		if err := rows.Scan(
			&cfg.ActionDependencyID,
			&cfg.ProjectID,
			&cfg.RelayKey,
			&cfg.PublicURL,
			&cfg.WebhookSecretHash,
			&cfg.CountryRelaysJSON,
			&cfg.AllowedUpdates,
			&cfg.TimeoutSeconds,
			&cfg.Limit,
			&cfg.NextOffset,
			&cfg.DeleteWebhook,
			&cfg.DropPendingUpdates,
			&cfg.WebhookDeletedAt,
		); err != nil {
			return nil, err
		}
		encrypted, err := depqueries.EncryptedCredential(ctx, p.P, cfg.ActionDependencyID)
		if err != nil {
			return nil, err
		}
		token, err := cryptoapp.Open(encrypted, p.MasterHEX)
		if err != nil {
			_ = p.finishConfig(context.Background(), cfg.ActionDependencyID, false, nil, nil, "bot token decrypt failed")
			continue
		}
		cfg.BotToken = string(token)
		out = append(out, cfg)
	}
	return out, rows.Err()
}

func (p *Poller) pollOne(ctx context.Context, cfg pollingConfig) {
	var nextOffset *int64
	var lastUpdateID *int64
	errText := ""
	ok := false
	defer func() {
		if err := p.finishConfig(context.Background(), cfg.ActionDependencyID, ok, nextOffset, lastUpdateID, errText); err != nil {
			slog.Warn("telegram polling finish failed", slog.String("relay_key", cfg.RelayKey), slog.Any("err", err))
		}
	}()
	if cfg.DeleteWebhook && cfg.WebhookDeletedAt == nil {
		if err := p.deleteWebhook(ctx, cfg); err != nil {
			errText = err.Error()
			return
		}
		now := time.Now()
		cfg.WebhookDeletedAt = &now
	}
	updates, err := p.getUpdates(ctx, cfg)
	if err != nil {
		errText = err.Error()
		if strings.Contains(strings.ToLower(errText), "webhook") {
			_ = p.resetWebhookDeleted(context.Background(), cfg.ActionDependencyID)
		}
		return
	}
	relay := &Relay{P: p.P, E: p.E, HTTPClient: p.HTTPClient}
	for _, update := range updates {
		raw, _ := json.Marshal(update)
		if _, err := relay.IngestPayload(ctx, &cfg.relayConfig, raw, update, Origin{
			Ingress:   "polling",
			Headers:   map[string]any{"X-Riders-Relay-Ingress": "polling"},
			UserAgent: "riders-pulse-telegram-poller",
		}); err != nil {
			errText = err.Error()
			return
		}
		if updateID := int64At(update, "update_id"); updateID != nil {
			if lastUpdateID == nil || *updateID > *lastUpdateID {
				value := *updateID
				lastUpdateID = &value
				offset := value + 1
				nextOffset = &offset
			}
		}
	}
	ok = true
}

func (p *Poller) deleteWebhook(ctx context.Context, cfg pollingConfig) error {
	_, err := p.callBotAPI(ctx, cfg, "deleteWebhook", map[string]any{
		"drop_pending_updates": cfg.DropPendingUpdates,
	})
	if err != nil {
		return err
	}
	_, err = p.P.Exec(ctx, `
UPDATE telegram_relay_configs
   SET polling_webhook_deleted_at = now()
 WHERE action_dependency_id = $1`, cfg.ActionDependencyID)
	return err
}

func (p *Poller) getUpdates(ctx context.Context, cfg pollingConfig) ([]map[string]any, error) {
	body := map[string]any{
		"timeout":         cfg.TimeoutSeconds,
		"limit":           cfg.Limit,
		"allowed_updates": cfg.AllowedUpdates,
	}
	if cfg.NextOffset != nil {
		body["offset"] = *cfg.NextOffset
	}
	raw, err := p.callBotAPI(ctx, cfg, "getUpdates", body)
	if err != nil {
		return nil, err
	}
	var updates []map[string]any
	if err := json.Unmarshal(raw, &updates); err != nil {
		return nil, fmt.Errorf("getUpdates result json: %w", err)
	}
	return updates, nil
}

func (p *Poller) callBotAPI(ctx context.Context, cfg pollingConfig, method string, body map[string]any) (json.RawMessage, error) {
	client := p.HTTPClient
	if client == nil {
		client = httptransport.NewDefaultClient(time.Duration(cfg.TimeoutSeconds+15) * time.Second)
	}
	payload, _ := json.Marshal(body)
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutSeconds+15)*time.Second)
	defer cancel()
	req, err := stdhttp.NewRequestWithContext(reqCtx, stdhttp.MethodPost, telegramBotAPIURL(cfg.BotToken, method), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	var env botAPIEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("%s invalid response: %w", method, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !env.OK {
		if env.Description != "" {
			return nil, fmt.Errorf("%s failed: %s", method, env.Description)
		}
		return nil, fmt.Errorf("%s failed: HTTP %d", method, resp.StatusCode)
	}
	return env.Result, nil
}

func telegramBotAPIURL(token, method string) string {
	return "https://api.telegram.org/bot" + token + "/" + method
}

func (p *Poller) finishConfig(ctx context.Context, depID string, ok bool, nextOffset, lastUpdateID *int64, errText string) error {
	if ok {
		_, err := p.P.Exec(ctx, `
UPDATE telegram_relay_configs
   SET polling_locked_until = NULL,
       polling_last_ok_at = now(),
       polling_last_error = NULL,
       polling_next_offset = COALESCE($2, polling_next_offset),
       polling_last_update_id = COALESCE($3, polling_last_update_id)
 WHERE action_dependency_id = $1`, depID, int64OrNil(nextOffset), int64OrNil(lastUpdateID))
		return err
	}
	_, err := p.P.Exec(ctx, `
UPDATE telegram_relay_configs
   SET polling_locked_until = NULL,
       polling_last_error = $2
 WHERE action_dependency_id = $1`, depID, nullIfEmpty(errText))
	return err
}

func (p *Poller) resetWebhookDeleted(ctx context.Context, depID string) error {
	_, err := p.P.Exec(ctx, `
UPDATE telegram_relay_configs
   SET polling_webhook_deleted_at = NULL
 WHERE action_dependency_id = $1`, depID)
	return err
}
