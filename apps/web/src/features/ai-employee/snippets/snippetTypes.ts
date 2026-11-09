import type { Dispatch, SetStateAction } from "react";

import type { AIChatMessage, AIToolCall } from "../../../lib/api";
import type { ProposalDrafts } from "../aiEmployeeTypes";

export type ProposalSnippetProps = {
  message: AIChatMessage;
  call?: AIToolCall;
  proposalDrafts: ProposalDrafts;
  approving: string;
  setProposalDrafts: Dispatch<SetStateAction<ProposalDrafts>>;
  updateToolCall: (toolCallId: string, inputJSON: Record<string, unknown>, reason?: string) => Promise<boolean>;
  approveToolCall: (toolCallId: string) => Promise<void>;
  rejectToolCall: (toolCallId: string) => Promise<void>;
};
