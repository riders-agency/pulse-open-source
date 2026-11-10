"use client";

import { useState } from "react";
import { ArrowLeft, MessageSquare, Info, Bell, MonitorSmartphone, Phone, PhoneOff, RefreshCw, Search, Settings, Users, Wifi, WifiOff } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import type { SiteChatCallSession } from "../../../lib/api";
import type { VisitorFilter } from "../model/types";
import type { SiteChatState } from "../model/useSiteChatState";
import { CallPanel } from "./call-panel";
import { ChatPanel } from "./chat-panel";
import { CobrowseViewport } from "./cobrowse-viewport";
import { EventTimeline } from "./event-timeline";
import { SiteChatFullscreenLayer } from "./SiteChatFullscreenLayer";
import type { SiteChatConsoleActions, SiteChatConsoleVoice } from "./site-chat-console-view-types";
import { VisitorCard, VisitorHeader } from "./visitor";
import { VisitorFacts } from "./visitor-facts";
import { WidgetSettings } from "./widget-settings";

export function SiteChatConsoleView({
  actions,
  load,
  state,
  voice,
}: {
  actions: SiteChatConsoleActions;
  load: () => Promise<void>;
  state: SiteChatState;
  voice: SiteChatConsoleVoice;
}) {
  const [mobileConversation, setMobileConversation] = useState(false);
  return (
    <main className="page live-assist-page" data-feature="live-assist">
      <header className="live-assist-topbar">
        <div>
          <h1>Чат на сайте</h1>
          <p className="muted">{state.activeProject?.name ?? "Выберите проект"} · диалоги с посетителями</p>
        </div>
        <div className="live-assist-statusbar">
          <Badge tone={state.wsStatus === "online" ? "success" : "warning"}>{state.wsStatus === "online" ? <Wifi size={14} /> : <WifiOff size={14} />} {state.wsStatus === "online" ? "На связи" : state.wsStatus === "connecting" ? "Подключение…" : "Обновляем каждые 4 секунды"}</Badge>
          <Badge className="live-counter"><Users size={14} /> {state.counters.online} на сайте</Badge>
          <Badge className="live-counter">{state.counters.chat} в чате</Badge>
          <Badge className="live-counter">{state.counters.calls} звонков</Badge>
          <Button size="sm" variant="secondary" onClick={actions.openSettings}><Settings size={16} /> Виджет</Button>
          <Button size="sm" variant="secondary" onClick={() => void load().catch((err) => state.setError(err instanceof Error ? err.message : "Не удалось обновить чат"))}><RefreshCw size={16} /> Обновить</Button>
        </div>
      </header>
      {state.error ? <div className="notice danger" role="alert">{state.error}</div> : null}
      {state.copied ? <div className="notice success">{state.copied}</div> : null}
      {state.mediaError ? <div className="notice danger">{state.mediaError}</div> : null}
      {!state.activeProjectId ? <div className="empty-state">Выберите проект, чтобы открыть чат.</div> : (
        <section className={`live-assist-shell ${mobileConversation ? "is-conversation-open" : ""}`}>
          <VisitorsAside actions={actions} state={state} voice={voice} onSelect={() => setMobileConversation(true)} />
          <LiveWorkspace actions={actions} state={state} voice={voice} onBack={() => setMobileConversation(false)} />
        </section>
      )}
      {state.settingsOpen ? (
        <WidgetSettings
          widgets={state.widgets}
          form={state.widgetForm}
          setForm={state.setWidgetForm}
          selectedWidget={state.selectedWidget}
          onCreate={() => void actions.createWidget()}
          onSave={(widget) => void actions.saveWidget(widget)}
          onToggle={(widget) => void actions.toggleWidget(widget)}
          onCopy={(value, label) => void actions.copyText(value, label)}
          onClose={() => state.setSettingsOpen(false)}
        />
      ) : null}
      <SiteChatFullscreenLayer actions={actions} state={state} voice={voice} />
    </main>
  );
}

function VisitorsAside({ actions, state, voice, onSelect }: { actions: SiteChatConsoleActions; state: SiteChatState; voice: SiteChatConsoleVoice; onSelect: () => void }) {
  return (
    <aside className="live-visitors">
      <div className="live-search">
        <Search size={16} />
        <Input value={state.query} onChange={(event) => state.setQuery(event.target.value)} aria-label="Найти посетителя" placeholder="Имя, контакт или страница" />
      </div>
      <div className="live-filters">
        {(["all", "online", "offline", "invited", "chat", "call", "monitoring", "intent"] as VisitorFilter[]).map((item) => (
          <Button size="sm" variant={state.filter === item ? "secondary" : "ghost"} key={item} aria-pressed={state.filter === item} onClick={() => state.setFilter(item)}>{visitorFilterLabels[item]}</Button>
        ))}
      </div>
      <div className="live-visitor-list">
        {state.filteredVisitors.length === 0 ? <div className="empty-state">Посетителей в этом фильтре нет</div> : null}
        {state.filteredVisitors.map((session) => (
          <VisitorCard key={session.visit_session_id} session={session} active={session.visit_session_id === state.selectedVisitor?.visit_session_id} onClick={() => { actions.selectVisitor(session); onSelect(); }} onCall={() => void voice.callVisitor(session)} />
        ))}
      </div>
    </aside>
  );
}

function LiveWorkspace({ actions, state, voice, onBack }: { actions: SiteChatConsoleActions; state: SiteChatState; voice: SiteChatConsoleVoice; onBack: () => void }) {
  const [pane, setPane] = useState("chat");
  if (!state.selectedVisitor) return <section className="live-welcome"><MessageSquare size={32} /><h2>Здесь начинаются диалоги</h2><p>Выберите посетителя слева, чтобы прочитать историю или ответить. Новые обращения появятся автоматически.</p></section>;
  return (
    <section className="live-workspace">
      <Button className="live-mobile-back" variant="ghost" onClick={onBack}><ArrowLeft size={16} /> Все диалоги</Button>
      <VisitorHeader
        session={state.selectedVisitor}
        onInvite={() => state.selectedVisitor && void actions.invite(state.selectedVisitor)}
        onCall={() => state.selectedVisitor && void voice.callVisitor(state.selectedVisitor)}
        onStartMonitoring={() => state.selectedVisitor && void actions.startMonitoring(state.selectedVisitor)}
        onStopMonitoring={() => state.selectedVisitor && void actions.stopMonitoring(state.selectedVisitor)}
        onFullscreen={() => state.setFullscreen(true)}
      />
      {state.incomingCalls.length ? <IncomingCalls calls={state.incomingCalls} voice={voice} /> : null}
      <div className="live-workspace-tabs" role="tablist" aria-label="Раздел диалога">
        {([{ id: "chat", label: "Переписка", icon: MessageSquare }, { id: "screen", label: "Экран", icon: MonitorSmartphone }, { id: "details", label: "Сведения", icon: Info }]).map(({ id, label, icon: Icon }) => <Button key={id} role="tab" id={`chat-tab-${id}`} aria-controls={`chat-pane-${id}`} aria-selected={pane === id} variant={pane === id ? "secondary" : "ghost"} onClick={() => setPane(id)}><Icon size={15} />{label}</Button>)}
      </div>
      <div className="live-main-grid">
        <div className="live-primary-panel live-chat-area" id="chat-pane-chat" role="tabpanel" aria-labelledby="chat-tab-chat" hidden={pane !== "chat"}>
          <ChatPanel
            messages={state.messages}
            draft={state.draft}
            setDraft={state.setDraft}
            pendingAttachments={state.pendingAttachments}
            onFiles={actions.uploadAttachmentFiles}
            onRemoveAttachment={actions.removePendingAttachment}
            onSend={() => void actions.sendMessage()}
            disabled={!state.selectedConversationId || state.sending || state.pendingAttachments.some((item) => item.uploading)}
          />
        </div>
        <div className="live-primary-panel live-webvisor-area" id="chat-pane-screen" role="tabpanel" aria-labelledby="chat-tab-screen" hidden={pane !== "screen"}>
          {state.fullscreen ? (
            <div className="cobrowse-parked"><MonitorSmartphone size={18} /><strong>Веб-визор открыт на весь экран</strong></div>
          ) : (
            <CobrowseViewport key={`inline-${state.selectedVisitor?.visit_session_id ?? "none"}`} snapshot={state.selectedSnapshot} frames={state.frames} fullscreen={false} />
          )}
        </div>
        <aside className="live-side-panel" id="chat-pane-details" role="tabpanel" aria-labelledby="chat-tab-details" hidden={pane !== "details"}>
          <CallPanel call={state.activeCall} recordings={state.callRecordings} muted={state.muted} connectionStatus={state.callConnectionStatus} remoteAudioReady={state.remoteAudioReady} remoteAudioRef={state.remoteAudioRef} onPlayRemoteAudio={voice.playRemoteAudio} onMute={voice.toggleMute} onEnd={() => void voice.endCall()} />
          <VisitorFacts snapshot={state.selectedSnapshot} session={state.selectedVisitor} />
          <EventTimeline frames={state.frames} messages={state.messages} compact />
        </aside>
      </div>
    </section>
  );
}

function IncomingCalls({ calls, voice }: { calls: SiteChatCallSession[]; voice: SiteChatConsoleVoice }) {
  return (
    <div className="incoming-call-strip">
      {calls.map((call) => (
        <div key={call.call_session_id}>
          <Bell size={16} />
          <strong>Посетитель просит звонок</strong>
          <span>{call.media_kind === "video" ? "Видеозвонок" : "Аудиозвонок"}</span>
          <Button size="sm" onClick={() => void voice.acceptCall(call)}><Phone size={14} /> Принять</Button>
          <Button size="sm" variant="secondary" onClick={() => void voice.rejectCall(call)}><PhoneOff size={14} /> Отклонить</Button>
        </div>
      ))}
    </div>
  );
}

const visitorFilterLabels: Record<VisitorFilter, string> = { all: "Все", online: "На сайте", offline: "Ушли", invited: "Приглашены", chat: "В чате", call: "Звонок", monitoring: "Показ экрана", intent: "Нужна помощь" };
