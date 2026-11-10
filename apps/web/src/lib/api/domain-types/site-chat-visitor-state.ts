import type { SiteChatMirrorSnapshot, VisitorViewportContext } from "./site-chat-snapshot";

export type VisitorStateSnapshot = {
  visitor_id?: string;
  visit_session_id?: string;
  chat_widget_id?: string;
  conversation_id?: string;
  conversation_url?: string;
  panel_url?: string;
  call_url?: string;
  current_url?: string;
  current_title?: string;
  viewport?: { width?: number; height?: number; layout_width?: number; layout_height?: number; scale?: number; offset_x?: number; offset_y?: number };
  document?: { width?: number; height?: number };
  scroll?: { x?: number; y?: number; document_width?: number; document_height?: number };
  cursor?: { x?: number; y?: number };
  last_click?: { x?: number; y?: number };
  analytics?: {
    referrer?: string;
    browser?: string;
    os?: string;
    device_type?: string;
    screen?: { width?: number; height?: number };
    language?: string;
    timezone?: string;
    metrika?: { client_id?: string; user_id?: string; session_id?: string };
  };
  activity?: {
    last_seen_at?: string;
    dwell_seconds?: number;
    clicks_last_minute?: number;
    is_active?: boolean;
    connection_status?: "online" | "offline" | "reconnecting" | string;
    call_status?: string;
    call_session_id?: string;
  };
  signals?: {
    is_first_visit?: boolean;
    has_rage_click?: boolean;
    has_dead_click?: boolean;
    is_high_intent?: boolean;
    chat_open?: boolean;
    operator_invited?: boolean;
    monitoring_active?: boolean;
  };
  suggested_actions?: string[];
  monitoring_status?: string;
  touch?: { x?: number; y?: number };
  widget_state?: { open?: boolean; active_tab?: string; url?: string; title?: string };
  viewport_context?: VisitorViewportContext;
  mirror?: SiteChatMirrorSnapshot;
};
