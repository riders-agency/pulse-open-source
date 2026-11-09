"use client";

import { useCallback } from "react";

import type { SiteChatCallSession } from "../../../lib/api";
import { uploadConversationAttachment } from "../api/siteChatConsoleApi";
import { recordingMimeType } from "./media";
import type { SiteChatState } from "./useSiteChatState";

export function useSiteChatRecording(state: SiteChatState, loadCallRecordings: (callSessionId?: string) => Promise<void>) {
  const uploadCallRecording = useCallback(async (call: SiteChatCallSession | null, blob: Blob) => {
    if (!state.token || !call?.conversation_id || !blob.size) return;
    const attachment = await uploadConversationAttachment(state.token, call.conversation_id, blob, `call-${call.call_session_id}.webm`, call.call_session_id);
    state.setCallRecordings((prev) => [attachment, ...prev.filter((item) => item.attachment_id !== attachment.attachment_id)]);
    state.wsRef.current?.send("call.recording", { attachment_id: attachment.attachment_id, call_session_id: call.call_session_id, uploaded_by_kind: "operator" }, call);
    await loadCallRecordings(call.call_session_id);
  }, [loadCallRecordings, state]);

  const stopOperatorRecording = useCallback(() => {
    const recorder = state.recorderRef.current;
    if (recorder && recorder.state !== "inactive") {
      try {
        recorder.stop();
        return;
      } catch {
        state.recorderRef.current = null;
      }
    }
    const ctx = state.recordingMixContextRef.current;
    state.recordingMixContextRef.current = null;
    void ctx?.close().catch(() => undefined);
  }, [state]);

  const startOperatorRecording = useCallback((call: SiteChatCallSession | null, local: MediaStream | null, remote: MediaStream | null) => {
    if (!call?.recording_enabled || state.recorderRef.current || !local || !remote || typeof MediaRecorder === "undefined") return;
    const localAudio = local.getAudioTracks().filter((track) => track.readyState === "live");
    const remoteAudio = remote.getAudioTracks().filter((track) => track.readyState === "live");
    if (!localAudio.length && !remoteAudio.length) return;
    try {
      const AudioCtor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
      if (!AudioCtor) return;
      const ctx = new AudioCtor();
      const destination = ctx.createMediaStreamDestination();
      const connect = (tracks: MediaStreamTrack[], gainValue: number) => {
        if (!tracks.length) return;
        const source = ctx.createMediaStreamSource(new MediaStream(tracks));
        const gain = ctx.createGain();
        gain.gain.value = gainValue;
        source.connect(gain);
        gain.connect(destination);
      };
      connect(localAudio, 0.85);
      connect(remoteAudio, 1);
      if (!destination.stream.getAudioTracks().length) {
        void ctx.close().catch(() => undefined);
        return;
      }
      state.recordingMixContextRef.current = ctx;
      state.recordedChunksRef.current = [];
      state.recordingCallRef.current = call;
      const mime = recordingMimeType();
      const recorder = mime ? new MediaRecorder(destination.stream, { mimeType: mime }) : new MediaRecorder(destination.stream);
      state.recorderRef.current = recorder;
      recorder.ondataavailable = (event) => {
        if (event.data?.size) state.recordedChunksRef.current.push(event.data);
      };
      recorder.onstop = () => {
        const chunks = state.recordedChunksRef.current;
        const recordedCall = state.recordingCallRef.current;
        const mix = state.recordingMixContextRef.current;
        state.recorderRef.current = null;
        state.recordedChunksRef.current = [];
        state.recordingCallRef.current = null;
        state.recordingMixContextRef.current = null;
        void mix?.close().catch(() => undefined);
        if (chunks.length) {
          void uploadCallRecording(recordedCall, new Blob(chunks, { type: "audio/webm" })).catch(() => state.setMediaError("Запись разговора не загрузилась."));
        }
      };
      recorder.start(1000);
    } catch {
      state.recorderRef.current = null;
      state.recordedChunksRef.current = [];
      state.recordingCallRef.current = null;
      const ctx = state.recordingMixContextRef.current;
      state.recordingMixContextRef.current = null;
      void ctx?.close().catch(() => undefined);
    }
  }, [state, uploadCallRecording]);

  return { startOperatorRecording, stopOperatorRecording };
}
