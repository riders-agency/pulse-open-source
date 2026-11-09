import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { ActionConfig, RetryPolicy } from "../model/types";
import { defaultConfig } from "../registry/actions";

export function RetrySettings({ config, updateRetry }: { config: ActionConfig; updateRetry: (patch: Partial<RetryPolicy>) => void }) {
  const retry = config.retry_policy ?? defaultConfig("telegram.send_message").retry_policy!;
  return (
    <Card>
      <CardHeader>
        <CardTitle>Повторы</CardTitle>
        <CardDescription>По умолчанию 3 попытки, жесткий лимит 10.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Максимум попыток">
          <Input type="number" min={0} max={10} value={retry.max_attempts} onChange={(event) => updateRetry({ max_attempts: Number(event.target.value) })} />
        </Field>
        <Field label="Стратегия">
          <Select value={retry.strategy} onChange={(event) => updateRetry({ strategy: event.target.value as RetryPolicy["strategy"] })}>
            <option value="fixed">Постоянный интервал</option>
            <option value="exponential">С увеличением интервала</option>
          </Select>
        </Field>
        <Field label="Начальный интервал">
          <Input value={retry.initial_interval} onChange={(event) => updateRetry({ initial_interval: event.target.value })} />
        </Field>
        <Field label="Время ожидания">
          <Input value={retry.timeout} onChange={(event) => updateRetry({ timeout: event.target.value })} />
        </Field>
      </CardContent>
    </Card>
  );
}
