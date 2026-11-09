import Link from "next/link";
import { Building2, FolderKanban, PlayCircle, Workflow } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";

type DashboardQuickActionsProps = {
  activeCompanyId?: string | null;
  activeProjectId?: string | null;
};

export function DashboardQuickActions({ activeCompanyId, activeProjectId }: DashboardQuickActionsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Быстрые действия</CardTitle>
        <CardDescription>Продолжите работу в нужном разделе.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Link className="between" href="/companies">
          <span className="row"><Building2 size={18} /> Компании</span>
          <Badge>Открыть</Badge>
        </Link>
        {activeCompanyId ? (
          <Link className="between" href="/projects">
            <span className="row"><FolderKanban size={18} /> Проекты компании</span>
            <Badge>Открыть</Badge>
          </Link>
        ) : null}
        {activeProjectId ? (
          <>
            <Link className="between" href="/project-workflows">
              <span className="row"><Workflow size={18} /> Сценарии проекта</span>
              <Badge>Открыть</Badge>
            </Link>
            <Link className="between" href="/project-integrations">
              <span className="row"><PlayCircle size={18} /> Интеграции</span>
              <Badge>Открыть</Badge>
            </Link>
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}
