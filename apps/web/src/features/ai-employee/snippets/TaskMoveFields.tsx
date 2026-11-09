import { textValue } from "../aiEmployeeUtils";

export function TaskMoveFields({ display, inputJSON }: { display: Record<string, unknown>; inputJSON: Record<string, unknown> }) {
  return (
    <div className="ai-proposal-fields" data-testid="ai-proposal-human-card">
      <label>
        <span>Задача</span>
        <strong>{textValue(display.task_key) ? `${textValue(display.task_key)} · ${textValue(display.title)}` : textValue(display.title) || textValue(inputJSON.title)}</strong>
      </label>
      <label>
        <span>Действие</span>
        <strong>{textValue(display.human_action_label) || "Стартовать задачу"}</strong>
      </label>
      <label>
        <span>Новый статус</span>
        <strong>{textValue(display.to_status_name) || textValue(inputJSON.to_status_name) || "В работе"}</strong>
      </label>
      <label>
        <span>Комментарий</span>
        <strong>{textValue(display.comment) || textValue(inputJSON.comment) || "Без комментария"}</strong>
      </label>
    </div>
  );
}
