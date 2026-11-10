import { Pause, Play, Copy, Trash2 } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import { Sheet, SheetDescription, SheetHeader, SheetTitle } from "../../../components/ui/sheet";
import type { Integration, Project, ProviderCatalogEntry } from "../../../lib/api";
import { supportsRetryPolicy } from "../model/actionRules";
import { actionLabel } from "../model/editorDraft";
import type { ActionConfig, ActionNodeType, RetryPolicy } from "../model/types";
import { ActionSettings } from "./action-settings";
import {
  RetrySettings,
  SuccessCriteriaSettings,
  TestRunPanel,
} from "./panels";

type EditorSettingsSheetProps = {
  selectedNode: ActionNodeType | null;
  selectedSubtitle: string;
  project: Project | null;
  token: string;
  integrations: Integration[];
  webhookSlug: string;
  templateVariables: string[];
  testInputText: string;
  stepsOverrideText: string;
  testBusy: "full" | "through_action" | "selected_only" | "";
  providerCatalog: ProviderCatalogEntry[];
  duplicateActionNode: (nodeId: string) => void;
  deleteSelectedNode: () => void;
  updateSelectedLabel: (label: string) => void;
  updateSelectedConfig: (patch: Partial<ActionConfig>) => void;
  updateRetry: (patch: Partial<RetryPolicy>) => void;
  setTestInputText: (value: string) => void;
  setStepsOverrideText: (value: string) => void;
  runTest: (mode: "full" | "through_action" | "selected_only") => void;
};

export function EditorSettingsSheet({
  selectedNode,
  selectedSubtitle,
  project,
  token,
  integrations,
  webhookSlug,
  templateVariables,
  testInputText,
  stepsOverrideText,
  testBusy,
  providerCatalog,
  duplicateActionNode,
  deleteSelectedNode,
  updateSelectedLabel,
  updateSelectedConfig,
  updateRetry,
  setTestInputText,
  setStepsOverrideText,
  runTest,
}: EditorSettingsSheetProps) {
  return (
    <Sheet open>
      <SheetHeader>
        <div className="between gap-2">
          <div>
            <SheetTitle>{selectedNode?.data.label ?? "Действие"}</SheetTitle>
            <SheetDescription>{selectedSubtitle}</SheetDescription>
          </div>
          {selectedNode ? (
            <div className="row">
              <Button
                variant="outline"
                size="sm"
                title="Дублировать действие"
                onClick={() => duplicateActionNode(selectedNode.id)}
              >
                <Copy size={15} />
                Копия
              </Button>
              <Button
                variant="destructive"
                size="sm"
                title="Удалить действие из черновика"
                onClick={deleteSelectedNode}
              >
                <Trash2 size={15} />
                Удалить
              </Button>
            </div>
          ) : null}
        </div>
        {selectedNode ? (
          <div className="flow-sheet-note muted">
            Удаление меняет только черновик. Уже выполненные запуски и их действия остаются в истории.
          </div>
        ) : null}
      </SheetHeader>
      {selectedNode ? (
        <div className="flow-settings-stack stack">
          <Card>
            <CardHeader>
              <CardTitle>Действие</CardTitle>
              <CardDescription>{actionLabel(selectedNode.data.action_type_key)}</CardDescription>
            </CardHeader>
            <CardContent className="stack">
              {!(selectedNode.data.action_type_key.endsWith(".trigger") || selectedNode.data.action_type_key.endsWith("_trigger")) ? <div className="stack">
                <Button variant={selectedNode.data.config.disabled ? "secondary" : "outline"} aria-pressed={!selectedNode.data.config.disabled} onClick={() => updateSelectedConfig({ disabled: !selectedNode.data.config.disabled })}>
                  {selectedNode.data.config.disabled ? <Play size={15} /> : <Pause size={15} />}
                  {selectedNode.data.config.disabled ? "Включить действие" : "Отключить действие"}
                </Button>
                <p className="muted text-xs">{selectedNode.data.config.disabled ? "Действие будет пропущено. Настройки сохранены. Зависимые шаги не получат результат этого действия." : "Включённые действия выполняются при следующем запуске опубликованной версии."}</p>
              </div> : null}
              <Field label="Название действия">
                <Input
                  value={selectedNode.data.label}
                  onChange={(event) => updateSelectedLabel(event.target.value)}
                  placeholder={actionLabel(selectedNode.data.action_type_key)}
                />
              </Field>
              <div className="muted text-xs">{selectedSubtitle}</div>
            </CardContent>
          </Card>
          <ActionSettings
            selectedNode={selectedNode}
            project={project}
            token={token}
            integrations={integrations}
            webhookSlug={webhookSlug}
            templateVariables={templateVariables}
            updateSelectedConfig={updateSelectedConfig}
          />
          <TestRunPanel
            selectedNode={selectedNode}
            testInputText={testInputText}
            stepsOverrideText={stepsOverrideText}
            busy={testBusy}
            setTestInputText={setTestInputText}
            setStepsOverrideText={setStepsOverrideText}
            runTest={runTest}
          />
          {supportsRetryPolicy(selectedNode.data.action_type_key) ? (
            <RetrySettings config={selectedNode.data.config} updateRetry={updateRetry} />
          ) : null}
          <SuccessCriteriaSettings
            actionType={selectedNode.data.action_type_key}
            providerCatalog={providerCatalog}
          />
        </div>
      ) : (
        <div className="flow-empty-state muted rounded-md border border-border p-3 text-sm">
          Сценарий пустой. Добавьте действие слева или перетащите его на поле.
        </div>
      )}
    </Sheet>
  );
}
