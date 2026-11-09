import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import type { ActionConfig } from "../model/types";

export function FioNormalizeSettings({
  config,
  webhookSlug,
  update,
}: {
  config: ActionConfig;
  webhookSlug: string;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>ФИО: нормализация</CardTitle>
        <CardDescription>
          Разбивает строку ФИО на u_name / u_surname / u_sname по словарям Orion Travel. Результат доступен как{" "}
          <code>{"{{ steps.<action_key>.u_name }}"}</code>.
        </CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Вход (input)" hint={`Шаблон, напр. {{ ${webhookSlug}.name }}`}>
          <Input
            value={config.input ?? `{{ ${webhookSlug}.name }}`}
            onChange={(event) => update({ input: event.target.value })}
          />
        </Field>
        <Field
          label="output_fields JSON"
          hint='Опциональные алиасы expose, напр. {"client_name":"u_name"}'
        >
          <Textarea
            value={config.output_fields ? JSON.stringify(config.output_fields, null, 2) : "{}"}
            onChange={(event) => {
              try {
                update({ output_fields: JSON.parse(event.target.value || "{}") as Record<string, string> });
              } catch {
                /* ignore invalid json while typing */
              }
            }}
          />
        </Field>
      </CardContent>
    </Card>
  );
}
