import Link from "next/link";

import { Button } from "../../../components/ui/button";

type DashboardHeaderProps = {
  activeCompanyId?: string | null;
  activeProjectId?: string | null;
};

export function DashboardHeader({ activeCompanyId, activeProjectId }: DashboardHeaderProps) {
  return (
    <div className="page-title">
      <div>
        <h1>Обзор</h1>
        <p className="muted">Задачи, сценарии и события вашего проекта.</p>
      </div>
      <div className="row">
        <Link href="/companies/new">
          <Button>Создать компанию</Button>
        </Link>
        {activeCompanyId ? (
          <Link href="/projects/new">
            <Button variant="secondary">Создать проект</Button>
          </Link>
        ) : null}
        {activeProjectId ? (
          <Link href="/project-workflows/new">
            <Button variant="outline">Создать сценарий</Button>
          </Link>
        ) : null}
      </div>
    </div>
  );
}
