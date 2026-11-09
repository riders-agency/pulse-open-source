import { useEffect, useEffectEvent } from "react";
import { AI_EMPLOYEE_TEST_IDS } from "../aiEmployeeConstants";

export function useAIEmployeeFocus(open: boolean, close: () => void) {
  const onClose = useEffectEvent(close);
  useEffect(() => {
    if (!open) return;
    const previous = document.activeElement as HTMLElement | null;
    document.querySelector<HTMLTextAreaElement>(`[data-testid="${AI_EMPLOYEE_TEST_IDS.composerInput}"]`)?.focus({ preventScroll: true });
    const onKey = (event: KeyboardEvent) => { if (event.key === "Escape") onClose(); };
    window.addEventListener("keydown", onKey);
    return () => { window.removeEventListener("keydown", onKey); previous?.focus({ preventScroll: true }); };
  }, [open]);
}
