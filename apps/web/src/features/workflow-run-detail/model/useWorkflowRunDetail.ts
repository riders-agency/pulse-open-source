"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import type { ActionRun, Run, RunEvent } from "../../../lib/api";
import { usePulseStore } from "../../../store/pulse-store";
import { normalizeActionOutput, normalizeDiagnostics, pickLatestAttempt } from "../../run-integration-panels/parseOutputs";
import { getWorkflowRunDetail, replayWorkflowRun, retryWorkflowRunAction, type WorkflowRunGraph } from "../api/workflowRunApi";
import { buildRuntimeEdges, buildRuntimeNodes } from "../ui/runtimeGraph";

export function useWorkflowRunDetail(workflowRunId: string) {
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const [run, setRun] = useState<Run | null>(null);
  const [actions, setActions] = useState<ActionRun[]>([]);
  const [events, setEvents] = useState<RunEvent[]>([]);
  const [graph, setGraph] = useState<WorkflowRunGraph | null>(null);
  const [selectedKey, setSelectedKey] = useState("");
  const [error, setError] = useState("");
  const [replayAllBusy, setReplayAllBusy] = useState(false);
  const [retryActionBusy, setRetryActionBusy] = useState(false);

  const load = useCallback(async () => {
    if (!token) return;
    const detail = await getWorkflowRunDetail(workflowRunId, token);
    setRun(detail.run);
    if (detail.run.company_id) setActiveCompany(detail.run.company_id);
    if (detail.run.project_id) setActiveProject(detail.run.project_id);
    setActions(detail.actions);
    setEvents(detail.events);
    setGraph(detail.graph);
  }, [setActiveCompany, setActiveProject, token, workflowRunId]);

  useEffect(() => {
    load().catch((err) => setError(err.message));
    const id = window.setInterval(() => load().catch(() => undefined), 2500);
    return () => window.clearInterval(id);
  }, [load]);

  useEffect(() => {
    if (actions.length === 0) return;
    if (selectedKey && actions.some((action) => action.action_key === selectedKey)) return;
    const preferred =
      actions.find((action) => action.status === "failed") ??
      actions.find((action) => action.status === "retrying" || action.status === "waiting") ??
      actions.find((action) => action.status === "succeeded") ??
      actions[0];
    setSelectedKey(preferred?.action_key ?? "");
  }, [actions, selectedKey]);

  const nodes = useMemo(() => buildRuntimeNodes(actions, graph), [actions, graph]);
  const edges = useMemo(() => buildRuntimeEdges(nodes, graph), [graph, nodes]);
  const selected = actions.find((action) => action.action_key === selectedKey) ?? actions[0];
  const selectedNorm = useMemo(() => normalizeActionOutput(selected?.output_json), [selected?.output_json]);
  const latestAttempt = useMemo(() => (selected ? pickLatestAttempt(selected) : undefined), [selected]);
  const latestAttemptDiagnostics = useMemo(
    () => normalizeDiagnostics(latestAttempt?.diagnostics) ?? normalizeActionOutput(latestAttempt?.output_json).diagnostics,
    [latestAttempt],
  );
  const hasExpose = !!selectedNorm.expose && typeof selectedNorm.expose === "object" && Object.keys(selectedNorm.expose).length > 0;

  const replayFull = async () => {
    if (!token) return;
    setReplayAllBusy(true);
    try {
      const out = await replayWorkflowRun(workflowRunId, token);
      router.push(`/runs/${out.workflow_run_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось повторить запуск");
    } finally {
      setReplayAllBusy(false);
    }
  };

  const replayOneAction = async () => {
    if (!token || !selected?.action_key) return;
    setRetryActionBusy(true);
    try {
      const out = await retryWorkflowRunAction(workflowRunId, token, selected.action_key);
      router.push(`/runs/${out.workflow_run_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось выполнить повтор действия");
    } finally {
      setRetryActionBusy(false);
    }
  };

  return {
    edges,
    error,
    events,
    hasExpose,
    latestAttempt,
    latestAttemptDiagnostics,
    nodes,
    replayAllBusy,
    replayFull,
    replayOneAction,
    retryActionBusy,
    run,
    selected,
    setSelectedKey,
  };
}
