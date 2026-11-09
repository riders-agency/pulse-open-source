import { Field, Input } from "../../../components/ui/input";
import type { Integration, Project } from "../../../lib/api";
import type { ActionConfig, ActionNodeType } from "../model/types";
import {
  APINETNormalizeSettings,
  BotStateSettings,
  ClientPaymentCycleCollectSettings,
  ClientPaymentDueSettings,
  FioNormalizeSettings,
  HttpRequestSettings,
  IntegrationSettings,
  PhoneNormalizeSettings,
  ScheduleSettings,
  TelegramRelaySettings,
  UonLeadSettings,
  WebhookSettings,
  YClientsCertificateSettings,
} from "./panels";

export function ActionSettings({
  selectedNode,
  project,
  token,
  integrations,
  webhookSlug,
  templateVariables,
  updateSelectedConfig,
}: {
  selectedNode: ActionNodeType;
  project: Project | null;
  token: string;
  integrations: Integration[];
  webhookSlug: string;
  templateVariables: string[];
  updateSelectedConfig: (patch: Partial<ActionConfig>) => void;
}) {
  const actionType = selectedNode.data.action_type_key;
  if (actionType === "webhook.trigger") {
    return <WebhookSettings project={project} config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "telegram.relay.trigger") {
    return (
      <TelegramRelaySettings
        project={project}
        token={token}
        integrations={integrations}
        config={selectedNode.data.config}
        update={updateSelectedConfig}
      />
    );
  }
  if (actionType === "schedule.trigger") {
    return <ScheduleSettings config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "clients.payment_due.trigger") {
    return <ClientPaymentDueSettings config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "clients.payment_cycle.collect") {
    return <ClientPaymentCycleCollectSettings config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "http.request") {
    return <HttpRequestSettings config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "bot.set_state") {
    return <BotStateSettings config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "ai.apinet.normalize_text") {
    return <APINETNormalizeSettings integrations={integrations} config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "uon.create_lead") {
    return <UonLeadSettings integrations={integrations} config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "yclients.certificate.issue") {
    return <YClientsCertificateSettings integrations={integrations} config={selectedNode.data.config} update={updateSelectedConfig} />;
  }
  if (actionType === "person.fio.normalize") {
    return <FioNormalizeSettings config={selectedNode.data.config} webhookSlug={webhookSlug} update={updateSelectedConfig} />;
  }
  if (actionType === "person.phone.normalize") {
    return <PhoneNormalizeSettings config={selectedNode.data.config} webhookSlug={webhookSlug} update={updateSelectedConfig} />;
  }
  if (actionType === "telegram.send_document" || actionType.includes("send")) {
    return (
      <IntegrationSettings
        integrations={integrations}
        type={actionType}
        config={selectedNode.data.config}
        webhookSlug={webhookSlug}
        templateVariables={templateVariables}
        update={updateSelectedConfig}
      />
    );
  }
  if (actionType === "condition.branch") {
    return (
      <Field label="Условие">
        <Input
          value={selectedNode.data.config.text ?? `${webhookSlug}.order.total > 10000`}
          onChange={(event) => updateSelectedConfig({ text: event.target.value })}
        />
      </Field>
    );
  }
  return <p className="muted">Для этого действия дополнительных настроек пока нет.</p>;
}
