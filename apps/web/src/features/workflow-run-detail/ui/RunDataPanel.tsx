import { Badge } from "../../../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import type { ActionAttempt, ActionRun, Run } from "../../../lib/api";
import { jsonText } from "../../../lib/api";
import { dataBadge } from "../model/workflowRunDiagnostics";
import { formatMoney, paymentRunInfo } from "../model/workflowRunPayment";
import { actionTypeLabel, sourceTypeLabel, statusLabel } from "../model/workflowRunLabels";

function RunDataBlock({
  title,
  description,
  data,
  defaultOpen = true,
}: {
  title: string;
  description: string;
  data: unknown;
  defaultOpen?: boolean;
}) {
  return (
    <details className="run-data-block" open={defaultOpen}>
      <summary>
        <span>
          <strong>{title}</strong>
          <small>{description}</small>
        </span>
        <Badge tone={data == null ? "muted" : "info"}>{dataBadge(data)}</Badge>
      </summary>
      <pre className="code-block">{jsonText(data)}</pre>
    </details>
  );
}

export function RunDataPanel({
  run,
  selected,
  latestAttempt,
}: {
  run: Run | null;
  selected?: ActionRun;
  latestAttempt?: ActionAttempt;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Данные запуска</CardTitle>
        <CardDescription>Входящие данные и фактические параметры выбранного действия.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <div className="run-data-summary">
          <div><span>Запуск</span><strong>{run ? statusLabel(run.status) : "загрузка"}</strong></div>
          <div><span>Источник</span><strong>{sourceTypeLabel(run?.source_type)}</strong></div>
          <div><span>Действие</span><strong>{selected ? actionTypeLabel(selected.action_type_key) : "—"}</strong></div>
        </div>
        <PaymentRunSummary run={run} selected={selected} />
        <RunDataBlock title="Вход запуска" description="JSON, который пришел в сценарий." data={run?.input_json} />
        {selected ? (
          <>
            <RunDataBlock title="Отправлено действию" description="Поля и данные, которые исполнитель передал выбранному шагу" data={selected.input_json} />
            <RunDataBlock title="Результат действия" description="Нормализованный результат для интерфейса и следующих шагов" data={selected.output_json} defaultOpen={selected.status === "failed"} />
            <RunDataBlock title="Ответ провайдера" description="Сырой обезличенный ответ провайдера" data={latestAttempt?.provider_response_json ?? selected.provider_response_json} defaultOpen={selected.status === "failed"} />
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}

function PaymentRunSummary({ run, selected }: { run?: Run | null; selected?: ActionRun }) {
  const info = paymentRunInfo(run?.input_json, selected);
  if (!info) return null;
  if (info.kind === "report") {
    return (
      <section className="payment-run-summary">
        <div className="payment-run-head">
          <strong>Отчет по оплатам</strong>
          <span>Сначала смотрим очереди, JSON ниже нужен только для отладки.</span>
        </div>
        <div className="payment-run-groups">
          <div><span>Ждем денег</span><strong>{info.waitingHint}</strong></div>
          <div><span>Ждут счета</span><strong>{info.invoiceHint}</strong></div>
        </div>
        {info.digest ? <p>{info.digest}</p> : null}
      </section>
    );
  }
  return (
    <section className="payment-run-summary">
      <div className="payment-run-head">
        <strong>{info.clientName}</strong>
        <span>{info.serviceName || "Оплата клиента"}</span>
      </div>
      <div className="payment-run-groups">
        <div><span>Сумма</span><strong>{formatMoney(info.amount)}</strong></div>
        <div><span>Состояние</span><strong>{info.statusLabel}</strong></div>
      </div>
      <p>{info.nextStep}</p>
    </section>
  );
}
