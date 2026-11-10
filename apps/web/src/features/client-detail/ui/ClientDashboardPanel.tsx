import Link from "next/link";
import { CreditCard, Plus, Save, UserPlus } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { formatDate } from "../../../lib/dates";
import { billingPeriodLabel, statusLabel } from "../../../lib/client-labels";
import { formatMoney } from "../model/clientDetailModel";
import type { ClientDetailState } from "../model/useClientDetailState";

export function ClientDashboardPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel section-panel-spaced">
      <header>
        <div>
          <h2>Дашборд клиента</h2>
          <p>Сводка по услугам, периодам и оплатам клиента.</p>
        </div>
        <div className="form-actions">
          <a className="ui-button ui-button--secondary ui-button--md" href="#client-responsibles">
            <UserPlus size={18} />
            <span>Назначить ответственного</span>
          </a>
          <Link className="ui-button ui-button--default ui-button--md" href={state.clientId ? `/tasks/all?client_id=${state.clientId}&create=1` : "/tasks/all"}>
            <Plus size={18} />
            <span>Создать задачу</span>
          </Link>
          <Link className="ui-button ui-button--secondary ui-button--md" href={state.clientId ? `/clients/${state.clientId}/services/new` : "/clients"}>
            <Plus size={18} />
            <span>Добавить услугу</span>
          </Link>
          <Button onClick={state.recordPayment} disabled={state.paymentBusy || state.services.length === 0}>
            <CreditCard size={18} />
            <span>Счет оплачен</span>
          </Button>
          <Button variant="secondary" onClick={state.saveClient} disabled={state.busy || !state.edit?.name.trim()}>
            <Save size={18} />
            <span>Сохранить клиента</span>
          </Button>
        </div>
      </header>
      <div className="section-panel-body">
        <div className="grid-3">
          <div className="metric"><span>Статус</span><strong>{statusLabel(state.client?.status)}</strong></div>
          <div className="metric"><span>Последняя оплата</span><strong>{state.latestPayment ? formatMoney(state.latestPayment.total_amount) : "—"}</strong></div>
          <div className="metric"><span>Следующий период</span><strong>{formatDate(state.nextService?.next_due_date)}</strong></div>
        </div>
        <div className="compact-list compact-list-spaced">
          {state.activeServices.slice(0, 3).map((item) => (
            <button className="compact-item text-left" type="button" key={item.client_service_id} onClick={() => state.selectPaymentService(item)}>
              <div className="between">
                <strong>{item.name}</strong>
                <span className="muted">{formatMoney(item.amount)}</span>
              </div>
              <span className="muted">
                {billingPeriodLabel(item.billing_period)} · следующий период {formatDate(item.next_due_date, "не задан")}
              </span>
            </button>
          ))}
          {state.activeServices.length === 0 ? (
            <div className="empty-state">У клиента пока нет услуг. Чтобы внести оплату по услуге, сначала добавьте услугу.</div>
          ) : null}
        </div>
      </div>
    </section>
  );
}
