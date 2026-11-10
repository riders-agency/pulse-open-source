"use client";

import type { Dispatch, SetStateAction } from "react";

import { Button } from "../../components/ui/button";
import { Field, Input, Textarea } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import type { WorkStatus } from "../../lib/api";
import { dateInputProps } from "../../lib/dates";
import { WorkClientPicker } from "./work-client-picker";
import { taskTypes, type WorkTaskDraft } from "./work-task-types";
import { priorityLabel, statusLabel, taskTypeLabel } from "./work-utils";

export function WorkTaskCreateForm({
  companyId,
  createTask,
  draft,
  setDraft,
  setError,
  statuses,
}: {
  companyId?: string;
  createTask: () => Promise<void>;
  draft: WorkTaskDraft;
  setDraft: Dispatch<SetStateAction<WorkTaskDraft>>;
  setError: (value: string) => void;
  statuses: WorkStatus[];
}) {
  return (
    <form
      className="work-create-panel"
      onSubmit={(event) => {
        event.preventDefault();
        void createTask().catch((err) => setError(err.message));
      }}
    >
      <div className="work-create-grid">
        <Field label="Тип">
          <Select value={draft.type} onChange={(event) => setDraft({ ...draft, type: event.target.value })}>
            {taskTypes.map((item) => <option key={item} value={item}>{taskTypeLabel(item)}</option>)}
          </Select>
        </Field>
        <Field label="Статус">
          <Select value={draft.status_id} onChange={(event) => setDraft({ ...draft, status_id: event.target.value })}>
            {statuses.map((status) => <option key={status.task_workflow_status_id} value={status.task_workflow_status_id}>{statusLabel(status.name)}</option>)}
          </Select>
        </Field>
        <Field label="Приоритет">
          <Select value={draft.priority} onChange={(event) => setDraft({ ...draft, priority: event.target.value })}>
            {["none", "low", "medium", "high", "urgent"].map((item) => <option key={item} value={item}>{priorityLabel(item)}</option>)}
          </Select>
        </Field>
        <WorkClientPicker companyId={companyId} value={draft.client_id} onChange={(clientId) => setDraft({ ...draft, client_id: clientId })} />
        <Field label="Срок">
          <Input {...dateInputProps} value={draft.due_date} onChange={(event) => setDraft({ ...draft, due_date: event.target.value })} />
        </Field>
        <Field label="Оценка, минут">
          <Input type="number" value={draft.estimate_minutes} onChange={(event) => setDraft({ ...draft, estimate_minutes: event.target.value })} />
        </Field>
        <Field label="Название" className="span-2">
          <Input value={draft.title} onChange={(event) => setDraft({ ...draft, title: event.target.value })} />
        </Field>
        <Field label="Описание" className="span-2">
          <Textarea value={draft.description} onChange={(event) => setDraft({ ...draft, description: event.target.value })} />
        </Field>
      </div>
      <div className="form-actions">
        <Button type="submit">Создать и показать в задачах</Button>
      </div>
    </form>
  );
}
