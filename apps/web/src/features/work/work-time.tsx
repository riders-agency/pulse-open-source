"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Play, Plus, Square, Timer } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Field, Input } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import type { WorkAccess, WorkTask, WorkTimeEntry, WorkTimer } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { createWorkTimeEntry, getActiveWorkTimers, getWorkBoard, listWorkBoards, listWorkTime, startWorkTaskTimer, stopWorkTimer } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { ActiveWorkTimers, WorkTimeEntriesTable } from "./work-time-panels";
import { minutesLabel } from "./work-utils";

export function WorkTimePage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [entries, setEntries] = useState<WorkTimeEntry[]>([]);
  const [timer, setTimer] = useState<WorkTimer | null>(null);
  const [timers, setTimers] = useState<WorkTimer[]>([]);
  const [tasks, setTasks] = useState<WorkTask[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [taskId, setTaskId] = useState("");
  const [minutes, setMinutes] = useState(60);
  const [description, setDescription] = useState("");
  const [now, setNow] = useState(Date.now());
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (!token || !projectId) return;
    const [timeRows, timerRow, boardList] = await Promise.all([
      listWorkTime(projectId, token),
      getActiveWorkTimers(token),
      listWorkBoards(projectId, token),
    ]);
    setAccess(timeRows.access);
    setEntries(timeRows.items || []);
    setTimer(timerRow.timer || null);
    setTimers(timerRow.timers || []);
    const board = boardList.items[0];
    if (board) {
      const detail = await getWorkBoard(board.task_board_id, token);
      const flat = (detail.board.columns || []).flatMap((column) => column.tasks);
      setTasks(flat);
      if (!taskId && flat[0]) setTaskId(flat[0].task_id);
    }
    setError("");
  }, [projectId, taskId, token]);

  useEffect(() => {
    void load().catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить время"));
  }, [load]);

  useEffect(() => {
    const interval = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(interval);
  }, []);

  const total = useMemo(() => entries.reduce((sum, row) => sum + row.duration_minutes, 0), [entries]);
  const activeTotalMs = useMemo(() => {
    return timers.reduce((sum, row) => {
      const startedMs = Date.parse(row.started_at);
      const elapsedMs = Number.isFinite(startedMs) ? Math.max(row.elapsed_minutes * 60_000, now - startedMs) : row.elapsed_minutes * 60_000;
      return sum + elapsedMs;
    }, 0);
  }, [now, timers]);
  if (!projectId) return <WorkNoProject />;

  const stop = async (stopTaskId?: string) => {
    if (!token) return;
    await stopWorkTimer(token, stopTaskId || "");
    await load();
  };

  const start = async () => {
    if (!token || !taskId) return;
    await startWorkTaskTimer(taskId, token);
    await load();
  };

  const add = async () => {
    if (!token || !taskId) return;
    await createWorkTimeEntry(token, {
      task_id: taskId,
      duration_minutes: minutes,
      description,
      entry_date: new Date().toISOString().slice(0, 10),
      billable: true,
    });
    setDescription("");
    setShowForm(false);
    await load();
  };

  return (
    <WorkFrame
      access={access}
      title="Таймтрекер"
      description="Запуск таймеров по задачам, параллельные активные таймеры и ручные записи времени."
      actions={<Button variant="secondary" onClick={() => setShowForm(!showForm)}><Plus size={16} /> Добавить время</Button>}
    >
      {error ? <div className="form-error">{error}</div> : null}
      <section className="work-timer-strip">
        <div>
          <Timer size={18} />
          <span>{timer ? `${timers.length} активных · ${timer.task_title}` : "Активных таймеров нет. Выберите задачу ниже и нажмите Старт."}</span>
        </div>
        <strong>{timers.length ? clockFromMs(activeTotalMs) : minutesLabel(total)}</strong>
        {timer ? <Button variant="destructive" size="sm" onClick={() => void stop().catch((err) => setError(err.message))}><Square size={15} /> Остановить последний</Button> : null}
      </section>
      <form
        className="work-manual-time"
        onSubmit={(event) => {
          event.preventDefault();
          void start().catch((err) => setError(err.message));
        }}
      >
        <Field label="Запустить таймер по задаче">
          <Select value={taskId} onChange={(event) => setTaskId(event.target.value)} disabled={tasks.length === 0}>
            {tasks.length === 0 ? <option value="">Нет задач на доске</option> : null}
            {tasks.map((task) => <option key={task.task_id} value={task.task_id}>{task.title}</option>)}
          </Select>
        </Field>
        <div className="form-actions">
          <Button type="submit" disabled={!taskId}><Play size={16} /> Старт</Button>
        </div>
      </form>
      <ActiveWorkTimers now={now} timers={timers} onStop={(id) => void stop(id).catch((err) => setError(err.message))} />
      {showForm ? (
        <form
          className="work-manual-time"
          onSubmit={(event) => {
            event.preventDefault();
            void add().catch((err) => setError(err.message));
          }}
        >
          <Field label="Задача">
            <Select value={taskId} onChange={(event) => setTaskId(event.target.value)}>
              {tasks.map((task) => <option key={task.task_id} value={task.task_id}>{task.title}</option>)}
            </Select>
          </Field>
          <Field label="Минуты">
            <Input type="number" min={1} value={minutes} onChange={(event) => setMinutes(Number(event.target.value))} />
          </Field>
          <Field label="Описание">
            <Input value={description} onChange={(event) => setDescription(event.target.value)} placeholder="Что было сделано" />
          </Field>
          <div className="form-actions">
            <Button type="submit">Добавить</Button>
          </div>
        </form>
      ) : null}
      <WorkTimeEntriesTable entries={entries} />
    </WorkFrame>
  );
}

function clockFromMs(value: number) {
  const totalSeconds = Math.max(0, Math.floor(value / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}
