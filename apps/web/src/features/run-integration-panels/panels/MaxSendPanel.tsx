"use client";

import { MessageSquare } from "lucide-react";

import { CollapsibleRawJson } from "../CollapsibleRawJson";
import { CriterionChecklist } from "../CriterionChecklist";
import { ForwardedKeysTable } from "../ForwardedKeysTable";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";
import { normalizeActionOutput, normalizeProvider } from "../parseOutputs";

export function MaxSendPanel({ action, omitExposeSummary }: IntegrationPanelCtx) {
  const out = normalizeActionOutput(action.output_json);
  const provider = normalizeProvider(action.provider_response_json);

  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2">
        <MessageSquare className="text-violet-500" size={20} />
        <div>
          <div className="text-sm font-medium">MAX</div>
          <div className="muted text-xs">{action.action_key}</div>
        </div>
      </div>
      {typeof provider?.status_code === "number" ? <div className="muted text-xs">HTTP {provider.status_code}</div> : null}
      <div className="grid gap-2 rounded-md border border-border bg-muted/35 p-3 text-xs">
        <div className="between"><span>HTTP</span><code>{provider?.status_code != null ? String(provider.status_code) : "—"}</code></div>
        <div className="between"><span>message_id</span><code>{provider?.message_id != null ? String(provider.message_id) : "—"}</code></div>
        <div className="between"><span>Описание</span><code>{provider?.description != null ? String(provider.description) : provider?.error != null ? String(provider.error) : "—"}</code></div>
      </div>
      <CriterionChecklist results={out.criteriaResults} />
      {omitExposeSummary ? null : <ForwardedKeysTable expose={out.expose} />}

      <CollapsibleRawJson title="Сырой ответ MAX" data={action.provider_response_json} />
      <CollapsibleRawJson title="output_json" data={action.output_json} />
    </div>
  );
}
