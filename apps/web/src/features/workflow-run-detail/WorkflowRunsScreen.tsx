"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Table } from "../../components/ui/table";
import { sourceTypeLabel, statusLabel, statusTone } from "./model/workflowRunLabels";
import { apiGet } from "../../lib/api";
import type { Run } from "../../lib/api";
import { listWorkflowRuns, startManualWorkflowRun } from "./api/workflowRunApi";
import { formatDateTime } from "../../lib/utils";
import { usePulseStore } from "../../store/pulse-store";

export default function WorkflowRunsPage() {
  const { workflowId } = useParams<{ workflowId: string }>();
  const token = usePulseStore((s) => s.token);
  const [office, setOffice] = useState(false);
  const [rows, setRows] = useState<Run[]>([]);
  const [error, setError] = useState("");
  useEffect(() => { if (token) void apiGet<{ slug: string }>(`/api/v1/workflows/${workflowId}`, token).then((row) => setOffice(row.slug === "inbox-ai-departments")).catch((err: Error) => setError(err.message)); }, [token, workflowId]);

  const load = () => {
    if (!token) return;
    listWorkflowRuns(workflowId, token)
      .then((data) => setRows(data))
      .catch((err) => setError(err.message));
  };

  useEffect(() => { load(); const timer = window.setInterval(load, 3000); return () => window.clearInterval(timer); }, [token, workflowId]);

  const manualRun = async () => {
    setError("");
    try {
      await startManualWorkflowRun(workflowId, token);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось запустить сценарий");
    }
  };

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Запуски сценария</h1>
          <p className="muted">{office ? "Последние 30 исполнений. Опрос почты каждые 30 секунд." : "Откройте запуск, чтобы увидеть действия и результат."}</p>
        </div>
        <div className="row">
          <Link href={`/workflows/${workflowId}/edit`}><Button variant="outline">Редактор</Button></Link>
          {!office && <Button onClick={manualRun}>Запустить вручную</Button>}
        </div>
      </div>
      {error ? <Badge tone="danger">{error}</Badge> : null}
      <Card>
        <CardHeader>
          <CardTitle>История запусков</CardTitle>
          <CardDescription>Откройте запуск, чтобы увидеть карту выполнения.</CardDescription>
        </CardHeader>
        <CardContent>
          {rows.length === 0 ? (
            <div className="empty-state">Запусков пока нет</div>
          ) : (
            <Table>
              <thead><tr><th>Запуск</th><th>Источник</th><th>Статус</th><th>Начало</th><th></th></tr></thead>
              <tbody>
                {rows.map((run) => (
                  <tr key={run.workflow_run_id}>
                    <td className="mono">{run.workflow_run_id}</td>
                    <td><Badge tone={run.source_type === "test" ? "warning" : "muted"}>{sourceTypeLabel(run.source_type)}</Badge></td>
                    <td><Badge tone={statusTone(run.status)}>{statusLabel(run.status)}</Badge></td>
                    <td>{formatDateTime(run.started_at)}</td>
                    <td><Link className="action-link" href={`/runs/${run.workflow_run_id}`}>Карта выполнения</Link></td>
                  </tr>
                ))}
              </tbody>
            </Table>
          )}
        </CardContent>
      </Card>
    </main>
  );
}
