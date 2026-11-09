import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Table } from "../../../components/ui/table";
import type { ActionAttempt, ActionRun } from "../../../lib/api";
import { jsonText } from "../../../lib/api";
import { formatDate } from "../../../lib/utils";
import { CollapsibleRawJson } from "../../run-integration-panels/CollapsibleRawJson";
import { CriterionChecklist } from "../../run-integration-panels/CriterionChecklist";
import { diagnosticsForActionRun } from "../../run-integration-panels/diagnostics/resolve";
import { ForwardedKeysTable } from "../../run-integration-panels/ForwardedKeysTable";
import { normalizeActionOutput, normalizeDiagnostics } from "../../run-integration-panels/parseOutputs";
import { ResolvedIntegrationPanel } from "../../run-integration-panels/registry";
import { diagnosticCriteria, diagnosticTone, diagnosticsForAction } from "../model/workflowRunDiagnostics";
import { actionTypeLabel, statusLabel } from "../model/workflowRunLabels";

export function RunActionDetailPanel({
  latestAttempt,
  latestAttemptDiagnostics,
  replayAllBusy,
  replayOneAction,
  retryActionBusy,
  selected,
}: {
  latestAttempt?: ActionAttempt;
  latestAttemptDiagnostics: ReturnType<typeof normalizeDiagnostics> | undefined;
  replayAllBusy: boolean;
  replayOneAction: () => Promise<void>;
  retryActionBusy: boolean;
  selected?: ActionRun;
}) {
  if (!selected) return <div className="empty-state">Действий пока нет</div>;
  const selectedNorm = normalizeActionOutput(selected.output_json);
  const selectedDiagnostics = diagnosticsForAction(selected);
  const selectedCriteria = diagnosticCriteria(selected);
  const hasExpose = !!selectedNorm.expose && typeof selectedNorm.expose === "object" && Object.keys(selectedNorm.expose).length > 0;

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{actionTypeLabel(selected.action_type_key)}</CardTitle>
          <CardDescription>Что произошло, почему и что делать дальше.</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          <div className="between"><span>Статус</span><Badge tone={diagnosticTone(selected.status, selectedDiagnostics)}>{statusLabel(selected.status)}</Badge></div>
          <div className="between"><span>Попытки</span><span>{selected.attempt_count ?? 0}/{selected.max_attempts ?? 10}</span></div>
          <div className="between"><span>Начало</span><span>{formatDate(selected.started_at)}</span></div>
          {selectedDiagnostics ? <SelectedDiagnosticsCard selected={selected} /> : null}
          <Button variant="outline" disabled={!selected.action_key || retryActionBusy || replayAllBusy} onClick={() => replayOneAction()}>
            {retryActionBusy ? "Создание запуска…" : "Повторить только это действие"}
          </Button>
          <p className="muted text-xs">
            Новый запуск с теми же входными данными исполнит только выбранный шаг. Ошибка конфигурации завершит новый запуск с сообщением об ошибке.
          </p>
          {selected.error ? <Badge tone="danger">{selected.error}</Badge> : null}
        </CardContent>
      </Card>
      {hasExpose ? <ForwardedOutputCard expose={selectedNorm.expose} /> : null}
      <Card>
        <CardHeader>
          <CardTitle>Критерии успеха</CardTitle>
          <CardDescription>Что именно проверял раннер для этого действия.</CardDescription>
        </CardHeader>
        <CardContent>
          {selectedCriteria.length > 0 ? <CriterionChecklist results={selectedCriteria} /> : <div className="muted text-sm">Для этого действия критерии не заданы.</div>}
        </CardContent>
      </Card>
      <ProviderResponseCard selected={selected} hasExpose={hasExpose} />
      <details className="rounded-md border border-border">
        <summary className="cursor-pointer px-3 py-2 text-xs font-medium text-muted-foreground">
          Отладочные JSON блоки: вход, результат, провайдер
        </summary>
        <div className="stack border-t p-3">
          <pre className="code-block">{jsonText(selected.input_json)}</pre>
          <pre className="code-block">{jsonText(selected.output_json)}</pre>
          <pre className="code-block">{jsonText(selected.provider_response_json)}</pre>
        </div>
      </details>
      <AttemptsCard latestAttempt={latestAttempt} latestAttemptDiagnostics={latestAttemptDiagnostics} selected={selected} />
    </>
  );
}

function SelectedDiagnosticsCard({ selected }: { selected: ActionRun }) {
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
      <div className="row flex-wrap">
        {selectedDiagnostics.provider ? <Badge>{selectedDiagnostics.provider}</Badge> : null}
        <Badge tone={selectedDiagnostics.retryable ? "warning" : "muted"}>
          {selectedDiagnostics.retryable ? "можно повторить" : "повтор не поможет без исправления"}
        </Badge>
      </div>
    </div>
  );
}

function ForwardedOutputCard({ expose }: { expose: Record<string, unknown> }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Выход для следующего шага</CardTitle>
        <CardDescription>Плоские ключи из <code>expose</code> для связки действий.</CardDescription>
      </CardHeader>
      <CardContent>
        <ForwardedKeysTable expose={expose} compact />
      </CardContent>
    </Card>
  );
}

function ProviderResponseCard({ selected, hasExpose }: { selected: ActionRun; hasExpose: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Ответ провайдера и интерпретация</CardTitle>
        <CardDescription>Специальная панель Telegram / MAX / SMTP плюс сырой redacted ответ.</CardDescription>
      </CardHeader>
      <CardContent>
        {selected.action_type_key ? <ResolvedIntegrationPanel action={selected} omitExposeSummary={hasExpose} /> : <div className="muted text-sm">Нет типа действия</div>}
      </CardContent>
    </Card>
  );
}

function AttemptsCard({ latestAttempt, latestAttemptDiagnostics, selected }: { latestAttempt?: ActionAttempt; latestAttemptDiagnostics: ReturnType<typeof normalizeDiagnostics> | undefined; selected: ActionRun }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Попытки</CardTitle>
        <CardDescription>История выполнения действия, тип ошибки и интерпретация последней попытки.</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <thead><tr><th>#</th><th>Статус</th><th>Длительность</th><th>Причина</th><th>Ошибка</th></tr></thead>
          <tbody>
            {(selected.attempts ?? []).map((attempt) => <AttemptRow attempt={attempt} key={attempt.action_attempt_id || attempt.attempt_no} selected={selected} />)}
            {(selected.attempts ?? []).length === 0 ? <tr><td colSpan={5}>Попыток пока нет</td></tr> : null}
          </tbody>
        </Table>
        {latestAttempt ? (
          <div className="stack mt-4">
            {latestAttemptDiagnostics?.human_reason ? (
              <div className="rounded-md border border-border bg-muted/35 p-3 text-sm">Последняя попытка: {latestAttemptDiagnostics.human_reason}</div>
            ) : null}
            <CollapsibleRawJson title="Сырой ответ последней попытки" data={latestAttempt.provider_response_json} defaultOpen={selected.status === "failed"} />
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}

function AttemptRow({ attempt, selected }: { attempt: ActionAttempt; selected: ActionRun }) {
  const d = normalizeDiagnostics(attempt.diagnostics) ?? normalizeActionOutput(attempt.output_json).diagnostics ?? diagnosticsForActionRun({
    ...selected,
    status: attempt.status,
    error: attempt.error,
    output_json: attempt.output_json,
    provider_response_json: attempt.provider_response_json,
  });
  return (
    <tr>
      <td>{attempt.attempt_no}</td>
      <td><Badge tone={diagnosticTone(attempt.status, d)}>{attempt.status}</Badge></td>
      <td>{attempt.duration_ms ?? 0} мс</td>
      <td>{d?.reason_code || "—"}</td>
      <td>{attempt.error || "—"}</td>
    </tr>
  );
}
