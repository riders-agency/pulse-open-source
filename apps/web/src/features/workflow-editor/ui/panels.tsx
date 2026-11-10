import { Play } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Textarea } from "../../../components/ui/input";
import type { WorkflowVersion } from "../../../lib/api";
import { formatDate } from "../../../lib/dates";
import type { ActionNodeType } from "../model/types";

export { IntegrationSettings } from "./integration-settings";
export { UonLeadSettings } from "./uon-lead-settings";
export { YClientsCertificateSettings } from "./yclients-certificate-settings";
export { FioNormalizeSettings } from "./fio-normalize-settings";
export { PhoneNormalizeSettings } from "./phone-normalize-settings";
export {
  APINETNormalizeSettings,
  BotStateSettings,
  ClientPaymentCycleCollectSettings,
  ClientPaymentDueSettings,
  HttpRequestSettings,
  RetrySettings,
  ScheduleSettings,
} from "./runtime-action-settings";
export { SuccessCriteriaSettings } from "./success-criteria-settings";
export { TelegramRelaySettings } from "./telegram-relay-settings";
export { WebhookSettings } from "./webhook-settings";
export { ScenarioExportPanel } from "./scenario-export-panel";

export function ChangeDescriptionControl({
  value,
  open,
  setValue,
  setOpen,
}: {
  value: string;
  open: boolean;
  setValue: (value: string) => void;
  setOpen: (value: boolean) => void;
}) {
  return (
    <div className={`version-description-control ${open ? "version-description-open" : ""}`}>
      {open ? (
        <Textarea
          value={value}
          onChange={(event) => setValue(event.target.value)}
          onBlur={() => {
            if (!value.trim()) setOpen(false);
          }}
          placeholder="Описание изменений для версии. Можно пустым."
        />
      ) : (
        <button className="version-description-toggle" type="button" onClick={() => setOpen(true)}>
          {value.trim() || "Описание версии"}
        </button>
      )}
    </div>
  );
}

export function VersionHistoryPanel({
  versions,
  busy,
  restoreVersion,
}: {
  versions: WorkflowVersion[];
  busy: boolean;
  restoreVersion: (version: WorkflowVersion) => void;
}) {
  return (
    <section className="palette-section version-history">
      <h3>Версии</h3>
      {versions.length ? (
        <div className="version-list">
          {versions.map((version) => (
            <div className="version-row" key={version.workflow_version_id}>
              <div>
                <div className="version-title">
                  v{version.version_no}
                  {version.is_current ? <Badge tone="success">текущая</Badge> : null}
                </div>
                {version.description ? <p>{version.description}</p> : <p className="muted">Без описания</p>}
                <small>{formatDate(version.published_at, "")}</small>
              </div>
              <Button variant="outline" size="sm" disabled={busy || version.is_current} onClick={() => restoreVersion(version)}>
                Откатить
              </Button>
            </div>
          ))}
        </div>
      ) : (
        <p className="muted text-xs">Опубликованных версий пока нет.</p>
      )}
    </section>
  );
}

export function TestRunPanel({
  selectedNode,
  testInputText,
  stepsOverrideText,
  busy,
  setTestInputText,
  setStepsOverrideText,
  runTest,
}: {
  selectedNode: ActionNodeType;
  testInputText: string;
  stepsOverrideText: string;
  busy: "full" | "through_action" | "selected_only" | "";
  setTestInputText: (value: string) => void;
  setStepsOverrideText: (value: string) => void;
  runTest: (mode: "full" | "through_action" | "selected_only") => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Тестовый запуск черновика</CardTitle>
        <CardDescription>Создает тестовую версию из текущего черновика и не меняет опубликованный сценарий.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Тестовые данные JSON">
          <Textarea value={testInputText} onChange={(event) => setTestInputText(event.target.value)} />
        </Field>
        <div className="grid gap-2">
          <Button variant="secondary" disabled={!!busy} onClick={() => runTest("full")}>
            <Play size={15} />
            {busy === "full" ? "Запуск…" : "Весь сценарий"}
          </Button>
          <Button variant="outline" disabled={!!busy || !selectedNode.id} onClick={() => runTest("through_action")}>
            До выбранного действия
          </Button>
        </div>
        <details className="rounded-md border border-border">
          <summary className="cursor-pointer px-3 py-2 text-xs font-medium text-muted-foreground">Дополнительно: только выбранное действие</summary>
          <div className="stack border-t p-3">
            <Field label="Подстановка результатов шагов">
              <Textarea value={stepsOverrideText} onChange={(event) => setStepsOverrideText(event.target.value)} />
            </Field>
            <Button variant="outline" disabled={!!busy || !selectedNode.id} onClick={() => runTest("selected_only")}>
              {busy === "selected_only" ? "Запуск…" : "Только выбранное действие"}
            </Button>
          </div>
        </details>
      </CardContent>
    </Card>
  );
}
