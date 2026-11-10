"use client";

import Link from "next/link";
import { AlertTriangle, FileUp, ReceiptText, WalletCards } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  formatMoney,
  serviceCountLabel,
} from "./model/paymentsModel";
import { invoiceScenarioLabel } from "../../lib/client-labels";
import { formatDate } from "../../lib/dates";
import { useProjectPayments } from "./model/useProjectPayments";
import { DocumentSlots, PaymentActionCard, PaymentFocusCard } from "./ui/ProjectPaymentCards";

export default function ProjectPaymentsPage() {
  const payments = useProjectPayments();

  return (
    <main className="page payments-page" data-feature="payments">
      <div className="page-title payments-title">
        <div>
          <h1>Оплаты</h1>
          <p className="muted">{payments.project?.name ?? "Деньги, счета и документы проекта"}</p>
        </div>
      </div>
      {payments.error ? <Badge tone="danger">{payments.error}</Badge> : null}

      <section className="payment-focus" aria-label="Сводка оплат">
        <PaymentFocusCard
          icon={<WalletCards size={22} />}
          title="Ждем денег"
          value={`${payments.totals.waitingCount} · ${formatMoney(payments.totals.waitingAmount)}`}
          hint="Счета уже отправлены, осталось получить оплату."
          tone="money"
        />
        <PaymentFocusCard
          icon={<ReceiptText size={22} />}
          title="Ждут счета"
          value={`${payments.totals.invoiceCount} · ${formatMoney(payments.totals.invoiceAmount)}`}
          hint="План по услугам, где счет еще не прошел сценарий услуги."
          tone="invoice"
        />
        <PaymentFocusCard
          icon={<FileUp size={22} />}
          title="Нужны документы"
          value={`${payments.totals.documentCount}`}
          hint="В списке ниже видно, чего не хватает по каждому клиенту."
          tone="docs"
        />
      </section>

      <section className="payment-board" aria-label="Очередь оплат">
        <section className="payment-lane payment-lane--primary">
          <header>
            <div>
              <h2>Ждем денег</h2>
              <p>Клиент уже получил счет. Дальше только оплата и закрывающие документы.</p>
            </div>
            <Badge tone="success">{payments.waitingMoney.length}</Badge>
          </header>
          <div className="payment-card-list">
            {payments.waitingMoney.length === 0 ? (
              <div className="empty-state">Нет счетов, по которым ждем деньги</div>
            ) : payments.waitingMoney.map((item) => (
              <PaymentActionCard
                key={item.payment_action_id}
                action={item}
                onClient={() => payments.selectClient(item.client_id)}
                onPayment={() => payments.selectPaymentAction(item.payment_action_id)}
                uploadingDocumentKey={payments.uploadingDocumentKey}
                onDocument={(kind, file) => void payments.uploadActionDocument(item.payment_action_id, kind, file)}
              />
            ))}
          </div>
        </section>

        <section className="payment-lane">
          <header>
            <div>
              <h2>Ждут счета</h2>
              <p>Запланированные услуги. Счет появится после загрузки файла или интеграции по сценарию услуги.</p>
            </div>
            <Badge tone="warning">{payments.needInvoice.length}</Badge>
          </header>
          <div className="payment-card-list">
            {payments.needInvoice.length === 0 ? (
              <div className="empty-state">Нет услуг, которые ждут счет</div>
            ) : payments.needInvoice.map((item) => {
              const busyKey = `${item.client_id}:${item.due_date}`;
              const preparedBusyKey = `${item.client_id}:${item.due_date}:prepared`;
              const plannedAction = item.payment_action_id ? payments.actionById.get(item.payment_action_id) : undefined;
              const invoiceBusyKey = item.payment_action_id ? `${item.payment_action_id}:invoice` : `schedule:${item.client_id}:${item.due_date}:invoice`;
              const invoiceUploading = payments.uploadingDocumentKey === invoiceBusyKey;
              const invoicePrepared = plannedAction?.status === "sent" || plannedAction?.status === "snoozed";
              return (
                <article className="payment-card payment-card--plan" key={`${item.client_id}:${item.due_date}`}>
                  <div className="payment-card-main">
                    <Link className="payment-client" href={`/clients/${item.client_id}`} onClick={() => payments.selectClient(item.client_id)}>
                      {item.client_name}
                    </Link>
                    <strong>{formatMoney(item.total_amount)}</strong>
                    <span>{formatDate(item.due_date)} · {serviceCountLabel(item.service_count)} · {invoiceScenarioLabel(item.invoice_scenario_key)}</span>
                  </div>
                  <div className="payment-next-step">
                    <AlertTriangle size={16} aria-hidden />
                    <span>{invoicePrepared ? "Счет подготовлен. Загрузите файл счета, если его нет." : "Счет еще не подготовлен по сценарию услуги."}</span>
                  </div>
                  {item.payment_action_id ? (
                    <DocumentSlots
                      documents={plannedAction?.documents ?? []}
                      paid={plannedAction?.status === "paid"}
                      paymentActionId={item.payment_action_id}
                      uploadingDocumentKey={payments.uploadingDocumentKey}
                      onDocument={(kind, file) => void payments.uploadActionDocument(item.payment_action_id || "", kind, file)}
                    />
                  ) : null}
                  <div className="payment-card-actions">
                    {!invoicePrepared ? (
                      <Button size="sm" onClick={() => void payments.markInvoicePrepared(item)} disabled={payments.busyActionKey === preparedBusyKey}>
                        {payments.busyActionKey === preparedBusyKey ? "Сохраняем..." : "Счет подготовлен"}
                      </Button>
                    ) : null}
                    {!item.payment_action_id ? (
                      <label className={`document-upload-button ${invoiceUploading ? "is-busy" : ""}`}>
                        <FileUp size={14} aria-hidden />
                        {invoiceUploading ? "Загрузка..." : "Загрузить счет"}
                        <input
                          type="file"
                          hidden
                          onChange={(event) => {
                            void payments.uploadScheduleInvoice(item, event.target.files?.[0]);
                            event.currentTarget.value = "";
                          }}
                        />
                      </label>
                    ) : null}
                      <Button size="sm" variant="secondary" onClick={() => void payments.openPaymentAction(item)} disabled={payments.busyActionKey === busyKey}>
                        {payments.busyActionKey === busyKey ? "Открываем..." : item.payment_action_id ? "Открыть счёт" : "Подготовить счёт"}
                      </Button>
                  </div>
                </article>
              );
            })}
          </div>
        </section>
      </section>
    </main>
  );
}
