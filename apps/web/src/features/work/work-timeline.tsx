"use client";

import { useEffect, useMemo, useState } from "react";

import type { WorkAccess, WorkTaskList } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { listWorkTasks } from "./api/workApi";
import { timelineOffsetClass, timelineSpanClass } from "./model/workStyle";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { shortDate } from "./work-utils";

export function WorkTimelinePage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [tasks, setTasks] = useState<WorkTaskList>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token || !projectId) return;
    listWorkTasks(projectId, token, { limit: 120 })
      .then((payload) => {
        setAccess(payload.access);
        setTasks(payload.tasks);
        setError("");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить таймлайн"));
  }, [projectId, token]);

  const rows = useMemo(() => (tasks?.items || []).filter((task) => task.start_date || task.due_date).slice(0, 60), [tasks]);
  if (!projectId) return <WorkNoProject />;

  return (
    <WorkFrame access={access} title="План сроков" description="Календарная линия задач: старт, дедлайн, просрочки и зависимости по срокам.">
      {error ? <div className="form-error">{error}</div> : null}
      <section className="work-timeline">
        {rows.map((task, index) => (
          <article className="work-timeline-row" key={task.task_id}>
            <div>
              <strong>{task.title}</strong>
            </div>
            <div className={`work-timeline-bar ${timelineOffsetClass(index)} ${timelineSpanClass(index)}`}>
              <span>{shortDate(task.start_date)} - {shortDate(task.due_date)}</span>
            </div>
          </article>
        ))}
      </section>
    </WorkFrame>
  );
}
