package http

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	aictx "riders.agency/pulse/api/internal/modules/ai/context"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	"riders.agency/pulse/api/internal/platform/problem"
)

func CreateSession(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		var in airepo.SessionInput
		if !decode(w, r, &in) {
			return
		}
		in.PageContextJSON = aictx.NormalizePageContext(in.PageContextJSON).RawJSON()
		out, err := ai.CreateSession(r.Context(), projectID, userID, in)
		write(w, http.StatusCreated, out, err)
	}
}

func GetSession(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "session_id", "")
			return
		}
		out, err := ai.GetSession(r.Context(), projectID, sessionID, userID)
		if err == nil {
			out = attachLangGraphThreadState(r.Context(), out, projectID, sessionID, userID)
		}
		write(w, http.StatusOK, out, err)
	}
}

func PostMessage(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "session_id", "")
			return
		}
		var in airepo.MessageInput
		if !decode(w, r, &in) {
			return
		}
		normalizeMessageContext(&in)
		if !requireAIMessageCapability(w, r, in) {
			return
		}
		out, err := streamLangGraphTurn(r.Context(), httptest.NewRecorder(), nil, ai, projectID, sessionID, userID, in)
		write(w, http.StatusOK, out, err)
	}
}

func PostMessageStream(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "session_id", "")
			return
		}
		var in airepo.MessageInput
		if !decode(w, r, &in) {
			return
		}
		normalizeMessageContext(&in)
		if !requireAIMessageCapability(w, r, in) {
			return
		}

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)

		runtime := aiEmployeeRuntime()
		if runtime == "langgraph" {
			out, err := streamLangGraphTurn(r.Context(), w, flusher, ai, projectID, sessionID, userID, in)
			if err != nil {
				writeSSE(w, flusher, "error", map[string]any{
					"type":    "ERROR",
					"message": aiStreamErrorMessage(err),
					"code":    err.Error(),
				})
				if out == nil {
					if snapshot, snapErr := ai.GetSession(context.Background(), projectID, sessionID, userID); snapErr == nil {
						out = attachLangGraphThreadState(context.Background(), snapshot, projectID, sessionID, userID)
					}
				}
				if out != nil {
					writeSSE(w, flusher, "session", map[string]any{
						"type":    "MESSAGE_FINAL",
						"runtime": runtime,
						"session": out,
					})
				}
				writeSSE(w, flusher, "done", map[string]any{"type": "DONE", "ok": false})
				return
			}
			writeSSE(w, flusher, "session", map[string]any{
				"type":    "MESSAGE_FINAL",
				"runtime": runtime,
				"session": out,
			})
			writeSSE(w, flusher, "done", map[string]any{"type": "DONE", "ok": true})
			return
		}

		writeSSE(w, flusher, "error", map[string]any{"type": "ERROR", "message": "Сотрудник недоступен: требуется LangGraph.", "code": "runtime_unavailable"})
		writeSSE(w, flusher, "done", map[string]any{"type": "DONE", "ok": false})
	}
}

func CancelSession(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "session_id", "")
			return
		}
		runtimeCtx, err := ai.SessionRuntimeContext(r.Context(), projectID, sessionID)
		if err != nil {
			write(w, http.StatusOK, nil, err)
			return
		}
		if _, err := cancelLangGraphSession(r.Context(), runtimeCtx, userID); err != nil {
			write(w, http.StatusOK, nil, err)
			return
		}
		out, err := ai.CancelSession(r.Context(), projectID, sessionID, userID)
		if err == nil {
			out = attachLangGraphThreadState(r.Context(), out, projectID, sessionID, userID)
		}
		write(w, http.StatusOK, out, err)
	}
}

func normalizeMessageContext(in *airepo.MessageInput) {
	if in == nil {
		return
	}
	in.PageContextJSON = aictx.NormalizePageContext(in.PageContextJSON).RawJSON()
}
