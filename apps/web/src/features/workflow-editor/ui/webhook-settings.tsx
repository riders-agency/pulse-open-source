import { useMemo, useState } from "react";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Project } from "../../../lib/api";
import type { ActionConfig, JsonField } from "../model/types";
import { buildWebhookSnippets } from "./webhook-snippets";

export function WebhookSettings({
  project,
  config,
  update,
}: {
  project: Project | null;
  config: ActionConfig;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const [snippetKey, setSnippetKey] = useState("curl");
  const fields = config.json_fields ?? [];
  const hookUrl = `/hooks/${project?.hook_key ?? "{project_key}"}/${config.endpoint_key ?? "{endpoint_key}"}`;
  const absoluteHookUrl = `https://pulse.riders.agency${hookUrl}`;
  const samplePayload = config.sample_payload_json?.trim() || "{}";
  const snippets = useMemo(
    () => buildWebhookSnippets({ absoluteHookUrl, authType: config.auth_type, samplePayload }),
    [absoluteHookUrl, config.auth_type, samplePayload],
  );
  const selectedSnippet = snippets.find((snippet) => snippet.key === snippetKey) ?? snippets[0];
  const setField = (index: number, patch: Partial<JsonField>) => {
    update({ json_fields: fields.map((field, i) => (i === index ? { ...field, ...patch } : field)) });
  };
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Сниппеты</CardTitle>
          <CardDescription>Готовые варианты для отправки события во входящий webhook.</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          <div className="segmented" role="tablist" aria-label="Сниппеты webhook">
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
        </CardContent>
      </Card>
      <Field label="URL вебхука">
        <Input value={hookUrl} readOnly />
      </Field>
      <Field label="Ключ точки входа">
        <Input value={config.endpoint_key ?? ""} onChange={(event) => update({ endpoint_key: event.target.value })} />
      </Field>
      <Field label="Авторизация">
        <Select value={config.auth_type ?? "project_secret"} onChange={(event) => update({ auth_type: event.target.value as ActionConfig["auth_type"] })}>
          <option value="project_secret">Секрет проекта</option>
          <option value="hmac">Подпись HMAC</option>
          <option value="bearer">Токен Bearer</option>
          <option value="none">none только dev</option>
        </Select>
      </Field>
      <Field label="Тестовый JSON">
        <Textarea value={config.sample_payload_json ?? ""} onChange={(event) => update({ sample_payload_json: event.target.value })} />
      </Field>
      <Card>
        <CardHeader>
          <CardTitle>Конструктор JSON</CardTitle>
          <CardDescription>Опишите обязательные и необязательные поля входящего payload.</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          {fields.map((field, index) => (
            <div className="grid-2" key={`${field.path}-${index}`}>
              <Input value={field.path} onChange={(event) => setField(index, { path: event.target.value })} placeholder="order.total" />
              <Select value={field.type} onChange={(event) => setField(index, { type: event.target.value as JsonField["type"] })}>
                <option value="string">Строка</option>
                <option value="number">Число</option>
                <option value="boolean">Логическое значение</option>
                <option value="object">Объект</option>
                <option value="array">Массив</option>
              </Select>
              <Input value={field.description} onChange={(event) => setField(index, { description: event.target.value })} placeholder="Описание" />
              <Input value={field.example} onChange={(event) => setField(index, { example: event.target.value })} placeholder="Пример" />
              <label className="row">
                <input type="checkbox" checked={field.required} onChange={(event) => setField(index, { required: event.target.checked })} />
                обязательное
              </label>
            </div>
          ))}
          <Button
            variant="outline"
            onClick={() =>
              update({
                json_fields: [...fields, { path: "", type: "string", required: false, description: "", example: "" }],
              })
            }
          >
            Добавить поле
          </Button>
        </CardContent>
      </Card>
    </>
  );
}
