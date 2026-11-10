package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	webhookssvc "riders.agency/pulse/api/internal/modules/webhooks/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	hrepo "riders.agency/pulse/api/internal/modules/webhooks/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func ListByWorkflow(h *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		wf, err := uuid.Parse(chi.URLParam(r, "workflow_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_id", "")
			return
		}
		items, err := h.ListByWorkflow(r.Context(), wf, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		problem.JSON(w, http.StatusOK, items)
	}
}

func Create(h *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		wf, err := uuid.Parse(chi.URLParam(r, "workflow_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_id", "")
			return
		}
		var in hrepo.CreateEndpointInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if in.Name == "" || in.EndpointKey == "" || in.WorkflowVersionID == "" {
			problem.Write(w, http.StatusBadRequest, "required fields missing", "")
			return
		}
		out, err := h.Create(r.Context(), wf, c.UserID, in)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "create", err.Error())
			return
		}
		problem.JSON(w, http.StatusCreated, out)
	}
}

func Snippets(h *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		we, err := uuid.Parse(chi.URLParam(r, "webhook_endpoint_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_endpoint_id", "")
			return
		}
		hk, ek, err := h.Get(r.Context(), we, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		base := hrepo.PublicAPIBase()
		url := fmt.Sprintf("%s/hooks/%s/%s", base, hk, ek)
		j := `{"demo":true,"order":{"id":"123","total":15000}}`
		ts := "2026-06-25T08:00:00Z"
		idem := "event-123"

		out := map[string]string{
			"hook_url":    url,
			"curl":        fmt.Sprintf("curl -X POST '%s' -H 'Authorization: Bearer PROJECT_SECRET_HERE' -H 'Content-Type: application/json' -H 'X-Riders-Timestamp: %s' -H 'Idempotency-Key: %s' -d '%s'", url, ts, idem, strings.ReplaceAll(j, "'", `'\"'\"'`)),
			"javascript":  fmt.Sprintf("fetch(%q,{method:'POST',headers:{Authorization:'Bearer PROJECT_SECRET_HERE','Content-Type':'application/json','X-Riders-Timestamp':new Date().toISOString(),'Idempotency-Key':'event-123'},body:JSON.stringify(%s)})", url, j),
			"node_fetch":  fmt.Sprintf(`import fetch from "node-fetch"; fetch(%q,{method:'POST',headers:{Authorization:'Bearer PROJECT_SECRET_HERE','Content-Type':'application/json','X-Riders-Timestamp':new Date().toISOString(),'Idempotency-Key':'event-123'},body:JSON.stringify(%s)})`, url, j),
			"php":         fmt.Sprintf("$ch=curl_init(%q);curl_setopt_array($ch,[CURLOPT_POST=>true,CURLOPT_HTTPHEADER=>['Authorization: Bearer PROJECT_SECRET_HERE','Content-Type: application/json','X-Riders-Timestamp: '.gmdate('c'),'Idempotency-Key: event-123'],CURLOPT_POSTFIELDS=>%q]);curl_exec($ch);", url, j),
			"php_bitrix":  "// Bitrix: use \\Bitrix\\Main\\Web\\HttpClient similarly; keep secrets server-side.",
			"python":      fmt.Sprintf("import datetime, requests\nheaders={'Authorization':'Bearer PROJECT_SECRET_HERE','Content-Type':'application/json','X-Riders-Timestamp':datetime.datetime.now(datetime.UTC).isoformat(),'Idempotency-Key':'event-123'}\nrequests.post(%q, headers=headers, json=%s)\n", url, j),
			"go":          fmt.Sprintf("req,_:=http.NewRequest(\"POST\",%q,bytes.NewBufferString(%q))\nreq.Header.Set(\"Authorization\",\"Bearer PROJECT_SECRET_HERE\")\nreq.Header.Set(\"Content-Type\",\"application/json\")\nreq.Header.Set(\"X-Riders-Timestamp\",time.Now().UTC().Format(time.RFC3339Nano))\nreq.Header.Set(\"Idempotency-Key\",\"event-123\")\n", url, j),
			"csharp":      "// new HttpClient + PostAsync with StringContent and Bearer header",
			"idempotency": "Idempotency-Key or X-Riders-Event-Id is required; reuse the same value for safe retries of one external event.",
			"freshness":   "X-Riders-Timestamp is required for authenticated endpoints and must be inside the replay window.",
			"hmac_v1":     "Signature base: riders-webhook.v1 + LF + METHOD + LF + request_uri + LF + X-Riders-Timestamp + LF + X-Riders-Event-Id + LF + exact raw body; send hex HMAC-SHA256 as X-Riders-Signature: v1=<hex>.",
			"warning":     "Never ship real webhook secrets to browsers.",
		}
		problem.JSON(w, http.StatusOK, out)
	}
}

func GetEndpoint(wh *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "webhook_endpoint_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_endpoint_id", "")
			return
		}
		d, err := wh.GetEndpointDetail(r.Context(), id, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, d)
	}
}

func PatchEndpoint(wh *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "webhook_endpoint_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_endpoint_id", "")
			return
		}
		var body hrepo.EndpointPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if err := wh.PatchEndpoint(r.Context(), id, c.UserID, body); err != nil {
			if errors.Is(err, errs.ErrForbidden) {
				problem.Write(w, http.StatusForbidden, "forbidden", "")
				return
			}
			problem.Write(w, http.StatusBadRequest, "patch", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func RotateEndpointSecret(wh *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "webhook_endpoint_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_endpoint_id", "")
			return
		}
		sec, err := wh.RotateEndpointSecret(r.Context(), id, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "rotate", "")
			return
		}
		problem.JSON(w, http.StatusOK, map[string]string{"bearer_plain": sec, "note": "shown once"})
	}
}

func EndpointEvents(wh *hrepo.Webhooks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "webhook_endpoint_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_endpoint_id", "")
			return
		}
		rows, err := wh.ListWebhookEvents(r.Context(), id, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "list", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

func GetWebhookEvent(ev *webhookssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "webhook_event_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_event_id", "")
			return
		}
		d, err := ev.Get(r.Context(), id, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, d)
	}
}

func ReplayWebhookEvent(ev *webhookssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "webhook_event_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "webhook_event_id", "")
			return
		}
		run, err := ev.Replay(r.Context(), id, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "replay", "")
			return
		}
		problem.JSON(w, http.StatusAccepted, map[string]string{"workflow_run_id": run})
	}
}
