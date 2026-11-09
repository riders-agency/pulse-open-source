import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type { AIChatSession, AIEmployeeEnvelope, AIPageContext } from "../../../lib/api";
import {
  createAIEmployeeSession,
  evaluateAIModels,
  getAIEmployee,
  getAIEmployeeSession,
} from "../api/aiEmployeeApi";
import type { ProposalDrafts } from "../aiEmployeeTypes";
import {
  activePendingCall,
  legacyProjectSessionStorageKey,
  modelFromEmployee,
  proposalMessageIndexes,
  projectSessionStorageKey,
} from "../aiEmployeeUtils";
import { useAIEmployeeToolCalls } from "./useAIEmployeeToolCalls";
import { useAIEmployeeRequestScope } from "./useAIEmployeeRequestScope";

type UseAIEmployeeSessionInput = {
  activeProjectId?: string | null;
  token?: string | null;
  open: boolean;
  pageContext: AIPageContext;
  scope: string;
  onEntityChanged: () => void;
};

export function useAIEmployeeSession({
  activeProjectId,
  token,
  open,
  pageContext,
  scope,
  onEntityChanged,
}: UseAIEmployeeSessionInput) {
  const captureScope = useAIEmployeeRequestScope(activeProjectId, token);
  const sessionRequest = useRef(0);
  const [approving, setApproving] = useState("");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [employee, setEmployee] = useState<AIEmployeeEnvelope | null>(null);
  const [session, setSession] = useState<AIChatSession | null>(null);
  const [proposalDrafts, setProposalDrafts] = useState<ProposalDrafts>({});

  const { dependency, model } = useMemo(() => modelFromEmployee(employee), [employee]);
  const calls = session?.tool_calls ?? [];
  const pendingCall = useMemo(() => activePendingCall(calls), [calls]);

  const latestProposalIndexes = useMemo(() => proposalMessageIndexes(session?.messages), [session?.messages]);

  const persistSessionID = useCallback((updated: AIChatSession) => {
    sessionRequest.current++;
    if (!activeProjectId) return;
    window.localStorage.setItem(projectSessionStorageKey(activeProjectId), updated.ai_chat_session_id);
  }, [activeProjectId]);
  const { approveToolCall, rejectToolCall, updateToolCall } = useAIEmployeeToolCalls({
    activeProjectId,
    onEntityChanged,
    persistSessionID,
    setApproving,
    setError,
    setNotice,
    setProposalDrafts,
    setSession,
    token,
  });

  const loadEmployee = useCallback(async () => {
    if (!token || !activeProjectId) return;
    const isCurrent = captureScope();
    setError("");
    try {
      const out = await getAIEmployee(activeProjectId, token);
      if (!isCurrent()) return;
      setEmployee(out);
    } catch (err) {
      if (!isCurrent()) return;
      setError(err instanceof Error ? err.message : "Не удалось открыть AI-сотрудника");
    }
  }, [activeProjectId, captureScope, token]);

  const restoreProjectSession = useCallback(async () => {
    if (!token || !activeProjectId || session?.project_id === activeProjectId) return;
    const isCurrent = captureScope();
    const storageKey = projectSessionStorageKey(activeProjectId);
    const request = sessionRequest.current;
    const legacyStorageKey = legacyProjectSessionStorageKey(activeProjectId);
    const sessionId = window.localStorage.getItem(storageKey) || window.localStorage.getItem(legacyStorageKey);
    if (!sessionId) return;
    try {
      const restored = await getAIEmployeeSession(activeProjectId, sessionId, token);
      if (!isCurrent() || request !== sessionRequest.current) return;
      setSession(restored);
      window.localStorage.setItem(storageKey, restored.ai_chat_session_id);
    } catch {
      if (!isCurrent() || request !== sessionRequest.current) return;
      window.localStorage.removeItem(storageKey);
      window.localStorage.removeItem(legacyStorageKey);
    }
  }, [activeProjectId, captureScope, session?.project_id, token]);

  const ensureSession = useCallback(async () => {
    if (!token || !activeProjectId) return null;
    if (session?.project_id === activeProjectId) return session;
    sessionRequest.current++;
    const isCurrent = captureScope();
    const created = await createAIEmployeeSession(activeProjectId, token, {
      screen_scope: scope,
      screen_entity_type: pageContext.pageKind,
      screen_entity_id: pageContext.workflowId || pageContext.workflowRunId || pageContext.taskId || pageContext.clientId || "",
      workflow_id: pageContext.workflowId,
      workflow_run_id: pageContext.workflowRunId,
      task_id: pageContext.taskId,
      client_id: pageContext.clientId,
      action_dependency_id: dependency?.action_dependency_id,
      model_id: model,
      page_context_json: pageContext,
    });
    if (!isCurrent()) return null;
    setSession(created);
    persistSessionID(created);
    return created;
  }, [activeProjectId, captureScope, dependency?.action_dependency_id, model, pageContext, persistSessionID, scope, session, token]);

  const evaluateModels = useCallback(async () => {
    if (!token || !activeProjectId || !dependency?.action_dependency_id) return;
    const isCurrent = captureScope();
    setError("");
    setNotice("");
    try {
      const out = await evaluateAIModels(activeProjectId, dependency.action_dependency_id, token);
      if (!isCurrent()) return;
      setNotice(`Модель выбрана: ${out.selected_model_id || "см. интеграцию"}`);
      await loadEmployee();
    } catch (err) {
      if (!isCurrent()) return;
      setError(err instanceof Error ? err.message : "Не удалось оценить модели");
    }
  }, [activeProjectId, captureScope, dependency?.action_dependency_id, loadEmployee, token]);

  useEffect(() => {
    if (!open || !activeProjectId || !token) return;
    void loadEmployee();
    void restoreProjectSession();
  }, [activeProjectId, loadEmployee, open, restoreProjectSession, token]);

  useEffect(() => {
    setSession(null);
    setEmployee(null); setError(""); setNotice(""); setApproving("");
    setProposalDrafts({});
  }, [activeProjectId, token]);

  return {
    approving,
    calls,
    dependency,
    employee,
    ensureSession,
    error,
    evaluateModels,
    latestProposalIndexes,
    notice,
    pendingCall,
    persistSessionID,
    proposalDrafts,
    rejectToolCall,
    approveToolCall,
    session,
    setError,
    setNotice,
    setProposalDrafts,
    setSession,
    updateToolCall,
  };
}
