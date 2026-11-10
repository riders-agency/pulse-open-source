import { AlertTriangle, Check, CheckCircle2, ChevronDown, Loader2, XCircle } from "lucide-react";

import {
  jsonPreview,
  primaryActionLabel,
  proposalDisplay,
  proposalEntitySnippets,
  proposalInput,
  proposalMissing,
  statusLabel,
  stringList,
  textValue,
  toolLabel,
} from "../aiEmployeeUtils";
import { AIEntitySnippets } from "./AIEntitySnippets";
import type { ProposalSnippetProps } from "./snippetTypes";
import "./proposal-history.css";
import { TaskCreateFields } from "./TaskCreateFields";
import { TaskMoveFields } from "./TaskMoveFields";
import { ToolProposalFields } from "./ToolProposalFields";

export function AIProposalSnippet({
  message,
  call,
  proposalDrafts,
  approving,
  setProposalDrafts,
  updateToolCall,
  approveToolCall,
  rejectToolCall,
}: ProposalSnippetProps) {
  const inputJSON = proposalInput(call, message);
  const display = proposalDisplay(call, message);
  const missing = proposalMissing(message);
  const snippets = proposalEntitySnippets(call, message);
  const callId = call?.ai_tool_call_id || "";
  const draft = proposalDrafts[callId] || inputJSON;
  const editing = call?.status === "pending_approval";
  const updateDraft = (patch: Record<string, unknown>) => setProposalDrafts((current) => ({ ...current, [callId]: { ...draft, ...patch } }));
  const toolKey = call?.tool_key || String(message.payload_json?.tool_key || "");

  if (toolKey === "task.create" && call?.status === "succeeded") {
    return <details className="ai-completed-proposal" data-testid="ai-proposal-card-task-create">
      <summary><CheckCircle2 size={16} /><span>Подтверждено: {textValue(display.title) || textValue(inputJSON.title)}</span><ChevronDown size={15} /></summary>
      <TaskCreateFields draft={draft} display={display} inputJSON={inputJSON} editing={false} updateDraft={updateDraft} />
    </details>;
  }

  return (
    <div
      className={`ai-proposal-card ai-risk-${call?.risk_level || message.payload_json?.risk_level || "read"}`}
      data-testid={toolKey === "task.create" ? "ai-proposal-card-task-create" : "ai-proposal-card"}
    >
      <div className="ai-proposal-head">
        <div>
          <strong>{toolLabel(toolKey)}</strong>
          <p>{call?.status === "pending_approval" ? "Проверьте поля перед применением" : "Состояние действия"}</p>
        </div>
        <span>{statusLabel(call?.status)}</span>
      </div>
      {missing.length ? (
        <div className="ai-missing-fields" data-testid="ai-proposal-missing-fields">
          <AlertTriangle size={15} />
          <span>{String(missing[0]?.message || "Нужно уточнить обязательные поля")}</span>
        </div>
      ) : null}
      <AIEntitySnippets snippets={snippets} />
      {toolKey === "task.move" ? (
        <TaskMoveFields display={display} inputJSON={inputJSON} />
      ) : toolKey === "task.create" ? (
        <TaskCreateFields draft={draft} display={display} inputJSON={inputJSON} editing={editing} updateDraft={updateDraft} />
      ) : <ToolProposalFields input={draft} />}
      {textValue(display.description) || textValue(inputJSON.description) ? (
        <div className="ai-proposal-note">{textValue(display.description) || textValue(inputJSON.description)}</div>
      ) : null}
      {stringList(display.checklist || inputJSON.checklist).length ? (
        <ul className="ai-proposal-checklist">
          {stringList(display.checklist || inputJSON.checklist).map((item) => <li key={item}><CheckCircle2 size={14} /> {item}</li>)}
        </ul>
      ) : null}
      <details className="ai-dev-details">
        <summary>Технические детали</summary>
        <pre>{jsonPreview(draft)}</pre>
      </details>
      {call?.status === "pending_approval" ? (
        <div className="ai-proposal-actions" data-testid="ai-approval-card">
          <button type="button" className="ai-approve" data-testid="ai-approval-apply" onClick={() => void (async () => {
            const saved = await updateToolCall(call.ai_tool_call_id, draft, "user edited proposal before apply");
            if (saved) await approveToolCall(call.ai_tool_call_id);
          })()} disabled={approving === call.ai_tool_call_id}>
            {approving === call.ai_tool_call_id ? <Loader2 size={14} className="spin" /> : <Check size={14} />}
            {primaryActionLabel(call.tool_key)}
          </button>
          <button type="button" className="ai-reject" data-testid="ai-approval-reject" onClick={() => void rejectToolCall(call.ai_tool_call_id)} disabled={approving === call.ai_tool_call_id}>
            <XCircle size={14} />
            Отклонить
          </button>
        </div>
      ) : null}
    </div>
  );
}
