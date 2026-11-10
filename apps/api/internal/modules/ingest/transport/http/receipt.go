package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	problem "riders.agency/pulse/api/internal/platform/problem"
)

func (h *Webhook) Receipt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.receiptKeys(
			w,
			r,
			chi.URLParam(r, "hook_key"),
			chi.URLParam(r, "endpoint_key"),
			chi.URLParam(r, "workflow_run_id"),
		)
	}
}

func (h *Webhook) receiptKeys(w http.ResponseWriter, r *http.Request, hookKey, endpointKey, runID string) {
	parsedRunID, err := uuid.Parse(strings.TrimSpace(runID))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "invalid run id", "")
		return
	}
	endpoint, err := lookupWebhookEndpoint(r.Context(), h.P, hookKey, endpointKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			problem.Write(w, http.StatusNotFound, "receipt not found", "")
			return
		}
		problem.Write(w, http.StatusInternalServerError, "lookup", "")
		return
	}
	if strings.TrimSpace(endpoint.Auth.authType) == "none" {
		problem.Write(w, http.StatusUnauthorized, "receipt authorization required", "")
		return
	}
	now := time.Now().UTC()
	identity, err := parseWebhookIdentity(r, now, webhookReplayWindow(endpoint.Auth.replayWindowSeconds), true)
	if err != nil {
		problem.Write(w, http.StatusUnauthorized, "receipt rejected", "")
		return
	}
	if _, err := endpoint.Auth.verify(r, nil, identity, h.MasterHEX, now); err != nil {
		problem.Write(w, http.StatusUnauthorized, "receipt rejected", "")
		return
	}

	var owned bool
	err = h.P.QueryRow(r.Context(), `
SELECT true
  FROM webhook_events
 WHERE workflow_run_id = $1
   AND project_id = $2
   AND webhook_endpoint_id = $3`, parsedRunID, endpoint.ProjectID, endpoint.EndpointID).Scan(&owned)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			problem.Write(w, http.StatusNotFound, "receipt not found", "")
			return
		}
		problem.Write(w, http.StatusInternalServerError, "receipt lookup", "")
		return
	}
	if !owned || h.Receipts == nil {
		problem.Write(w, http.StatusInternalServerError, "receipt unavailable", "")
		return
	}
	receipt, err := h.Receipts.WebhookReceipt(
		r.Context(), parsedRunID.String(), endpoint.ProjectID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			problem.Write(w, http.StatusNotFound, "receipt not found", "")
			return
		}
		problem.Write(w, http.StatusInternalServerError, "receipt lookup", "")
		return
	}
	receipt.Terminal = webhookRunTerminal(receipt.Status)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	problem.JSON(w, http.StatusOK, receipt)
}

func webhookRunTerminal(status string) bool {
	switch strings.TrimSpace(status) {
	case "succeeded", "failed", "cancelled":
		return true
	default:
		return false
	}
}
