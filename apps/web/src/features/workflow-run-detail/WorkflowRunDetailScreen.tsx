"use client";

import { graphAriaLabels } from "../ai-employee/graph/graphLabels";
import { OfficeDecisionPanel } from "./ui/OfficeDecisionPanel";
import { OfficeUsagePanel } from "./ui/OfficeUsagePanel";
import { Background, Controls, MiniMap, ReactFlow } from "@xyflow/react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Sheet, SheetDescription, SheetHeader, SheetTitle } from "../../components/ui/sheet";
import { Table } from "../../components/ui/table";
import { formatDate } from "../../lib/utils";
import { useWorkflowRunDetail } from "./model/useWorkflowRunDetail";
import { actionTypeLabel, sourceTypeLabel, statusLabel } from "./model/workflowRunLabels";
import { RunActionDetailPanel } from "./ui/RunActionDetailPanel";
import { RunDataPanel } from "./ui/RunDataPanel";
import { nodeTypes } from "./ui/runtimeGraph";

export function WorkflowRunDetailScreen({ workflowRunId }: { workflowRunId: string }) {
  const detail = useWorkflowRunDetail(workflowRunId);

  const approval = detail.actions.find((action) => action.action_type_key === "office.approve" && action.status === "waiting");
  const officeRun = detail.actions.some((action) => action.action_type_key === "mailbox.incoming.trigger");
  return (
    <main className="runtime-shell">
      <section className="runtime-canvas">
        <div className="flow-toolbar">
          <div>
            <strong>Карта выполнения</strong>
            <div className="muted runtime-run-subtitle">
              Запуск {workflowRunId} · {sourceTypeLabel(detail.run?.source_type)} · {detail.run ? statusLabel(detail.run.status) : "загрузка"}
            </div>
          </div>
          <div className="row">
            {detail.error ? <Badge tone="danger">{detail.error}</Badge> : null}
            {!officeRun && <Button variant="secondary" disabled={detail.replayAllBusy} onClick={detail.replayFull}>
              {detail.replayAllBusy ? "Создание запуска…" : "Повторить запуск"}
            </Button>}
          </div>
        </div>
        <ReactFlow
          nodes={detail.nodes}
          edges={detail.edges}
          nodeTypes={nodeTypes}
          ariaLabelConfig={graphAriaLabels}
          onNodeClick={(_, node) => detail.setSelectedKey(node.id)}
          fitView
          minZoom={0.2}
          fitViewOptions={{ padding: 0.15 }}
          nodesDraggable={false}
        >
          <Background />
          {!officeRun && <MiniMap ariaLabel="Мини-карта" />}
          <Controls />
        </ReactFlow>
      </section>

      <Sheet open>
        <SheetHeader>
          <SheetTitle>Результат действия</SheetTitle>
          <SheetDescription>{detail.selected ? actionTypeLabel(detail.selected.action_type_key) : "Выберите блок на карте"}</SheetDescription>
        </SheetHeader>
        <div className="stack runtime-detail-stack">
          {approval && detail.run?.project_id && <OfficeDecisionPanel key={workflowRunId} projectId={detail.run.project_id} runId={workflowRunId} output={approval.output_json} />}
          {officeRun && <OfficeUsagePanel actions={detail.actions} />}
          <details open={officeRun ? undefined : true}><summary>Подробности шагов и журнал</summary>
          <RunDataPanel run={detail.run} selected={detail.selected} latestAttempt={detail.latestAttempt} />
          <RunActionDetailPanel
            allowReplay={!officeRun}
            latestAttempt={detail.latestAttempt}
            latestAttemptDiagnostics={detail.latestAttemptDiagnostics}
            replayAllBusy={detail.replayAllBusy}
            replayOneAction={detail.replayOneAction}
            retryActionBusy={detail.retryActionBusy}
            selected={detail.selected}
          />
          <Card>
            <CardHeader>
              <CardTitle>События запуска</CardTitle>
              <CardDescription>События сохраняются во время выполнения.</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <thead><tr><th>Событие</th><th>Действие</th><th>Время</th></tr></thead>
                <tbody>
                  {detail.events.map((event) => (
                    <tr key={event.run_event_id || `${event.event_type}-${event.created_at}`}>
                      <td>{event.event_type}</td>
                      <td>{event.action_key || "—"}</td>
                      <td>{formatDate(event.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </CardContent>
          </Card>
          </details>
        </div>
      </Sheet>
    </main>
  );
}
