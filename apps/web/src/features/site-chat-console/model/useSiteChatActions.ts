"use client";

import { useCallback, useRef } from "react";

import type { SiteChatWidget, SiteVisitorSession } from "../../../lib/api";
import {
  createSiteChatWidget,
  inviteSiteVisitor,
  sendConversationMessage,
  startVisitorMonitoring,
  stopVisitorMonitoring,
  updateSiteChatWidget,
  uploadConversationAttachment,
} from "../api/siteChatConsoleApi";
import type { PendingAttachment } from "./types";
import type { SiteChatState } from "./useSiteChatState";
import { initialWidget, widgetFormFromWidget, widgetSettingsPayload } from "./widget";

export function useSiteChatActions(
  state: SiteChatState,
  loaders: {
    loadMessages: (conversationId?: string) => Promise<void>;
    loadVisitors: () => Promise<void>;
    loadWidgets: () => Promise<void>;
  },
) {
  const sendingRef = useRef(false);
  const selectVisitor = (session: SiteVisitorSession) => {
    if (sendingRef.current) return;
    if (session.conversation_id !== state.selectedConversationId) {
      state.setMessages([]);
    }
    state.setSelectedVisitSessionId(session.visit_session_id);
    state.setSelectedConversationId(session.conversation_id);
    state.setFrames([]);
  };

  const createWidget = async () => {
    if (!state.token || !state.activeProjectId) return;
    await createSiteChatWidget(state.activeProjectId, state.token, {
      ...state.widgetForm,
      heavy_tracking_default: false,
      sampling_rate: 1,
      settings_json: widgetSettingsPayload(state.widgetForm),
    });
    state.setWidgetForm(initialWidget);
    await loaders.loadWidgets();
  };

  const saveWidget = async (widget: SiteChatWidget) => {
    if (!state.token) return;
    await updateSiteChatWidget(widget.chat_widget_id, state.token, {
      name: state.widgetForm.name,
      title: state.widgetForm.title,
      subtitle: state.widgetForm.subtitle,
      accent_color: state.widgetForm.accent_color,
      position: state.widgetForm.position,
      settings_json: widgetSettingsPayload(state.widgetForm, widget.settings_json),
    });
    await loaders.loadWidgets();
    flashCopied(state, "Настройки виджета сохранены");
  };

  const toggleWidget = async (widget: SiteChatWidget) => {
    if (!state.token) return;
    await updateSiteChatWidget(widget.chat_widget_id, state.token, { status: widget.status === "active" ? "paused" : "active" });
    await loaders.loadWidgets();
  };

  const invite = async (session: SiteVisitorSession) => {
    if (!state.token) return;
    const out = await inviteSiteVisitor(session.visit_session_id, state.token);
    state.setSelectedVisitSessionId(session.visit_session_id);
    state.setSelectedConversationId(out.conversation_id);
    await Promise.all([loaders.loadVisitors(), loaders.loadMessages(out.conversation_id)]);
  };

  const startMonitoring = async (session: SiteVisitorSession) => {
    if (!state.token) return;
    await startVisitorMonitoring(session.visit_session_id, state.token);
    await loaders.loadVisitors();
  };

  const stopMonitoring = async (session: SiteVisitorSession) => {
    if (!state.token) return;
    await stopVisitorMonitoring(session.visit_session_id, state.token);
    await loaders.loadVisitors();
  };

  const uploadAttachmentFiles = useCallback((files: FileList | File[]) => {
    if (!state.token || !state.selectedConversationId) return;
    Array.from(files).slice(0, 8).forEach((file) => {
      if (file.size > 25 * 1024 * 1024) {
        const id = `tmp_${crypto.randomUUID()}`;
        state.setPendingAttachments((prev) => [...prev, { id, file, uploading: false, error: "До 25 МБ" }]);
        return;
      }
      const id = `tmp_${crypto.randomUUID()}`;
      const localUrl = file.type.startsWith("image/") ? URL.createObjectURL(file) : undefined;
      const pending: PendingAttachment = { id, file, localUrl, uploading: true };
      state.setPendingAttachments((prev) => [...prev, pending]);
      uploadConversationAttachment(state.token!, state.selectedConversationId, file, file.name, state.activeCall?.call_session_id)
        .then((attachment) => {
          state.setPendingAttachments((prev) => prev.map((item) => item.id === id ? { ...item, uploading: false, attachment } : item));
        })
        .catch(() => {
          state.setPendingAttachments((prev) => prev.map((item) => item.id === id ? { ...item, uploading: false, error: "Не загрузился" } : item));
        });
    });
  }, [state]);

  const removePendingAttachment = (id: string) => {
    state.setPendingAttachments((prev) => prev.filter((item) => {
      if (item.id !== id) return true;
      if (item.localUrl) URL.revokeObjectURL(item.localUrl);
      return false;
    }));
  };

  const sendMessage = async () => {
    const attachmentIds = state.pendingAttachments.flatMap((item) => item.attachment?.attachment_id ? [item.attachment.attachment_id] : []);
    const hasUploading = state.pendingAttachments.some((item) => item.uploading);
    if (!state.token || !state.selectedConversationId || hasUploading || (!state.draft.trim() && attachmentIds.length === 0)) return;
    if (sendingRef.current) return;
    sendingRef.current = true;
    state.setSending(true);
    state.setError("");
    try {
      await sendConversationMessage(state.selectedConversationId, state.token, state.draft.trim(), attachmentIds);
      state.setDraft("");
      state.pendingAttachments.forEach((item) => { if (item.localUrl) URL.revokeObjectURL(item.localUrl); });
      state.setPendingAttachments([]);
      await loaders.loadMessages(state.selectedConversationId);
    } catch (error) {
      state.setError(error instanceof Error ? error.message : "Не удалось отправить сообщение. Текст сохранён, попробуйте ещё раз.");
    } finally {
      sendingRef.current = false;
      state.setSending(false);
    }
  };

  const copyText = async (value: string, label: string) => {
    if (!value) return;
    await navigator.clipboard?.writeText(value);
    flashCopied(state, label);
  };

  const openSettings = () => {
    state.setWidgetForm(state.selectedWidget ? widgetFormFromWidget(state.selectedWidget) : initialWidget);
    state.setSettingsOpen(true);
  };

  return {
    copyText,
    createWidget,
    invite,
    openSettings,
    removePendingAttachment,
    saveWidget,
    selectVisitor,
    sendMessage,
    startMonitoring,
    stopMonitoring,
    toggleWidget,
    uploadAttachmentFiles,
  };
}

function flashCopied(state: SiteChatState, label: string) {
  state.setCopied(label);
  window.setTimeout(() => state.setCopied(""), 1400);
}
