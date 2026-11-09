"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Table } from "../../components/ui/table";
import type { Run } from "../../lib/api";
import { listWorkflowRuns, startManualWorkflowRun } from "./api/workflowRunApi";
import { formatDate } from "../../lib/utils";
import { usePulseStore } from "../../store/pulse-store";

export default function WorkflowRunsPage() {
  const { workflowId } = useParams<{ workflowId: string }>();
  const token = usePulseStore((s) => s.token);
  const [rows, setRows] = useState<Run[]>([]);
  const [error, setError] = useState("");

  const load = () => {
    if (!token) return;
    listWorkflowRuns(workflowId, token)
      .then((data) => setRows(data))
      .catch((err) => setError(err.message));
  };

  useEffect(load, [token, workflowId]);

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
          <p className="muted">История запусков отдельно от визуального редактора.</p>
        </div>
        <div className="row">
          <Link href={`/workflows/${workflowId}/edit`}><Button variant="outline">Редактор</Button></Link>
          <Button onClick={manualRun}>Запустить вручную</Button>
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
                    <td><Badge tone={run.source_type === "test" ? "warning" : "muted"}>{run.source_type ?? "—"}</Badge></td>
                    <td><Badge tone={run.status === "succeeded" ? "success" : run.status === "failed" ? "danger" : "info"}>{run.status}</Badge></td>
                    <td>{formatDate(run.started_at)}</td>
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
