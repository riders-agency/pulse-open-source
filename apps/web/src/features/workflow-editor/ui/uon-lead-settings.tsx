import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Integration } from "../../../lib/api";
import type { ActionConfig } from "../model/types";

export function UonLeadSettings({
  integrations,
  config,
  update,
}: {
  integrations: Integration[];
  config: ActionConfig;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const filtered = integrations.filter(
    (item) => item.status === "active" && item.dependency_type_key === "uon.account",
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>U-ON: обращение</CardTitle>
        <CardDescription>
          Создаёт лид в U-ON.Travel через POST /lead/create.json. Поля поддерживают шаблоны из входа сценария.
        </CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Интеграция U-ON">
          <Select value={config.action_dependency_id ?? ""} onChange={(event) => update({ action_dependency_id: event.target.value })}>
            <option value="">Выберите интеграцию</option>
            {filtered.map((item) => (
              <option value={item.action_dependency_id} key={item.action_dependency_id}>
                {item.name}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Канал (channel)" hint="Попадёт в source как «lead_source_prefix / channel».">
          <Input value={config.channel ?? "посетитель на сайте"} onChange={(event) => update({ channel: event.target.value })} />
        </Field>
        <Field label="Текст обращения (note)">
          <Textarea value={config.note ?? ""} onChange={(event) => update({ note: event.target.value })} />
        </Field>
        <div className="grid-2">
          <Field label="Имя (u_name)">
            <Input value={config.u_name ?? ""} onChange={(event) => update({ u_name: event.target.value })} />
          </Field>
          <Field label="Фамилия (u_surname)">
            <Input value={config.u_surname ?? ""} onChange={(event) => update({ u_surname: event.target.value })} />
          </Field>
        </div>
        <div className="grid-2">
          <Field label="Отчество (u_sname)">
            <Input value={config.u_sname ?? ""} onChange={(event) => update({ u_sname: event.target.value })} />
          </Field>
          <Field label="Телефон (u_phone)">
            <Input value={config.u_phone ?? ""} onChange={(event) => update({ u_phone: event.target.value })} />
          </Field>
        </div>
        <Field label="Электронная почта (u_email)">
          <Input value={config.u_email ?? ""} onChange={(event) => update({ u_email: event.target.value })} />
        </Field>
        <Field
          label="Доп. поля JSON"
          hint='tourist_count, date_from, travel_type_id, r_u_id, r_co_id, extended_fields и др. Пример: {"r_u_id":"123","extended_fields":{"193320":"{{ current_title }}"}}'
        >
          <Textarea
            value={config.extra_fields_json ?? "{\n  \"extended_fields\": {\n    \"193320\": \"{{ current_title }}\"\n  }\n}"}
            onChange={(event) => update({ extra_fields_json: event.target.value })}
          />
        </Field>
      </CardContent>
    </Card>
  );
}
