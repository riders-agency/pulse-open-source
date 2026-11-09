"use client";

import { useSiteChatActions } from "./model/useSiteChatActions";
import { useSiteChatLifecycle } from "./model/useSiteChatLifecycle";
import { useSiteChatLiveSnapshot } from "./model/useSiteChatLiveSnapshot";
import { useSiteChatLoaders } from "./model/useSiteChatLoaders";
import { useSiteChatState } from "./model/useSiteChatState";
import { useSiteChatVoice } from "./model/useSiteChatVoice";
import { SiteChatConsoleView } from "./ui/SiteChatConsoleView";

export function SiteChatConsole() {
  const state = useSiteChatState();
  const loaders = useSiteChatLoaders(state);
  const live = useSiteChatLiveSnapshot(state);
  const voice = useSiteChatVoice(state, loaders.loadCallRecordings);
  const actions = useSiteChatActions(state, loaders);

  useSiteChatLifecycle(state, loaders, live, voice);

  return (
    <SiteChatConsoleView
      actions={actions}
      load={loaders.load}
      state={state}
      voice={voice}
    />
  );
}
