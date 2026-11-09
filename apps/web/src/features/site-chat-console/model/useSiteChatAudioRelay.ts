"use client";

import { useCallback } from "react";

import type { SiteChatCallSession } from "../../../lib/api";
import { audioRelayMimeType, base64ToArrayBuffer, blobToBase64 } from "./media";
import type { SiteChatState } from "./useSiteChatState";

type SendOperatorSignal = (
  call: Partial<SiteChatCallSession> | null,
  type: string,
  payload: Record<string, unknown>,
) => void;

export function useSiteChatAudioRelay({
  playRemoteAudio,
  sendOperatorSignal,
  state,
}: {
  playRemoteAudio: () => void;
  sendOperatorSignal: SendOperatorSignal;
  state: SiteChatState;
}) {
  const stopAudioRelay = useCallback(() => {
    const recorder = state.relayRecorderRef.current;
    if (recorder && recorder.state !== "inactive") {
      try {
        recorder.stop();
      } catch {}
    }
    state.relayRecorderRef.current = null;
    state.relayQueueRef.current = [];
    state.relayFallbackQueueRef.current = [];
    state.relayFallbackPlayingRef.current = false;
    state.relayPlaybackActiveRef.current = false;
    state.relaySourceBufferRef.current = null;
    state.relayMediaSourceRef.current = null;
    if (state.relayObjectUrlRef.current) {
      URL.revokeObjectURL(state.relayObjectUrlRef.current);
      state.relayObjectUrlRef.current = "";
    }
    if (state.remoteAudioRef.current) {
      state.remoteAudioRef.current.removeAttribute("src");
      state.remoteAudioRef.current.load();
    }
  }, [state]);

  const startAudioRelay = useCallback((call: Partial<SiteChatCallSession> | null, stream: MediaStream | null) => {
    if (!call?.call_session_id || !stream || state.relayRecorderRef.current || typeof MediaRecorder === "undefined" || !stream.getAudioTracks().some((track) => track.readyState === "live")) return;
    try {
      const mime = audioRelayMimeType();
      const recorder = mime ? new MediaRecorder(stream, { mimeType: mime }) : new MediaRecorder(stream);
      state.relayRecorderRef.current = recorder;
      recorder.ondataavailable = (event) => {
        if (!event.data?.size || event.data.size > 220 * 1024) return;
        void blobToBase64(event.data)
          .then((data) => {
            if (!data) return;
            state.relaySequenceRef.current += 1;
            sendOperatorSignal(call, "call.audio_chunk", {
              data,
              mime_type: event.data.type || mime || "audio/webm",
              sequence: state.relaySequenceRef.current,
            });
          })
          .catch(() => undefined);
      };
      recorder.onstop = () => {
        if (state.relayRecorderRef.current === recorder) state.relayRecorderRef.current = null;
      };
      recorder.start(500);
    } catch {
      state.relayRecorderRef.current = null;
    }
  }, [sendOperatorSignal, state]);

  const flushRelayQueue = useCallback(() => {
    const sourceBuffer = state.relaySourceBufferRef.current;
    if (!sourceBuffer || sourceBuffer.updating || state.relayQueueRef.current.length === 0) return;
    const buffer = state.relayQueueRef.current.shift();
    if (!buffer) return;
    try {
      sourceBuffer.appendBuffer(buffer);
      playRemoteAudio();
    } catch {
      window.setTimeout(flushRelayQueue, 80);
    }
  }, [playRemoteAudio, state]);

  const ensureRelayPlayback = useCallback((mime: string) => {
    const audio = state.remoteAudioRef.current, MediaSourceCtor = window.MediaSource;
    if (!audio || !mime || !MediaSourceCtor || (MediaSourceCtor.isTypeSupported && !MediaSourceCtor.isTypeSupported(mime))) return false;
    if (state.relayMediaSourceRef.current || state.relaySourceBufferRef.current) return true;
    try {
      const mediaSource = new MediaSourceCtor(), url = URL.createObjectURL(mediaSource);
      state.relayMediaSourceRef.current = mediaSource;
      state.relayObjectUrlRef.current = url;
      state.relayPlaybackActiveRef.current = true;
      audio.srcObject = null;
      audio.src = url;
      audio.muted = false;
      audio.volume = 1;
      audio.controls = true;
      mediaSource.addEventListener("sourceopen", () => {
        try {
          const sourceBuffer = mediaSource.addSourceBuffer(mime);
          sourceBuffer.mode = "sequence";
          state.relaySourceBufferRef.current = sourceBuffer;
          sourceBuffer.addEventListener("updateend", flushRelayQueue);
          flushRelayQueue();
          playRemoteAudio();
        } catch {
          state.relaySourceBufferRef.current = null;
        }
      }, { once: true });
      return true;
    } catch {
      return false;
    }
  }, [flushRelayQueue, playRemoteAudio, state]);

  const playRelayFallbackQueue = useCallback(() => {
    const audio = state.remoteAudioRef.current;
    if (!audio || state.relayFallbackPlayingRef.current || state.relayFallbackQueueRef.current.length === 0) return;
    const item = state.relayFallbackQueueRef.current.shift();
    if (!item) return;
    const url = URL.createObjectURL(new Blob([item.buffer], { type: item.mime || "audio/webm" }));
    state.relayFallbackPlayingRef.current = true;
    state.relayPlaybackActiveRef.current = true;
    audio.srcObject = null;
    audio.src = url;
    audio.muted = false;
    audio.volume = 1;
    audio.controls = true;
    audio.onended = () => {
      state.relayFallbackPlayingRef.current = false;
      URL.revokeObjectURL(url);
      playRelayFallbackQueue();
    };
    void audio.play()
      .then(() => {
        state.setRemoteAudioReady(true);
        state.setMediaError("");
      })
      .catch(() => {
        state.relayFallbackPlayingRef.current = false;
        state.setMediaError("Нажмите «Включить звук» в панели звонка.");
      });
  }, [state]);

  const playAudioRelayChunk = useCallback((payload: Record<string, unknown>) => {
    if (payload.sender === "operator" || !payload.data) return;
    let buffer: ArrayBuffer;
    try {
      buffer = base64ToArrayBuffer(payload.data);
    } catch {
      return;
    }
    const mime = String(payload.mime_type || "audio/webm");
    if (ensureRelayPlayback(mime)) {
      state.relayQueueRef.current.push(buffer);
      flushRelayQueue();
      state.setCallConnectionStatus("connected");
      return;
    }
    state.relayFallbackQueueRef.current.push({ mime, buffer });
    playRelayFallbackQueue();
    state.setCallConnectionStatus("connected");
  }, [ensureRelayPlayback, flushRelayQueue, playRelayFallbackQueue, state]);

  return { playAudioRelayChunk, startAudioRelay, stopAudioRelay };
}
