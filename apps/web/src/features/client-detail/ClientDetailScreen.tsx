"use client";

import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { formatDate } from "../../lib/dates";
import { clientTypeLabel } from "../../lib/client-labels";
import { formatMoney } from "./model/clientDetailModel";
import { useClientDetailState } from "./model/useClientDetailState";
import { ClientDashboardPanel } from "./ui/ClientDashboardPanel";
import { ClientEditPanel } from "./ui/ClientEditPanel";
import { ClientPaymentsPanel } from "./ui/ClientPaymentsPanel";
import { ClientServicesPanel } from "./ui/ClientServicesPanel";
import { ClientSidebarPanels } from "./ui/ClientSidebarPanels";

export default function ClientDetailPage() {
  const state = useClientDetailState();

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>{state.client?.name ?? "Клиент"}</h1>
          <p className="muted">
            {state.client?.inn ? `ИНН ${state.client.inn}` : state.client ? clientTypeLabel(state.client.client_type) : "Карточка клиента"}
            {state.primaryResponsible ? ` · ${state.primaryResponsible.display_name}` : ""}
          </p>
        </div>
        {state.client?.company_id ? (
          <Link className="ui-button ui-button--secondary ui-button--md" href="/clients">
            <ArrowLeft size={18} />
            <span>К списку</span>
          </Link>
        ) : null}
      </div>

      {state.error ? <div className="form-error">{state.error}</div> : null}

      <section className="grid-3">
        <div className="metric"><span>Активные услуги</span><strong>{state.activeServices.length}</strong></div>
        <div className="metric"><span>Регулярно</span><strong>{formatMoney(state.recurringTotal)}</strong></div>
        <div className="metric"><span>Просрочка</span><strong>{state.overdueServices.length ? `${state.overdueServices.length}` : "нет"}</strong></div>
      </section>

      <ClientDashboardPanel state={state} />

      <div className="detail-layout detail-layout-spaced">
        <div className="stack">
          <ClientEditPanel state={state} />
          <ClientServicesPanel state={state} />
          <ClientPaymentsPanel state={state} />
          <section className="section-panel">
            <header>
              <div>
                <h2>Задачи</h2>
                <p>Будущая рабочая область задач клиента.</p>
              </div>
            </header>
            <div className="section-panel-body">
              <div className="empty-state">
                Задачи клиента будут подключены следующим модулем. Следующий период: {formatDate(state.nextService?.next_due_date, "не задан")}
              </div>
            </div>
          </section>
        </div>
        <ClientSidebarPanels state={state} />
      </div>
    </main>
  );
}
