package http

import (
	"github.com/go-chi/chi/v5"
	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"
)

func MountWidgetScriptRoute(r chi.Router) {
	r.Get("/site-chat/widget.js", PublicWidgetJS())
}

func MountPublicRoutes(r chi.Router, siteChat *sitechatsvc.Commands, hub *Hub) {
	r.Get("/public/site-chat/widgets/{public_key}/config", PublicWidgetConfig(siteChat))
	r.Get("/public/site-chat/voice-config", PublicVoiceConfig())
	r.Get("/public/site-chat/ws", PublicWS(hub))
	r.Post("/public/site-chat/presence", Presence(siteChat))
	r.Post("/public/site-chat/events", Events(siteChat))
	r.Post("/public/site-chat/conversations", PublicConversation(siteChat))
	r.Get("/public/site-chat/conversations/{conversation_id}/messages", PublicMessages(siteChat))
	r.Post("/public/site-chat/conversations/{conversation_id}/attachments", PublicAttachmentUpload(siteChat))
	r.Post("/public/site-chat/messages", PublicMessage(siteChat, hub))
	r.Post("/public/site-chat/calls/request", PublicCallRequest(siteChat, hub))
	r.Post("/public/site-chat/calls/{call_session_id}/accept", PublicCallAccept(siteChat, hub))
	r.Post("/public/site-chat/calls/{call_session_id}/reject", PublicCallReject(siteChat, hub))
	r.Post("/public/site-chat/calls/{call_session_id}/end", PublicCallEnd(siteChat, hub))
	r.Get("/site-chat/attachments/{attachment_id}/{download_token}", AttachmentDownload(siteChat))
}

func MountOperatorRoutes(r chi.Router, siteChat *sitechatsvc.Commands, hub *Hub) {
	r.Patch("/site-chat-widgets/{chat_widget_id}", PatchWidget(siteChat))
	r.Get("/site-chat/ws/operator", OperatorWS(hub))
	r.Get("/site-chat/voice-config", OperatorVoiceConfig())
	r.Post("/site-visit-sessions/{visit_session_id}/monitoring/start", StartMonitoring(siteChat, hub))
	r.Post("/site-visit-sessions/{visit_session_id}/monitoring/stop", StopMonitoring(siteChat, hub))
	r.Post("/site-visit-sessions/{visit_session_id}/invite", Invite(siteChat, hub))
	r.Post("/site-visit-sessions/{visit_session_id}/call-offer", OperatorCallOffer(siteChat, hub))
	r.Get("/site-visit-sessions/{visit_session_id}/events", VisitEvents(siteChat))
	r.Get("/site-visit-sessions/{visit_session_id}/snapshot", VisitSnapshot(siteChat))
	r.Post("/site-chat/calls/{call_session_id}/accept", OperatorCallAccept(siteChat, hub))
	r.Post("/site-chat/calls/{call_session_id}/reject", OperatorCallReject(siteChat, hub))
	r.Post("/site-chat/calls/{call_session_id}/end", OperatorCallEnd(siteChat, hub))
	r.Get("/site-chat/calls/{call_session_id}/recordings", OperatorCallRecordings(siteChat))
	r.Get("/site-chat-conversations/{conversation_id}/messages", ConversationMessages(siteChat))
	r.Post("/site-chat-conversations/{conversation_id}/attachments", OperatorAttachmentUpload(siteChat))
	r.Post("/site-chat-conversations/{conversation_id}/messages", OperatorMessage(siteChat, hub))
}

func MountProjectRoutes(r chi.Router, siteChat *sitechatsvc.Commands) {
	r.Get("/site-chat-widgets", ListWidgets(siteChat))
	r.Post("/site-chat-widgets", CreateWidget(siteChat))
	r.Get("/site-visitors/active", ActiveVisitors(siteChat))
}
