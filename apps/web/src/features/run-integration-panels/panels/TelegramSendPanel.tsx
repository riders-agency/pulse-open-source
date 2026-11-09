"use client";

import { MessageCircle } from "lucide-react";

import { CollapsibleRawJson } from "../CollapsibleRawJson";
import { CriterionChecklist } from "../CriterionChecklist";
import { ForwardedKeysTable } from "../ForwardedKeysTable";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";
import { normalizeActionOutput, normalizeProvider } from "../parseOutputs";

export function TelegramSendPanel({ action, omitExposeSummary }: IntegrationPanelCtx) {
  const out = normalizeActionOutput(action.output_json);
  const provider = normalizeProvider(action.provider_response_json);

  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2">
        <MessageCircle className="text-sky-500" size={20} />
        <div>
          <div className="text-sm font-medium">Telegram</div>
          <div className="muted text-xs">{action.action_key}</div>
        </div>
      </div>
      {typeof provider?.ok === "boolean" ? (
        <div className="text-sm">
          API ok:{" "}
          <code>{String(provider.ok)}</code>
          {provider?.description != null ? <span className="muted"> — {String(provider.description)}</span> : null}
        </div>
      ) : null}
      <div className="grid gap-2 rounded-md border border-border bg-muted/35 p-3 text-xs">
        <div className="between"><span>HTTP</span><code>{provider?.status_code != null ? String(provider.status_code) : "—"}</code></div>
        <div className="between"><span>message_id</span><code>{provider?.message_id != null ? String(provider.message_id) : "—"}</code></div>
        <div className="between"><span>retry_after</span><code>{(provider?.parameters as Record<string, unknown> | undefined)?.retry_after != null ? String((provider?.parameters as Record<string, unknown>).retry_after) : "—"}</code></div>
      </div>
      <CriterionChecklist results={out.criteriaResults} />
      {omitExposeSummary ? null : <ForwardedKeysTable expose={out.expose} />}

      <CollapsibleRawJson title="Сырой ответ Telegram" data={action.provider_response_json} />
      <CollapsibleRawJson title="output_json" data={action.output_json} />
    </div>
  );
}
