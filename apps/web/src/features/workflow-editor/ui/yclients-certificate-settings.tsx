import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Integration } from "../../../lib/api";
import type { ActionConfig } from "../model/types";

export function YClientsCertificateSettings({
  integrations,
  config,
  update,
}: {
  integrations: Integration[];
  config: ActionConfig;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const filtered = integrations.filter(
    (item) => item.status === "active" && item.dependency_type_key === "yclients.account",
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>YClients: выпуск сертификата</CardTitle>
        <CardDescription>
          Получает только постоплатную команду DemoShop. Платёж создаётся и подтверждается исключительно в DemoShop.
        </CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Интеграция YClients">
          <Select value={config.action_dependency_id ?? ""} onChange={(event) => update({ action_dependency_id: event.target.value })}>
            <option value="">Выберите интеграцию</option>
            {filtered.map((item) => (
              <option value={item.action_dependency_id} key={item.action_dependency_id}>{item.name}</option>
            ))}
          </Select>
        </Field>
        <div className="grid-2">
          <Field label="Ключ операции"><Input value={config.operation_key ?? ""} onChange={(event) => update({ operation_key: event.target.value })} /></Field>
          <Field label="Номер сертификата"><Input value={config.certificate_number ?? ""} onChange={(event) => update({ certificate_number: event.target.value })} /></Field>
        </div>
        <div className="grid-2">
          <Field label="Идентификатор заказа"><Input value={config.order_id ?? ""} onChange={(event) => update({ order_id: event.target.value })} /></Field>
          <Field label="Идентификатор позиции"><Input value={config.order_item_id ?? ""} onChange={(event) => update({ order_item_id: event.target.value })} /></Field>
        </div>
        <Field label="Идентификатор экземпляра"><Input value={config.instance_id ?? ""} onChange={(event) => update({ instance_id: event.target.value })} /></Field>
        <div className="grid-2">
          <Field label="Тип продукта"><Input value={config.product_kind ?? ""} onChange={(event) => update({ product_kind: event.target.value })} /></Field>
          <Field label="Номинал, ₽"><Input value={config.amount_rub ?? ""} onChange={(event) => update({ amount_rub: event.target.value })} /></Field>
        </div>
        <Field label="Название продукта"><Input value={config.product_title ?? ""} onChange={(event) => update({ product_title: event.target.value })} /></Field>
        <Field label="Идентификаторы услуг YClients (JSON)" hint="Массив ID услуг, например [29212605]. Для свободного номинала оставьте [].">
          <Textarea value={config.service_ids_json ?? "[]"} onChange={(event) => update({ service_ids_json: event.target.value })} />
        </Field>
      </CardContent>
    </Card>
  );
}
