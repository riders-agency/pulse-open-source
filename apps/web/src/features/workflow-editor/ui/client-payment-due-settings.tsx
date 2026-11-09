import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import type { ActionConfig } from "../model/types";

export function ClientPaymentDueSettings({ config, update }: { config: ActionConfig; update: (patch: Partial<ActionConfig>) => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Счета к оплате</CardTitle>
        <CardDescription>Сценарий каждый день проверяет услуги проекта, переводит счет в статус Счет подготовлен по сценарию услуги и готовит письмо ответственному со стороны заказчика.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <label className="row">
          <input type="checkbox" checked={config.enabled ?? true} onChange={(event) => update({ enabled: event.target.checked })} />
          включено
        </label>
        <div className="grid-2">
          <Field label="За сколько дней до оплаты">
            <Input min={0} max={90} type="number" value={config.days_before_due ?? 7} onChange={(event) => update({ days_before_due: Number(event.target.value) })} />
          </Field>
          <Field label="Напоминать снова через, дней">
            <Input min={1} max={365} type="number" value={config.remind_every_days ?? 1} onChange={(event) => update({ remind_every_days: Number(event.target.value) })} />
          </Field>
        </div>
        <div className="grid-2">
          <Field label="Активно с">
            <Input type="date" value={config.active_from ?? ""} onChange={(event) => update({ active_from: event.target.value })} />
          </Field>
          <Field label="Активно по">
            <Input type="date" value={config.active_until ?? ""} onChange={(event) => update({ active_until: event.target.value })} />
          </Field>
        </div>
        <Field label="Часовой пояс">
          <Input value={config.timezone ?? "Europe/Moscow"} onChange={(event) => update({ timezone: event.target.value })} />
        </Field>
      </CardContent>
    </Card>
  );
}
