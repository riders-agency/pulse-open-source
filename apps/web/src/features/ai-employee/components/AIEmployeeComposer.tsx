import { Send, XCircle } from "lucide-react";

import { AI_EMPLOYEE_TEST_IDS } from "../aiEmployeeConstants";

export function AIEmployeeComposer({
  input,
  disabled,
  sending,
  onInputChange,
  onSend,
  onStop,
}: {
  input: string;
  disabled?: boolean;
  sending: boolean;
  onInputChange: (value: string) => void;
  onSend: (value: string) => void;
  onStop: () => void;
}) {
  return (
    <form
      className="ai-composer"
      onSubmit={(event) => {
        event.preventDefault();
        if (!disabled && !sending && input.trim()) onSend(input);
      }}
    >
      <textarea
        aria-label="Сообщение ИИ-сотруднику"
        value={input}
        disabled={disabled || sending}
        onChange={(event) => onInputChange(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing && !sending && !disabled) {
            event.preventDefault();
            onSend(event.currentTarget.value);
          }
        }}
        placeholder={disabled ? "ИИ-сотрудник недоступен" : "Напишите, чем помочь…"}
        rows={2}
        data-testid={AI_EMPLOYEE_TEST_IDS.composerInput}
      />
      {sending ? (
        <button
          type="button"
          className="ai-stop"
          onClick={onStop}
          aria-label="Остановить запрос"
          data-testid={AI_EMPLOYEE_TEST_IDS.composerStop}
        >
          <XCircle size={17} />
          <span>Стоп</span>
        </button>
      ) : (
        <button type="submit" disabled={disabled || !input.trim()} aria-label="Отправить" data-testid={AI_EMPLOYEE_TEST_IDS.composerSend}>
          <Send size={17} />
          <span>Отправить</span>
        </button>
      )}
    </form>
  );
}
