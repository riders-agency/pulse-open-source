export function AIEmployeeFooter({
  disabled,
  quickActions,
  sending,
  onQuickAction,
}: {
  disabled?: boolean;
  quickActions: string[];
  sending: boolean;
  onQuickAction: (label: string) => void;
}) {
  return (
    <div className="ai-command-row" aria-label="Быстрые действия">
      {quickActions.slice(0, 3).map((label) => (
        <button key={label} type="button" onClick={() => onQuickAction(label)} disabled={disabled || sending}>
          {label}
        </button>
      ))}
    </div>
  );
}
