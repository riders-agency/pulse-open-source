export type { VisitorStateSnapshot } from "./site-chat-visitor-state";
import type { VisitorStateSnapshot } from "./site-chat-visitor-state";

export type SiteChatWidget = {
  chat_widget_id: string;
  company_id: string;
  project_id: string;
  public_key: string;
  name: string;
  status: "active" | "paused" | "archived";
  title: string;
  subtitle: string;
  accent_color: string;
  position: "left" | "right";
  heavy_tracking_default: boolean;
  sampling_rate: number;
  settings_json?: Record<string, unknown>;
  embed_script?: string;
  created_at?: string;
  updated_at?: string;
};

export type SiteVisitorSession = {
  visitor_id: string;
  visit_session_id: string;
  chat_widget_id: string;
  project_id: string;
  conversation_id: string;
  visitor_name?: string;
  visitor_contact?: string;
  conversation_url: string;
  panel_url: string;
  call_url: string;
  current_url: string;
  current_title: string;
  status: "active" | "idle" | "ended" | string;
  started_at: string;
  ended_at?: string;
  last_activity_at: string;
  duration_seconds: number;
  visitor_message_count: number;
  monitoring_status: "idle" | "requested" | "active" | "stopped";
  operator_invited: boolean;
  chat_open: boolean;
  visitor_state_snapshot: VisitorStateSnapshot;
};

export type SiteChatConversation = {
  conversation_id: string;
  project_id: string;
  chat_widget_id: string;
  visitor_id: string;
  visit_session_id?: string;
  status: "open" | "closed";
  title: string;
  visitor_name?: string;
  visitor_contact?: string;
  operator_url?: string;
  call_url?: string;
  created_at: string;
  updated_at: string;
};

export type SiteChatAttachment = {
  attachment_id: string;
  conversation_id: string;
  message_id?: string;
  call_session_id?: string;
  uploaded_by_kind: "visitor" | "operator" | "system";
  uploaded_by_user_id?: string;
  file_name: string;
  content_type: string;
  size_bytes: number;
  kind: "image" | "audio" | "video" | "file";
  public_url: string;
  metadata_json?: Record<string, unknown>;
  created_at: string;
};

export type SiteChatMessage = {
  message_id: string;
  conversation_id: string;
  author_kind: "visitor" | "operator" | "system";
  author_user_id?: string;
  author_name: string;
  body: string;
  attachments?: SiteChatAttachment[];
  created_at: string;
};

export type SiteChatCallSession = {
  call_session_id: string;
  company_id: string;
  project_id: string;
  chat_widget_id: string;
  visitor_id: string;
  visit_session_id?: string;
  conversation_id?: string;
  direction: "operator_to_visitor" | "visitor_to_operator";
  requested_by: "operator" | "visitor";
  requested_by_user_id?: string;
  status: "offered" | "ringing" | "accepted" | "rejected" | "missed" | "ended" | "failed" | "cancelled";
  privacy_consent_accepted_at?: string;
  recording_consent_accepted_at?: string;
  recording_enabled: boolean;
  media_kind: "audio" | "video";
  started_at?: string;
  ended_at?: string;
  last_signal_at?: string;
  end_reason?: string;
  metadata_json?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type SiteCobrowseFrame = {
  frame_id: string;
  visit_session_id: string;
  sequence: number;
  kind: "snapshot" | "mutation" | "cursor" | "touch" | "scroll" | "viewport" | "navigation" | "click" | "diagnostic";
  occurred_at: string;
  payload_json: Record<string, unknown>;
  created_at: string;
};

export type SiteChatWSEnvelope = {
  id: string;
  type: string;
  sent_at: string;
  project_id?: string;
  chat_widget_id?: string;
  visitor_id?: string;
  visit_session_id?: string;
  conversation_id?: string;
  call_session_id?: string;
  payload?: Record<string, unknown>;
};
