"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";

import { usePulseStore } from "../../store/pulse-store";
import { AI_EMPLOYEE_TEST_IDS } from "./aiEmployeeConstants";
import {
  buildPageContext,
  mergeOptimisticMessages,
  quickActionsFor,
  screenScope,
} from "./aiEmployeeUtils";
import { AIEmployeeComposer } from "./components/AIEmployeeComposer";
import { AIEmployeeFooter } from "./components/AIEmployeeFooter";
import { AIEmployeeHeader } from "./components/AIEmployeeHeader";
import { AIEmployeeLauncher } from "./components/AIEmployeeLauncher";
import { AIMessageList } from "./components/AIMessageList";
import { AIProgressLane } from "./components/AIProgressLane";
import { useAIEmployeeFocus } from "./hooks/useAIEmployeeFocus";
import { useAIEmployeeScroll } from "./hooks/useAIEmployeeScroll";
import { useAIEmployeeSession } from "./hooks/useAIEmployeeSession";
import { useAIEmployeeStream } from "./hooks/useAIEmployeeStream";

export function ProjectAIEmployeeWidget() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const [open, setOpen] = useState(false);
  const [fullscreen, setFullscreen] = useState(false);
  const chatLogRef = useRef<HTMLDivElement | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const pageContext = useMemo(
    () => buildPageContext(pathname, activeProjectId, searchParams?.toString() ?? ""),
    [activeProjectId, pathname, searchParams],
  );
  const disabled = !activeProjectId;
  const scope = useMemo(() => screenScope(pageContext), [pageContext]);
  const quickActions = useMemo(() => quickActionsFor(pageContext), [pageContext]);
  const {
    approving,
    calls,
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
  } = useAIEmployeeSession({
    activeProjectId,
    token,
    open,
    pageContext,
    scope,
    onEntityChanged: () => router.refresh(),
  });
  const scrollToBottom = useAIEmployeeScroll(chatLogRef, open);

  const {
    input,
    optimisticMessages,
    sendMessage,
    sending,
    setInput,
    stopCurrentRequest,
    streamingStatuses,
    streamingText,
  } = useAIEmployeeStream({
    activeProjectId,
    token,
    pageContext,
    pathname,
    fullscreen,
    session,
    pendingCall,
    ensureSession,
    persistSessionID,
    setError,
    setNotice,
    setSession,
    scrollToBottom,
  });
  const messages = useMemo(() => mergeOptimisticMessages(session, optimisticMessages), [optimisticMessages, session]);
  const capability = employee?.capability;
  const aiDisabled = capability?.state === "disabled";
  const capabilityNotice = capability && capability.state !== "ready" ? (capability.state === "disabled" ? "ИИ-сотрудник отключён. Проверьте настройки интеграции." : "Сейчас доступен ограниченный режим. Некоторые действия могут быть недоступны.") : "";

  const runQuickAction = useCallback((label: string) => {
    if (/оцени модели/i.test(label)) { void evaluateModels(); return; }
    void sendMessage(label);
  }, [evaluateModels, sendMessage]);

  useLayoutEffect(() => {
    if (!open) return;
    scrollToBottom("auto");
  }, [approving, calls.length, error, fullscreen, messages.length, notice, open, proposalDrafts, scrollToBottom, sending, streamingStatuses.length, streamingText]);

  useAIEmployeeFocus(open, () => setOpen(false));

  if (!token) return null;

  return (
    <div
      className={`ai-employee ${open ? "is-open" : ""} ${fullscreen ? "is-fullscreen" : ""}`}
      data-feature="ai-employee"
      data-testid={AI_EMPLOYEE_TEST_IDS.widget}
    >
      <AIEmployeeLauncher disabled={disabled} open={open} onToggle={() => setOpen((value) => !value)} />

      {open ? (
        <aside
          className="ai-employee-drawer"
          role="dialog"
          aria-modal="false"
          aria-label="ИИ сотрудник проекта"
          data-testid={AI_EMPLOYEE_TEST_IDS.drawer}
        >
          <AIEmployeeHeader
            fullscreen={fullscreen}
            onToggleFullscreen={() => setFullscreen((value) => !value)}
            onClose={() => setOpen(false)}
          />

          {employee?.setup_required ? (
            <div className="ai-setup-warning">
              {capability?.allows_proposals
                ? "Провайдер модели не подключён. Доступны сценарные команды; свободный диалог требует настройки интеграции."
                : "Подключите провайдера ИИ в интеграциях проекта, чтобы начать разговор."}
            </div>
          ) : null}
          {capabilityNotice ? <div className="ai-setup-warning">{capabilityNotice}</div> : null}

          <AIMessageList
            messages={messages}
            calls={calls}
            latestProposalIndexes={latestProposalIndexes}
            proposalDrafts={proposalDrafts}
            approving={approving}
            chatLogRef={chatLogRef}
            messagesEndRef={messagesEndRef}
            setProposalDrafts={setProposalDrafts}
            updateToolCall={updateToolCall}
            approveToolCall={approveToolCall}
            rejectToolCall={rejectToolCall}
          />

          <AIProgressLane sending={sending} statuses={streamingStatuses} streamingText={streamingText} />

          {notice ? <div className="ai-notice">{notice}</div> : null}
          {error ? <div className="ai-error" data-testid="ai-error-banner">{error}</div> : null}

          <AIEmployeeFooter disabled={aiDisabled} quickActions={quickActions} sending={sending} onQuickAction={runQuickAction} />
          <AIEmployeeComposer
            disabled={aiDisabled}
            input={input}
            sending={sending}
            onInputChange={setInput}
            onSend={(value) => void sendMessage(value)}
            onStop={stopCurrentRequest}
          />
        </aside>
      ) : null}
    </div>
  );
}
