import type { IntegrationDocSearchResult, WorkflowVersion } from "../../../lib/api";
import type { ActionKind } from "../model/types";
import { iconFor } from "../model/editorDraft";
import { palette } from "../registry/actions";
import { ScenarioExportPanel, VersionHistoryPanel } from "./panels";

type EditorPaletteSidebarProps = {
  exportedScenarioJSON: string;
  exportedAIPrompt: string;
  aiTaskText: string;
  integrationDocResults: IntegrationDocSearchResult[];
  selectedIntegrationDocKeys: string[];
  integrationDocsBusy: boolean;
  scenarioImportText: string;
  versions: WorkflowVersion[];
  versionsBusy: boolean;
  addActionNode: (type: ActionKind) => void;
  setAiTaskText: (value: string) => void;
  setSelectedIntegrationDocKeys: (value: string[] | ((current: string[]) => string[])) => void;
  setScenarioImportText: (value: string) => void;
  applyScenarioImport: () => void;
  copyText: (value: string, label: string) => void;
  restoreVersion: (version: WorkflowVersion) => void;
};

export function EditorPaletteSidebar({
  exportedScenarioJSON,
  exportedAIPrompt,
  aiTaskText,
  integrationDocResults,
  selectedIntegrationDocKeys,
  integrationDocsBusy,
  scenarioImportText,
  versions,
  versionsBusy,
  addActionNode,
  setAiTaskText,
  setSelectedIntegrationDocKeys,
  setScenarioImportText,
  applyScenarioImport,
  copyText,
  restoreVersion,
}: EditorPaletteSidebarProps) {
  return (
    <aside className="flow-palette">
      {palette.map((group) => (
        <section className="palette-section" key={group.group}>
          <h3>{group.group}</h3>
          {group.items.map((item) => (
            <button
              className="palette-item"
              draggable
              key={item.type}
              onClick={() => addActionNode(item.type)}
              onDragStart={(event) => event.dataTransfer.setData("application/rp-action", item.type)}
            >
              {iconFor(item.type)}
              {item.label}
            </button>
          ))}
        </section>
      ))}
      <ScenarioExportPanel
        jsonText={exportedScenarioJSON}
        promptText={exportedAIPrompt}
        aiTaskText={aiTaskText}
        setAiTaskText={setAiTaskText}
        integrationDocs={integrationDocResults}
        selectedDocKeys={selectedIntegrationDocKeys}
        setSelectedDocKeys={setSelectedIntegrationDocKeys}
        docsBusy={integrationDocsBusy}
        importText={scenarioImportText}
        setImportText={setScenarioImportText}
        applyImport={applyScenarioImport}
        copyText={copyText}
      />
      <VersionHistoryPanel versions={versions} busy={versionsBusy} restoreVersion={restoreVersion} />
    </aside>
  );
}
