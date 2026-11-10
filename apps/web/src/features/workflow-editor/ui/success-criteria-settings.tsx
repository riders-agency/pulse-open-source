import { CheckCircle2 } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import { jsonText, type ProviderCatalogEntry } from "../../../lib/api";
import type { ActionKind } from "../model/types";

export function SuccessCriteriaSettings({
  actionType,
  providerCatalog,
}: {
  actionType: ActionKind;
  providerCatalog: ProviderCatalogEntry[];
}) {
  const provider = providerCatalog.find((entry) => entry.action_type_key === actionType);
  const jsonPath = providerCatalog.find((entry) => entry.action_type_key === "json_path");
  const criteria = Array.isArray(provider?.criteria) ? provider.criteria : [];
  const failureCategories = Array.isArray(provider?.failure_categories) ? provider.failure_categories : [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Проверки доставки</CardTitle>
        <CardDescription>API берет этот контракт из реестра провайдеров.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        {provider ? (
          <>
            <div className="rounded-md border border-border bg-muted/35 p-3">
              <div className="text-sm font-medium">{provider.display_name}</div>
              {provider.retry?.summary ? <div className="muted mt-1 text-xs">{provider.retry.summary}</div> : null}
            </div>
            <div className="stack gap-2">
              {criteria.length ? (
                criteria.map((criterion) => (
                  <div key={criterion.type} className="flex gap-2 rounded-md border border-border p-3 text-sm">
                    <CheckCircle2 className="mt-0.5 shrink-0 text-emerald-600" size={16} />
                    <div className="min-w-0">
                      <div className="font-medium">{criterion.label}</div>
                      {criterion.description ? <div className="muted mt-1 text-xs">{criterion.description}</div> : null}
                      <div className="muted mt-1 text-xs">
                        {criterion.expected ? <>Ожидание: <code>{criterion.expected}</code>. </> : null}
                        {criterion.source ? <>Источник: <code>{criterion.source}</code>.</> : null}
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="muted rounded-md border border-border p-3 text-sm">
                  Для этого действия отдельных delivery checks нет. Backend все равно хранит результат и diagnostics запуска.
                </div>
              )}
            </div>
            {failureCategories.length ? (
              <details className="rounded-md border border-border">
                <summary className="cursor-pointer px-3 py-2 text-xs font-medium text-muted-foreground">Причины ошибок и повторы</summary>
                <div className="stack border-t p-3">
                  {failureCategories.map((item) => (
                    <div key={item.code} className="rounded-md border border-border p-2 text-xs">
                      <div className="between gap-2">
                        <code>{item.code}</code>
                        <Badge tone={item.retryable ? "warning" : "muted"}>{item.retryable ? "повтор" : "финал"}</Badge>
                      </div>
                      <div className="mt-1">{item.label}</div>
                      {item.next_action ? <div className="muted mt-1">{item.next_action}</div> : null}
                    </div>
                  ))}
                </div>
              </details>
            ) : null}
            <details className="rounded-md border border-border">
              <summary className="cursor-pointer px-3 py-2 text-xs font-medium text-muted-foreground">Схема ответа провайдера</summary>
              <pre className="code-block m-0 border-t">{jsonText(provider.raw_response_schema ?? {})}</pre>
            </details>
          </>
        ) : actionType === "condition.branch" && jsonPath ? (
          <div className="grid gap-3 rounded-md border border-border p-3 text-sm">
            <div className="font-medium">{jsonPath.display_name}</div>
            <div className="grid-2">
              <Field label="Источник">
                <Select value="input" disabled>
                  <option value="input">Входные данные</option>
                </Select>
              </Field>
              <Field label="Условие">
                <Select value="equals" disabled>
                  <option value="equals">Равно</option>
                </Select>
              </Field>
              <Field label="Путь к значению">
                <Input value="order.total" readOnly />
              </Field>
              <Field label="Ожидаемое значение">
                <Input value="12500" readOnly />
              </Field>
            </div>
          </div>
        ) : (
          <div className="muted text-sm">Для этого действия нет специальных проверок провайдера.</div>
        )}
      </CardContent>
    </Card>
  );
}
