"use client";

import { type SetStateAction, useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";

import type { SiteChatAttachment, SiteChatCallSession, SiteChatMessage, SiteChatWidget, SiteCobrowseFrame, SiteVisitorSession } from "../../../lib/api";
import type { LiveAssistSocket } from "../../../lib/site-chat-ws";
import { usePulseStore } from "../../../store/pulse-store";
import { isOnline } from "./format";
import type { CallConnectionStatus, PendingAttachment, SiteChatVoiceConfig, VisitorFilter, WidgetForm } from "./types";
import { initialWidget } from "./widget";

export function useSiteChatState() {
  const searchParams = useSearchParams();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const projects = usePulseStore((s) => s.projects);
  const activeProject = useMemo(() => projects.find((project) => project.project_id === activeProjectId), [activeProjectId, projects]);
  const [widgets, setWidgets] = useState<SiteChatWidget[]>([]);
  const [visitors, setVisitors] = useState<SiteVisitorSession[]>([]);
  const [messages, setMessages] = useState<SiteChatMessage[]>([]);
  const [frames, setFrames] = useState<SiteCobrowseFrame[]>([]);
  const [selectedVisitSessionId, setSelectedVisitSessionId] = useState(searchParams.get("visit_session_id") ?? "");
  const [selectedConversationId, setSelectedConversationId] = useState(searchParams.get("conversation_id") ?? "");
  const [filter, setFilter] = useState<VisitorFilter>("all");
  const [query, setQuery] = useState("");
  const [sending, setSending] = useState(false);
  const [drafts, setDrafts] = useState<Record<string, string>>({});
  const [attachmentDrafts, setAttachmentDrafts] = useState<Record<string, PendingAttachment[]>>({});
  const draftKey = `${activeProjectId}:${selectedConversationId}`;
  const draft = drafts[draftKey] || "";
  const pendingAttachments = attachmentDrafts[draftKey] || [];
  const setDraft = (value: string) => setDrafts((prev) => ({ ...prev, [draftKey]: value }));
  const setPendingAttachments = (value: SetStateAction<PendingAttachment[]>) => setAttachmentDrafts((prev) => ({ ...prev, [draftKey]: typeof value === "function" ? value(prev[draftKey] || []) : value }));
  const attachmentDraftsRef = useRef(attachmentDrafts);
  useEffect(() => { attachmentDraftsRef.current = attachmentDrafts; }, [attachmentDrafts]);
  useEffect(() => () => { Object.values(attachmentDraftsRef.current).flat().forEach((item) => { if (item.localUrl) URL.revokeObjectURL(item.localUrl); }); }, []);
  const [wsStatus, setWsStatus] = useState<"connecting" | "online" | "fallback">("connecting");
  const [error, setError] = useState("");
  const [copied, setCopied] = useState("");
  const [fullscreen, setFullscreen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [widgetForm, setWidgetForm] = useState<WidgetForm>(initialWidget);
  const [activeCall, setActiveCall] = useState<SiteChatCallSession | null>(null);
  const [incomingCalls, setIncomingCalls] = useState<SiteChatCallSession[]>([]);
  const [callRecordings, setCallRecordings] = useState<SiteChatAttachment[]>([]);
  const [mediaError, setMediaError] = useState("");
  const [muted, setMuted] = useState(false);
  const [_localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [voiceConfig, setVoiceConfig] = useState<SiteChatVoiceConfig | null>(null);
  const [callConnectionStatus, setCallConnectionStatus] = useState<CallConnectionStatus>("idle");
  const [remoteAudioReady, setRemoteAudioReady] = useState(false);
  const wsRef = useRef<LiveAssistSocket | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const peerRef = useRef<RTCPeerConnection | null>(null);
  const pendingIceRef = useRef<RTCIceCandidateInit[]>([]);
  const offeredCallIdRef = useRef("");
  const voiceRef = useRef({ makingOffer: false, ignoreOffer: false, settingRemoteAnswerPending: false, polite: true });
  const remoteAudioRef = useRef<HTMLAudioElement | null>(null);
  const remoteStreamRef = useRef<MediaStream | null>(null);
  const ringAudioCtxRef = useRef<AudioContext | null>(null);
  const ringTimerRef = useRef<number | null>(null);
  const recorderRef = useRef<MediaRecorder | null>(null);
  const recordedChunksRef = useRef<Blob[]>([]);
  const recordingCallRef = useRef<SiteChatCallSession | null>(null);
  const recordingMixContextRef = useRef<AudioContext | null>(null);
  const relayRecorderRef = useRef<MediaRecorder | null>(null);
  const relaySequenceRef = useRef(0);
  const relayMediaSourceRef = useRef<MediaSource | null>(null);
  const relaySourceBufferRef = useRef<SourceBuffer | null>(null);
  const relayQueueRef = useRef<ArrayBuffer[]>([]);
  const relayObjectUrlRef = useRef("");
  const relayFallbackQueueRef = useRef<{ mime: string; buffer: ArrayBuffer }[]>([]);
  const relayFallbackPlayingRef = useRef(false);
  const relayPlaybackActiveRef = useRef(false);

  const selectedVisitor =
    visitors.find((session) => session.visit_session_id === selectedVisitSessionId) ??
    visitors.find((session) => session.conversation_id === selectedConversationId) ??
    visitors[0] ??
    null;
  const selectedWidget = widgets[0] ?? null;
  const selectedSnapshot = selectedVisitor?.visitor_state_snapshot ?? {};
  const lastFrameSequence = frames[frames.length - 1]?.sequence ?? 0;
  const filteredVisitors = useMemo(() => filterVisitors(visitors, filter, query), [filter, query, visitors]);
  const counters = useMemo(() => visitorCounters(visitors, incomingCalls.length), [incomingCalls.length, visitors]);

  return {
    sending, setSending, activeCall, activeProject, activeProjectId, callConnectionStatus, callRecordings, copied, counters, draft, error,
    filter, filteredVisitors, frames, fullscreen, incomingCalls, lastFrameSequence, mediaError, messages, muted,
    pendingAttachments, query, remoteAudioReady, selectedConversationId, selectedSnapshot, selectedVisitSessionId,
    selectedVisitor, selectedWidget, settingsOpen, token, visitors, voiceConfig, widgets, widgetForm, wsStatus,
    localStreamRef, offeredCallIdRef, peerRef, pendingIceRef, recordedChunksRef, recorderRef, recordingCallRef,
    recordingMixContextRef, relayFallbackPlayingRef, relayFallbackQueueRef, relayMediaSourceRef, relayObjectUrlRef,
    relayPlaybackActiveRef, relayQueueRef, relayRecorderRef, relaySequenceRef, relaySourceBufferRef, remoteAudioRef,
    remoteStreamRef, ringAudioCtxRef, ringTimerRef, voiceRef, wsRef,
    setActiveCall, setCallConnectionStatus, setCallRecordings, setCopied, setDraft, setError, setFilter, setFrames,
    setFullscreen, setIncomingCalls, setLocalStream, setMediaError, setMessages, setMuted, setPendingAttachments,
    setQuery, setRemoteAudioReady, setSelectedConversationId, setSelectedVisitSessionId, setSettingsOpen, setVisitors,
    setVoiceConfig, setWidgets, setWidgetForm, setWsStatus,
  };
}

export type SiteChatState = ReturnType<typeof useSiteChatState>;

function filterVisitors(visitors: SiteVisitorSession[], filter: VisitorFilter, query: string) {
  const q = query.trim().toLowerCase();
  return visitors.filter((session) => {
    const snapshot = session.visitor_state_snapshot ?? {};
    const signals = snapshot.signals ?? {};
    const activity = snapshot.activity ?? {};
    const online = isOnline(session);
    if (filter === "online" && !online) return false;
    if (filter === "offline" && online) return false;
    if (filter === "invited" && !signals.operator_invited) return false;
    if (filter === "chat" && !signals.chat_open) return false;
    if (filter === "call" && !activity.call_status) return false;
    if (filter === "monitoring" && !signals.monitoring_active) return false;
    if (filter === "intent" && !signals.is_high_intent && !signals.has_rage_click) return false;
    if (!q) return true;
    return [session.current_url, session.current_title, session.visitor_name, session.visitor_contact, session.conversation_id, snapshot.analytics?.referrer].some((value) => String(value ?? "").toLowerCase().includes(q));
  });
}

function visitorCounters(visitors: SiteVisitorSession[], incomingCallCount: number) {
  const online = visitors.filter(isOnline).length;
  return {
    online,
    offline: Math.max(visitors.length - online, 0),
    chat: visitors.filter((v) => v.visitor_state_snapshot?.signals?.chat_open).length,
    calls: visitors.filter((v) => v.visitor_state_snapshot?.activity?.call_status).length + incomingCallCount,
    monitoring: visitors.filter((v) => v.visitor_state_snapshot?.signals?.monitoring_active).length,
    intent: visitors.filter((v) => v.visitor_state_snapshot?.signals?.is_high_intent || v.visitor_state_snapshot?.signals?.has_rage_click).length,
  };
}
