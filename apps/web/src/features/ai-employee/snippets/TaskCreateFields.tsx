import { CalendarDays } from "lucide-react";

import {
  boolLabel,
  dateInputToISO,
  dateInputValue,
  dateLabel,
  priorityLabel,
  textValue,
} from "../aiEmployeeUtils";

export function TaskCreateFields({
  draft,
  display,
  inputJSON,
  editing,
  updateDraft,
}: {
  draft: Record<string, unknown>;
  display: Record<string, unknown>;
  inputJSON: Record<string, unknown>;
  editing: boolean;
  updateDraft: (patch: Record<string, unknown>) => void;
}) {
  return (
    <div className="ai-proposal-fields" data-testid="ai-proposal-human-card">
      <label>
        <span>Название</span>
        {editing ? <input value={textValue(draft.title)} onChange={(event) => updateDraft({ title: event.target.value })} /> : <strong>{textValue(display.title) || textValue(inputJSON.title) || "Без названия"}</strong>}
      </label>
      <label>
        <span>Клиент</span>
        {editing ? <input value={textValue(draft.client_name) || textValue(draft.client_query)} onChange={(event) => updateDraft({ client_query: event.target.value, client_id: "", client_name: "", client_resolved: false })} /> : <strong>{textValue(display.client_name) || "Не указан"}</strong>}
      </label>
      <label>
        <span>Срок</span>
        {editing ? (
          <input inputMode="numeric" placeholder="дд.мм.гггг" value={dateInputValue(draft.due_date)} onChange={(event) => updateDraft({ due_date: dateInputToISO(event.target.value) })} />
        ) : (
          <strong><CalendarDays size={14} /> {textValue(display.due_date_label) || dateLabel(inputJSON.due_date)}</strong>
        )}
      </label>
      <label>
        <span>Приоритет</span>
        {editing ? (
          <select value={textValue(draft.priority) || "medium"} onChange={(event) => updateDraft({ priority: event.target.value })}>
            <option value="low">Низкий</option>
            <option value="medium">Средний</option>
            <option value="high">Высокий</option>
            <option value="urgent">Срочный</option>
          </select>
        ) : (
          <strong>{priorityLabel(display.priority || inputJSON.priority)}</strong>
        )}
      </label>
      <label>
        <span>Оплата</span>
        {editing ? (
          <select value={draft.billable === true ? "true" : "false"} onChange={(event) => updateDraft({ billable: event.target.value === "true" })}>
            <option value="false">Нет</option>
            <option value="true">Да</option>
          </select>
        ) : (
          <strong>{boolLabel(display.billable ?? inputJSON.billable)}</strong>
        )}
      </label>
      <label>
        <span>Таймер</span>
        {editing ? (
          <select value={(draft.start_timer_after_create ?? inputJSON.start_timer_after_create) === true ? "true" : "false"} onChange={(event) => updateDraft({ start_timer_after_create: event.target.value === "true" })}>
            <option value="false">Не запускать</option>
            <option value="true">Запустить после создания</option>
          </select>
        ) : (
          <strong>{(display.start_timer_after_create ?? inputJSON.start_timer_after_create) === true ? "Запустить после создания" : "Не запускать"}</strong>
        )}
      </label>
    </div>
  );
}
