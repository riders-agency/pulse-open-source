package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	relayruntime "riders.agency/pulse/api/internal/modules/telegramrelay/runtime"
	"riders.agency/pulse/api/internal/platform/httputil"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

type Relay = relayruntime.Relay

func Post(h *Relay) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		relayKey := strings.TrimSpace(chi.URLParam(r, "relay_key"))
		if relayKey == "" {
			problem.Write(w, http.StatusBadRequest, "relay_key", "")
			return
		}
		cfg, err := h.LoadConfig(r.Context(), relayKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				problem.Write(w, http.StatusNotFound, "unknown relay", "")
				return
			}
			problem.Write(w, http.StatusInternalServerError, "lookup", "")
			return
		}
		credential := relayCredential(r)
		if credential == "" || bcrypt.CompareHashAndPassword([]byte(cfg.WebhookSecretHash), []byte(credential)) != nil {
			problem.Write(w, http.StatusUnauthorized, "invalid relay credential", "")
			return
		}

		raw, _ := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
			problem.Write(w, http.StatusBadRequest, "telegram update json", "")
			return
		}
		out, err := h.IngestPayload(r.Context(), cfg, raw, payload, relayruntime.Origin{
			Ingress:   "webhook",
			Headers:   requestHeaders(r),
			UserAgent: r.Header.Get("User-Agent"),
		})
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, err.Error(), "")
			return
		}
		problem.JSON(w, http.StatusAccepted, out)
	}
}

func relayCredential(r *http.Request) string {
	for _, header := range []string{"X-Telegram-Bot-Api-Secret-Token", "X-Relay-Secret"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	return httputil.Bearer(r.Header.Get("Authorization"))
}

func requestHeaders(r *http.Request) map[string]any {
	out := map[string]any{}
	for key, values := range r.Header {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "authorization") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "password") {
			out[key] = "***"
			continue
		}
		if len(values) == 1 {
			out[key] = values[0]
		} else {
			out[key] = values
		}
	}
	return out
}
