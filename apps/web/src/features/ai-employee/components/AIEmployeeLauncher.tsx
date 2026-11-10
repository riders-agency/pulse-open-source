import { Bot } from "lucide-react";

import { AI_EMPLOYEE_TEST_IDS } from "../aiEmployeeConstants";

export function AIEmployeeLauncher({
  disabled,
  open,
  onToggle,
}: {
  disabled: boolean;
  open: boolean;
  onToggle: () => void;
}) {
  return (
    <div className="ai-employee-controls">
      <button
        type="button"
        className="ai-employee-bubble"
        disabled={disabled}
        onClick={onToggle}
        title={disabled ? "Выберите проект" : "ИИ сотрудник проекта"}
        aria-label="ИИ сотрудник"
        data-testid={AI_EMPLOYEE_TEST_IDS.launcher}
        aria-expanded={open}
      >
        <Bot size={18} />
        <span>ИИ сотрудник</span>
      </button>
    </div>
  );
}
