"use client";

import { Globe } from "lucide-react";

import { CollapsibleRawJson } from "../CollapsibleRawJson";
import { CriterionChecklist } from "../CriterionChecklist";
import { ForwardedKeysTable } from "../ForwardedKeysTable";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";
import { normalizeActionOutput } from "../parseOutputs";

export function HttpRequestPanel({ action, omitExposeSummary }: IntegrationPanelCtx) {
  const out = normalizeActionOutput(action.output_json);

  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2 text-muted-foreground">
        <Globe size={18} />
        <span className="text-sm font-medium">HTTP запрос</span>
      </div>
      {out.criteriaResults.length === 0 && (!out.expose || Object.keys(out.expose).length === 0) ? (
        <p className="muted text-sm">
          Для мокового действия возможен только упрощённый вывод. После перевода запросов на живой клиент здесь будут код ответа и критерии.
        </p>
      ) : null}
      <CriterionChecklist results={out.criteriaResults} />
      {omitExposeSummary ? null : <ForwardedKeysTable expose={out.expose} />}
      <CollapsibleRawJson title="Технические: provider_response_json" data={action.provider_response_json} />
      <CollapsibleRawJson title="Технические: output_json" data={action.output_json} />
    </div>
  );
}
