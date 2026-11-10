"use client";

import { useCallback, useEffect, useRef } from "react";

import type { SiteChatCallSession, SiteChatWSEnvelope } from "../../../lib/api";
import { connectOperatorLiveAssist } from "../../../lib/site-chat-ws";
import { payloadString } from "./live-snapshot";
import type { SiteChatState } from "./useSiteChatState";

const frameMessageTypes = ["event.batch", "cursor.move", "touch.move", "click", "scroll", "viewport", "page.navigate", "monitoring.snapshot", "monitoring.incremental", "widget_state"];

export function useSiteChatLifecycle(
  state: SiteChatState,
  loaders: {
    load: () => Promise<void>;
    loadCallRecordings: (callSessionId?: string) => Promise<void>;
    loadFrames: (visitSessionId?: string, since?: number) => Promise<void>;
    loadMessages: (conversationId?: string) => Promise<void>;
    loadVisitors: () => Promise<void>;
  },
  live: {
    appendLiveFrames: (message: SiteChatWSEnvelope) => void;
    patchVisitorConnectionStatus: (message: SiteChatWSEnvelope, status: "online" | "offline") => void;
    patchVisitorLiveSnapshot: (message: SiteChatWSEnvelope) => void;
  },
  voice: {
    handleWebRTCSignal: (message: SiteChatWSEnvelope) => Promise<void>;
    playAudioRelayChunk: (payload: Record<string, unknown>) => void;
    playOperatorRing: () => void;
    startOperatorOffer: (call: Partial<SiteChatCallSession> | null) => Promise<void>;
    stopOperatorPeer: (stopMedia?: boolean) => void;
    stopOperatorRing: () => void;
  },
) {
  const stateRef = useRef(state);
  const loadedProject = useRef(state.activeProjectId);
  const loadersRef = useRef(loaders);
  const liveRef = useRef(live);
  const voiceRef = useRef(voice);
  useEffect(() => {
    stateRef.current = state;
    loadersRef.current = loaders;
    liveRef.current = live;
    voiceRef.current = voice;
  });

  const handleWSMessage = useCallback((message: SiteChatWSEnvelope) => {
    const currentState = stateRef.current;
    const currentLoaders = loadersRef.current;
    const currentLive = liveRef.current;
    const currentVoice = voiceRef.current;
    if (message.type === "presence.update") {
      const status = message.payload?.connection_status;
      if (status === "online" || status === "offline") currentLive.patchVisitorConnectionStatus(message, status);
      void currentLoaders.loadVisitors().catch(() => undefined);
    }
    if (message.type === "visitor.left") {
      currentLive.patchVisitorConnectionStatus(message, "offline");
      void currentLoaders.loadVisitors().catch(() => undefined);
    }
    if (message.type === "chat.message" && (!currentState.selectedConversationId || message.conversation_id === currentState.selectedConversationId)) {
      void currentLoaders.loadMessages(message.conversation_id).catch(() => undefined);
    }
    if (message.type === "call.request" || message.type === "call.offer") {
      const call = message.payload?.call_session as SiteChatCallSession | undefined;
      if (call) {
        currentState.setActiveCall(call);
        if (call.direction === "visitor_to_operator" && call.status === "ringing") {
          currentVoice.playOperatorRing();
          currentState.setIncomingCalls((prev) => [call, ...prev.filter((item) => item.call_session_id !== call.call_session_id)].slice(0, 5));
        }
      }
    }
    if (message.type === "call.accept" || message.type === "call.reject" || message.type === "call.end") {
      const call = message.payload?.call_session as SiteChatCallSession | undefined;
      if (call) {
        currentState.setActiveCall(call);
        currentState.setIncomingCalls((prev) => prev.filter((item) => item.call_session_id !== call.call_session_id));
      }
      currentVoice.stopOperatorRing();
      if (message.type === "call.accept" && call) void currentVoice.startOperatorOffer(call).catch(() => undefined);
      if (message.type === "call.end" || message.type === "call.reject") currentVoice.stopOperatorPeer(true);
    }
    if (message.type === "call.recording") {
      const callSessionId = message.call_session_id || payloadString(message.payload ?? {}, "call_session_id");
      if (callSessionId && currentState.activeCall?.call_session_id === callSessionId) void currentLoaders.loadCallRecordings(callSessionId).catch(() => undefined);
    }
    if (message.type === "call.audio_chunk") currentVoice.playAudioRelayChunk(message.payload ?? {});
    if (message.type === "webrtc.offer" || message.type === "webrtc.answer" || message.type === "webrtc.ice") void currentVoice.handleWebRTCSignal(message);
    if (frameMessageTypes.includes(message.type)) {
      currentLive.patchVisitorLiveSnapshot(message);
      currentLive.appendLiveFrames(message);
      if (["event.batch", "page.navigate", "monitoring.snapshot", "monitoring.incremental"].includes(message.type)) void currentLoaders.loadVisitors().catch(() => undefined);
      if (message.visit_session_id === currentState.selectedVisitor?.visit_session_id && !["cursor.move", "touch.move", "scroll", "viewport"].includes(message.type)) {
        void currentLoaders.loadFrames(message.visit_session_id, currentState.lastFrameSequence).catch(() => undefined);
      }
    }
  }, []);
  useEffect(() => {
    if (loadedProject.current !== state.activeProjectId) resetProjectConversation(stateRef.current);
    loadedProject.current = state.activeProjectId;
    if (!state.token || !state.activeProjectId) return;
    void loadersRef.current.load().catch((err) => stateRef.current.setError(err instanceof Error ? err.message : "Не удалось загрузить чат"));
  }, [state.activeProjectId, state.token]);
  useEffect(() => {
    const visitor = state.selectedVisitor;
    if (!visitor) return;
    const currentState = stateRef.current;
    if (currentState.selectedVisitSessionId !== visitor.visit_session_id) {
      currentState.setSelectedVisitSessionId(visitor.visit_session_id);
      currentState.setFrames([]);
    }
    if (currentState.selectedConversationId !== visitor.conversation_id) {
      currentState.setSelectedConversationId(visitor.conversation_id);
    }
    void loadersRef.current.loadMessages(visitor.conversation_id).catch(() => undefined);
    void loadersRef.current.loadFrames(visitor.visit_session_id, 0).catch(() => undefined);
  }, [state.selectedVisitor?.conversation_id, state.selectedVisitor?.visit_session_id]);
  useEffect(() => {
    if (!state.token || !state.activeProjectId) return;
    const currentState = stateRef.current;
    currentState.wsRef.current?.close();
    currentState.setWsStatus("connecting");
    const socket = connectOperatorLiveAssist({
      token: state.token,
      projectId: state.activeProjectId,
      handlers: {
        onOpen: () => stateRef.current.setWsStatus("online"),
        onClose: () => stateRef.current.setWsStatus("fallback"),
        onError: () => stateRef.current.setWsStatus("fallback"),
        onMessage: handleWSMessage,
      },
    });
    currentState.wsRef.current = socket;
    return () => {
      socket.close();
      if (stateRef.current.wsRef.current === socket) stateRef.current.wsRef.current = null;
    };
  }, [handleWSMessage, state.activeProjectId, state.token]);
  useEffect(() => {
    if (state.wsStatus === "online") return;
    const timer = window.setInterval(() => {
      const currentLoaders = loadersRef.current;
      const currentState = stateRef.current;
      void currentLoaders.loadVisitors().catch(() => undefined);
      void currentLoaders.loadMessages().catch(() => undefined);
      void currentLoaders.loadFrames(undefined, currentState.lastFrameSequence).catch(() => undefined);
    }, 4000);
    return () => window.clearInterval(timer);
  }, [state.wsStatus]);
  useEffect(() => {
    void loadersRef.current.loadCallRecordings(state.activeCall?.call_session_id).catch(() => undefined);
  }, [state.activeCall?.call_session_id]);
  useEffect(() => () => voiceRef.current.stopOperatorPeer(true), []);
  useEffect(() => {
    if (!state.fullscreen) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") stateRef.current.setFullscreen(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [state.fullscreen]);
  useEffect(() => {
    const currentState = stateRef.current;
    const stream = currentState.remoteStreamRef.current;
    const audio = currentState.remoteAudioRef.current;
    if (!stream || !audio) return;
    audio.srcObject = stream;
    audio.muted = false;
    audio.volume = 1;
    audio.controls = true;
  }, [state.activeCall?.call_session_id, state.callConnectionStatus, state.fullscreen, state.remoteAudioReady]);
}

function resetProjectConversation(state: SiteChatState) {
  state.setVisitors([]); state.setMessages([]); state.setFrames([]);
  state.setSelectedVisitSessionId(""); state.setSelectedConversationId("");
}
