"use client";

import { useEffect, useState } from "react";

import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/card";
import type { WorkAccess, WorkReport, WorkReportBucket } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { getWorkReport } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { minutesLabel } from "./work-utils";

export function WorkReportsPage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [report, setReport] = useState<WorkReport>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token || !projectId) return;
    getWorkReport(projectId, token)
      .then((payload) => {
        setAccess(payload.access);
        setReport(payload.report);
        setError("");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить отчеты"));
  }, [projectId, token]);

  if (!projectId) return <WorkNoProject />;

  return (
    <WorkFrame access={access} title="Отчеты" description="Время, просрочки и bottlenecks по текущему проекту.">
      {error ? <div className="form-error">{error}</div> : null}
      {!access ? (
        <section className="work-empty"><h2>Загрузка отчетов</h2><p>Собираем время, статусы и клиентские срезы проекта.</p></section>
      ) : !access.can_see_reports ? (
        <section className="work-empty"><h2>Отчеты недоступны</h2><p>Для этой роли внутренние отчеты проекта закрыты.</p></section>
      ) : (
        <>
          <section className="work-metrics">
            <Metric label="Создано" value={report?.created_tasks || 0} />
            <Metric label="Закрыто" value={report?.completed_tasks || 0} />
            <Metric label="Просрочено" value={report?.overdue_tasks || 0} />
            <Metric label="Оплачиваемое время" value={minutesLabel(report?.billable_minutes)} />
          </section>
          <section className="work-report-grid">
            <Bucket title="По клиентам" rows={report?.by_client || []} />
            <Bucket title="По сотрудникам" rows={report?.by_employee || []} />
            <Bucket title="По статусам" rows={report?.by_status || []} />
            <Bucket title="По проекту" rows={report?.by_project || []} />
          </section>
        </>
      )}
    </WorkFrame>
  );
}

function Metric({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="work-metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function Bucket({ title, rows }: { title: string; rows: WorkReportBucket[] }) {
  return (
    <Card>
      <CardHeader><CardTitle>{title}</CardTitle></CardHeader>
      <CardContent className="work-buckets">
        {rows.map((row) => (
          <div key={`${title}-${row.id || row.name}`}>
            <span>{row.name}</span>
            <strong>{minutesLabel(row.logged_minutes)}</strong>
          </div>
        ))}
        {rows.length === 0 ? <div className="empty-state">Нет данных</div> : null}
      </CardContent>
    </Card>
  );
}
