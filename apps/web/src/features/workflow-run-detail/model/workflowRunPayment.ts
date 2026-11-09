import type { ActionRun } from "../../../lib/api";

type PaymentRunInfo =
  | { kind: "event"; clientName: string; serviceName: string; amount: number; statusLabel: string; nextStep: string }
  | { kind: "report"; waitingHint: string; invoiceHint: string; digest: string };

export function paymentRunInfo(runInput: unknown, selected?: ActionRun): PaymentRunInfo | null {
  const input = asRecord(runInput);
  const selectedInput = asRecord(selected?.input_json);
  const output = asRecord(selected?.output_json);
  const current =
    recordAt(input, "current_payment") ??
    recordAt(input, "payment_event") ??
    recordAt(input, "payment_action") ??
    recordAt(selectedInput, "current_payment") ??
    recordAt(selectedInput, "payment_event") ??
    recordAt(output, "current_payment");
  if (current) {
    const status = firstText(stringAt(current, "status_label"), statusText(stringAt(current, "status")));
    const amount = firstNumber(numberAt(current, "amount"), numberAt(current, "total"), numberAt(current, "total_amount"));
    const client = firstText(stringAt(current, "client_name"), stringAt(recordAt(current, "client"), "name"), "Клиент");
    const service = firstText(stringAt(current, "service_name"), stringAt(current, "service_summary"));
    return {
      kind: "event",
      clientName: client,
      serviceName: service,
      amount,
      statusLabel: status || "Оплата",
      nextStep: status === "Оплачен" || stringAt(current, "status") === "paid"
        ? "Счет оплачен. Проверьте акт / УПД в карточке оплаты."
        : "Счет в очереди “Ждем денег”. Следующий шаг — отметить оплату или загрузить подтверждение.",
    };
  }
  const isPaymentReport =
    selected?.action_type_key === "clients.payment_due.trigger" ||
    selected?.action_type_key === "clients.payment_cycle.collect" ||
    stringAt(recordAt(input, "_pulse"), "source_type") === "client_payment_due";
  if (!isPaymentReport) return null;
  return {
    kind: "report",
    waitingHint: "клиенты со счетами",
    invoiceHint: "клиенты без счета",
    digest: compactDigest(firstText(stringAt(output, "digest_plain"), stringAt(output, "digest_max"), stringAt(output, "digest_telegram"))),
  };
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

function recordAt(root: Record<string, unknown>, key: string) {
  const value = root[key];
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : null;
}

function stringAt(root: Record<string, unknown> | null, key: string) {
  const value = root?.[key];
  return typeof value === "string" ? value.trim() : "";
}

function numberAt(root: Record<string, unknown>, key: string) {
  const value = root[key];
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return 0;
}

function firstText(...values: string[]) {
  return values.find((value) => value.trim())?.trim() ?? "";
}

function firstNumber(...values: number[]) {
  return values.find((value) => value > 0) ?? 0;
}

function statusText(status: string) {
  const labels: Record<string, string> = {
    paid: "Оплачен",
    sent: "Ждем денег",
    snoozed: "Отложено",
    pending: "Ждет счет",
  };
  return labels[status] ?? status;
}

function compactDigest(value: string) {
  const normalized = value.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
  if (normalized.length <= 220) return normalized;
  return `${normalized.slice(0, 220)}...`;
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB", maximumFractionDigits: 0 }).format(value || 0);
}
