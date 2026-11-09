"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Plus, Search, SlidersHorizontal } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import type { WorkAccess, WorkBoard, WorkStatus, WorkTaskList } from "../../lib/api";
import { toISODate } from "../../lib/dates";
import { usePulseStore } from "../../store/pulse-store";
import { restoreWorkTask, createWorkTask, getWorkBoard, listWorkBoards, listWorkTasks } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { WorkTaskCreateForm } from "./work-task-create-form";
import { WorkTaskTable, columnLabel } from "./work-task-table";
import { emptyWorkTaskDraft, type WorkTaskDraft } from "./work-task-types";

const allColumns = ["title", "status", "assignee", "priority", "type", "client", "due", "estimate", "spent", "checklist"];

export function WorkTasksPage() {
  const token = usePulseStore((s) => s.token);
  const companyId = usePulseStore((s) => s.activeCompanyId);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [list, setList] = useState<WorkTaskList>();
  const [boards, setBoards] = useState<WorkBoard[]>([]);
  const [columns, setColumns] = useState<string[]>(allColumns);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState("all");
  const [workflowFilter, setWorkflowFilter] = useState("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [draft, setDraft] = useState<WorkTaskDraft>(emptyWorkTaskDraft());
  const [error, setError] = useState("");

  useEffect(() => {
    const saved = window.localStorage.getItem("work-table-columns");
    try { if (saved) { const parsed: unknown = JSON.parse(saved); if (Array.isArray(parsed)) setColumns(parsed.filter((column: unknown): column is string => typeof column === "string" && allColumns.includes(column))); } } catch { /* Fall back to default columns when browser storage is corrupt. */ }
  }, []);

  const load = useCallback(async () => {
    if (!token || !projectId) return;
    const [tasks, boardPayload] = await Promise.all([
      listWorkTasks(projectId, token, {
        limit: 120,
        archived: filter === "archived",
        search,
        assignee: filter === "me" || filter === "unassigned" ? filter : undefined,
        due: filter === "overdue" ? "overdue" : undefined,
        taskWorkflowId: workflowFilter !== "all" ? workflowFilter : undefined,
      }),
      listWorkBoards(projectId, token),
    ]);
    let boardItems = boardPayload.items || [];
    if (boardItems[0]?.task_board_id) {
      const detail = await getWorkBoard(boardItems[0].task_board_id, token);
      boardItems = [detail.board, ...boardItems.slice(1)];
    }
    setAccess(tasks.access);
    setList(tasks.tasks);
    setBoards(boardItems);
    const firstStatus = boardItems[0]?.columns?.[0]?.status.task_workflow_status_id;
    if (!draft.status_id && firstStatus) setDraft((value) => ({ ...value, status_id: firstStatus }));
    setError("");
  }, [companyId, draft.status_id, filter, projectId, search, token, workflowFilter]);

  useEffect(() => {
    void load().catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить задачи"));
  }, [load]);

  useEffect(() => {
    const onAIEntityChanged = () => {
      void load().catch((err) => setError(err instanceof Error ? err.message : "Не удалось обновить задачи"));
    };
    window.addEventListener("pulse:ai-entity-changed", onAIEntityChanged);
    return () => window.removeEventListener("pulse:ai-entity-changed", onAIEntityChanged);
  }, [load]);

  useEffect(() => {
    const clientId = new URLSearchParams(window.location.search).get("client_id") || "";
    if (!clientId) return;
    setCreateOpen(true);
    setDraft((value) => ({ ...value, client_id: clientId }));
  }, []);

  const statuses = useMemo(() => {
    const map = new Map<string, WorkStatus>();
    for (const board of boards) {
      for (const column of board.columns || []) map.set(column.status.task_workflow_status_id, column.status);
    }
    return [...map.values()];
  }, [boards]);

  const workflowOptions = useMemo(() => {
    const map = new Map<string, string>();
    for (const board of boards) {
      if (board.task_workflow_id) map.set(board.task_workflow_id, board.name);
    }
    return [...map.entries()];
  }, [boards]);

  if (!projectId) return <WorkNoProject />;

  const toggle = (column: string) => {
    const next = columns.includes(column) ? columns.filter((item) => item !== column) : [...columns, column];
    setColumns(next);
    window.localStorage.setItem("work-table-columns", JSON.stringify(next));
  };

  const createTask = async () => {
    if (!token || !projectId || !draft.title.trim()) return;
    await createWorkTask(token, {
      project_id: projectId,
      task_workflow_status_id: draft.status_id || statuses[0]?.task_workflow_status_id,
      title: draft.title.trim(),
      description: draft.description,
      type: draft.type,
      priority: draft.priority,
      severity: draft.priority === "urgent" ? "critical" : "medium",
      client_id: draft.client_id || undefined,
      due_date: toISODate(draft.due_date) || undefined,
      estimate_minutes: Number(draft.estimate_minutes) || 0,
      billable: true,
      labels: draft.type === "bug" ? ["bug"] : [],
    });
    setDraft(emptyWorkTaskDraft(statuses[0]?.task_workflow_status_id || ""));
    setCreateOpen(false);
    await load();
  };

  return (
    <WorkFrame
      access={access}
      title="Все задачи"
      description="Единый список задач проекта: фильтры, быстрое создание, настройка колонок и карточка задачи."
      actions={access?.can_write_tasks && !access.is_client_guest ? <Button onClick={() => setCreateOpen(!createOpen)}><Plus size={16} /> Создать задачу</Button> : null}
    >
      {error ? <div className="form-error">{error}</div> : null}
      <section className="work-task-toolbar">
        <label className="work-search"><Search size={16} /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Поиск по названию, описанию и тегам" /></label>
        <Select value={filter} onChange={(event) => setFilter(event.target.value)} aria-label="Фильтр задач">
          <option value="all">Все задачи</option>
          <option value="me">Мои</option>
          <option value="unassigned">Без ответственного</option>
          <option value="overdue">Просроченные</option>
          {!access?.is_client_guest ? <option value="archived">Удалённые</option> : null}
        </Select>
        <Select value={workflowFilter} onChange={(event) => setWorkflowFilter(event.target.value)} aria-label="Фильтр процесса">
          <option value="all">Все процессы</option>
          {workflowOptions.map(([id, label]) => <option key={id} value={id}>{label}</option>)}
        </Select>
        <span>{list?.total || 0} задач</span>
      </section>
      {createOpen ? (
        <WorkTaskCreateForm companyId={companyId} createTask={createTask} draft={draft} setDraft={setDraft} setError={setError} statuses={statuses} />
      ) : null}
      <div className="work-column-picker" aria-label="Колонки таблицы">
        <SlidersHorizontal size={16} />
        {allColumns.map((column) => (
          <label key={column}>
            <input type="checkbox" checked={columns.includes(column)} onChange={() => toggle(column)} />
            <span>{columnLabel(column)}</span>
          </label>
        ))}
      </div>
      {filter === "archived" ? <section className="stack" aria-label="Удалённые задачи">
        {!list?.items.length ? <p className="empty-state">Удалённых задач нет</p> : null}
        {list?.items.map((task) => <div className="between" key={task.task_id}><span>{task.title}</span>{access?.can_write_tasks ? <Button variant="secondary" onClick={() => { if (token) void restoreWorkTask(task.task_id, token).then(load).catch((err) => setError(err.message)); }}>Восстановить</Button> : null}</div>)}
      </section> : <WorkTaskTable columns={columns} list={list} />}
    </WorkFrame>
  );
}
