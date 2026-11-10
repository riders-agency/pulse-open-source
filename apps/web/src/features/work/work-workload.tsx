"use client";

import { useEffect, useMemo, useState } from "react";

import { Badge } from "../../components/ui/badge";
import type { WorkAccess, WorkEmployee, WorkTaskList } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { listWorkEmployees, listWorkTasks } from "./api/workApi";
import { meterFillClass } from "./model/workStyle";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { minutesLabel } from "./work-utils";

export function WorkWorkloadPage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [employees, setEmployees] = useState<WorkEmployee[]>([]);
  const [tasks, setTasks] = useState<WorkTaskList>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token || !projectId) return;
    Promise.all([
      listWorkEmployees(projectId, token),
      listWorkTasks(projectId, token, { limit: 160 }),
    ])
      .then(([people, taskPayload]) => {
        setAccess(people.access);
        setEmployees(people.items || []);
        setTasks(taskPayload.tasks);
        setError("");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить загрузку сотрудников"));
  }, [projectId, token]);

  const planned = useMemo(() => {
    const map = new Map<string, number>();
    for (const task of tasks?.items || []) {
      if (!task.assignee_user_id) continue;
      map.set(task.assignee_user_id, (map.get(task.assignee_user_id) || 0) + (task.estimate_minutes || 0));
    }
    return map;
  }, [tasks]);

  if (!projectId) return <WorkNoProject />;

  return (
    <WorkFrame access={access} title="Загрузка" description="Загрузка сотрудников: доступность, план и фактическое время.">
      {error ? <div className="form-error">{error}</div> : null}
      <section className="workload-grid">
        {employees.map((employee) => {
          const plannedMinutes = planned.get(employee.user_id) || 0;
          const capacity = employee.capacity_hours_per_week * 60;
          const loadPercent = Math.min(100, Math.round((plannedMinutes / Math.max(1, capacity)) * 100));
          const over = plannedMinutes > capacity;
          return (
            <article className="workload-row" key={employee.employee_profile_id}>
              <div>
                <strong>{employee.full_name}</strong>
                <span>{employee.role} · {employee.department || "Без отдела"}</span>
              </div>
              <div className="workload-meter"><span className={meterFillClass(loadPercent, "workload-meter-fill")} /></div>
              <span>{minutesLabel(plannedMinutes)} / {employee.capacity_hours_per_week}ч</span>
              {over ? <Badge tone="danger">Перегруз</Badge> : <Badge tone="success">Норма</Badge>}
            </article>
          );
        })}
      </section>
    </WorkFrame>
  );
}
