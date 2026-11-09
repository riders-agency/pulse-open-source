"use client";

import { useCallback, useEffect, useRef } from "react";

import {
  getSiteChatVoiceConfig,
  listActiveVisitors,
  listCallRecordings,
  listCobrowseFrames,
  listConversationMessages,
  listSiteChatWidgets,
} from "../api/siteChatConsoleApi";
import type { SiteChatState } from "./useSiteChatState";

export function useSiteChatLoaders(state: SiteChatState) {
  const currentStateRef = useRef(state);
  useEffect(() => { currentStateRef.current = state; }, [state]);
  const visitorsInFlightRef = useRef(false);
  const messageLoadsRef = useRef(new Set<string>());
  const frameLoadsRef = useRef(new Set<string>());

  const loadVisitors = useCallback(async () => {
    if (!state.token || !state.activeProjectId) return;
    if (visitorsInFlightRef.current) return;
    visitorsInFlightRef.current = true;
    try {
      const rows = await listActiveVisitors(state.activeProjectId, state.token);
      if (currentStateRef.current.activeProjectId !== state.activeProjectId) return;
      state.setVisitors(rows);
      if (!state.selectedVisitSessionId && !state.selectedConversationId && rows[0]) {
        state.setSelectedVisitSessionId(rows[0].visit_session_id);
        state.setSelectedConversationId(rows[0].conversation_id);
      }
    } finally {
      visitorsInFlightRef.current = false;
    }
  }, [state]);

  const loadWidgets = useCallback(async () => {
    if (!state.token || !state.activeProjectId) return;
    const rows = await listSiteChatWidgets(state.activeProjectId, state.token);
    if (currentStateRef.current.activeProjectId === state.activeProjectId) state.setWidgets(rows);
  }, [state]);

  const loadMessages = useCallback(async (conversationId?: string) => {
    const cid = conversationId || state.selectedConversationId || state.selectedVisitor?.conversation_id || "";
    if (!state.token || !cid) {
      state.setMessages([]);
      return;
    }
    if (messageLoadsRef.current.has(cid)) return;
    messageLoadsRef.current.add(cid);
    try {
      const rows = await listConversationMessages(cid, state.token);
      const current = currentStateRef.current;
      if (current.activeProjectId === state.activeProjectId && (current.selectedConversationId || current.selectedVisitor?.conversation_id) === cid) state.setMessages(rows.slice(-220));
    } finally {
      messageLoadsRef.current.delete(cid);
    }
  }, [state]);

  const loadCallRecordings = useCallback(async (callSessionId?: string) => {
    const cid = callSessionId || state.activeCall?.call_session_id || "";
    if (!state.token || !cid) {
      state.setCallRecordings([]);
      return;
    }
    state.setCallRecordings(await listCallRecordings(cid, state.token));
  }, [state]);

  const loadVoiceConfig = useCallback(async () => {
    if (!state.token) return;
    state.setVoiceConfig(await getSiteChatVoiceConfig(state.token));
  }, [state]);

  const loadFrames = useCallback(async (visitSessionId?: string, since = 0) => {
    const sid = visitSessionId || state.selectedVisitor?.visit_session_id || "";
    if (!state.token || !sid) return;
    const key = `${sid}:${since}`;
    if (frameLoadsRef.current.has(key)) return;
    frameLoadsRef.current.add(key);
    try {
      const rows = await listCobrowseFrames(sid, since, state.token);
      if (currentStateRef.current.activeProjectId !== state.activeProjectId || currentStateRef.current.selectedVisitor?.visit_session_id !== sid) return;
      if (rows.length) state.setFrames((prev) => [...prev, ...rows].slice(-500));
    } finally {
      frameLoadsRef.current.delete(key);
    }
  }, [state]);

  const load = useCallback(async () => {
    state.setError("");
    await Promise.all([loadWidgets(), loadVisitors(), loadVoiceConfig()]);
  }, [loadVisitors, loadVoiceConfig, loadWidgets, state]);

  return { load, loadCallRecordings, loadFrames, loadMessages, loadVisitors, loadVoiceConfig, loadWidgets };
}
