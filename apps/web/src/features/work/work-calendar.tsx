"use client";

import { useEffect, useMemo, useState } from "react";

import { Badge } from "../../components/ui/badge";
import type { WorkAccess, WorkTaskList } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { listWorkTasks } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { priorityLabel, priorityTone, shortDate } from "./work-utils";

export function WorkCalendarPage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [tasks, setTasks] = useState<WorkTaskList>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token || !projectId) return;
    listWorkTasks(projectId, token, { limit: 160 })
      .then((payload) => {
        setAccess(payload.access);
        setTasks(payload.tasks);
        setError("");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить календарь"));
  }, [projectId, token]);

  const byDate = useMemo(() => {
    const map = new Map<string, NonNullable<typeof tasks>["items"]>();
    for (const task of tasks?.items || []) {
      const key = task.due_date ? shortDate(task.due_date) : "Без срока";
      map.set(key, [...(map.get(key) || []), task]);
    }
    return [...map.entries()];
  }, [tasks]);

  if (!projectId) return <WorkNoProject />;

  return (
    <WorkFrame access={access} title="Календарь задач" description="Задачи проекта, сгруппированные по сроку.">
      {error ? <div className="form-error">{error}</div> : null}
      <section className="work-calendar-grid">
        {byDate.map(([date, rows]) => (
          <article className="work-date-card" key={date}>
            <header><strong>{date}</strong><span>{rows.length}</span></header>
            {rows.slice(0, 8).map((task) => (
              <div className="work-date-task" key={task.task_id}>
                <strong>{task.title}</strong>
                <Badge tone={priorityTone(task.priority) as any}>{priorityLabel(task.priority)}</Badge>
              </div>
            ))}
          </article>
        ))}
      </section>
    </WorkFrame>
  );
}
