import { useCallback } from "react";
import type { SiteChatCallSession, SiteVisitorSession } from "../../../lib/api";
import { acceptSiteChatCall, endSiteChatCall, offerVisitorCall, rejectSiteChatCall } from "../api/siteChatConsoleApi";
import { callActionErrorMessage } from "./format";
import { operatorAudioConstraints, rtcConfiguration } from "./media";
import { useSiteChatAudioRelay } from "./useSiteChatAudioRelay";
import { useSiteChatRecording } from "./useSiteChatRecording";
import { useSiteChatRing } from "./useSiteChatRing";
import type { SiteChatState } from "./useSiteChatState";

export function useSiteChatVoice(state: SiteChatState, loadCallRecordings: (callSessionId?: string) => Promise<void>) {
  const { playOperatorRing, stopOperatorRing } = useSiteChatRing(state);
  const { startOperatorRecording, stopOperatorRecording } = useSiteChatRecording(state, loadCallRecordings);

  const playRemoteAudio = useCallback(() => {
    const stream = state.remoteStreamRef.current, audio = state.remoteAudioRef.current;
    if (!audio) return;
    if (!state.relayPlaybackActiveRef.current && stream && audio.srcObject !== stream) audio.srcObject = stream;
    audio.muted = false; audio.volume = 1; audio.controls = true;
    void audio.play().then(() => { state.setMediaError(""); state.setRemoteAudioReady(true); }).catch(() => state.setMediaError("Не удалось включить звук. Проверьте системную громкость и разрешения браузера."));
  }, [state]);

  const sendOperatorSignal = useCallback((call: Partial<SiteChatCallSession> | null, type: string, payload: Record<string, unknown>) => {
    if (!call?.call_session_id) return;
    state.wsRef.current?.send(type, { ...payload, sender: "operator" }, call);
  }, [state]);
  const { playAudioRelayChunk, startAudioRelay, stopAudioRelay } = useSiteChatAudioRelay({ playRemoteAudio, sendOperatorSignal, state });

  const stopOperatorPeer = useCallback((stopMedia = false) => {
    stopOperatorRing(); stopOperatorRecording(); stopAudioRelay(); state.peerRef.current?.close(); state.peerRef.current = null;
    state.offeredCallIdRef.current = ""; state.pendingIceRef.current = []; state.voiceRef.current = { makingOffer: false, ignoreOffer: false, settingRemoteAnswerPending: false, polite: true };
    state.setRemoteAudioReady(false);
    if (state.remoteAudioRef.current) state.remoteAudioRef.current.srcObject = null;
    state.remoteStreamRef.current = null;
    if (stopMedia) { state.localStreamRef.current?.getTracks().forEach((track) => track.stop()); state.localStreamRef.current = null; state.setLocalStream(null); }
    state.setCallConnectionStatus("idle");
  }, [state, stopAudioRelay, stopOperatorRecording, stopOperatorRing]);

  const ensureOperatorStream = useCallback(async (call?: { media_kind?: SiteChatCallSession["media_kind"] } | null) => {
    const existing = state.localStreamRef.current;
    if (existing?.getAudioTracks().some((track) => track.readyState === "live")) return existing;
    existing?.getTracks().forEach((track) => track.stop()); state.localStreamRef.current = null; state.setLocalStream(null); state.setCallConnectionStatus("permission");
    const stream = await navigator.mediaDevices.getUserMedia({ ...operatorAudioConstraints(state.voiceConfig), video: call?.media_kind === "video" });
    if (!stream.getAudioTracks().some((track) => track.readyState === "live")) { stream.getTracks().forEach((track) => track.stop()); throw new Error("operator microphone unavailable"); }
    state.localStreamRef.current = stream; state.setLocalStream(stream); return stream;
  }, [state]);

  const attachRemoteAudio = useCallback((stream: MediaStream, autoplay = true) => {
    const audio = state.remoteAudioRef.current;
    state.remoteStreamRef.current = stream;
    if (state.relayPlaybackActiveRef.current || !audio) return;
    audio.srcObject = stream; audio.muted = false; audio.volume = 1; audio.controls = true;
    if (!autoplay) return;
    void audio.play().then(() => { state.setMediaError(""); state.setRemoteAudioReady(true); }).catch(() => { state.setRemoteAudioReady(false); state.setMediaError("Chrome заблокировал звук. Нажмите «Включить звук» в панели звонка."); });
  }, [state]);

  const addOperatorTracks = useCallback((pc: RTCPeerConnection, stream: MediaStream) => Promise.all(stream.getTracks().map(async (track) => {
    track.enabled = track.kind === "audio" ? !state.muted : true;
    const transceiver = pc.getTransceivers().find((item) => (item.sender.track?.kind === track.kind) || (item.receiver.track?.kind === track.kind && (!item.sender.track || item.sender.track.readyState === "ended")));
    if (transceiver) { transceiver.direction = "sendrecv"; await transceiver.sender.replaceTrack(track); return; }
    try { pc.addTrack(track, stream); } catch {}
  })).then(() => undefined), [state]);

  const ensureOperatorPeer = useCallback((call: Partial<SiteChatCallSession> | null) => {
    if (state.peerRef.current) return state.peerRef.current;
    const pc = new RTCPeerConnection(rtcConfiguration(state.voiceConfig));
    state.peerRef.current = pc; state.pendingIceRef.current = []; state.voiceRef.current = { makingOffer: false, ignoreOffer: false, settingRemoteAnswerPending: false, polite: true };
    pc.onicecandidate = (event) => { if (event.candidate && call?.call_session_id) sendOperatorSignal(call, "webrtc.ice", { candidate: event.candidate.toJSON() }); };
    pc.ontrack = (event) => {
      const stream = event.streams[0] ?? new MediaStream([event.track]);
      stream.getAudioTracks().forEach((track) => { track.enabled = true; track.onunmute = () => attachRemoteAudio(stream); });
      attachRemoteAudio(stream); startOperatorRecording(call as SiteChatCallSession, state.localStreamRef.current, stream); state.setCallConnectionStatus("connected");
    };
    pc.oniceconnectionstatechange = () => { if (pc.iceConnectionState === "failed") { try { pc.restartIce(); } catch { state.setCallConnectionStatus("failed"); } } };
    pc.onconnectionstatechange = () => { if (pc.connectionState === "connected") state.setCallConnectionStatus("connected"); if (pc.connectionState === "failed" || pc.connectionState === "disconnected") state.setCallConnectionStatus("failed"); };
    return pc;
  }, [attachRemoteAudio, sendOperatorSignal, startOperatorRecording, state]);

  const drainOperatorIce = useCallback(async (pc: RTCPeerConnection) => {
    if (!pc.remoteDescription) return;
    const pending = state.pendingIceRef.current.splice(0);
    await Promise.all(pending.map((candidate) => pc.addIceCandidate(candidate).catch(() => undefined)));
  }, [state]);

  const startOperatorOffer = useCallback(async (call: Partial<SiteChatCallSession> | null) => {
    if (!call?.call_session_id) return;
    const stream = await ensureOperatorStream(call), pc = ensureOperatorPeer(call);
    startAudioRelay(call, stream);
    if (state.offeredCallIdRef.current === call.call_session_id && pc.localDescription) return;
    state.offeredCallIdRef.current = call.call_session_id;
    const voice = state.voiceRef.current;
    state.setCallConnectionStatus("connecting");
    try {
      await addOperatorTracks(pc, stream);
      if (pc.signalingState !== "stable") return;
      voice.makingOffer = true;
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      sendOperatorSignal(call, "webrtc.offer", { description: pc.localDescription });
    } catch (error) {
      state.offeredCallIdRef.current = ""; state.setCallConnectionStatus("failed"); state.setMediaError("WebRTC-разговор не запустился. Проверьте микрофон и сеть."); throw error;
    } finally { voice.makingOffer = false; }
  }, [addOperatorTracks, ensureOperatorPeer, ensureOperatorStream, sendOperatorSignal, startAudioRelay, state]);

  const handleWebRTCSignal = useCallback(async (message: { payload?: Record<string, unknown>; type: string; call_session_id?: string; visitor_id?: string; visit_session_id?: string; conversation_id?: string }) => {
    const payload = message.payload ?? {};
    if (payload.sender === "operator") return;
    const call = state.activeCall ?? state.incomingCalls.find((item) => item.call_session_id === message.call_session_id) ?? null;
    const signalCall = call ?? { call_session_id: message.call_session_id, visitor_id: message.visitor_id, visit_session_id: message.visit_session_id, conversation_id: message.conversation_id, media_kind: "audio" as const };
    if (!signalCall.call_session_id) return;
    try {
      if (message.type === "webrtc.ice" && payload.candidate) {
        const candidate = payload.candidate as RTCIceCandidateInit, pc = state.peerRef.current;
        if (!pc || !pc.remoteDescription) { state.pendingIceRef.current.push(candidate); return; }
        await pc.addIceCandidate(candidate); return;
      }
      if (!payload.description) return;
      const pc = ensureOperatorPeer(signalCall), description = payload.description as RTCSessionDescriptionInit, voice = state.voiceRef.current;
      const readyForOffer = !voice.makingOffer && (pc.signalingState === "stable" || voice.settingRemoteAnswerPending);
      const offerCollision = description.type === "offer" && !readyForOffer;
      voice.ignoreOffer = !voice.polite && offerCollision;
      if (voice.ignoreOffer) return;
      state.setCallConnectionStatus("connecting"); voice.settingRemoteAnswerPending = description.type === "answer";
      await pc.setRemoteDescription(description); voice.settingRemoteAnswerPending = false; await drainOperatorIce(pc);
      if (description.type === "offer") {
        const stream = await ensureOperatorStream(signalCall);
        startAudioRelay(signalCall, stream); await addOperatorTracks(pc, stream);
        const answer = await pc.createAnswer(); await pc.setLocalDescription(answer);
        sendOperatorSignal(signalCall, "webrtc.answer", { description: pc.localDescription });
      }
    } catch {
      state.voiceRef.current.settingRemoteAnswerPending = false; state.setCallConnectionStatus("failed"); state.setMediaError("WebRTC-разговор не подключился. Проверьте доступ к микрофону и сеть.");
    }
  }, [addOperatorTracks, drainOperatorIce, ensureOperatorPeer, ensureOperatorStream, sendOperatorSignal, startAudioRelay, state]);

  const callVisitor = useCallback(async (session: SiteVisitorSession) => {
    if (!state.token) return;
    state.setMediaError("");
    try {
      stopOperatorPeer(true);
      const stream = await ensureOperatorStream({ media_kind: "audio" });
      state.localStreamRef.current = stream; state.setLocalStream(stream);
      const call = await offerVisitorCall(session, state.token);
      state.setActiveCall(call); state.setCallConnectionStatus("connecting");
    } catch (error) { state.setMediaError(callActionErrorMessage(error)); state.setCallConnectionStatus("failed"); }
  }, [ensureOperatorStream, state, stopOperatorPeer]);

  const acceptCall = useCallback(async (call: SiteChatCallSession) => {
    if (!state.token) return;
    state.setMediaError("");
    try {
      stopOperatorRing(); stopOperatorPeer(true);
      const stream = await ensureOperatorStream(call);
      state.localStreamRef.current = stream; state.setLocalStream(stream);
      const next = await acceptSiteChatCall(call.call_session_id, state.token);
      state.setActiveCall(next); state.setIncomingCalls((prev) => prev.filter((item) => item.call_session_id !== call.call_session_id)); state.setCallConnectionStatus("connecting");
      await startOperatorOffer(next);
    } catch (error) { state.setMediaError(callActionErrorMessage(error)); state.setCallConnectionStatus("failed"); }
  }, [ensureOperatorStream, startOperatorOffer, state, stopOperatorPeer, stopOperatorRing]);

  const rejectCall = useCallback(async (call: SiteChatCallSession) => {
    if (!state.token) return;
    stopOperatorRing();
    const next = await rejectSiteChatCall(call.call_session_id, state.token);
    state.setActiveCall(next); state.setIncomingCalls((prev) => prev.filter((item) => item.call_session_id !== call.call_session_id));
  }, [state, stopOperatorRing]);

  const endCall = useCallback(async () => {
    if (!state.token || !state.activeCall) return;
    stopOperatorPeer(true);
    state.setActiveCall(await endSiteChatCall(state.activeCall.call_session_id, state.token));
  }, [state, stopOperatorPeer]);

  const toggleMute = useCallback(() => {
    state.localStreamRef.current?.getAudioTracks().forEach((track) => { track.enabled = state.muted; });
    state.setMuted(!state.muted);
  }, [state]);

  return { acceptCall, callVisitor, endCall, handleWebRTCSignal, playAudioRelayChunk, playOperatorRing, playRemoteAudio, rejectCall, startOperatorOffer, stopOperatorPeer, stopOperatorRing, toggleMute };
}
