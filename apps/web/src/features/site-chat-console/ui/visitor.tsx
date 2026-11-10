import { Expand, Phone, Send } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import type { SiteVisitorSession } from "../../../lib/api";
import { formatDateTime } from "../../../lib/utils";
import { formatDuration, isOnline, shortURL, visitorTitle } from "../model/format";

export function VisitorCard({ session, active, onClick, onCall }: { session: SiteVisitorSession; active: boolean; onClick: () => void; onCall: () => void }) {
  const snapshot = session.visitor_state_snapshot ?? {};
  const signals = snapshot.signals ?? {};
  const activity = snapshot.activity ?? {};
  const online = isOnline(session);
  return (
    <div
      role="button"
      tabIndex={0}
      className={`live-visitor-card ${active ? "is-active" : ""}`}
      onClick={onClick}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onClick();
        }
      }}
    >
      <span className={`live-dot ${online ? "is-online" : "is-offline"}`} />
      <strong>{visitorTitle(session)}</strong>
      <small>{shortURL(session.current_url)}</small>
      <span>{formatDuration(session.duration_seconds)} · старт {formatDateTime(session.started_at)}</span>
      <div className="live-card-badges">
        {!online ? <Badge tone="warning">Не на сайте</Badge> : null}
        {session.visitor_message_count > 0 ? <Badge>{session.visitor_message_count} сообщ.</Badge> : null}
        {session.visitor_message_count > 0 && !online ? <Badge tone="success">архив</Badge> : null}
        {signals.chat_open ? <Badge>В чате</Badge> : null}
        {signals.operator_invited ? <Badge tone="warning">Приглашён</Badge> : null}
        {signals.monitoring_active ? <Badge tone="success">Показ экрана</Badge> : null}
        {activity.call_status ? <Badge tone="danger">Звонок</Badge> : null}
        {signals.has_rage_click ? <Badge tone="danger">Частые клики</Badge> : null}
      </div>
      <Button size="sm" variant="ghost" disabled={!online} onClick={(event) => { event.stopPropagation(); if (online) onCall(); }}><Phone size={14} /> Позвонить</Button>
    </div>
  );
}

export function VisitorHeader({
  session,
  onInvite,
  onCall,
  onStartMonitoring,
  onStopMonitoring,
  onFullscreen,
}: {
  session: SiteVisitorSession | null;
  onInvite: () => void;
  onCall: () => void;
  onStartMonitoring: () => void;
  onStopMonitoring: () => void;
  onFullscreen: () => void;
}) {
  const monitoring = session?.visitor_state_snapshot?.signals?.monitoring_active;
  return (
    <header className="visitor-workspace-head">
      <div>
        <strong>{visitorTitle(session)}</strong>
        <span>{session?.current_url ?? "Нет выбранной сессии"}</span>
        <div className="row">
          <Badge tone={session && isOnline(session) ? "success" : "warning"}>{session && isOnline(session) ? "На сайте" : "Не на сайте"}</Badge>
          {session?.visitor_message_count ? <Badge tone="success">{session.visitor_message_count} сообщ.</Badge> : null}
          <Badge>{formatDuration(session?.duration_seconds)}</Badge>
        </div>
      </div>
      <div className="visitor-actions">
        <Button size="sm" variant="secondary" disabled={!session || !isOnline(session)} onClick={onInvite}><Send size={14} /> Пригласить</Button>
        <Button size="sm" disabled={!session || !isOnline(session)} onClick={onCall}><Phone size={14} /> Позвонить</Button>
        <Button size="sm" variant="secondary" disabled={!session} onClick={monitoring ? onStopMonitoring : onStartMonitoring}>{monitoring ? "Показ включен" : "Включить показ"}</Button>
        <Button size="sm" variant="secondary" disabled={!session} onClick={onFullscreen}><Expand size={14} /> Развернуть</Button>
      </div>
    </header>
  );
}
