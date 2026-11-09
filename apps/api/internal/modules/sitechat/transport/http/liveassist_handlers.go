package http

import (
	"net/http"
	"strconv"

	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"

	sitechat "riders.agency/pulse/api/internal/modules/sitechat/repository"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

func OperatorCallOffer(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
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
		var body sitechat.CallInput
		if !decodeJSON(w, r, &body) {
			return
		}
		call, err := s.OperatorCallOffer(r.Context(), sessionID, claims.UserID, body)
		if writeSiteChatErr(w, err) {
			return
		}
		broadcastCall(hub, call, "call.offer")
		problem.JSON(w, http.StatusCreated, call)
	}
}

func PublicCallRequest(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body sitechat.CallInput
		if !decodeJSON(w, r, &body) {
			return
		}
		call, err := s.PublicCallRequest(r.Context(), body)
		if writeSiteChatErr(w, err) {
			return
		}
		broadcastCall(hub, call, "call.request")
		problem.JSON(w, http.StatusCreated, call)
	}
}

func PublicCallAccept(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return callTransitionHandler(s, hub, "visitor", "accept")
}

func PublicCallReject(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return callTransitionHandler(s, hub, "visitor", "reject")
}

func PublicCallEnd(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return callTransitionHandler(s, hub, "visitor", "end")
}

func OperatorCallAccept(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return callTransitionHandler(s, hub, "operator", "accept")
}

func OperatorCallReject(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return callTransitionHandler(s, hub, "operator", "reject")
}

func OperatorCallEnd(s *sitechatsvc.Commands, hub *Hub) http.HandlerFunc {
	return callTransitionHandler(s, hub, "operator", "end")
}

func callTransitionHandler(s *sitechatsvc.Commands, hub *Hub, actor, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callID, ok := pathUUID(w, r, "call_session_id")
		if !ok {
			return
		}
		var body sitechat.CallInput
		if r.ContentLength != 0 && !decodeJSON(w, r, &body) {
			return
		}
		var (
			call *sitechat.CallSessionRow
			err  error
		)
		switch action {
		case "accept":
			call, err = s.AcceptCall(r.Context(), callID, actor, body)
		case "reject":
			call, err = s.RejectCall(r.Context(), callID, actor, body)
		case "end":
			if actor == "visitor" {
				call, err = s.PublicEndCall(r.Context(), callID, body)
			} else {
				call, err = s.EndCall(r.Context(), callID, body)
			}
		default:
			call, err = s.EndCall(r.Context(), callID, body)
		}
		if writeSiteChatErr(w, err) {
			return
		}
		broadcastCall(hub, call, "call."+action)
		problem.JSON(w, http.StatusOK, call)
	}
}

func VisitEvents(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := pathUUID(w, r, "visit_session_id")
		if !ok {
			return
		}
		since := int64(0)
		if raw := r.URL.Query().Get("since_sequence"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err == nil && parsed > 0 {
				since = parsed
			}
		}
		rows, err := s.CobrowseFrames(r.Context(), sessionID, since)
		writeSiteChat(w, http.StatusOK, rows, err)
	}
}

func VisitSnapshot(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := pathUUID(w, r, "visit_session_id")
		if !ok {
			return
		}
		out, err := s.Snapshot(r.Context(), sessionID)
		writeSiteChat(w, http.StatusOK, out, err)
	}
}

func broadcastCall(hub *Hub, call *sitechat.CallSessionRow, kind string) {
	if hub == nil || call == nil {
		return
	}
	scope := callScope(call)
	msg := envelope(kind, scope, map[string]any{"call_session": call})
	msg.CallSessionID = call.CallSessionID
	hub.broadcastConversation(scope, msg)
}

func callScope(call *sitechat.CallSessionRow) sitechat.LiveAssistScope {
	return sitechat.LiveAssistScope{
		CompanyID:      call.CompanyID,
		ProjectID:      call.ProjectID,
		ChatWidgetID:   call.ChatWidgetID,
		VisitorID:      call.VisitorID,
		VisitSessionID: call.VisitSessionID,
		ConversationID: call.ConversationID,
	}
}
