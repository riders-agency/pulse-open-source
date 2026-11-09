"use client";

import { useEffect, useState, type Dispatch, type SetStateAction } from "react";

import { Button } from "../../components/ui/button";
import { Field, Input, Textarea } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import type { WorkTaskDetail } from "../../lib/api";
import { dateInputProps, formatDate, formatDateTime, toISODate } from "../../lib/dates";
import { WorkClientPicker } from "./work-client-picker";
import { minutesLabel, priorityLabel, statusLabel, taskTypeLabel } from "./work-utils";
import { taskPriorities, taskTypes } from "./work-task-types";

export function WorkTaskCardTab({
  activeCompanyId,
  addChecklist,
  canWrite,
  checklistTitle,
  patch,
  progress,
  setChecklistTitle,
  setError,
  setTask,
  task,
  toggleChecklist,
  updateCustomField,
}: {
  activeCompanyId: string | null;
  addChecklist: () => Promise<void>;
  canWrite: boolean;
  checklistTitle: string;
  patch: (body: Record<string, unknown>) => Promise<void>;
  progress: string;
  setChecklistTitle: (value: string) => void;
  setError: (value: string) => void;
  setTask: Dispatch<SetStateAction<WorkTaskDetail | undefined>>;
  task: WorkTaskDetail;
  toggleChecklist: (itemId: string, next: boolean) => Promise<void>;
  updateCustomField: (fieldId: string, value: unknown) => Promise<void>;
}) {
  return (
    <section className="work-drawer-grid">
      <div className="work-drawer-main">
        <section className="work-drawer-section">
          <header><strong>Основное</strong></header>
          <Field label="Название">
            <Input disabled={!canWrite} value={task.title} onChange={(event) => setTask({ ...task, title: event.target.value })} onBlur={() => canWrite && void patch({ title: task.title })} />
          </Field>
          <Field label="Описание">
            <Textarea disabled={!canWrite} value={task.description || ""} onChange={(event) => setTask({ ...task, description: event.target.value })} onBlur={() => canWrite && void patch({ description: task.description || "" })} />
          </Field>
        </section>
        <section className="work-drawer-section">
          <header><strong>Чеклист</strong><span>{progress}</span></header>
          <div className="work-checklist">
            {(task.checklist || []).map((item) => (
              <label className="work-checklist-item" key={item.task_checklist_item_id}>
                <input disabled={!canWrite} checked={item.is_done} type="checkbox" onChange={(event) => void toggleChecklist(item.task_checklist_item_id, event.target.checked)} />
                <span>{item.title}</span>
              </label>
            ))}
            {canWrite ? (
              <form
                className="work-inline-create"
                onSubmit={(event) => {
                  event.preventDefault();
                  void addChecklist().catch((err) => setError(err.message));
                }}
              >
                <Input value={checklistTitle} onChange={(event) => setChecklistTitle(event.target.value)} placeholder="Добавить пункт" />
                <Button size="sm" type="submit">ОК</Button>
              </form>
            ) : null}
          </div>
        </section>
      </div>
      <aside className="work-properties-panel">
        <Property label="Статус" value={statusLabel(task.status_name)} />
        <Field label="Тип">
          <Select disabled={!canWrite} value={task.type} onChange={(event) => void patch({ type: event.target.value })}>
            {taskTypes.map((item) => <option key={item} value={item}>{taskTypeLabel(item)}</option>)}
          </Select>
        </Field>
        <Field label="Приоритет">
          <Select disabled={!canWrite} value={task.priority} onChange={(event) => void patch({ priority: event.target.value })}>
            {taskPriorities.map((item) => <option key={item} value={item}>{priorityLabel(item)}</option>)}
          </Select>
        </Field>
        <Field label="Старт">
          <TaskDateInput disabled={!canWrite} value={task.start_date} onCommit={(value) => void patch({ start_date: value })} />
        </Field>
        <Field label="Срок">
          <TaskDateInput disabled={!canWrite} value={task.due_date} onCommit={(value) => void patch({ due_date: value })} />
        </Field>
        <Field label="Оценка, минут">
          <Input disabled={!canWrite} type="number" value={task.estimate_minutes || 0} onChange={(event) => void patch({ estimate_minutes: Number(event.target.value) || 0 })} />
        </Field>
        <Property label="Ответственный" value={task.assignee_name || "Не назначен"} />
        <WorkClientPicker
          companyId={activeCompanyId}
          currentName={task.client_name}
          disabled={!canWrite}
          value={task.client_id || ""}
          onChange={(clientId) => void patch({ client_id: clientId })}
        />
        <Property label="Время" value={`${minutesLabel(task.spent_minutes_cached)} / ${minutesLabel(task.estimate_minutes)}`} />
        <Property label="Создано" value={formatDateTime(task.created_at)} />
        {(task.custom_fields || []).map((field) => (
          <Field key={field.task_custom_field_id} label={field.name}>
            {field.type === "select" ? (
              <Select disabled={!canWrite} value={String(field.value || "")} onChange={(event) => void updateCustomField(field.task_custom_field_id, event.target.value)}>
                <option value="">Не заполнено</option>
                {(field.options || []).map((option) => <option key={option} value={option}>{option}</option>)}
              </Select>
            ) : (
              <Input disabled={!canWrite} defaultValue={String(field.value || "")} onBlur={(event) => void updateCustomField(field.task_custom_field_id, event.target.value)} />
            )}
          </Field>
        ))}
      </aside>
    </section>
  );
}

function Property({ label, value }: { label: string; value: string }) {
  return (
    <div className="work-property">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function TaskDateInput({ disabled, value, onCommit }: { disabled: boolean; value?: string; onCommit: (value: string | null) => void }) {
  const [draft, setDraft] = useState(formatDate(value, ""));

  useEffect(() => {
    setDraft(formatDate(value, ""));
  }, [value]);

  return (
    <Input
      {...dateInputProps}
      disabled={disabled}
      value={draft}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={() => onCommit(toISODate(draft) || null)}
    />
  );
}
