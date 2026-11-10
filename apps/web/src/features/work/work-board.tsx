"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Plus, Settings2 } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Field } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import type { WorkAccess, WorkBoard, WorkBoardTemplate } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { createWorkBoard, createWorkTask, getWorkBoard, getWorkPresets, listWorkBoards, moveWorkTask, startWorkTaskTimer } from "./api/workApi";
import { boardColumnsClass } from "./model/workStyle";
import { WorkBoardCreateForm } from "./work-board-create-form";
import { boardGroupLabel } from "./work-board-utils";
import { WorkBoardColumn } from "./work-board-column";
import { WorkFrame, WorkNoProject } from "./work-frame";

export function WorkBoardPage() {
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [boards, setBoards] = useState<WorkBoard[]>([]);
  const [boardId, setBoardId] = useState("");
  const [board, setBoard] = useState<WorkBoard>();
  const [presets, setPresets] = useState<WorkBoardTemplate[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [boardName, setBoardName] = useState("");
  const [templateKey, setTemplateKey] = useState("kanban_status");
  const [creatingStatus, setCreatingStatus] = useState("");
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  const loadBoards = useCallback(async () => {
    if (!token || !projectId) return;
    const [list, presetPayload] = await Promise.all([
      listWorkBoards(projectId, token),
      getWorkPresets(token),
    ]);
    setAccess(list.access);
    const items = list.items || [];
    setBoards(items);
    setPresets(presetPayload.board_templates || []);
    const nextBoardID = boardId && items.some((item) => item.task_board_id === boardId) ? boardId : items[0]?.task_board_id || "";
    setBoardId(nextBoardID);
    setError("");
  }, [boardId, projectId, token]);

  const loadBoard = useCallback(async (id: string) => {
    if (!token || !id) {
      setBoard(undefined);
      return;
    }
    const detail = await getWorkBoard(id, token);
    setBoard(detail.board);
    setAccess(detail.access);
  }, [token]);

  useEffect(() => {
    void loadBoards().catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить доски"));
  }, [projectId, token]);

  useEffect(() => {
    void loadBoard(boardId).catch((err) => setError(err instanceof Error ? err.message : "Не удалось открыть доску"));
  }, [boardId, loadBoard]);

  useEffect(() => {
    const onAIEntityChanged = () => {
      void loadBoards().catch((err) => setError(err instanceof Error ? err.message : "Не удалось обновить доски"));
      if (boardId) {
        void loadBoard(boardId).catch((err) => setError(err instanceof Error ? err.message : "Не удалось обновить доску"));
      }
    };
    window.addEventListener("pulse:ai-entity-changed", onAIEntityChanged);
    return () => window.removeEventListener("pulse:ai-entity-changed", onAIEntityChanged);
  }, [boardId, loadBoard, loadBoards]);

  const columns = useMemo(() => board?.columns || [], [board]);
  if (!projectId) return <WorkNoProject />;

  const selectedPreset = presets.find((preset) => preset.key === templateKey);

  const createBoard = async () => {
    if (!token || !projectId) return;
    const created = await createWorkBoard(token, {
      project_id: projectId,
      template_key: templateKey,
      name: boardName.trim() || selectedPreset?.name,
    });
    setCreateOpen(false);
    setBoardName("");
    await loadBoards();
    setBoardId(created.task_board_id);
  };

  const createTask = async (statusId: string) => {
    if (!token || !projectId || !title.trim()) return;
    await createWorkTask(token, {
      project_id: projectId,
      task_workflow_status_id: statusId,
      title: title.trim(),
      type: "task",
      priority: "medium",
      severity: "medium",
      billable: true,
    });
    setTitle("");
    setCreatingStatus("");
    await loadBoard(boardId);
  };

  const moveTask = async (taskId: string, statusId: string, rank: number, expectedRevision?: number) => {
    if (!token) return;
    await moveWorkTask(token, taskId, statusId, rank, expectedRevision);
    await loadBoard(boardId);
  };

  const startTimer = async (taskId: string) => {
    if (!token) return;
    await startWorkTaskTimer(taskId, token);
    await loadBoard(boardId);
  };

  return (
    <WorkFrame
      access={access}
      title="Доски задач"
      description="Настраиваемые представления задач: канбан, исполнители, приоритеты и клиентская приемка."
      actions={access?.can_manage_processes ? <Button variant="secondary" onClick={() => setCreateOpen(!createOpen)}><Plus size={16} /> Создать доску</Button> : null}
    >
      {error ? <div className="form-error">{error}</div> : null}
      <section className="work-board-toolbar">
        <Field label="Доска">
          <Select value={boardId} onChange={(event) => setBoardId(event.target.value)}>
            {boards.map((item) => <option key={item.task_board_id} value={item.task_board_id}>{item.name}</option>)}
          </Select>
        </Field>
        <div className="work-board-meta">
          <Settings2 size={16} />
          <span>Группировка: {boardGroupLabel(board?.group_by)} · дорожки: {boardGroupLabel(board?.swimlane_by)}</span>
        </div>
      </section>
      {createOpen ? (
        <WorkBoardCreateForm
          boardName={boardName}
          createBoard={createBoard}
          presets={presets}
          selectedPreset={selectedPreset}
          setBoardName={setBoardName}
          setError={setError}
          setTemplateKey={setTemplateKey}
          templateKey={templateKey}
        />
      ) : null}
      <section className={`work-board ${boardColumnsClass(columns.length)}`} aria-label="Канбан доска">
        {columns.map((column) => (
          <WorkBoardColumn
            access={access}
            column={column}
            creatingStatus={creatingStatus}
            key={column.status.task_workflow_status_id}
            onCreateTask={createTask}
            onMoveTask={moveTask}
            onOpenTask={(taskId) => router.push(`/tasks/${taskId}`)}
            onStartTimer={startTimer}
            setCreatingStatus={setCreatingStatus}
            setError={setError}
            setTitle={setTitle}
            title={title}
          />
        ))}
      </section>
    </WorkFrame>
  );
}
