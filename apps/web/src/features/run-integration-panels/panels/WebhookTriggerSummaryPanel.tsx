"use client";

import { Webhook } from "lucide-react";

import { CollapsibleRawJson } from "../CollapsibleRawJson";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";

export function WebhookTriggerSummaryPanel({ action }: IntegrationPanelCtx) {
  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2">
        <Webhook className="text-slate-600" size={20} />
        <div className="text-sm font-medium">Триггер: вебхук</div>
      </div>
      <CollapsibleRawJson title="payload / input запуска" data={action.input_json} />
      <CollapsibleRawJson title="output триггера" data={action.output_json} defaultOpen />
    </div>
  );
}
