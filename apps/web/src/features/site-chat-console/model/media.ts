import type { SiteChatVoiceConfig } from "./types";

export function operatorAudioConstraints(config?: SiteChatVoiceConfig | null): MediaStreamConstraints {
  return {
    audio: {
      echoCancellation: config?.audio?.echo_cancellation ?? true,
      noiseSuppression: config?.audio?.noise_suppression ?? true,
      autoGainControl: config?.audio?.auto_gain_control ?? true,
      channelCount: config?.audio?.channel_count ?? 1,
    },
    video: false,
  };
}

export function rtcConfiguration(config?: SiteChatVoiceConfig | null): RTCConfiguration {
  return {
    iceServers: config?.ice_servers?.length ? config.ice_servers : [{ urls: "stun:stun.l.google.com:19302" }],
    iceTransportPolicy: config?.ice_transport_policy === "relay" ? "relay" : "all",
    bundlePolicy: "balanced",
    rtcpMuxPolicy: "require",
  };
}

export function recordingMimeType() {
  const types = ["audio/webm;codecs=opus", "audio/webm", "audio/ogg;codecs=opus"];
  if (typeof MediaRecorder === "undefined" || !MediaRecorder.isTypeSupported) return "";
  return types.find((type) => MediaRecorder.isTypeSupported(type)) ?? "";
}

export function audioRelayMimeType() {
  return recordingMimeType();
}

export function blobToBase64(blob: Blob) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => {
      const value = String(reader.result || "");
      resolve(value.includes(",") ? value.split(",").pop() ?? "" : value);
    };
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}

export function base64ToArrayBuffer(value: unknown) {
  const binary = atob(String(value || ""));
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
  return bytes.buffer;
}
