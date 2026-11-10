import type { SiteChatState } from "../model/useSiteChatState";
import { FullscreenCobrowse } from "./fullscreen-cobrowse";
import type { SiteChatConsoleActions, SiteChatConsoleVoice } from "./site-chat-console-view-types";

export function SiteChatFullscreenLayer({
  actions,
  state,
  voice,
}: {
  actions: SiteChatConsoleActions;
  state: SiteChatState;
  voice: SiteChatConsoleVoice;
}) {
  if (!state.fullscreen || !state.selectedVisitor) return null;
  return (
    <FullscreenCobrowse
      session={state.selectedVisitor}
      snapshot={state.selectedSnapshot}
      frames={state.frames}
      messages={state.messages}
      incomingCalls={state.incomingCalls}
      call={state.activeCall}
      recordings={state.callRecordings}
      draft={state.draft}
      setDraft={state.setDraft}
      pendingAttachments={state.pendingAttachments}
      onFiles={actions.uploadAttachmentFiles}
      onRemoveAttachment={actions.removePendingAttachment}
      onSend={() => void actions.sendMessage()}
      chatDisabled={state.sending || !state.selectedConversationId || state.pendingAttachments.some((item) => item.uploading)}
      muted={state.muted}
      callConnectionStatus={state.callConnectionStatus}
      remoteAudioReady={state.remoteAudioReady}
      remoteAudioRef={state.remoteAudioRef}
      onPlayRemoteAudio={voice.playRemoteAudio}
      onAcceptCall={(call) => void voice.acceptCall(call)}
      onRejectCall={(call) => void voice.rejectCall(call)}
      onMute={voice.toggleMute}
      onClose={() => state.setFullscreen(false)}
      onEndCall={() => void voice.endCall()}
    />
  );
}
