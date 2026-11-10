"use client";

import { CheckCircle2, History, MessageSquare, Play, Shield, Square, Timer } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Field, Textarea } from "../../components/ui/input";
import type { WorkAccess, WorkTaskDetail, WorkTimer } from "../../lib/api";
import { formatDate, formatDateTime } from "../../lib/dates";
import { minutesLabel, timeSourceLabel, timerClock, visibilityLabel } from "./work-utils";
import type { TaskTab } from "./work-task-types";

export { WorkTaskAttachmentsTab } from "./work-task-attachments-tab";

const tabs: Array<[TaskTab, (task: WorkTaskDetail) => string]> = [
  ["card", () => "Карточка"],
  ["comments", (task) => `Комментарии ${task.comments?.length || 0}`],
  ["time", () => "Время"],
  ["activity", () => "Действия"],
  ["access", () => "Доступы"],
  ["attachments", (task) => `Вложения ${task.attachments?.length || 0}`],
];

export function WorkTaskTabsNav({ setTab, tab, task }: { setTab: (value: TaskTab) => void; tab: TaskTab; task: WorkTaskDetail }) {
  return (
    <nav className="work-drawer-tabs" aria-label="Разделы задачи">
      {tabs.map(([key, label]) => (
        <button className={tab === key ? "is-active" : ""} key={key} type="button" onClick={() => setTab(key)}>
          {label(task)}
        </button>
      ))}
    </nav>
  );
}

export function WorkTaskCommentsTab({
  addComment,
  comment,
  commentLabel,
  setComment,
  setError,
  task,
}: {
  addComment: () => Promise<void>;
  comment: string;
  commentLabel: string;
  setComment: (value: string) => void;
  setError: (value: string) => void;
  task: WorkTaskDetail;
}) {
  return (
    <section className="work-comments">
      {(task.comments || []).map((item) => (
        <article key={item.task_comment_id}>
          <header><strong>{item.author_name}</strong><span>{formatDateTime(item.created_at)} · {visibilityLabel(item.visibility)}</span></header>
          <p>{item.body}</p>
        </article>
      ))}
      <form
        className="work-comment-form"
        onSubmit={(event) => {
          event.preventDefault();
          void addComment().catch((err) => setError(err.message));
        }}
      >
        <Field label={commentLabel}>
          <Textarea value={comment} onChange={(event) => setComment(event.target.value)} />
        </Field>
        <Button type="submit"><MessageSquare size={16} /> Добавить комментарий</Button>
      </form>
    </section>
  );
}

export function WorkTaskTimeTab({
  activeTaskTimer,
  canUseTime,
  now,
  setError,
  startTimer,
  stopTimer,
  task,
  taskTimers,
}: {
  activeTaskTimer?: WorkTimer;
  canUseTime: boolean;
  now: number;
  setError: (value: string) => void;
  startTimer: () => Promise<void>;
  stopTimer: () => Promise<void>;
  task: WorkTaskDetail;
  taskTimers: WorkTimer[];
}) {
  return (
    <section className="work-task-tab-grid">
      <div className="work-timer-strip">
        <div>
          <Timer size={18} />
          <span>{activeTaskTimer ? "Таймер по этой задаче идет" : "Запуск таймера не останавливает другие активные задачи."}</span>
        </div>
        <strong>{activeTaskTimer ? timerClock(activeTaskTimer.started_at, activeTaskTimer.elapsed_minutes, now) : minutesLabel(task.spent_minutes_cached)}</strong>
        {canUseTime && activeTaskTimer ? <Button variant="destructive" size="sm" type="button" onClick={() => void stopTimer().catch((err) => setError(err.message))}><Square size={15} /> Стоп</Button> : null}
        {canUseTime && !activeTaskTimer ? <Button size="sm" type="button" onClick={() => void startTimer().catch((err) => setError(err.message))}><Play size={15} /> Старт</Button> : null}
      </div>
      {taskTimers.length > 1 ? (
        <div className="work-success-note">По этой задаче запущено {taskTimers.length} таймера. Кнопка остановит последний активный.</div>
      ) : null}
      <div className="work-time-list">
        {(task.time_entries || []).map((entry) => (
          <article key={entry.time_entry_id}>
            <CheckCircle2 size={16} />
            <strong>{minutesLabel(entry.duration_minutes)}</strong>
            <span>{entry.user_name} · {formatDate(entry.entry_date)} · {timeSourceLabel(entry.source)}</span>
          </article>
        ))}
        {(task.time_entries || []).length === 0 ? <div className="work-empty">По этой задаче еще нет записей времени.</div> : null}
      </div>
    </section>
  );
}

export function WorkTaskActivityTab({ task }: { task: WorkTaskDetail }) {
  return (
    <section className="work-activity work-task-tab-grid">
      {(task.activity || []).map((event, index) => (
        <span key={`${event.event_type}-${index}`}><History size={14} /> {event.event_type} · {event.actor_name} · {formatDateTime(event.created_at)}</span>
      ))}
    </section>
  );
}

export function WorkTaskAccessTab({ canWrite, effectiveAccess, patch, task }: { canWrite: boolean; effectiveAccess?: WorkAccess; patch: (body: Record<string, unknown>) => Promise<void>; task: WorkTaskDetail }) {
  return (
    <section className="work-task-tab-grid">
      <div className="work-access-card">
        <Shield size={18} />
        <div>
          <strong>{effectiveAccess?.role_name || effectiveAccess?.role || "Роль не определена"}</strong>
          <span>{effectiveAccess?.is_client_guest ? "Клиентский доступ" : "Внутренний доступ"} · {effectiveAccess?.company_name}</span>
        </div>
      </div>
      <div className="work-permission-grid">
        <Permission label="Читать задачи" value={effectiveAccess?.can_read} />
        <Permission label="Редактировать задачи" value={effectiveAccess?.can_write_tasks} />
        <Permission label="Процессы" value={effectiveAccess?.can_manage_processes} />
        <Permission label="Участники" value={effectiveAccess?.can_manage_members} />
        <Permission label="Отчеты" value={effectiveAccess?.can_see_reports} />
        <Permission label="Утверждать время" value={effectiveAccess?.can_approve_time} />
      </div>
      <div className="work-drawer-section">
        <header><strong>Видимость для клиента</strong></header>
        <label className="row">
          <input disabled={!canWrite} checked={task.client_visible} type="checkbox" onChange={(event) => void patch({ client_visible: event.target.checked })} />
          <span>Клиент видит эту задачу в гостевом доступе</span>
        </label>
        <label className="row">
          <input disabled={!canWrite} checked={task.billable} type="checkbox" onChange={(event) => void patch({ billable: event.target.checked })} />
          <span>Время по задаче оплачиваемое</span>
        </label>
      </div>
    </section>
  );
}

function Permission({ label, value }: { label: string; value?: boolean }) {
  return (
    <div className="work-permission">
      <input checked={Boolean(value)} readOnly type="checkbox" />
      <span>{label}<small>{value ? "Разрешено" : "Запрещено"}</small></span>
    </div>
  );
}
