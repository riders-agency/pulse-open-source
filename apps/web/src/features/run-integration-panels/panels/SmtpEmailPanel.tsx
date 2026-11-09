"use client";

import { MailCheck } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { CollapsibleRawJson } from "../CollapsibleRawJson";
import { CriterionChecklist } from "../CriterionChecklist";
import { ForwardedKeysTable } from "../ForwardedKeysTable";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";
import { normalizeActionOutput, normalizeProvider } from "../parseOutputs";

function SmtpStagesTimeline({ envelope }: { envelope: Record<string, unknown> | null }) {
  if (!envelope) return null;
  const stages = envelope.stages;
  const reply = envelope.last_server_reply != null ? String(envelope.last_server_reply) : null;
  const failedStage = envelope.failed_stage != null ? String(envelope.failed_stage) : null;
  const smtpCode = envelope.smtp_code != null ? String(envelope.smtp_code) : null;
  const enhanced = envelope.enhanced_status_code != null ? String(envelope.enhanced_status_code) : null;
  if (!Array.isArray(stages) || stages.length === 0) return null;

  return (
    <div className="stack gap-2">
      <div className="text-sm font-medium text-foreground">Этапы SMTP</div>
      <ol className="list-inside list-decimal text-xs leading-relaxed text-muted-foreground">
        {stages.map((s, i) => (
          <li key={i}>{typeof s === "string" ? s : JSON.stringify(s)}</li>
        ))}
      </ol>
      {reply ? (
        <div className="text-xs">
          Последний ответ: <code>{reply}</code>
        </div>
      ) : null}
      {(failedStage || smtpCode || enhanced) ? (
        <div className="grid gap-1 rounded-md border border-border bg-muted/35 p-2 text-xs">
          <div>Этап сбоя: <code>{failedStage || "—"}</code></div>
          <div>SMTP код: <code>{smtpCode || "—"}</code></div>
          <div>Enhanced code: <code>{enhanced || "—"}</code></div>
        </div>
      ) : null}
      {typeof envelope.accepted === "boolean" ? (
        <Badge tone={envelope.accepted ? "success" : "danger"}>accepted={String(envelope.accepted)}</Badge>
      ) : null}
    </div>
  );
}

export function SmtpEmailPanel({ action, omitExposeSummary }: IntegrationPanelCtx) {
  const out = normalizeActionOutput(action.output_json);
  const provider = normalizeProvider(action.provider_response_json);
  const smtp = provider?.smtp != null && typeof provider.smtp === "object" && !Array.isArray(provider.smtp) ? (provider.smtp as Record<string, unknown>) : null;

  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2">
        <MailCheck className="text-amber-600" size={20} />
        <div>
          <div className="text-sm font-medium">Письмо (SMTP)</div>
          <div className="muted text-xs">{action.action_key}</div>
        </div>
      </div>
      <SmtpStagesTimeline envelope={smtp} />
      <CriterionChecklist results={out.criteriaResults} />
      {omitExposeSummary ? null : <ForwardedKeysTable expose={out.expose} />}

      <CollapsibleRawJson title="Сырые данные провайдера" data={action.provider_response_json} />
      <CollapsibleRawJson title="output_json" data={action.output_json} />
    </div>
  );
}
