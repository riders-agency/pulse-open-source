"use client";

import { Send } from "lucide-react";

import type { ActionRun } from "../../../lib/api";
import { CollapsibleRawJson } from "../CollapsibleRawJson";
import { CriterionChecklist } from "../CriterionChecklist";
import { ForwardedKeysTable } from "../ForwardedKeysTable";
import { normalizeActionOutput, normalizeProvider } from "../parseOutputs";

export type IntegrationPanelCtx = { action: ActionRun; omitExposeSummary?: boolean };

export function GenericIntegrationPanel({ action, omitExposeSummary }: IntegrationPanelCtx) {
  const out = normalizeActionOutput(action.output_json);
  const provider = normalizeProvider(action.provider_response_json);

  const statusHttp = typeof provider?.status_code === "number" ? provider.status_code : undefined;

  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2 text-muted-foreground">
        <Send size={18} />
        <span className="text-sm font-medium">{action.action_type_key}</span>
        {typeof statusHttp === "number" ? <span className="text-xs text-muted-foreground">HTTP {statusHttp}</span> : null}
      </div>

      <CriterionChecklist results={out.criteriaResults} />
      {omitExposeSummary ? null : <ForwardedKeysTable expose={out.expose} />}

      <CollapsibleRawJson title="Технические: output_json" data={action.output_json} />
      <CollapsibleRawJson title="Технические: provider_response_json" data={action.provider_response_json} />
      <CollapsibleRawJson title="Технические: input_json" data={action.input_json} />
    </div>
  );
}
