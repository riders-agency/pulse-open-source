"use client";

import { AlertTriangle, MessageSquare, Paperclip, Play } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import type { WorkAccess, WorkTask } from "../../lib/api";
import { meterFillClass } from "./model/workStyle";
import { minutesLabel, priorityLabel, priorityTone, shortDate, taskIsOverdue, taskTypeLabel } from "./work-utils";

export function TaskCard({ task, access, onOpen, onStartTimer }: { task: WorkTask; access?: WorkAccess; onOpen: () => void; onStartTimer: () => void }) {
  const estimate = task.estimate_minutes || 0;
  const spent = task.spent_minutes_cached || 0;
  const timeProgress = estimate ? Math.min(100, Math.round((spent / estimate) * 100)) : 0;
  return (
    <article
      className={`work-task-card ${taskIsOverdue(task) ? "is-overdue" : ""}`}
      draggable={Boolean(access?.can_write_tasks && !access.is_client_guest)}
      onDragStart={(event) => {
        event.dataTransfer.setData("text/task-id", task.task_id);
        event.dataTransfer.setData("text/task-revision", String(task.revision));
      }}
    >
      <button className="work-task-card-open" type="button" onClick={onOpen} aria-label={`Открыть карточку задачи ${task.title}`}>
        <strong>{task.title}</strong>
        <span>{task.description?.trim() || "Описание не заполнено"}</span>
      </button>
      <div className="work-task-badges">
        <Badge tone="muted">{taskTypeLabel(task.type)}</Badge>
        <Badge tone={priorityTone(task.priority) as any}>{priorityLabel(task.priority)}</Badge>
        {task.is_blocked ? <span className="work-danger"><AlertTriangle size={14} /> заблокирована</span> : null}
      </div>
      <div className="work-task-facts">
        <span>{task.assignee_name || "Не назначен"}</span>
        <span>{shortDate(task.due_date)}</span>
        <span>{task.client_name || "Без клиента"}</span>
      </div>
      <div className="work-task-progress">
        <div><span>Время</span><strong>{minutesLabel(spent)} / {minutesLabel(estimate)}</strong></div>
        <i className={meterFillClass(timeProgress, "work-task-progress-fill")} />
      </div>
      <div className="work-task-icons">
        {task.checklist_total_count ? <span>{task.checklist_done_count}/{task.checklist_total_count}</span> : null}
        {task.comments_count ? <span><MessageSquare size={13} />{task.comments_count}</span> : null}
        {task.attachments_count ? <span><Paperclip size={13} />{task.attachments_count}</span> : null}
        {access?.can_write_tasks && !access.is_client_guest ? (
          <button type="button" title="Запустить таймер" onClick={onStartTimer}>
            <Play size={13} />
          </button>
        ) : null}
      </div>
    </article>
  );
}
