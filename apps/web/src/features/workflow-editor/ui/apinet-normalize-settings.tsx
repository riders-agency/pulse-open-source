import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Integration } from "../../../lib/api";
import type { ActionConfig } from "../model/types";

export function APINETNormalizeSettings({ integrations, config, update }: {
  integrations: Integration[];
  config: ActionConfig;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const filtered = integrations.filter((item) => item.status === "active" && item.dependency_type_key === "ai.apinet");
  return (
    <Card>
      <CardHeader>
        <CardTitle>APINET нормализация</CardTitle>
        <CardDescription>Редактирует исходный текст и отдает результат в steps текущего действия.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Интеграция">
          <Select value={config.action_dependency_id ?? ""} onChange={(event) => update({ action_dependency_id: event.target.value })}>
            <option value="">Выберите APINET</option>
            {filtered.map((item) => (
              <option value={item.action_dependency_id} key={item.action_dependency_id}>{item.name}</option>
            ))}
          </Select>
        </Field>
        <Field label="source_text">
          <Textarea value={config.source_text ?? "{{ telegram_relay.text }}"} onChange={(event) => update({ source_text: event.target.value })} />
        </Field>
        <Field label="Инструкция">
          <Textarea value={config.instructions ?? ""} onChange={(event) => update({ instructions: event.target.value })} />
        </Field>
        <div className="grid-2">
          <Field label="Модель">
            <Input value={config.model ?? "gpt-5-chat-latest"} onChange={(event) => update({ model: event.target.value })} />
          </Field>
          <Field label="max_tokens">
            <Input min={1} type="number" value={config.max_tokens ?? 800} onChange={(event) => update({ max_tokens: Number(event.target.value) })} />
          </Field>
        </div>
        <Field label="Вариативность ответа">
          <Input min={0} max={2} step={0.1} type="number" value={config.temperature ?? 0.2} onChange={(event) => update({ temperature: Number(event.target.value) })} />
        </Field>
        <div className="template-help">
          <div className="row">
            <code>{"{{ telegram_relay.text }}"}</code>
            <code>{"{{ steps.<action_key>.normalized_text }}"}</code>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
