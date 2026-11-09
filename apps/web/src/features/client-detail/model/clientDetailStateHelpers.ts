import type { ClientService, PaymentAction } from "../../../lib/api";
import { isBeforeDate, todayDateInput } from "../../../lib/dates";
import { emptyPaymentForm, paymentCycleFields } from "./clientDetailModel";

export function filterPayments(payments: PaymentAction[], paymentFilter: string) {
  if (paymentFilter === "all") return payments;
  if (paymentFilter === "manual") return payments.filter((item) => item.source === "manual");
  if (paymentFilter === "bank") return payments.filter((item) => item.source === "bank");
  if (paymentFilter === "email") return payments.filter((item) => item.source === "email" || item.source === "incoming_email");
  if (paymentFilter === "integration") return payments.filter((item) => ["integration", "api_key", "kontur_elba", "other_integration"].includes(item.source));
  if (paymentFilter === "overdue") return payments.filter((item) => item.status !== "paid" && isBeforeDate(item.due_date, todayDateInput()));
  return payments.filter((item) => item.status === paymentFilter);
}

export function defaultPaymentForm(current: ReturnType<typeof emptyPaymentForm>, services: ClientService[]) {
  if (current.client_service_id || services.length === 0) return current;
  const first = services.find((item) => item.status === "active") ?? services[0];
  return {
    ...current,
    client_service_id: first.client_service_id,
    amount: String(first.amount || ""),
    ...paymentCycleFields(first, first.next_due_date || current.period_start || todayDateInput()),
  };
}

export async function runClientAction(action: () => Promise<void>, setError: (value: string) => void, fallback: string) {
  setError("");
  try {
    await action();
  } catch (err) {
    setError(err instanceof Error ? err.message : fallback);
  }
}

export function validatePayment(clientId: string, services: ClientService[], paymentForm: ReturnType<typeof emptyPaymentForm>) {
  if (!clientId) return "";
  if (services.length === 0) return "У клиента пока нет услуг. Чтобы внести оплату по услуге, сначала добавьте услугу.";
  if (!paymentForm.client_service_id || !paymentForm.period_start || !paymentForm.period_end || !paymentForm.amount) return "Выберите услугу, сумму и период оплаты.";
  if (isBeforeDate(paymentForm.period_end, paymentForm.period_start)) return "Дата окончания периода не может быть раньше даты начала.";
  const amount = Number(paymentForm.amount.replace(",", "."));
  if (!Number.isFinite(amount) || amount <= 0) return "Сумма оплаты должна быть больше нуля.";
  return "";
}
