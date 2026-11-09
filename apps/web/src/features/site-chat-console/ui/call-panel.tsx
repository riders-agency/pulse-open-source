import type { RefObject } from "react";
import { Mic, MicOff, Phone, PhoneOff, Volume2 } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import type { SiteChatAttachment, SiteChatCallSession } from "../../../lib/api";
import { callLabel, callStatusLabel } from "../model/format";
import type { CallConnectionStatus } from "../model/types";
import { AttachmentGallery } from "./chat-panel";

function callConnectionLabel(status: CallConnectionStatus, remoteAudioReady: boolean) {
  if (remoteAudioReady || status === "connected") return "Звук подключен";
  if (status === "permission") return "Запрашиваем микрофон";
  if (status === "connecting") return "Устанавливаем соединение";
  if (status === "failed") return "Соединение не установлено";
  return "Ожидаем соединения";
}

export function CallPanel({
  call,
  recordings,
  muted,
  connectionStatus,
  remoteAudioReady,
  remoteAudioRef,
  onPlayRemoteAudio,
  onMute,
  onEnd,
}: {
  call: SiteChatCallSession | null;
  recordings: SiteChatAttachment[];
  muted: boolean;
  connectionStatus: CallConnectionStatus;
  remoteAudioReady: boolean;
  remoteAudioRef: RefObject<HTMLAudioElement | null>;
  onPlayRemoteAudio: () => void;
  onMute: () => void;
  onEnd: () => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle><Phone size={16} /> Звонок</CardTitle>
        <CardDescription>{callLabel(call)} · {call ? callConnectionLabel(connectionStatus, remoteAudioReady) : "нет соединения"}</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        {call ? (
          <>
            <Badge tone={call.status === "accepted" ? "success" : call.status === "rejected" || call.status === "ended" ? "warning" : "danger"}>{callStatusLabel(call.status)}</Badge>
            <Badge tone={remoteAudioReady || connectionStatus === "connected" ? "success" : connectionStatus === "failed" ? "danger" : "warning"}>{callConnectionLabel(connectionStatus, remoteAudioReady)}</Badge>
            <audio ref={remoteAudioRef} className="live-call-audio" controls autoPlay playsInline preload="auto" />
            {recordings.length ? (
              <div className="call-recordings">
                <strong>Записи разговора</strong>
                <AttachmentGallery attachments={recordings} />
              </div>
            ) : null}
            <div className="row">
              <Button variant="secondary" onClick={onPlayRemoteAudio}><Volume2 size={15} /> Включить звук</Button>
              <Button variant="secondary" onClick={onMute}>{muted ? <MicOff size={15} /> : <Mic size={15} />} {muted ? "Включить" : "Без звука"}</Button>
              <Button variant="destructive" onClick={onEnd} disabled={["ended", "rejected", "cancelled"].includes(call.status)}><PhoneOff size={15} /> Завершить</Button>
            </div>
          </>
        ) : <div className="empty-state">Активного звонка нет.</div>}
      </CardContent>
    </Card>
  );
}
