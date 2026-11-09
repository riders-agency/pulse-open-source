"use client";

import Link from "next/link";

import { Badge } from "../../components/ui/badge";
import { Table } from "../../components/ui/table";
import type { WorkTaskList } from "../../lib/api";
import { minutesLabel, priorityLabel, priorityTone, shortDate, statusLabel, taskTypeLabel } from "./work-utils";

export function WorkTaskTable({ columns, list }: { columns: string[]; list?: WorkTaskList }) {
  return (
    <Table>
      <thead>
        <tr>
          {columns.includes("title") ? <th>Задача</th> : null}
          {columns.includes("status") ? <th>Статус</th> : null}
          {columns.includes("assignee") ? <th>Ответственный</th> : null}
          {columns.includes("priority") ? <th>Приоритет</th> : null}
          {columns.includes("type") ? <th>Тип</th> : null}
          {columns.includes("client") ? <th>Клиент</th> : null}
          {columns.includes("due") ? <th>Срок</th> : null}
          {columns.includes("estimate") ? <th>Оценка</th> : null}
          {columns.includes("spent") ? <th>Факт</th> : null}
          {columns.includes("checklist") ? <th>Чеклист</th> : null}
        </tr>
      </thead>
      <tbody>
        {(list?.items || []).map((task) => (
          <tr key={task.task_id}>
            {columns.includes("title") ? <td><Link className="work-task-title-button" href={`/tasks/${task.task_id}`}>{task.title}</Link></td> : null}
            {columns.includes("status") ? <td>{statusLabel(task.status_name)}</td> : null}
            {columns.includes("assignee") ? <td>{task.assignee_name || "Не назначен"}</td> : null}
            {columns.includes("priority") ? <td><Badge tone={priorityTone(task.priority) as any}>{priorityLabel(task.priority)}</Badge></td> : null}
            {columns.includes("type") ? <td>{taskTypeLabel(task.type)}</td> : null}
            {columns.includes("client") ? <td>{task.client_name || "Без клиента"}</td> : null}
            {columns.includes("due") ? <td>{shortDate(task.due_date)}</td> : null}
            {columns.includes("estimate") ? <td>{minutesLabel(task.estimate_minutes)}</td> : null}
            {columns.includes("spent") ? <td>{minutesLabel(task.spent_minutes_cached)}</td> : null}
            {columns.includes("checklist") ? <td>{task.checklist_done_count}/{task.checklist_total_count}</td> : null}
          </tr>
        ))}
      </tbody>
    </Table>
  );
}

export function columnLabel(column: string) {
  return {
    title: "Название",
    status: "Статус",
    assignee: "Ответственный",
    priority: "Приоритет",
    type: "Тип",
    client: "Клиент",
    due: "Срок",
    estimate: "Оценка",
    spent: "Факт",
    checklist: "Чеклист",
  }[column] || column;
}
