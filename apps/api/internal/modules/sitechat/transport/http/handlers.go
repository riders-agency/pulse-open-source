package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"
	"riders.agency/pulse/api/internal/platform/httputil"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sitechat "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

type Handlers struct {
	S *sitechatsvc.Commands
}

func ListWidgets(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := pathUUID(w, r, "project_id")
		if !ok {
			return
		}
		rows, err := s.ListWidgets(r.Context(), projectID)
		writeSiteChat(w, http.StatusOK, rows, err)
	}
}

func CreateWidget(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		projectID, ok := pathUUID(w, r, "project_id")
		if !ok {
			return
		}
		var body sitechat.WidgetInput
		if !decodeJSON(w, r, &body) {
			return
		}
		id, err := s.CreateWidget(r.Context(), projectID, claims.UserID, body)
		writeSiteChat(w, http.StatusCreated, map[string]string{"chat_widget_id": id}, err)
	}
}

func PatchWidget(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		widgetID, ok := pathUUID(w, r, "chat_widget_id")
		if !ok {
			return
		}
		var body sitechat.WidgetPatch
		if !decodeJSON(w, r, &body) {
			return
		}
		if writeSiteChatErr(w, s.PatchWidget(r.Context(), widgetID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func PublicWidgetConfig(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicKey := chi.URLParam(r, "public_key")
		cfg, err := s.PublicWidgetConfig(r.Context(), publicKey, publicBase(r))
		if cfg != nil {
			cfg.Voice = voiceConfigPayload()
		}
		writeSiteChat(w, http.StatusOK, cfg, err)
	}
}

func PublicWidgetJS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(embeddedSiteChatWidgetJS))
	}
}

func Presence(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body sitechat.PresenceInput
		if !decodeJSON(w, r, &body) {
			return
		}
		if body.PublicKey == "" {
			body.PublicKey = r.URL.Query().Get("public_key")
		}
		body.IP = httputil.ClientIP(r)
		if body.UserAgent == "" {
			body.UserAgent = r.UserAgent()
		}
		out, err := s.RecordPresence(r.Context(), body)
		if out != nil {
			out.Voice = voiceConfigPayload()
		}
		writeSiteChat(w, http.StatusOK, out, err)
	}
}

func Events(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body sitechat.EventBatchInput
		if !decodeJSON(w, r, &body) {
			return
		}
		if writeSiteChatErr(w, s.RecordEvents(r.Context(), body)) {
			return
		}
		problem.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	}
}

func ActiveVisitors(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := pathUUID(w, r, "project_id")
		if !ok {
			return
		}
		rows, err := s.ActiveVisitors(r.Context(), projectID)
		writeSiteChat(w, http.StatusOK, rows, err)
	}
}

func StartMonitoring(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := pathUUID(w, r, "visit_session_id")
		if !ok {
			return
		}
		if writeSiteChatErr(w, s.SetMonitoring(r.Context(), sessionID, "active")) {
			return
		}
		if hub != nil {
			projectID, projectErr := s.ProjectForSession(r.Context(), sessionID)
			scope, err := s.ResolveOperatorScope(r.Context(), projectID, sessionID.String(), "")
			if projectErr == nil && err == nil {
				hub.broadcastVisitors(scope, envelope("monitoring.start", scope, map[string]any{}))
				hub.broadcastOperators(scope, envelope("presence.update", scope, map[string]any{"monitoring_status": "active"}))
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func StopMonitoring(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := pathUUID(w, r, "visit_session_id")
		if !ok {
			return
		}
		if writeSiteChatErr(w, s.SetMonitoring(r.Context(), sessionID, "stopped")) {
			return
		}
		if hub != nil {
			projectID, projectErr := s.ProjectForSession(r.Context(), sessionID)
			scope, err := s.ResolveOperatorScope(r.Context(), projectID, sessionID.String(), "")
			if projectErr == nil && err == nil {
				hub.broadcastVisitors(scope, envelope("monitoring.stop", scope, map[string]any{}))
				hub.broadcastOperators(scope, envelope("presence.update", scope, map[string]any{"monitoring_status": "stopped"}))
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func Invite(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		sessionID, ok := pathUUID(w, r, "visit_session_id")
		if !ok {
			return
		}
		var body sitechat.InviteInput
		if r.ContentLength != 0 && !decodeJSON(w, r, &body) {
			return
		}
		out, err := s.InviteWithInput(r.Context(), sessionID, claims.UserID, body)
		if err == nil && hub != nil && out != nil {
			scope := sitechat.LiveAssistScope{
				CompanyID:      "",
				ProjectID:      out.ProjectID,
				ChatWidgetID:   out.ChatWidgetID,
				VisitorID:      out.VisitorID,
				VisitSessionID: out.VisitSessionID,
				ConversationID: out.ConversationID,
			}
			hub.broadcastVisitors(scope, envelope("operator.invite", scope, map[string]any{
				"message":  body.Message,
				"template": body.Template,
				"force":    body.Force,
			}))
			hub.broadcastOperators(scope, envelope("presence.update", scope, map[string]any{"operator_invited": true, "chat_open": true}))
		}
		writeSiteChat(w, http.StatusOK, out, err)
	}
}

func ConversationMessages(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationID, ok := pathUUID(w, r, "conversation_id")
		if !ok {
			return
		}
		rows, err := s.Messages(r.Context(), conversationID)
		writeSiteChat(w, http.StatusOK, rows, err)
	}
}

func OperatorMessage(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		conversationID, ok := pathUUID(w, r, "conversation_id")
		if !ok {
			return
		}
		var body sitechat.MessageInput
		if !decodeJSON(w, r, &body) {
			return
		}
		id, err := s.OperatorMessage(r.Context(), conversationID, claims.UserID, body.Body, body.AttachmentIDs)
		if err == nil && hub != nil {
			conv, convErr := s.GetConversation(r.Context(), conversationID)
			if convErr == nil {
				scope := sitechat.LiveAssistScope{
					ProjectID:      conv.ProjectID,
					ChatWidgetID:   conv.ChatWidgetID,
					VisitorID:      conv.VisitorID,
					VisitSessionID: conv.VisitSessionID,
					ConversationID: conv.ConversationID,
				}
				hub.broadcastConversation(scope, envelope("chat.message", scope, map[string]any{
					"message_id":  id,
					"body":        body.Body,
					"author_kind": "operator",
					"attachments": body.AttachmentIDs,
				}))
			}
		}
		writeSiteChat(w, http.StatusCreated, map[string]string{"message_id": id}, err)
	}
}

func PublicConversation(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body sitechat.MessageInput
		if !decodeJSON(w, r, &body) {
			return
		}
		out, err := s.PublicConversation(r.Context(), body)
		writeSiteChat(w, http.StatusCreated, out, err)
	}
}

func PublicMessage(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body sitechat.MessageInput
		if !decodeJSON(w, r, &body) {
			return
		}
		id, err := s.PublicMessage(r.Context(), body)
		if err == nil && hub != nil {
			scope, scopeErr := s.ResolvePublicScope(r.Context(), body.PublicKey, body.VisitorID, body.VisitSessionID, body.ConversationID)
			if scopeErr == nil {
				hub.broadcastConversation(scope, envelope("chat.message", scope, map[string]any{
					"message_id":  id,
					"body":        body.Body,
					"author_kind": "visitor",
					"attachments": body.AttachmentIDs,
				}))
			}
		}
		writeSiteChat(w, http.StatusCreated, map[string]string{"message_id": id}, err)
	}
}

func PublicMessages(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationID, ok := pathUUID(w, r, "conversation_id")
		if !ok {
			return
		}
		publicKey := r.URL.Query().Get("public_key")
		visitorID := r.URL.Query().Get("visitor_id")
		rows, err := s.PublicMessages(r.Context(), publicKey, conversationID, visitorID)
		writeSiteChat(w, http.StatusOK, rows, err)
	}
}

func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		problem.WriteModule(w, http.StatusBadRequest, "site_chat", "bad_id", "Некорректный идентификатор", "Параметр "+name+" должен быть UUID.", "Проверьте ссылку и повторите действие.")
		return uuid.Nil, false
	}
	return id, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		problem.WriteModule(w, http.StatusBadRequest, "site_chat", "bad_json", "Некорректный JSON", "Тело запроса не удалось разобрать.", "Проверьте payload и повторите действие.")
		return false
	}
	return true
}

func writeSiteChat[T any](w http.ResponseWriter, status int, v T, err error) {
	if writeSiteChatErr(w, err) {
		return
	}
	problem.JSON(w, status, v)
}

func writeSiteChatErr(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, sitechat.ErrNotFound):
		problem.WriteModule(w, http.StatusNotFound, "site_chat", "not_found", "Ресурс чата не найден", "Виджет, посетитель, сессия или диалог не найдены.", "Проверьте идентификатор и активный проект.")
	case errors.Is(err, sitechat.ErrVisitorOffline):
		problem.WriteModule(w, http.StatusConflict, "site_chat", "visitor_offline", "Посетитель уже не на сайте", "Голосовой звонок можно начать только для активной сессии с живым подключением.", "Дождитесь нового визита или напишите сообщение в чат.")
	case errors.Is(err, sitechat.ErrBadInput):
		problem.WriteModule(w, http.StatusBadRequest, "site_chat", "bad_input", "Данные чата некорректны", "Не хватает обязательных полей или идентификатор имеет неверный формат.", "Проверьте payload и повторите действие.")
	case errors.Is(err, errs.ErrForbidden):
		problem.Write(w, http.StatusForbidden, "forbidden", "")
	default:
		problem.WriteModule(w, http.StatusInternalServerError, "site_chat", "internal_error", "Ошибка модуля чата", err.Error(), "Повторите действие позже или проверьте логи.")
	}
	return true
}

func publicBase(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	if xf := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xf != "" {
		scheme = strings.Split(xf, ",")[0]
	}
	return scheme + "://" + r.Host
}
