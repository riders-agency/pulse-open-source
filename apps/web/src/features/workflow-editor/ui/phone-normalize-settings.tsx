import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import type { ActionConfig } from "../model/types";

export function PhoneNormalizeSettings({
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
        <CardTitle>Телефон: нормализация</CardTitle>
        <CardDescription>
          Приводит российский номер к формату +7XXXXXXXXXX. Результат доступен как{" "}
          <code>{"{{ steps.phone_normalize.phone_e164 }}"}</code>.
        </CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Вход (input)" hint={`Шаблон, напр. {{ ${webhookSlug}.phone }} или {{ ${webhookSlug}.PHONE }}`}>
          <Input
            value={config.input ?? `{{ ${webhookSlug}.phone }}`}
            onChange={(event) => update({ input: event.target.value })}
          />
        </Field>
        <Field
          label="output_fields JSON"
          hint='Опциональные алиасы expose, напр. {"client_phone":"phone_e164"}'
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
