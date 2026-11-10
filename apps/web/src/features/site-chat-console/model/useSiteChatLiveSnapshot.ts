"use client";

import { useCallback } from "react";

import type { SiteChatWSEnvelope, SiteCobrowseFrame } from "../../../lib/api";
import { eventPayloads, frameKindFromEvent, mergeLiveSnapshotEvent, payloadString, sequenceFromPayload } from "./live-snapshot";
import type { SiteChatState } from "./useSiteChatState";

export function useSiteChatLiveSnapshot(state: SiteChatState) {
  const patchVisitorConnectionStatus = useCallback((message: SiteChatWSEnvelope, status: "online" | "offline") => {
    const payload = message.payload ?? {};
    const visitSessionId = message.visit_session_id || payloadString(payload, "visit_session_id") || "";
    const conversationId = message.conversation_id || payloadString(payload, "conversation_id") || "";
    if (!visitSessionId && !conversationId) return;
    const at = typeof payload.last_seen_at === "string" ? payload.last_seen_at : new Date().toISOString();
    state.setVisitors((prev) => prev.map((session) => {
      const matchesSession = visitSessionId && session.visit_session_id === visitSessionId;
      const matchesConversation = conversationId && session.conversation_id === conversationId;
      if (!matchesSession && !matchesConversation) return session;
      const snapshot = session.visitor_state_snapshot ?? {};
      return {
        ...session,
        last_activity_at: at,
        visitor_state_snapshot: {
          ...snapshot,
          activity: { ...snapshot.activity, connection_status: status, is_active: status === "online", last_seen_at: at },
        },
      };
    }));
  }, [state]);

  const patchVisitorLiveSnapshot = useCallback((message: SiteChatWSEnvelope) => {
    const payload = message.payload ?? {};
    const visitSessionId = message.visit_session_id || payloadString(payload, "visit_session_id") || "";
    const conversationId = message.conversation_id || payloadString(payload, "conversation_id") || "";
    if (!visitSessionId && !conversationId) return;
    const events = eventPayloads(message);
    if (!events.length) return;
    state.setVisitors((prev) => prev.map((session) => {
      const matchesSession = visitSessionId && session.visit_session_id === visitSessionId;
      const matchesConversation = conversationId && session.conversation_id === conversationId;
      if (!matchesSession && !matchesConversation) return session;
      const visitor_state_snapshot = events.reduce(
        (snapshot, event) => mergeLiveSnapshotEvent(snapshot, event.type, event.payload),
        session.visitor_state_snapshot ?? {},
      );
      return {
        ...session,
        current_url: visitor_state_snapshot.current_url ?? session.current_url,
        current_title: visitor_state_snapshot.current_title ?? session.current_title,
        last_activity_at: visitor_state_snapshot.activity?.last_seen_at ?? session.last_activity_at,
        visitor_state_snapshot,
      };
    }));
  }, [state]);

  const appendLiveFrames = useCallback((message: SiteChatWSEnvelope) => {
    const visitSessionId = message.visit_session_id || payloadString(message.payload ?? {}, "visit_session_id") || "";
    if (!visitSessionId || visitSessionId !== state.selectedVisitSessionId) return;
    const events = eventPayloads(message);
    if (!events.length) return;
    const at = message.sent_at || new Date().toISOString();
    const liveFrames = events.flatMap((event, index) => {
      const kind = frameKindFromEvent(event.type);
      if (!kind) return [];
      return [{
        frame_id: `live_${message.id}_${index}`,
        visit_session_id: visitSessionId,
        sequence: sequenceFromPayload(event.payload, Date.now() + index),
        kind,
        occurred_at: at,
        payload_json: event.payload,
        created_at: at,
      } satisfies SiteCobrowseFrame];
    });
    if (!liveFrames.length) return;
    state.setFrames((prev) => {
      const byKey = new Map(prev.map((frame) => [`${frame.visit_session_id}:${frame.sequence}:${frame.kind}`, frame]));
      liveFrames.forEach((frame) => byKey.set(`${frame.visit_session_id}:${frame.sequence}:${frame.kind}`, frame));
      return Array.from(byKey.values()).sort((a, b) => a.sequence - b.sequence).slice(-500);
    });
  }, [state]);

  return { appendLiveFrames, patchVisitorConnectionStatus, patchVisitorLiveSnapshot };
}
