import Link from "next/link";

import { Badge } from "../../../components/ui/badge";
import { Table } from "../../../components/ui/table";
import { paymentMethodLabel, paymentSourceLabel, paymentStatusLabel, statusTone } from "../../../lib/client-labels";
import { formatDate, formatDateRange } from "../../../lib/dates";
import { formatMoney, shortID } from "../model/clientDetailModel";
import type { ClientDetailState } from "../model/useClientDetailState";

const paymentFilters = [
  ["all", "Все"],
  ["paid", "Оплачено"],
  ["pending", "Ожидается"],
  ["overdue", "Просрочено"],
  ["manual", "Ручные"],
  ["bank", "Банк"],
  ["email", "Почта"],
  ["integration", "Интеграции"],
];

export function ClientPaymentsPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel">
      <header>
        <div>
          <h2>Оплаты</h2>
          <p>Оплаченные счета, счета к оплате, периоды и источники.</p>
        </div>
      </header>
      <div className="section-panel-body stack">
        <div className="segmented" role="tablist" aria-label="Фильтр оплат">
          {paymentFilters.map(([key, label]) => (
            <button className={`segment ${state.paymentFilter === key ? "segment-active" : ""}`} key={key} type="button" onClick={() => state.setPaymentFilter(key)}>
              {label}
            </button>
          ))}
        </div>
        {state.filteredPayments.length === 0 ? <div className="empty-state">Оплат и счетов по выбранному фильтру нет.</div> : <PaymentsTable state={state} />}
      </div>
    </section>
  );
}

function PaymentsTable({ state }: { state: ClientDetailState }) {
  return (
    <Table>
      <thead>
        <tr>
          <th>payment_id</th><th>client_id</th><th>client_service_id</th><th>Услуга</th><th>Сумма</th><th>Период</th>
          <th>Дата оплаты</th><th>Источник</th><th>Способ</th><th>Статус</th><th>Комментарий</th><th>Сценарий / действие</th>
        </tr>
      </thead>
      <tbody>
        {state.filteredPayments.map((item) => {
          const firstItem = item.items?.[0];
          const fact = item.facts?.[0];
          const selected =
            state.selectedPaymentActionId === item.payment_action_id ||
            Boolean(state.selectedPaymentId && item.facts?.some((paymentFact) => paymentFact.payment_fact_id === state.selectedPaymentId));
          return (
            <tr key={item.payment_action_id} className={selected ? "is-active" : undefined}>
              <td><Link className="action-link" href={`/payment-actions/${item.payment_action_id}`}>{shortID(item.payment_action_id)}</Link></td>
              <td className="mono">{shortID(item.client_id)}</td>
              <td className="mono">{firstItem?.client_service_id ? shortID(firstItem.client_service_id) : "—"}</td>
              <td>{firstItem?.title || "—"}</td>
              <td>{formatMoney(item.total_amount)}</td>
              <td>{formatDateRange(item.period_start, item.period_end)}</td>
              <td>{formatDate(item.confirmed_at)}</td>
              <td>{paymentSourceLabel(item.source)}</td>
              <td>{paymentMethodLabel(fact?.payment_method || item.payment_method)}</td>
              <td><Badge tone={statusTone(item.status)}>{paymentStatusLabel(item.status)}</Badge></td>
              <td>{item.comment || "—"}</td>
              <td>{item.workflow_run_id ? <Link className="action-link" href={`/runs/${item.workflow_run_id}`}>Запуск</Link> : "Плановый платёж"}</td>
            </tr>
          );
        })}
      </tbody>
    </Table>
  );
}
