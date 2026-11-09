"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { AlertTriangle, CheckCircle2, Clock3, Timer } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Card, CardContent } from "../../components/ui/card";
import type { WorkAccess, WorkOverview } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { getWorkOverview } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { minutesLabel } from "./work-utils";

export function WorkHomePage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [overview, setOverview] = useState<WorkOverview>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token || !projectId) return;
    getWorkOverview(projectId, token)
      .then((payload) => {
        setAccess(payload.access);
        setOverview(payload.overview);
        setError("");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить Work Tracker"));
  }, [projectId, token]);

  if (!projectId) return <WorkNoProject />;

  return (
    <WorkFrame access={access} title="Задачи" description="Единый модуль задач проекта: списки, доски, сроки, время и отчеты.">
      {error ? <div className="form-error">{error}</div> : null}
      <section className="work-metrics">
        <Metric icon={<Clock3 size={18} />} label="Сегодня" value={overview?.today_tasks || 0} />
        <Metric icon={<AlertTriangle size={18} />} label="Просрочено" value={overview?.overdue_tasks || 0} tone="danger" />
        <Metric icon={<CheckCircle2 size={18} />} label="На приемке" value={overview?.client_acceptance_tasks || 0} />
        <Metric icon={<Timer size={18} />} label="Время за неделю" value={minutesLabel(overview?.logged_week_minutes)} />
      </section>
      <section className="work-focus-grid">
        <Link className="work-focus" href="/tasks/all">
          <strong>Все задачи</strong>
          <span>Полный список с фильтрами, быстрым созданием и карточкой задачи.</span>
          <Badge>Открыть</Badge>
        </Link>
        <Link className="work-focus" href="/tasks/boards">
          <strong>Доски</strong>
          <span>Канбан и предустановленные представления под разные процессы.</span>
          <Badge>Открыть</Badge>
        </Link>
        <Link className="work-focus" href="/tasks/time">
          <strong>Таймтрекер</strong>
          <span>Таймер, ручные записи и табель времени проекта.</span>
          <Badge>Открыть</Badge>
        </Link>
        {access?.can_see_reports ? (
          <Link className="work-focus" href="/tasks/reports">
            <strong>Отчеты</strong>
            <span>Клиенты, сотрудники, статусы и оплачиваемое время.</span>
            <Badge>Открыть</Badge>
          </Link>
        ) : null}
      </section>
      <Card>
        <CardContent className="work-risk-list">
          <Risk label="Без ответственного" value={overview?.unassigned_tasks || 0} />
          <Risk label="Без оценки" value={overview?.unestimated_tasks || 0} />
          <Risk label="Заблокировано" value={overview?.blocked_tasks || 0} />
          <Risk label="На ревью" value={overview?.review_tasks || 0} />
        </CardContent>
      </Card>
    </WorkFrame>
  );
}

function Metric({ icon, label, value, tone }: { icon: React.ReactNode; label: string; value: string | number; tone?: string }) {
  return (
    <div className={`work-metric ${tone ? `is-${tone}` : ""}`}>
      {icon}
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function Risk({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
