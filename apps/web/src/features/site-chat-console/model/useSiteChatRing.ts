"use client";

import { useCallback } from "react";
import type { SiteChatState } from "./useSiteChatState";

export function useSiteChatRing(state: SiteChatState) {
  const stopOperatorRing = useCallback(() => {
    if (state.ringTimerRef.current) window.clearInterval(state.ringTimerRef.current);
    state.ringTimerRef.current = null;
  }, [state]);

  const playOperatorRing = useCallback(() => {
    if (state.ringTimerRef.current || typeof window === "undefined") return;
    const AudioCtor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!AudioCtor) return;
    try {
      state.ringAudioCtxRef.current = state.ringAudioCtxRef.current ?? new AudioCtor();
      void state.ringAudioCtxRef.current.resume().catch(() => undefined);
    } catch {
      return;
    }
    const playTone = (frequency: number, delay: number, duration: number, level: number) => {
      const ctx = state.ringAudioCtxRef.current;
      if (!ctx) return;
      try {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = "sine";
        osc.frequency.value = frequency;
        gain.gain.setValueAtTime(0.0001, ctx.currentTime + delay);
        gain.gain.exponentialRampToValueAtTime(level, ctx.currentTime + delay + 0.025);
        gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + delay + duration);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(ctx.currentTime + delay);
        osc.stop(ctx.currentTime + delay + duration + 0.03);
      } catch {}
    };
    const melody = () => {
      playTone(659.25, 0, 0.16, 0.07);
      playTone(783.99, 0.18, 0.16, 0.065);
      playTone(987.77, 0.36, 0.22, 0.055);
    };
    melody();
    state.ringTimerRef.current = window.setInterval(melody, 2600);
  }, [state]);

  return { playOperatorRing, stopOperatorRing };
}
