import type { SiteChatCallSession } from "../../../lib/api";
import type { SiteChatState } from "../model/useSiteChatState";

export type SiteChatConsoleActions = {
  copyText: (value: string, label: string) => Promise<void>;
  createWidget: () => Promise<void>;
  invite: (session: NonNullable<SiteChatState["selectedVisitor"]>) => Promise<void>;
  openSettings: () => void;
  removePendingAttachment: (id: string) => void;
  saveWidget: (widget: NonNullable<SiteChatState["selectedWidget"]>) => Promise<void>;
  selectVisitor: (session: NonNullable<SiteChatState["selectedVisitor"]>) => void;
  sendMessage: () => Promise<void>;
  startMonitoring: (session: NonNullable<SiteChatState["selectedVisitor"]>) => Promise<void>;
  stopMonitoring: (session: NonNullable<SiteChatState["selectedVisitor"]>) => Promise<void>;
  toggleWidget: (widget: NonNullable<SiteChatState["selectedWidget"]>) => Promise<void>;
  uploadAttachmentFiles: (files: FileList | File[]) => void;
};

export type SiteChatConsoleVoice = {
  acceptCall: (call: SiteChatCallSession) => Promise<void>;
  callVisitor: (session: NonNullable<SiteChatState["selectedVisitor"]>) => Promise<void>;
  endCall: () => Promise<void>;
  playRemoteAudio: () => void;
  rejectCall: (call: SiteChatCallSession) => Promise<void>;
  toggleMute: () => void;
};
