import type { AIChatSession, AIMessageInput, AIPageContext, AIToolCall } from "../../lib/api";
import { lastAssistantMessageID, localDateISO } from "./aiEmployeeUtils";

export function buildAIEmployeeMessageInput({
  body,
  current,
  fullscreen,
  pageContext,
  pathname,
  pendingCall,
  session,
}: {
  body: string;
  current: AIChatSession;
  fullscreen: boolean;
  pageContext: AIPageContext;
  pathname: string;
  pendingCall?: AIToolCall;
  session: AIChatSession | null;
}): AIMessageInput {
  const activeToolCallID = pendingCall?.ai_tool_call_id || "";
  const activePatchProposalID = session?.active_patch_proposal_id || "";
  return {
    body,
    page_context_json: pageContext,
    client_context_json: {
      active_tool_call_id: activeToolCallID,
      active_proposal_id: activeToolCallID,
      composer_mode: activeToolCallID ? "edit_pending_proposal" : "chat",
      source_surface: "widget",
    },
    ui_state_json: {
      source_surface: "widget",
      route: pageContext.route,
      pathname,
      open: true,
      pending_tool_call_id: activeToolCallID,
      selected_action_key: pageContext.selectedActionKey || "",
      visible_panel: pageContext.visiblePanel || "",
      fullscreen,
    },
    active_tool_call_id: activeToolCallID,
    active_patch_proposal_id: activePatchProposalID,
    reply_to_message_id: lastAssistantMessageID(current.messages ?? []),
    client_command: body.startsWith("/") ? body : "",
    local_date: localDateISO(),
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "Europe/Moscow",
  };
}
