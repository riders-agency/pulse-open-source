import { Sparkles } from "lucide-react";
import type { Dispatch, RefObject, SetStateAction } from "react";

import type { AIToolCall } from "../../../lib/api";
import { AI_EMPLOYEE_COPY_LABELS, AI_EMPLOYEE_TEST_IDS } from "../aiEmployeeConstants";
import type { ProposalDrafts, WidgetChatMessage } from "../aiEmployeeTypes";
import {
  messageKey,
  messageTimestampLabel,
  pendingCallForMessage,
  textValue,
} from "../aiEmployeeUtils";
import { AISnippetCatalog } from "../snippets/AISnippetCatalog";
import { ExecutionTrace } from "../graph/ExecutionTrace";

export function AIMessageList({
  messages,
  calls,
  latestProposalIndexes,
  proposalDrafts,
  approving,
  chatLogRef,
  messagesEndRef,
  setProposalDrafts,
  updateToolCall,
  approveToolCall,
  rejectToolCall,
}: {
  messages: WidgetChatMessage[];
  calls: AIToolCall[];
  latestProposalIndexes: Record<string, number>;
  proposalDrafts: ProposalDrafts;
  approving: string;
  chatLogRef: RefObject<HTMLDivElement | null>;
  messagesEndRef: RefObject<HTMLDivElement | null>;
  setProposalDrafts: Dispatch<SetStateAction<ProposalDrafts>>;
  updateToolCall: (toolCallId: string, inputJSON: Record<string, unknown>, reason?: string) => Promise<boolean>;
  approveToolCall: (toolCallId: string) => Promise<void>;
  rejectToolCall: (toolCallId: string) => Promise<void>;
}) {
  return (
    <div className="ai-chat-log" data-testid={AI_EMPLOYEE_TEST_IDS.chatLog} ref={chatLogRef}>
      {messages.length === 0 ? (
        <div className="ai-empty">
          <Sparkles size={24} />
          <h3>С чего начнём?</h3>
          <p>Помогу разобраться в задачах, клиентах и оплатах. Опишите, что нужно сделать — перед изменениями я попрошу подтверждение.</p>
        </div>
      ) : (
        messages.map((message, index) => {
          const call = pendingCallForMessage(message, calls);
          const proposalCallId = textValue(message.payload_json?.ai_tool_call_id);
          const isLatestProposal = !proposalCallId || latestProposalIndexes[proposalCallId] === index;
          return (
            <article
              className={`ai-message ai-message-${message.role} ${message.optimistic ? "is-optimistic" : ""}`}
              key={messageKey(message, index)}
              data-testid={
                message.role === "user"
                  ? AI_EMPLOYEE_TEST_IDS.userMessage
                  : message.role === "assistant"
                    ? AI_EMPLOYEE_TEST_IDS.assistantMessage
                    : AI_EMPLOYEE_TEST_IDS.genericMessage
              }
            >
              <p>
                <span className="ai-message-author">
                  {message.role === "user"
                    ? AI_EMPLOYEE_COPY_LABELS.user
                    : message.role === "assistant"
                      ? AI_EMPLOYEE_COPY_LABELS.assistant
                      : "Система"}
                  {messageTimestampLabel(message.created_at) ? ` [${messageTimestampLabel(message.created_at)}]` : ""}:{" "}
                </span>
                {message.body}
              </p>
              <AISnippetCatalog
                message={message}
                call={call}
                isLatestProposal={isLatestProposal}
                proposalDrafts={proposalDrafts}
                approving={approving}
                setProposalDrafts={setProposalDrafts}
                updateToolCall={updateToolCall}
                approveToolCall={approveToolCall}
                rejectToolCall={rejectToolCall}
              />
              <ExecutionTrace value={message.payload_json?.execution} />
            </article>
          );
        })
      )}
      <div data-testid={AI_EMPLOYEE_TEST_IDS.chatEnd} ref={messagesEndRef} />
    </div>
  );
}
