import type { AIChatMessage, AIToolCall } from "../../../lib/api";
import type { ProposalDrafts } from "../aiEmployeeTypes";
import { AIProposalSnippet } from "./AIProposalSnippet";
import { AITaskSearchResultSnippet } from "./AITaskSearchResultSnippet";
import { AIToolResultSnippet } from "./AIToolResultSnippet";
import type { ProposalSnippetProps } from "./snippetTypes";

type AISnippetCatalogProps = Omit<ProposalSnippetProps, "message" | "call"> & {
  message: AIChatMessage;
  call?: AIToolCall;
  isLatestProposal: boolean;
  proposalDrafts: ProposalDrafts;
};

export function AISnippetCatalog({
  message,
  call,
  isLatestProposal,
  proposalDrafts,
  approving,
  setProposalDrafts,
  updateToolCall,
  approveToolCall,
  rejectToolCall,
}: AISnippetCatalogProps) {
  return (
    <>
      <AIToolResultSnippet payload={message.payload_json} />
      <AITaskSearchResultSnippet payload={message.payload_json} />
      {message.payload_json?.kind === "proposal" && isLatestProposal ? (
        <AIProposalSnippet
          message={message}
          call={call}
          proposalDrafts={proposalDrafts}
          approving={approving}
          setProposalDrafts={setProposalDrafts}
          updateToolCall={updateToolCall}
          approveToolCall={approveToolCall}
          rejectToolCall={rejectToolCall}
        />
      ) : null}
    </>
  );
}
