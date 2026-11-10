"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import type { Run, Workflow as WorkflowRow } from "../../../lib/api";
import { getDashboardWorkflows } from "../api/dashboardApi";
import { messageFromError } from "./errors";

export function useDashboardWorkflowRuns(token: string | undefined, activeProjectId?: string | null) {
  const [workflows, setWorkflows] = useState<WorkflowRow[]>([]);
  const [runs, setRuns] = useState<Run[]>([]);
  const [error, setError] = useState("");
  const activeWorkflowCount = useMemo(
    () => workflows.filter((workflow) => workflow.status === "active").length,
    [workflows],
  );
  const latestErrors = useMemo(
    () => runs.filter((run) => run.status === "failed" || run.error).slice(0, 5),
    [runs],
  );

  const loadWorkflows = useCallback(async () => {
    if (!token || !activeProjectId) return;
    const { workflows: safe, runs: runRows } = await getDashboardWorkflows(activeProjectId, token);
    setWorkflows(safe);
    setRuns(runRows);
  }, [activeProjectId, token]);

  useEffect(() => {
    void loadWorkflows().catch((err) => setError(messageFromError(err, "Не удалось загрузить сценарии")));
  }, [loadWorkflows]);

  return {
    activeWorkflowCount,
    error,
    latestErrors,
    runs,
  };
}
