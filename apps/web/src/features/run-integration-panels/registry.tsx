"use client";

import type { ComponentType } from "react";

import type { ActionRun } from "../../lib/api";
import { GenericIntegrationPanel, type IntegrationPanelCtx } from "./panels/GenericIntegrationPanel";
import { HttpRequestPanel } from "./panels/HttpRequestPanel";
import { ManualTriggerSummaryPanel } from "./panels/ManualTriggerSummaryPanel";
import { MaxSendPanel } from "./panels/MaxSendPanel";
import { SmtpEmailPanel } from "./panels/SmtpEmailPanel";
import { TelegramRelayTriggerPanel } from "./panels/TelegramRelayTriggerPanel";
import { TelegramSendPanel } from "./panels/TelegramSendPanel";
import { WebhookTriggerSummaryPanel } from "./panels/WebhookTriggerSummaryPanel";

const REGISTRY: Record<string, ComponentType<IntegrationPanelCtx>> = {};

function registerComposite(keys: string[], C: ComponentType<IntegrationPanelCtx>) {
  for (const k of keys) REGISTRY[k] = C;
}

registerComposite(["telegram.send_message", "telegram.send_message::telegram.bot"], TelegramSendPanel);
registerComposite(["max.send_message", "max.send_message::max.bot"], MaxSendPanel);
registerComposite(["email.send", "email.send::smtp.server"], SmtpEmailPanel);
registerComposite(["http.request"], HttpRequestPanel);
registerComposite(["webhook.trigger"], WebhookTriggerSummaryPanel);
registerComposite(["telegram.relay.trigger", "telegram.relay.trigger::telegram.relay"], TelegramRelayTriggerPanel);
registerComposite(["manual.trigger"], ManualTriggerSummaryPanel);

function integrationProfileFromAction(action: ActionRun): string {
  const a = action as ActionRun & { integration_profile?: string; dependency_type_key?: string };
  const p = String(a.integration_profile ?? a.dependency_type_key ?? "").trim();
  return p;
}

function resolveKeys(actionTypeKey: string, profile?: string): string[] {
  const atk = actionTypeKey.trim();
  const out: string[] = [];
  const p = profile?.trim();
  if (atk && p) out.push(`${atk}::${p}`);
  if (atk) out.push(atk);
  return out;
}

/** Picks integration-specific panel component (action_type_key × dependency_type_key fallback). */
export function resolveIntegrationPanelComponent(action: ActionRun): ComponentType<IntegrationPanelCtx> {
  const profile = integrationProfileFromAction(action);
  for (const k of resolveKeys(action.action_type_key ?? "", profile)) {
    const C = REGISTRY[k];
    if (C) return C;
  }
  return GenericIntegrationPanel;
}

export function ResolvedIntegrationPanel({ action, omitExposeSummary }: { action: ActionRun; omitExposeSummary?: boolean }) {
  const Cmp = resolveIntegrationPanelComponent(action);
  return (
    <div className="integration-panel-root">
      <Cmp action={action} omitExposeSummary={omitExposeSummary} />
    </div>
  );
}
