import type { SiteChatAttachment } from "../../../lib/api";

export type VisitorFilter = "all" | "online" | "offline" | "invited" | "chat" | "call" | "monitoring" | "intent";

export type IframePreviewStatus = "idle" | "loading" | "loaded" | "unavailable";

export type PendingAttachment = {
  id: string;
  file: File;
  localUrl?: string;
  uploading: boolean;
  error?: string;
  attachment?: SiteChatAttachment;
};

export type CallConnectionStatus = "idle" | "permission" | "connecting" | "connected" | "failed";

export type SiteChatVoiceConfig = {
  ice_servers?: RTCIceServer[];
  ice_transport_policy?: RTCIceTransportPolicy;
  audio?: {
    echo_cancellation?: boolean;
    noise_suppression?: boolean;
    auto_gain_control?: boolean;
    channel_count?: number;
  };
};

export type WidgetForm = {
  name: string;
  title: string;
  subtitle: string;
  accent_color: string;
  position: "left" | "right";
  launcher_primary_text: string;
  launcher_call_text: string;
};
