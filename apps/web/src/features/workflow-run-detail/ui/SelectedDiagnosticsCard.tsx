import { Badge } from "../../../components/ui/badge";
import type { ActionRun } from "../../../lib/api";
import { diagnosticTone, diagnosticsForAction } from "../model/workflowRunDiagnostics";

export function SelectedDiagnosticsCard({ selected }: { selected: ActionRun }) {
  const selectedDiagnostics = diagnosticsForAction(selected);
  if (!selectedDiagnostics) return null;
  return (
    <div className="stack rounded-md border border-border bg-muted/35 p-3">
      <div className="between gap-3">
        <strong>{selectedDiagnostics.summary || "Диагностика действия"}</strong>
        {selectedDiagnostics.reason_code ? <Badge tone={diagnosticTone(selected.status, selectedDiagnostics)}>{selectedDiagnostics.reason_code}</Badge> : null}
      </div>
      {selectedDiagnostics.human_reason ? <p className="m-0 text-sm">{selectedDiagnostics.human_reason}</p> : null}
      {selectedDiagnostics.next_action ? <p className="muted m-0 text-sm">Что исправить: {selectedDiagnostics.next_action}</p> : null}
      {(selected.status === "failed" || selected.status === "retrying") && <div className="row flex-wrap">
        {selectedDiagnostics.provider ? <Badge>{selectedDiagnostics.provider}</Badge> : null}
        <Badge tone={selectedDiagnostics.retryable ? "warning" : "muted"}>
          {selectedDiagnostics.retryable ? "можно повторить" : "повтор не поможет без исправления"}
        </Badge>
      </div>}
    </div>
  );
}
