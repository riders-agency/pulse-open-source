"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Table } from "../../components/ui/table";
import { asArray, type Workflow } from "../../lib/api";
import { listProjectWorkflowsIfAllowed } from "./api/projectWorkflowsApi";
import { statusLabel, statusTone } from "../../lib/client-labels";
import { usePulseStore } from "../../store/pulse-store";

export default function WorkflowsPage() {
  const params = useParams<{ projectId?: string }>();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const [rows, setRows] = useState<Workflow[]>([]);
  const [error, setError] = useState("");
  const [denied, setDenied] = useState(false);
  const safeRows = useMemo((): Workflow[] => asArray<Workflow>(rows), [rows]);
  const projectId = params.projectId || activeProjectId;

  useEffect(() => {
    if (!token) return;
    if (!projectId) return;
    if (params.projectId) setActiveProject(projectId);
    setDenied(false);
    setError("");
    listProjectWorkflowsIfAllowed(projectId, token)
      .then(({ denied: accessDenied, workflows }) => {
        if (accessDenied) {
          setRows([]);
          setDenied(true);
          return;
        }
        setRows(workflows);
      })
      .catch((err) => {
        setRows([]);
        setDenied(true);
        setError(err instanceof Error ? err.message : "Нет доступа к сценариям проекта");
      });
  }, [projectId, params.projectId, setActiveProject, token]);

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Сценарии</h1>
          <p className="muted">Сценарий — главный объект продукта. Откройте редактор, чтобы собрать цепочку действий.</p>
        </div>
        {!denied ? (
          <Link href="/project-workflows/new">
            <Button>Создать сценарий</Button>
          </Link>
        ) : null}
      </div>
      {error && !denied ? <Badge tone="danger">{error}</Badge> : null}
      {denied ? (
        <Card>
          <CardHeader>
            <CardTitle>Сценарии скрыты</CardTitle>
            <CardDescription>Ваша роль может работать с задачами процесса, но не видеть и не редактировать общие сценарии проекта.</CardDescription>
          </CardHeader>
        </Card>
      ) : null}
      {!denied ? <Card>
        <CardHeader>
          <CardTitle>Сценарии проекта</CardTitle>
          <CardDescription>Редактор сценария открывается на отдельном рабочем экране.</CardDescription>
        </CardHeader>
        <CardContent>
          {safeRows.length === 0 ? (
            <div className="empty-state">Сценариев пока нет</div>
          ) : (
            <>
              <div className="workflow-table-view">
                <Table className="workflow-table">
                  <thead>
                    <tr>
                      <th>Название</th>
                      <th>Короткий адрес</th>
                      <th>Статус</th>
                      <th>Запуски</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {safeRows.map((workflow) => (
                      <tr key={workflow.workflow_id}>
                        <td>{workflow.name}</td>
                        <td>{workflow.slug}</td>
                        <td><Badge tone={statusTone(workflow.status)}>{statusLabel(workflow.status || "draft")}</Badge></td>
                        <td className="workflow-run-count">{formatRunCount(workflow.run_count)}</td>
                        <td>
                          <div className="workflow-row-actions">
                            <Link className="action-link" href={`/workflows/${workflow.workflow_id}/edit`}>Редактор</Link>
                            <Link className="action-link" href={`/workflows/${workflow.workflow_id}/runs`}>Запуски</Link>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
              <div className="workflow-mobile-list" aria-label="Сценарии проекта">
                {safeRows.map((workflow) => (
                  <article key={workflow.workflow_id}>
                    <div>
                      <strong>{workflow.name}</strong>
                      <span>{workflow.slug}</span>
                    </div>
                    <div className="workflow-mobile-meta">
                      <Badge tone={statusTone(workflow.status)}>{statusLabel(workflow.status || "draft")}</Badge>
                      <span>{formatRunCount(workflow.run_count)} запусков</span>
                    </div>
                    <div className="workflow-row-actions">
                      <Link className="action-link" href={`/workflows/${workflow.workflow_id}/edit`}>Редактор</Link>
                      <Link className="action-link" href={`/workflows/${workflow.workflow_id}/runs`}>Запуски</Link>
                    </div>
                  </article>
                ))}
              </div>
            </>
          )}
        </CardContent>
      </Card> : null}
    </main>
  );
}

function formatRunCount(value: number | undefined) {
  return typeof value === "number" && Number.isFinite(value) ? String(value) : "0";
}
