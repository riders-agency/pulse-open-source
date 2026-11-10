import Link from "next/link";
import { AlertTriangle } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import type { Run } from "../../../lib/api";

type DashboardErrorsCardProps = {
  latestErrors: Run[];
};

export function DashboardErrorsCard({ latestErrors }: DashboardErrorsCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Последние ошибки</CardTitle>
        <CardDescription>Сценарии, которым нужно ваше внимание.</CardDescription>
      </CardHeader>
      <CardContent>
        {latestErrors.length === 0 ? (
          <div className="empty-state">Ошибок пока нет</div>
        ) : (
          <div className="stack">
            {latestErrors.map((run) => (
              <Link className="between" href={`/runs/${run.workflow_run_id}`} key={run.workflow_run_id}>
                <span className="row"><AlertTriangle size={18} /> Запуск {run.workflow_run_id.slice(0, 8)}</span>
                <Badge tone="danger">{run.status}</Badge>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
