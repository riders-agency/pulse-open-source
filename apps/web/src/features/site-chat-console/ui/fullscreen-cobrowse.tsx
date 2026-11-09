import type { RefObject } from "react";
import { Bell, Phone, PhoneOff, X } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import type { SiteChatAttachment, SiteChatCallSession, SiteChatMessage, SiteCobrowseFrame, SiteVisitorSession, VisitorStateSnapshot } from "../../../lib/api";
import { callLabel, isOnline, visitorTitle } from "../model/format";
import type { CallConnectionStatus, PendingAttachment } from "../model/types";
import { CallPanel } from "./call-panel";
import { ChatPanel } from "./chat-panel";
import { CobrowseViewport } from "./cobrowse-viewport";

export function FullscreenCobrowse({
  session,
  snapshot,
  frames,
  messages,
  incomingCalls,
  call,
  recordings,
  draft,
  setDraft,
  pendingAttachments,
  onFiles,
  onRemoveAttachment,
  onSend,
  chatDisabled,
  muted,
  callConnectionStatus,
  remoteAudioReady,
  remoteAudioRef,
  onPlayRemoteAudio,
  onAcceptCall,
  onRejectCall,
  onMute,
  onClose,
  onEndCall,
}: {
  session: SiteVisitorSession;
  snapshot: VisitorStateSnapshot;
  frames: SiteCobrowseFrame[];
  messages: SiteChatMessage[];
  incomingCalls: SiteChatCallSession[];
  call: SiteChatCallSession | null;
  recordings: SiteChatAttachment[];
  draft: string;
  setDraft: (value: string) => void;
  pendingAttachments: PendingAttachment[];
  onFiles: (files: FileList | File[]) => void;
  onRemoveAttachment: (id: string) => void;
  onSend: () => void;
  chatDisabled: boolean;
  muted: boolean;
  callConnectionStatus: CallConnectionStatus;
  remoteAudioReady: boolean;
  remoteAudioRef: RefObject<HTMLAudioElement | null>;
  onPlayRemoteAudio: () => void;
  onAcceptCall: (call: SiteChatCallSession) => void;
  onRejectCall: (call: SiteChatCallSession) => void;
  onMute: () => void;
  onClose: () => void;
  onEndCall: () => void;
}) {
  const visibleIncomingCalls = incomingCalls.filter((item) => item.visit_session_id === session.visit_session_id || item.conversation_id === session.conversation_id);
  return (
    <div className="live-fullscreen" role="dialog" aria-modal="true" aria-label="Просмотр экрана на весь экран">
      <header>
        <div>
          <strong>{visitorTitle(session)}</strong>
          <span>{session.current_url}</span>
        </div>
        <div className="row">
          <Badge>{snapshot.viewport?.width ?? 0}×{snapshot.viewport?.height ?? 0}</Badge>
          <Badge tone={isOnline(session) ? "success" : "warning"}>{isOnline(session) ? "online" : "offline"}</Badge>
          {call ? <Badge tone={call.status === "accepted" ? "success" : "warning"}>{callLabel(call)}</Badge> : null}
          <Button variant="secondary" onClick={onClose}><X size={16} /> Закрыть</Button>
        </div>
      </header>
      <main>
        <section className="live-fullscreen-chat" aria-label="Чат с посетителем">
          {visibleIncomingCalls.length ? (
            <div className="incoming-call-strip live-fullscreen-incoming">
              {visibleIncomingCalls.map((item) => (
                <div key={item.call_session_id}>
                  <Bell size={16} />
                  <strong>Входящий звонок</strong>
                  <span>{item.media_kind}</span>
                  <Button size="sm" onClick={() => onAcceptCall(item)}><Phone size={14} /> Принять</Button>
                  <Button size="sm" variant="secondary" onClick={() => onRejectCall(item)}><PhoneOff size={14} /> Отклонить</Button>
                </div>
              ))}
            </div>
          ) : null}
          <ChatPanel
            messages={messages}
            draft={draft}
            setDraft={setDraft}
            pendingAttachments={pendingAttachments}
            onFiles={onFiles}
            onRemoveAttachment={onRemoveAttachment}
            onSend={onSend}
            disabled={chatDisabled}
          />
        </section>
        <section className="live-fullscreen-view" aria-label="Веб-визор посетителя">
          <CobrowseViewport key={`fullscreen-${session.visit_session_id}`} snapshot={snapshot} frames={frames} fullscreen />
          {call ? (
            <div className="live-fullscreen-call">
              <CallPanel call={call} recordings={recordings} muted={muted} connectionStatus={callConnectionStatus} remoteAudioReady={remoteAudioReady} remoteAudioRef={remoteAudioRef} onPlayRemoteAudio={onPlayRemoteAudio} onMute={onMute} onEnd={onEndCall} />
            </div>
          ) : null}
        </section>
      </main>
    </div>
  );
}
