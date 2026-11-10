import { useCallback, useEffect, useRef, useState } from "react";

import type { AIChatSession, AIPageContext, AIToolCall } from "../../../lib/api";
import { cancelAIEmployeeSession, postMessageStream } from "../api/aiEmployeeApi";
import { buildAIEmployeeMessageInput } from "../aiEmployeeMessageInput";
import type { WidgetChatMessage } from "../aiEmployeeTypes";

type UseAIEmployeeStreamInput = {
  activeProjectId?: string | null;
  token?: string | null;
  pageContext: AIPageContext;
  pathname: string;
  fullscreen: boolean;
  session: AIChatSession | null;
  pendingCall?: AIToolCall;
  ensureSession: () => Promise<AIChatSession | null>;
  persistSessionID: (session: AIChatSession) => void;
  setError: (value: string) => void;
  setNotice: (value: string) => void;
  setSession: (session: AIChatSession | null) => void;
  scrollToBottom: (behavior?: ScrollBehavior, force?: boolean) => void;
};

export function useAIEmployeeStream({
  activeProjectId,
  token,
  pageContext,
  pathname,
  fullscreen,
  session,
  pendingCall,
  ensureSession,
  persistSessionID,
  setError,
  setNotice,
  setSession,
  scrollToBottom,
}: UseAIEmployeeStreamInput) {
  const [sending, setSending] = useState(false);
  const [input, setInput] = useState("");
  const [optimisticMessages, setOptimisticMessages] = useState<WidgetChatMessage[]>([]);
  const [streamingText, setStreamingText] = useState("");
  const [streamingStatuses, setStreamingStatuses] = useState<string[]>([]);
  const streamAbortRef = useRef<AbortController | null>(null);
  const epochRef = useRef(0);

  const clearStreamingState = useCallback(() => {
    setStreamingText("");
    setStreamingStatuses([]);
  }, []);

  const sendMessage = useCallback(async (text: string) => {
    const body = text.trim();
    if (!body || !token || !activeProjectId || streamAbortRef.current) return;
    const controller = new AbortController();
    streamAbortRef.current = controller;
    const epoch = ++epochRef.current;
    const isCurrent = () => epoch === epochRef.current;
    setSending(true);
    setNotice("");
    setError("");
    setInput("");
    setOptimisticMessages([optimisticUserMessage(body)]);
    clearStreamingState();
    scrollToBottom("auto", true);
    try {
      const current = await ensureSession();
      if (!isCurrent()) return;
      if (!current) throw new Error("Не удалось открыть диалог. Повторите отправку.");
      const messageInput = buildAIEmployeeMessageInput({ body, current, fullscreen, pageContext, pathname, pendingCall, session });
      const updated = await postMessageStream(activeProjectId, current.ai_chat_session_id, token, messageInput, (event) => {
        if (!isCurrent()) return;
        if (event.status) {
          setStreamingStatuses((items) => [...items.slice(-5), event.status as string]);
        }
        if (event.text) {
          setStreamingText((value) => value + event.text);
        }
        if (event.message && event.type === "ERROR") {
          setStreamingStatuses((items) => [...items.slice(-5), event.message as string]);
        }
        scrollToBottom("auto");
      }, controller.signal);
      if (!isCurrent()) return;
      setSession(updated);
      setOptimisticMessages([]);
      persistSessionID(updated);
      clearStreamingState();
      scrollToBottom("auto");
    } catch (err) {
      if (!isCurrent()) return;
      setOptimisticMessages([]);
      if (err instanceof DOMException && err.name === "AbortError") {
        setNotice("Запрос остановлен.");
      } else {
        setInput(body);
        setError(err instanceof Error ? err.message : "Не удалось отправить сообщение");
      }
      clearStreamingState();
    } finally {
      if (streamAbortRef.current === controller) {
        streamAbortRef.current = null;
        setSending(false);
      }
    }
  }, [
    activeProjectId,
    clearStreamingState,
    ensureSession,
    fullscreen,
    pageContext,
    pathname,
    pendingCall?.ai_tool_call_id,
    persistSessionID,
    scrollToBottom,
    session?.active_patch_proposal_id,
    setError,
    setNotice,
    setSession,
    token,
  ]);

  const stopCurrentRequest = useCallback(() => {
    const epoch = ++epochRef.current;
    streamAbortRef.current?.abort();
    streamAbortRef.current = null;
    setSending(false);
    clearStreamingState();
    setOptimisticMessages([]);
    setNotice("Запрос остановлен.");
    if (token && activeProjectId && session?.ai_chat_session_id) {
      void cancelAIEmployeeSession(activeProjectId, session.ai_chat_session_id, token).then((updated) => {
        if (epoch !== epochRef.current) return;
        setSession(updated);
        persistSessionID(updated);
      }).catch(() => undefined);
    }
  }, [activeProjectId, clearStreamingState, persistSessionID, session?.ai_chat_session_id, setNotice, setSession, token]);

  useEffect(() => {
    epochRef.current++;
    streamAbortRef.current?.abort();
    streamAbortRef.current = null;
    setSending(false); setInput("");
    setOptimisticMessages([]);
    clearStreamingState();
    return () => { epochRef.current++; streamAbortRef.current?.abort(); streamAbortRef.current = null; };
  }, [activeProjectId, token, clearStreamingState]);

  return {
    input,
    optimisticMessages,
    sendMessage,
    sending,
    setInput,
    stopCurrentRequest,
    streamingStatuses,
    streamingText,
  };
}

function optimisticUserMessage(body: string): WidgetChatMessage {
  const now = new Date().toISOString();
  return {
    ai_chat_message_id: `optimistic-${Date.now()}`,
    role: "user",
    body,
    payload_json: { kind: "optimistic_user_message" },
    created_at: now,
    optimistic: true,
  };
}
