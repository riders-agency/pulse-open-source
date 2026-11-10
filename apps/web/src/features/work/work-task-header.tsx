"use client";

import Link from "next/link";
import { Play, Square, Timer, X } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import type { WorkAccess, WorkTaskDetail, WorkTimer } from "../../lib/api";
import { formatDate } from "../../lib/dates";
import { minutesLabel, priorityLabel, priorityTone, statusLabel, taskTypeLabel, timerClock } from "./work-utils";

export function WorkTaskHeader({
  activeTaskTimer,
  canUseTime,
  closeHref,
  effectiveAccess,
  now,
  onClose,
  setError,
  startTimer,
  stopTimer,
  task,
}: {
  activeTaskTimer?: WorkTimer;
  canUseTime: boolean;
  closeHref: string;
  effectiveAccess?: WorkAccess;
  now: number;
  onClose?: () => void;
  setError: (value: string) => void;
  startTimer: () => Promise<void>;
  stopTimer: () => Promise<void>;
  task: WorkTaskDetail;
}) {
  return (
    <header className="work-task-screen-header">
      <div className="work-task-screen-actions">
        {onClose ? (
          <button className="work-icon-button" type="button" onClick={onClose} aria-label="Закрыть задачу">
            <X size={18} />
          </button>
        ) : (
          <Link className="work-icon-button" href={closeHref} aria-label="К задачам">
            <X size={18} />
          </Link>
        )}
        <span className="work-drawer-eyebrow">Карточка задачи</span>
      </div>
      <div className="work-task-screen-title">
        <h1>{task.title}</h1>
        <p>{task.client_name || "Без клиента"} · {task.project_name || effectiveAccess?.project_name || "Проект не выбран"}</p>
      </div>
      <div className="work-task-hero-meta">
        <Badge tone="muted">{taskTypeLabel(task.type)}</Badge>
        <Badge tone={priorityTone(task.priority) as any}>{priorityLabel(task.priority)}</Badge>
        <span>{statusLabel(task.status_name)}</span>
        <span>{task.assignee_name || "Не назначен"}</span>
        <span>{formatDate(task.due_date, "Без срока")}</span>
        <span>{minutesLabel(task.spent_minutes_cached)} / {minutesLabel(task.estimate_minutes)}</span>
        {activeTaskTimer ? <span>Таймер {timerClock(activeTaskTimer.started_at, activeTaskTimer.elapsed_minutes, now)}</span> : null}
      </div>
      <div className="work-task-screen-buttons">
        {canUseTime && activeTaskTimer ? (
          <Button variant="destructive" size="sm" type="button" onClick={() => void stopTimer().catch((err) => setError(err.message))}>
            <Square size={15} /> Остановить таймер
          </Button>
        ) : null}
        {canUseTime && !activeTaskTimer ? (
          <Button size="sm" type="button" onClick={() => void startTimer().catch((err) => setError(err.message))}>
            <Play size={15} /> Запустить таймер
          </Button>
        ) : null}
        <Link className="ui-button ui-button--secondary ui-button--sm" href="/tasks/time">
          <Timer size={15} />
          <span>Таймтрекер</span>
        </Link>
      </div>
    </header>
  );
}
