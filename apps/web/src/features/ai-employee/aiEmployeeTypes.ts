import type { AIChatMessage, AIChatSession, AIMessageInput } from "../../lib/api";

export type AIStreamPayload = {
  type?: string;
  status?: string;
  text?: string;
  message?: string;
  code?: string;
  ok?: boolean;
  session?: AIChatSession;
};

export type WidgetChatMessage = AIChatMessage & {
  optimistic?: boolean;
};

export type ProposalDrafts = Record<string, Record<string, unknown>>;

export type SendAIMessage = (text: string) => void | Promise<void>;

export type PostMessageStream = (
  projectId: string,
  sessionId: string,
  token: string,
  input: AIMessageInput,
  onEvent: (payload: AIStreamPayload) => void,
  signal?: AbortSignal,
) => Promise<AIChatSession>;
