"use client";

import { Plus } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import type { WorkAccess, WorkBoard } from "../../lib/api";
import { TaskCard } from "./work-task-card";
import { statusLabel } from "./work-utils";

type WorkBoardColumnData = NonNullable<WorkBoard["columns"]>[number];

export function WorkBoardColumn({
  access,
  column,
  creatingStatus,
  onCreateTask,
  onMoveTask,
  onOpenTask,
  onStartTimer,
  setCreatingStatus,
  setError,
  setTitle,
  title,
}: {
  access?: WorkAccess;
  column: WorkBoardColumnData;
  creatingStatus: string;
  onCreateTask: (statusId: string) => Promise<void>;
  onMoveTask: (taskId: string, statusId: string, rank: number, expectedRevision?: number) => Promise<void>;
  onOpenTask: (taskId: string) => void;
  onStartTimer: (taskId: string) => Promise<void>;
  setCreatingStatus: (value: string) => void;
  setError: (value: string) => void;
  setTitle: (value: string) => void;
  title: string;
}) {
  const statusId = column.status.task_workflow_status_id;

  return (
    <div
      className="work-column"
      onDragOver={(event) => event.preventDefault()}
      onDrop={(event) => {
        const taskId = event.dataTransfer.getData("text/task-id");
        const revision = Number(event.dataTransfer.getData("text/task-revision"));
        if (taskId) void onMoveTask(taskId, statusId, Date.now() % 10_000_000, Number.isFinite(revision) ? revision : undefined);
      }}
    >
      <header>
        <div>
          <strong>{statusLabel(column.status.name)}</strong>
          <span>{column.count} задач</span>
        </div>
        {column.status.wip_limit ? <Badge tone={column.count > column.status.wip_limit ? "danger" : "muted"}>WIP {column.status.wip_limit}</Badge> : null}
      </header>
      <div className="work-column-list">
        {column.tasks.map((task) => (
          <TaskCard access={access} key={task.task_id} task={task} onOpen={() => onOpenTask(task.task_id)} onStartTimer={() => void onStartTimer(task.task_id)} />
        ))}
      </div>
      {access?.can_write_tasks && !access.is_client_guest ? (
        creatingStatus === statusId ? (
          <form
            className="work-inline-create"
            onSubmit={(event) => {
              event.preventDefault();
              void onCreateTask(statusId).catch((err) => setError(err.message));
            }}
          >
            <Input autoFocus value={title} onChange={(event) => setTitle(event.target.value)} placeholder="Название задачи" />
            <div className="row">
              <Button size="sm" type="submit">Создать</Button>
              <Button size="sm" type="button" variant="ghost" onClick={() => setCreatingStatus("")}>Отмена</Button>
            </div>
          </form>
        ) : (
          <button className="work-add-card" type="button" onClick={() => setCreatingStatus(statusId)}>
            <Plus size={16} /> Добавить задачу
          </button>
        )
      ) : null}
    </div>
  );
}
