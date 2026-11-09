import Link from "next/link";
import { ArrowRight, CreditCard, Plus, Trash2 } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Table } from "../../../components/ui/table";
import { billingPeriodLabel, invoiceScenarioFromMetadata, invoiceScenarioLabel, statusLabel, statusTone } from "../../../lib/client-labels";
import { formatDate } from "../../../lib/dates";
import { formatMoney, shortID } from "../model/clientDetailModel";
import type { ClientDetailState } from "../model/useClientDetailState";

export function ClientServicesPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel">
      <header>
        <div>
          <h2>Услуги</h2>
          <p>{state.activeProjectId ? `Только проект: ${state.activeProjectName || "выбранный"}` : "Все услуги клиента по проектам"}</p>
        </div>
        <Link className="ui-button ui-button--default ui-button--md" href={state.clientId ? `/clients/${state.clientId}/services/new` : "/clients"}>
          <Plus size={18} />
          <span>Добавить услугу</span>
        </Link>
      </header>
      <div className="section-panel-body">
        {state.services.length === 0 ? (
          <div className="empty-state">У клиента пока нет услуг. Чтобы внести оплату по услуге, сначала добавьте услугу.</div>
        ) : (
          <Table>
            <thead>
              <tr><th>ID</th><th>Название</th><th>Тип</th><th>Статус</th><th>Сумма</th><th>Период</th><th>Следующая дата</th><th>Сценарий счета</th><th /></tr>
            </thead>
            <tbody>
              {state.services.map((item) => (
                <tr key={item.client_service_id} className={state.selectedClientServiceId === item.client_service_id ? "is-active" : undefined}>
                  <td className="mono">{shortID(item.client_service_id)}</td>
                  <td>
                    <Link className="action-link" href={`/client-services/${item.client_service_id}`}>
                      <span>{item.name}</span>
                      <ArrowRight size={15} />
                    </Link>
                  </td>
                  <td>{item.service_type_name}</td>
                  <td><Badge tone={statusTone(item.status)}>{statusLabel(item.status)}</Badge></td>
                  <td>{formatMoney(item.amount)}</td>
                  <td>{billingPeriodLabel(item.billing_period)}</td>
                  <td>{formatDate(item.next_due_date)}</td>
                  <td>{invoiceScenarioLabel(invoiceScenarioFromMetadata(item.metadata_json))}</td>
                  <td>
                    <div className="form-actions">
                      <Button variant="ghost" size="sm" onClick={() => state.selectPaymentService(item)}><CreditCard size={16} /></Button>
                      <Button variant="ghost" size="sm" onClick={() => state.deleteService(item.client_service_id)}><Trash2 size={16} /></Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </div>
    </section>
  );
}
