package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

type Handlers struct {
	P *pgxpool.Pool
}

type startOut struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	ExpiresAt  string `json:"expires_at"`
	VerifyURL  string `json:"verify_url"`
}

func (h *Handlers) AuthStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		dc := uuid.NewString()[:8]
		uc := uuid.NewString()[:6]
		exp := time.Now().UTC().Add(15 * time.Minute)
		if _, err := h.P.Exec(r.Context(), `
INSERT INTO desktop_auth_requests (device_code, user_code, status, expires_at)
VALUES ($1,$2,'pending',$3)`, dc, uc, exp); err != nil {
			problem.Write(w, http.StatusConflict, "start", "")
			return
		}
		vu := "/desktop/activate"
		if o := r.Header.Get("Origin"); o != "" {
			vu = o + "?code=" + uc
		}
		problem.JSON(w, http.StatusOK, startOut{
			DeviceCode: dc,
			UserCode:   uc,
			ExpiresAt:  exp.Format(time.RFC3339),
			VerifyURL:  vu,
		})
	}
}

func (h *Handlers) Poll() http.HandlerFunc {
	type pollIn struct {
		DeviceCode string `json:"device_code"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body pollIn
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		var st string
		err := h.P.QueryRow(r.Context(), `
SELECT status::text FROM desktop_auth_requests WHERE device_code = $1 ORDER BY created_at DESC LIMIT 1`,
			body.DeviceCode).Scan(&st)
		if err != nil {
			problem.JSON(w, http.StatusOK, map[string]string{"status": "pending"})
			return
		}
		problem.JSON(w, http.StatusOK, map[string]string{"status": st})
	}
}

func (h *Handlers) ListMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		problem.JSON(w, http.StatusOK, []map[string]any{})
	}
}

func (h *Handlers) AckMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		_ = chi.URLParam(r, "desktop_message_id")
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handlers) MessagesStream(w http.ResponseWriter, r *http.Request) {
	if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
		problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	fl, ok := w.(http.Flusher)
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-t.C:
			if _, err := w.Write([]byte(": hb\n\n")); err != nil {
				return
			}
			if ok {
				fl.Flush()
			}
		}
	}
}
