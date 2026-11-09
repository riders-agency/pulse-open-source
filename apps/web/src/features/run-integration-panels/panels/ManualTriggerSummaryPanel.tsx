"use client";

import { MousePointer2 } from "lucide-react";

import { CollapsibleRawJson } from "../CollapsibleRawJson";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";

export function ManualTriggerSummaryPanel({ action }: IntegrationPanelCtx) {
  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2">
        <MousePointer2 className="text-slate-600" size={20} />
        <div className="text-sm font-medium">Триггер: ручной запуск</div>
      </div>
      <CollapsibleRawJson title="Вход запуска" data={action.input_json} defaultOpen />
      <CollapsibleRawJson title="output триггера" data={action.output_json} />
    </div>
  );
}
