import type { Dispatch, SetStateAction } from "react";
import { useCallback } from "react";

import type { AIChatSession } from "../../../lib/api";
import { approveAIToolCall, patchAIToolCall, rejectAIToolCall } from "../api/aiEmployeeApi";
import type { ProposalDrafts } from "../aiEmployeeTypes";
import { notifyEntityChanged } from "../aiEmployeeUtils";
import { useAIEmployeeRequestScope } from "./useAIEmployeeRequestScope";

export function useAIEmployeeToolCalls({
  activeProjectId,
  onEntityChanged,
  persistSessionID,
  setApproving,
  setError,
  setNotice,
  setProposalDrafts,
  setSession,
  token,
}: {
  activeProjectId?: string | null;
  onEntityChanged: () => void;
  persistSessionID: (updated: AIChatSession) => void;
  setApproving: (value: string) => void;
  setError: (value: string) => void;
  setNotice: (value: string) => void;
  setProposalDrafts: Dispatch<SetStateAction<ProposalDrafts>>;
  setSession: Dispatch<SetStateAction<AIChatSession | null>>;
  token?: string | null;
}) {
  const captureScope = useAIEmployeeRequestScope(activeProjectId, token);
  const approveToolCall = useCallback(async (toolCallId: string) => {
    if (!token || !activeProjectId) return;
    const isCurrent = captureScope();
    setApproving(toolCallId);
    setError("");
    setNotice("");
    try {
      const updated = await approveAIToolCall(activeProjectId, toolCallId, token);
      if (!isCurrent()) return;
      setSession(updated);
      persistSessionID(updated);
      const completed = (updated.tool_calls ?? []).find((call) => call.ai_tool_call_id === toolCallId);
      if (completed?.status === "failed") {
        setError(completed.error_text || "Действие не выполнено");
        return;
      }
      notifyEntityChanged(completed);
      onEntityChanged();
      setNotice(completed?.status === "succeeded" ? "Действие выполнено." : "Действие отправлено.");
    } catch (err) {
      if (!isCurrent()) return;
      setError(err instanceof Error ? err.message : "Не удалось выполнить действие");
    } finally {
      if (isCurrent()) setApproving("");
    }
  }, [activeProjectId, captureScope, onEntityChanged, persistSessionID, setApproving, setError, setNotice, setSession, token]);

  const rejectToolCall = useCallback(async (toolCallId: string) => {
    if (!token || !activeProjectId) return;
    const isCurrent = captureScope();
    setApproving(toolCallId);
    setError("");
    try {
      const updated = await rejectAIToolCall(activeProjectId, toolCallId, token);
      if (!isCurrent()) return;
      setSession(updated);
      persistSessionID(updated);
    } catch (err) {
      if (!isCurrent()) return;
      setError(err instanceof Error ? err.message : "Не удалось отклонить действие");
    } finally {
      if (isCurrent()) setApproving("");
    }
  }, [activeProjectId, captureScope, persistSessionID, setApproving, setError, setSession, token]);

  const updateToolCall = useCallback(async (toolCallId: string, inputJSON: Record<string, unknown>, reason = "user edited proposal card") => {
    if (!token || !activeProjectId) return false;
    const isCurrent = captureScope();
    setApproving(toolCallId);
    setError("");
    try {
      const updated = await patchAIToolCall(activeProjectId, toolCallId, token, inputJSON, reason);
      if (!isCurrent()) return false;
      setSession(updated);
      persistSessionID(updated);
      setProposalDrafts((current) => {
        const next = { ...current };
        delete next[toolCallId];
        return next;
      });
      setNotice("Черновик обновлен.");
      return true;
    } catch (err) {
      if (!isCurrent()) return false;
      setError(err instanceof Error ? err.message : "Не удалось обновить черновик");
      return false;
    } finally {
      if (isCurrent()) setApproving("");
    }
  }, [activeProjectId, captureScope, persistSessionID, setApproving, setError, setNotice, setProposalDrafts, setSession, token]);

  return { approveToolCall, rejectToolCall, updateToolCall };
}
