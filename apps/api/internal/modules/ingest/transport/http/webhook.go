package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"riders.agency/pulse/api/internal/modules/ingest/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/executor"
	runscontracts "riders.agency/pulse/api/internal/modules/runs/contracts"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

type Webhook struct {
	P         *pgxpool.Pool
	E         *executor.Engine
	MasterHEX string
	Receipts  runscontracts.RunReceiptReader
}

func (h *Webhook) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hk := chi.URLParam(r, "hook_key")
		ek := chi.URLParam(r, "endpoint_key")
		h.ingestKeys(w, r, hk, ek, "")
	}
}

func (h *Webhook) ingestKeys(w http.ResponseWriter, r *http.Request, hk, ek, pathSuffix string) {
	if hk == "" || ek == "" {
		problem.Write(w, http.StatusBadRequest, "path", "")
		return
	}
	base := "/hooks/" + hk + "/" + ek
	fullPath := base
	if pathSuffix != "" {
		fullPath = base + "/" + pathSuffix
	}
	endpoint, err := lookupWebhookEndpoint(r.Context(), h.P, hk, ek)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			problem.Write(w, http.StatusNotFound, "unknown hook", "")
			return
		}
		problem.Write(w, http.StatusInternalServerError, "lookup", "")
		return
	}
	pid := endpoint.ProjectID
	peid := endpoint.EndpointID
	wid := endpoint.WorkflowID
	wvid := endpoint.WorkflowVersionID

	raw, payload, err := readWebhookJSONBody(r)
	if err != nil {
		switch err {
		case errWebhookContentType:
			problem.Write(w, http.StatusUnsupportedMediaType, "webhook rejected", "")
		case errWebhookBodyTooLarge:
			problem.Write(w, http.StatusRequestEntityTooLarge, "webhook rejected", "")
		default:
			problem.Write(w, http.StatusBadRequest, "webhook rejected", "")
		}
		return
	}

	now := time.Now().UTC()
	authCtx := endpoint.Auth
	identity, err := parseWebhookIdentity(r, now, webhookReplayWindow(authCtx.replayWindowSeconds), authCtx.authType != "none")
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errWebhookTimestamp) {
			status = http.StatusUnauthorized
		}
		problem.Write(w, status, "webhook rejected", "")
		return
	}
	authResult, err := authCtx.verify(r, raw, identity, h.MasterHEX, now)
	if err != nil {
		problem.Write(w, http.StatusUnauthorized, "webhook rejected", "")
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "webhook rejected", "")
		return
	}
	receipt, err := service.Receive(r.Context(), h.P, service.WebhookInput{
		ProjectID: pid, EndpointID: peid, WorkflowID: wid, VersionID: wvid, Method: r.Method, Path: fullPath,
		Payload: payloadJSON, ContentType: r.Header.Get("Content-Type"), UserAgent: r.Header.Get("User-Agent"),
		IdempotencyKey: identity.idempotencyKey, EventID: identity.eventID, Timestamp: identity.timestamp,
		BodySHA: webhookBodySHA256(raw), AuthScheme: authResult.scheme, SignatureVerified: authResult.signatureVerified, EndpointKey: ek,
	})
	if err != nil {
		problem.Write(w, http.StatusInternalServerError, "webhook ingest", "")
		return
	}
	if h.E != nil && receipt["duplicate"] != "true" {
		h.E.Notify()
	}
	problem.JSON(w, http.StatusAccepted, receipt)
}
