import { apiBase, SiteChatWSEnvelope } from "./api";

type LiveAssistHandlers = {
  onOpen?: () => void;
  onClose?: () => void;
  onMessage?: (message: SiteChatWSEnvelope) => void;
  onError?: (error: Event) => void;
};

export type LiveAssistSocket = {
  close: () => void;
  send: (type: string, payload?: Record<string, unknown>, extra?: Partial<SiteChatWSEnvelope>) => void;
};

function wsBase() {
  const base = apiBase || (typeof window !== "undefined" ? window.location.origin : "");
  const url = new URL(base || "http://localhost");
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  return url;
}

export function connectOperatorLiveAssist({
  token,
  projectId,
  visitSessionId,
  conversationId,
  handlers,
}: {
  token: string;
  projectId: string;
  visitSessionId?: string;
  conversationId?: string;
  handlers: LiveAssistHandlers;
}): LiveAssistSocket {
  let socket: WebSocket | null = null;
  let closedByClient = false;
  let reconnectTimer: number | null = null;
  const pending: string[] = [];

  const flush = () => {
    if (!socket || socket.readyState !== WebSocket.OPEN || !pending.length) return;
    pending.splice(0).forEach((message) => socket?.send(message));
  };

  const open = () => {
    if (!projectId || !token || typeof window === "undefined") return;
    const url = wsBase();
    url.pathname = "/api/v1/site-chat/ws/operator";
    url.searchParams.set("project_id", projectId);
    // The browser sends its HttpOnly session cookie during the same-origin handshake.
    if (visitSessionId) url.searchParams.set("visit_session_id", visitSessionId);
    if (conversationId) url.searchParams.set("conversation_id", conversationId);
    socket = new WebSocket(url.toString());
    socket.onopen = () => {
      handlers.onOpen?.();
      flush();
    };
    socket.onerror = (event) => handlers.onError?.(event);
    socket.onmessage = (event) => {
      try {
        handlers.onMessage?.(JSON.parse(event.data) as SiteChatWSEnvelope);
      } catch {
        // Ignore malformed frames; the server still owns the durable state.
      }
    };
    socket.onclose = () => {
      handlers.onClose?.();
      socket = null;
      if (!closedByClient) {
        reconnectTimer = window.setTimeout(open, 2500);
      }
    };
  };

  open();

  return {
    close: () => {
      closedByClient = true;
      if (reconnectTimer) window.clearTimeout(reconnectTimer);
      socket?.close();
    },
    send: (type, payload = {}, extra = {}) => {
      const envelope: SiteChatWSEnvelope = {
        id: crypto.randomUUID ? crypto.randomUUID() : `msg_${Date.now()}`,
        type,
        sent_at: new Date().toISOString(),
        project_id: projectId,
        visit_session_id: visitSessionId,
        conversation_id: conversationId,
        payload,
        ...extra,
      };
      const serialized = JSON.stringify(envelope);
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        pending.push(serialized);
        pending.splice(0, Math.max(0, pending.length - 120));
        return;
      }
      socket.send(serialized);
    },
  };
}
