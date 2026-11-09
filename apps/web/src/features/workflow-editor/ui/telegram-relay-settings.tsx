import { useState } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Integration, Project } from "../../../lib/api";
import { isRecord } from "../model/editorDraft";
import type { ActionConfig } from "../model/types";
import { allTelegramUpdateTypes } from "../registry/actions";
import { TelegramRelayPollingCard } from "./telegram-relay-polling-card";
import { TelegramRelaySecretCard } from "./telegram-relay-secret-card";
import {
  buildTelegramRelaySnippets,
  defaultTelegramRelayPollingSnippetSettings,
  type TelegramRelayPollingSnippetSettings,
} from "./telegram-relay-snippets";

export function TelegramRelaySettings({
  project,
  token,
  integrations,
  config,
  update,
}: {
  project: Project | null;
  token: string;
  integrations: Integration[];
  config: ActionConfig;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const [snippetKey, setSnippetKey] = useState("curl");
  const [webhookSecret, setWebhookSecret] = useState("");
  const [pollingSnippetSettings, setPollingSnippetSettings] = useState<TelegramRelayPollingSnippetSettings>(
    defaultTelegramRelayPollingSnippetSettings,
  );
  const relayIntegrations = integrations.filter((item) => item.status === "active" && item.dependency_type_key === "telegram.relay");
  const selectedIntegration = relayIntegrations.find((item) => item.action_dependency_id === config.action_dependency_id);
  const meta = isRecord(selectedIntegration?.metadata_json) ? selectedIntegration.metadata_json : {};
  const relayKey = config.relay_key?.trim() || (typeof meta.relay_key === "string" ? meta.relay_key : "support_bot");
  const relayUrl = typeof meta.relay_url === "string" && meta.relay_url ? meta.relay_url : `https://pulse.riders.agency/api/v1/telegram/relay/${relayKey}`;
  const allowedUpdates = config.allowed_updates?.length ? config.allowed_updates : allTelegramUpdateTypes;
  const snippets = buildTelegramRelaySnippets({
    relayKey,
    relayUrl,
    secretValue: webhookSecret || "WEBHOOK_SECRET_FROM_INTEGRATION",
    allowedUpdates,
    polling: pollingSnippetSettings,
  });
  const selectedSnippet = snippets.find((snippet) => snippet.key === snippetKey) ?? snippets[0];
  const setAllowedUpdates = (value: string) => {
    update({ allowed_updates: value.split(",").map((item) => item.trim()).filter(Boolean) });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>События Telegram</CardTitle>
        <CardDescription>Принимает Bot API updates, сохраняет raw JSON и запускает downstream действия.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Интеграция">
          <Select
            value={config.action_dependency_id ?? ""}
            onChange={(event) => {
              const integration = relayIntegrations.find((item) => item.action_dependency_id === event.target.value);
              const integrationMeta = isRecord(integration?.metadata_json) ? integration.metadata_json : {};
              update({
                action_dependency_id: event.target.value,
                relay_key: typeof integrationMeta.relay_key === "string" ? integrationMeta.relay_key : config.relay_key,
              });
            }}
          >
            <option value="">Выберите telegram.relay</option>
            {relayIntegrations.map((item) => (
              <option value={item.action_dependency_id} key={item.action_dependency_id}>{item.name}</option>
            ))}
          </Select>
        </Field>
        <Field label="Ключ обработчика" hint="Префикс переменных: {{ telegram_relay.command }} и {{ <relay_key>.message.text }}.">
          <Input value={relayKey} onChange={(event) => update({ relay_key: event.target.value })} />
        </Field>
        <Field label="Адрес обработчика">
          <Input value={relayUrl} readOnly />
        </Field>
        <div className="segmented" role="tablist" aria-label="Примеры подключения Telegram">
          {snippets.map((snippet) => (
            <button
              className={`segment ${snippet.key === selectedSnippet.key ? "segment-active" : ""}`}
              key={snippet.key}
              onClick={() => setSnippetKey(snippet.key)}
              type="button"
            >
              {snippet.label}
            </button>
          ))}
        </div>
        <pre className="code-block">{selectedSnippet.code}</pre>
        <TelegramRelaySecretCard
          token={token}
          projectId={project?.project_id}
          actionDependencyId={config.action_dependency_id}
          relayKey={relayKey}
          onSecretChange={setWebhookSecret}
        />
        <TelegramRelayPollingCard
          token={token}
          projectId={project?.project_id}
          actionDependencyId={config.action_dependency_id}
          metadata={meta}
          onSnippetSettingsChange={setPollingSnippetSettings}
        />
        <Field label="Типы событий">
          <Input value={allowedUpdates.join(", ")} onChange={(event) => setAllowedUpdates(event.target.value)} />
        </Field>
        <div className="template-help">
          <div className="row">
            <code>{"{{ telegram_relay.command }}"}</code>
            <code>{"{{ telegram_relay.message.text }}"}</code>
            <code>{`{{ ${relayKey}.chat.id }}`}</code>
            <code>{"{{ system.link }}"}</code>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
