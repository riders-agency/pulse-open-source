"use client";
import { useCallback, useEffect, useMemo, useState } from "react";

import type { WorkAccess, WorkTaskDetail, WorkTimer } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import {
  addWorkTaskChecklistItem,
  addWorkTaskComment,
  getActiveWorkTimers,
  getWorkTask,
  patchWorkTask,
  setWorkTaskChecklistDone,
  setWorkTaskCustomField,
  startWorkTaskTimer,
  stopWorkTimer,
} from "./api/workApi";
import { useWorkTaskAttachments } from "./use-work-task-attachments";
import type { TaskTab } from "./work-task-types";

export function useWorkTaskDetail({ taskId, access, onChanged }: { taskId: string; access?: WorkAccess; onChanged?: () => void }) {
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const [task, setTask] = useState<WorkTaskDetail>();
  const [loadedAccess, setLoadedAccess] = useState<WorkAccess>();
  const [activeTimers, setActiveTimers] = useState<WorkTimer[]>([]);
  const [now, setNow] = useState(Date.now());
  const [tab, setTab] = useState<TaskTab>("card");
  const [comment, setComment] = useState("");
  const [checklistTitle, setChecklistTitle] = useState("");
  const [timerNotice, setTimerNotice] = useState("");
  const [error, setError] = useState("");

  const effectiveAccess = loadedAccess || access;
  const canWrite = Boolean(effectiveAccess?.can_write_tasks && !effectiveAccess.is_client_guest);
  const canUseTime = Boolean(canWrite || effectiveAccess?.can_approve_time);
  const taskTimers = useMemo(() => (task ? activeTimers.filter((timer) => timer.task_id === task.task_id) : []), [activeTimers, task]);
  const activeTaskTimer = taskTimers[0];

  const load = useCallback(async () => {
    if (!token || !taskId) return;
    const payload = await getWorkTask(taskId, token);
    setTask(payload.task);
    setLoadedAccess(payload.access);
    setError("");
  }, [taskId, token]);

  const loadActiveTimers = useCallback(async () => {
    if (!token || !canUseTime) {
      setActiveTimers([]);
      return;
    }
    const payload = await getActiveWorkTimers(token);
    setActiveTimers(payload.timers || (payload.timer ? [payload.timer] : []));
  }, [canUseTime, token]);

  useEffect(() => {
    setTask(undefined);
    setLoadedAccess(undefined);
    if (!taskId) return;
    void load().catch((err) => setError(err instanceof Error ? err.message : "Не удалось открыть задачу"));
  }, [load, taskId, token]);

  useEffect(() => {
    void loadActiveTimers().catch(() => setActiveTimers([]));
  }, [loadActiveTimers, taskId, token]);

  useEffect(() => {
    const interval = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(interval);
  }, []);

  const patch = async (body: Record<string, unknown>) => {
    if (!token || !task) return;
    const updated = await patchWorkTask(task.task_id, token, body);
    setTask({ ...task, ...updated });
    onChanged?.();
  };

  const addComment = async () => {
    if (!token || !task || !comment.trim()) return;
    const payload = await addWorkTaskComment(task.task_id, token, comment.trim(), effectiveAccess?.is_client_guest ? "client" : "internal");
    setTask(payload.task);
    setComment("");
    onChanged?.();
  };

  const addChecklist = async () => {
    if (!token || !task || !checklistTitle.trim()) return;
    const payload = await addWorkTaskChecklistItem(task.task_id, token, checklistTitle.trim());
    setTask(payload.task);
    setChecklistTitle("");
    onChanged?.();
  };

  const toggleChecklist = async (itemId: string, next: boolean) => {
    if (!token || !task) return;
    const payload = await setWorkTaskChecklistDone(task.task_id, itemId, token, next);
    setTask(payload.task);
    onChanged?.();
  };

  const updateCustomField = async (fieldId: string, value: unknown) => {
    if (!token || !task) return;
    const payload = await setWorkTaskCustomField(task.task_id, fieldId, token, value);
    setTask(payload.task);
    onChanged?.();
  };

  const startTimer = async () => {
    if (!token || !task) return;
    setTimerNotice("");
    await startWorkTaskTimer(task.task_id, token);
    setTimerNotice("Таймер запущен. Время считается прямо в карточке; другие активные задачи не останавливаются.");
    await loadActiveTimers();
    await load();
  };

  const stopTimer = async () => {
    if (!token || !task) return;
    setTimerNotice("");
    await stopWorkTimer(token, task.task_id);
    setTimerNotice("Таймер остановлен. Запись времени добавлена в задачу.");
    await loadActiveTimers();
    await load();
  };

  const { downloadAttachment, uploadAttachment, uploading } = useWorkTaskAttachments({ onChanged, setError, setTab, setTask, task, token });

  const progress = useMemo(() => {
    if (!task?.checklist_total_count) return "0/0";
    return `${task.checklist_done_count}/${task.checklist_total_count}`;
  }, [task]);
  const commentLabel = effectiveAccess?.is_client_guest ? "Комментарий клиента" : "Внутренний комментарий";

  return {
    activeCompanyId,
    activeTaskTimer,
    addChecklist,
    addComment,
    canUseTime,
    canWrite,
    checklistTitle,
    comment,
    commentLabel,
    downloadAttachment,
    effectiveAccess,
    error,
    now,
    patch,
    progress,
    setChecklistTitle,
    setComment,
    setError,
    setTab,
    setTask,
    startTimer,
    stopTimer,
    tab,
    task,
    taskTimers,
    timerNotice,
    toggleChecklist,
    updateCustomField,
    uploadAttachment,
    uploading,
  };
}
