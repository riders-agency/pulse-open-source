import Link from "next/link";

import { Badge } from "../../../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Table } from "../../../components/ui/table";
import type { Run } from "../../../lib/api";
import { formatDate } from "../../../lib/utils";

type DashboardRunsTableProps = {
  runs: Run[];
};

function runStatusTone(status: string) {
  if (status === "succeeded") return "success";
  if (status === "failed") return "danger";
  return "info";
}

export function DashboardRunsTable({ runs }: DashboardRunsTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Последние запуски</CardTitle>
        <CardDescription>История выполнения сценариев проекта.</CardDescription>
      </CardHeader>
      <CardContent>
        {runs.length === 0 ? <div className="empty-state">Запусков пока нет</div> : <RunsTable rows={runs} />}
      </CardContent>
    </Card>
  );
}

function RunsTable({ rows }: { rows: Run[] }) {
  return (
    <Table>
      <thead>
        <tr>
          <th>Запуск</th>
          <th>Статус</th>
          <th>Начало</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {rows.slice(0, 8).map((run) => (
          <tr key={run.workflow_run_id}>
            <td className="mono">{run.workflow_run_id}</td>
            <td><Badge tone={runStatusTone(run.status)}>{run.status}</Badge></td>
            <td>{formatDate(run.started_at)}</td>
            <td><Link className="action-link" href={`/runs/${run.workflow_run_id}`}>Карта выполнения</Link></td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
