package http

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	relayruntime "riders.agency/pulse/api/internal/modules/telegramrelay/runtime"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

const botAPIProxyMaxBody = 50 << 20

func BotAPIProxy(h *Relay) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		relayKey := strings.TrimPrefix(strings.TrimSpace(chi.URLParam(r, "relay_key")), "bot")
		method := strings.TrimSpace(chi.URLParam(r, "method"))
		if relayKey == "" || !relayruntime.ValidBotAPIMethod(method) {
			problem.Write(w, http.StatusBadRequest, "telegram bot api route", "")
			return
		}
		cfg, err := h.LoadBotAPIConfig(r.Context(), relayKey)
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
		raw, err := io.ReadAll(io.LimitReader(r.Body, botAPIProxyMaxBody+1))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "request body", "")
			return
		}
		if len(raw) > botAPIProxyMaxBody {
			problem.Write(w, http.StatusRequestEntityTooLarge, "request body too large", "")
			return
		}
		out, status, contentType, err := h.ForwardBotAPI(r.Context(), cfg.BotToken, method, raw, r.Header.Get("Content-Type"))
		if err != nil {
			problem.Write(w, http.StatusBadGateway, "telegram bot api", err.Error())
			return
		}
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		} else {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(status)
		_, _ = w.Write(out)
	}
}
