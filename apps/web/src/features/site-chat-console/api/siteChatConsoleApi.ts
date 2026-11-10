import {
  apiGet,
  apiPatch,
  apiPost,
  apiPostVoid,
  apiUpload,
  asArray,
  type SiteChatAttachment,
  type SiteChatCallSession,
  type SiteChatConversation,
  type SiteChatMessage,
  type SiteChatWidget,
  type SiteCobrowseFrame,
  type SiteVisitorSession,
} from "../../../lib/api";
import type { SiteChatVoiceConfig, WidgetForm } from "../model/types";

export async function listActiveVisitors(projectId: string, token: string) {
  return asArray<SiteVisitorSession>(await apiGet<SiteVisitorSession[]>(`/api/v1/projects/${projectId}/site-visitors/active`, token));
}

export async function listSiteChatWidgets(projectId: string, token: string) {
  return asArray<SiteChatWidget>(await apiGet<SiteChatWidget[]>(`/api/v1/projects/${projectId}/site-chat-widgets`, token));
}

export async function listConversationMessages(conversationId: string, token: string) {
  return asArray<SiteChatMessage>(await apiGet<SiteChatMessage[]>(`/api/v1/site-chat-conversations/${conversationId}/messages`, token));
}

export async function listCallRecordings(callSessionId: string, token: string) {
  return asArray<SiteChatAttachment>(await apiGet<SiteChatAttachment[]>(`/api/v1/site-chat/calls/${callSessionId}/recordings`, token));
}

export async function getSiteChatVoiceConfig(token: string) {
  return apiGet<SiteChatVoiceConfig>("/api/v1/site-chat/voice-config", token);
}

export async function listCobrowseFrames(visitSessionId: string, since: number, token: string) {
  return asArray<SiteCobrowseFrame>(await apiGet<SiteCobrowseFrame[]>(`/api/v1/site-visit-sessions/${visitSessionId}/events?since_sequence=${since}`, token));
}

export async function uploadConversationAttachment(token: string, conversationId: string, file: Blob, fileName: string, callSessionId?: string) {
  const form = new FormData();
  form.append("file", file, fileName);
  if (callSessionId) form.append("call_session_id", callSessionId);
  return apiUpload<SiteChatAttachment>(`/api/v1/site-chat-conversations/${conversationId}/attachments`, token, form);
}

export async function createSiteChatWidget(projectId: string, token: string, payload: WidgetForm & Record<string, unknown>) {
  return apiPost(`/api/v1/projects/${projectId}/site-chat-widgets`, token, payload);
}

export async function updateSiteChatWidget(widgetId: string, token: string, payload: Record<string, unknown>) {
  return apiPatch(`/api/v1/site-chat-widgets/${widgetId}`, token, payload);
}

export async function inviteSiteVisitor(visitSessionId: string, token: string) {
  return apiPost<SiteChatConversation>(`/api/v1/site-visit-sessions/${visitSessionId}/invite`, token, {
    message: "Команда на связи. Помочь на этой странице?",
    force: true,
  });
}

export async function startVisitorMonitoring(visitSessionId: string, token: string) {
  return apiPostVoid(`/api/v1/site-visit-sessions/${visitSessionId}/monitoring/start`, token);
}

export async function stopVisitorMonitoring(visitSessionId: string, token: string) {
  return apiPostVoid(`/api/v1/site-visit-sessions/${visitSessionId}/monitoring/stop`, token);
}

export async function offerVisitorCall(session: SiteVisitorSession, token: string) {
  return apiPost<SiteChatCallSession>(`/api/v1/site-visit-sessions/${session.visit_session_id}/call-offer`, token, {
    conversation_id: session.conversation_id,
    media_kind: "audio",
    recording_enabled: true,
  });
}

export async function acceptSiteChatCall(callSessionId: string, token: string) {
  return apiPost<SiteChatCallSession>(`/api/v1/site-chat/calls/${callSessionId}/accept`, token, {});
}

export async function rejectSiteChatCall(callSessionId: string, token: string) {
  return apiPost<SiteChatCallSession>(`/api/v1/site-chat/calls/${callSessionId}/reject`, token, { end_reason: "operator_rejected" });
}

export async function endSiteChatCall(callSessionId: string, token: string) {
  return apiPost<SiteChatCallSession>(`/api/v1/site-chat/calls/${callSessionId}/end`, token, { end_reason: "operator_ended" });
}

export async function sendConversationMessage(conversationId: string, token: string, body: string, attachmentIds: string[]) {
  return apiPost(`/api/v1/site-chat-conversations/${conversationId}/messages`, token, { body, attachment_ids: attachmentIds }, { inlineError: true });
}
