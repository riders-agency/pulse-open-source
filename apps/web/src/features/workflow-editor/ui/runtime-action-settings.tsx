import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { ActionConfig } from "../model/types";

export { APINETNormalizeSettings } from "./apinet-normalize-settings";
export { ClientPaymentDueSettings } from "./client-payment-due-settings";
export { RetrySettings } from "./retry-settings";

export function BotStateSettings({ config, update }: { config: ActionConfig; update: (patch: Partial<ActionConfig>) => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Состояние бота</CardTitle>
        <CardDescription>Сохраняет состояние Telegram диалога для следующей команды или сообщения.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="state_key" hint="Например waiting_for_email, menu.main или closed. Можно использовать шаблоны.">
          <Input value={config.state_key ?? ""} onChange={(event) => update({ state_key: event.target.value })} />
        </Field>
        <label className="row">
          <input
            type="checkbox"
            checked={config.merge_context ?? true}
            onChange={(event) => update({ merge_context: event.target.checked })}
          />
          merge context
        </label>
        <Field label="context_json" hint="JSON объект. Шаблоны берутся из telegram_relay, bot_session, steps и system.">
          <Textarea
            value={config.context_json ?? "{\n  \"last_command\": \"{{ telegram_relay.command }}\"\n}"}
            onChange={(event) => update({ context_json: event.target.value })}
          />
        </Field>
        <div className="template-help">
          <div className="row">
            <code>{"{{ telegram_relay.command }}"}</code>
            <code>{"{{ telegram_relay.chat.id }}"}</code>
            <code>{"{{ bot_session.state_key }}"}</code>
            <code>{"{{ system.link }}"}</code>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function ScheduleSettings({ config, update }: { config: ActionConfig; update: (patch: Partial<ActionConfig>) => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Расписание</CardTitle>
        <CardDescription>Сценарий будет запускаться по интервалу без внешнего вебхука.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <label className="row">
          <input
            type="checkbox"
            checked={config.enabled ?? true}
            onChange={(event) => update({ enabled: event.target.checked })}
          />
          включено
        </label>
        <Field label="Ключ расписания" hint="Будет доступен как {{ periodic.scheduled_at }} и {{ schedule.scheduled_at }}.">
          <Input value={config.schedule_key ?? "periodic"} onChange={(event) => update({ schedule_key: event.target.value })} />
        </Field>
        <Field label="Интервал, секунд" hint="Минимум 10 секунд; для ежедневного запуска укажите 86400.">
          <Input
            min={10}
            type="number"
            value={config.interval_seconds ?? 3600}
            onChange={(event) => update({ interval_seconds: Number(event.target.value) })}
          />
        </Field>
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

export function ClientPaymentCycleCollectSettings({ config, update }: { config: ActionConfig; update: (patch: Partial<ActionConfig>) => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Отчет по оплатам</CardTitle>
        <CardDescription>Отчет собирает две главные очереди: “Ждем денег” и “Ждут счета”. Документы проверяются в карточке оплаты.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Окно планирования, дней">
          <Input
            min={0}
            max={90}
            type="number"
            value={config.days_before_due ?? 30}
            onChange={(event) => update({ days_before_due: Number(event.target.value) })}
          />
        </Field>
      </CardContent>
    </Card>
  );
}

export function HttpRequestSettings({ config, update }: { config: ActionConfig; update: (patch: Partial<ActionConfig>) => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>HTTP запрос</CardTitle>
        <CardDescription>Ответ будет доступен в шаблонах следующих действий.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <div className="grid-2">
          <Field label="Метод">
            <Select value={config.method ?? "GET"} onChange={(event) => update({ method: event.target.value })}>
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
              <option value="DELETE">DELETE</option>
            </Select>
          </Field>
          <Field label="Timeout, секунд">
            <Input
              min={1}
              type="number"
              value={config.timeout_seconds ?? 30}
              onChange={(event) => update({ timeout_seconds: Number(event.target.value) })}
            />
          </Field>
        </div>
        <Field label="URL">
          <Input value={config.url ?? ""} onChange={(event) => update({ url: event.target.value })} />
        </Field>
        <Field label="Заголовки JSON">
          <Textarea
            value={config.headers_json ?? "{\n  \"Accept\": \"application/json\"\n}"}
            onChange={(event) => update({ headers_json: event.target.value })}
          />
        </Field>
        <Field label="Содержимое">
          <Textarea value={config.body ?? ""} onChange={(event) => update({ body: event.target.value })} />
        </Field>
        <div className="template-help">
          <div className="row">
            <code>{"{{ schedule.scheduled_at }}"}</code>
            <code>{"{{ system.link }}"}</code>
            <code>{"{{ steps.get_data.json.id }}"}</code>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
