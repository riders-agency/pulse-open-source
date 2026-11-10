"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { Edge } from "@xyflow/react";

import type { IntegrationDocSearchResult, Project, Workflow } from "../../../lib/api";
import { searchIntegrationDocs } from "../api/workflowEditorApi";
import { scenarioExportObject } from "./scenarioExport";
import type { ActionNodeType } from "./types";
import { workflowAIPrompt } from "./editorDraft";

export function useWorkflowEditorDocs({
  edges,
  nodes,
  project,
  setError,
  setNotice,
  templateVariables,
  testInputText,
  token,
  workflow,
}: {
  edges: Edge[];
  nodes: ActionNodeType[];
  project: Project | null;
  setError: (value: string) => void;
  setNotice: (value: string) => void;
  templateVariables: string[];
  testInputText: string;
  token: string;
  workflow: Workflow | null;
}) {
  const [scenarioImportText, setScenarioImportText] = useState("");
  const [aiTaskText, setAiTaskText] = useState("");
  const [integrationDocResults, setIntegrationDocResults] = useState<IntegrationDocSearchResult[]>([]);
  const [selectedIntegrationDocKeys, setSelectedIntegrationDocKeys] = useState<string[]>([]);
  const [integrationDocsBusy, setIntegrationDocsBusy] = useState(false);

  const selectedIntegrationDocs = useMemo(
    () => integrationDocResults.filter((doc) => selectedIntegrationDocKeys.includes(doc.doc_key)),
    [integrationDocResults, selectedIntegrationDocKeys],
  );
  const exportedScenario = useMemo(
    () => scenarioExportObject({ workflow, project, nodes, edges, testInputText, templateVariables, relevantDocs: selectedIntegrationDocs }),
    [edges, nodes, project, selectedIntegrationDocs, templateVariables, testInputText, workflow],
  );
  const exportedScenarioJSON = useMemo(() => JSON.stringify(exportedScenario, null, 2), [exportedScenario]);
  const exportedAIPrompt = useMemo(
    () => workflowAIPrompt({ task: aiTaskText, contextJSON: exportedScenarioJSON, docs: selectedIntegrationDocs }),
    [aiTaskText, exportedScenarioJSON, selectedIntegrationDocs],
  );

  const copyText = useCallback(async (value: string, label: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setNotice(`${label} скопирован.`);
      setError("");
    } catch {
      setError("Не удалось скопировать в буфер обмена");
    }
  }, [setError, setNotice]);

  useEffect(() => {
    if (!token) return;
    const timer = window.setTimeout(() => {
      setIntegrationDocsBusy(true);
      searchIntegrationDocs(token, aiTaskText, project?.project_id)
        .then((rows) => {
          setIntegrationDocResults(rows);
          setSelectedIntegrationDocKeys((current) => {
            const available = new Set(rows.map((doc) => doc.doc_key));
            const kept = current.filter((key) => available.has(key));
            const next = [...kept];
            for (const doc of rows) {
              if (next.length >= 5) break;
              if (!next.includes(doc.doc_key)) next.push(doc.doc_key);
            }
            return next;
          });
        })
        .catch(() => setIntegrationDocResults([]))
        .finally(() => setIntegrationDocsBusy(false));
    }, 350);
    return () => window.clearTimeout(timer);
  }, [aiTaskText, project?.project_id, token]);

  return {
    aiTaskText,
    copyText,
    exportedAIPrompt,
    exportedScenarioJSON,
    integrationDocResults,
    integrationDocsBusy,
    scenarioImportText,
    selectedIntegrationDocKeys,
    selectedIntegrationDocs,
    setAiTaskText,
    setScenarioImportText,
    setSelectedIntegrationDocKeys,
  };
}
