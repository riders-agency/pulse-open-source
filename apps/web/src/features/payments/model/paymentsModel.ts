import type { PaymentAction, PaymentDocument } from "../../../lib/api";

export type DocumentSlot = {
  kind: "invoice" | "receipt" | "act";
  title: string;
  hint: string;
};

export const documentSlots: DocumentSlot[] = [
  { kind: "invoice", title: "Счет клиенту", hint: "Файл, который отправили клиенту на оплату." },
  { kind: "receipt", title: "Платежка / подтверждение", hint: "Банк, платежка, чек или другое подтверждение поступления." },
  { kind: "act", title: "Акт / УПД", hint: "Закрывающий документ после оплаты." },
];

export function compareDueDate(a: PaymentAction, b: PaymentAction) {
  return String(a.due_date || "").localeCompare(String(b.due_date || ""));
}

export function missingDocumentLabel(action: PaymentAction) {
  const docs = action.documents ?? [];
  if (!hasDocument(docs, "invoice")) return "счет клиенту";
  if (!hasDocument(docs, "receipt") && action.status !== "paid") return "платежка или подтверждение";
  if (action.status === "paid" && !hasDocument(docs, "act")) return "акт / УПД";
  return "";
}

export function nextMissingDocument(documents: PaymentDocument[], status?: string) {
  if (!latestDocument(documents, "invoice")) return "Нужен счет клиенту";
  if (!latestDocument(documents, "receipt") && status !== "paid") return "Нужна платежка или подтверждение";
  if (status === "paid" && !latestDocument(documents, "act")) return "Нужен акт / УПД";
  return "";
}

export function hasDocument(documents: PaymentDocument[], kind: string) {
  return documents.some((item) => item.kind === kind);
}

export function latestDocument(documents: PaymentDocument[], kind: string) {
  return documents.find((item) => item.kind === kind);
}

export function itemSummary(action: PaymentAction) {
  const titles = (action.items ?? []).map((item) => item.title).filter(Boolean);
  if (!titles.length) return "услуги";
  if (titles.length === 1) return titles[0];
  return `${titles[0]} +${titles.length - 1}`;
}

export function serviceCountLabel(count: number) {
  const n = count || 0;
  if (n === 1) return "1 услуга";
  if (n > 1 && n < 5) return `${n} услуги`;
  return `${n} услуг`;
}

export function paymentActionPath(paymentActionId: string) {
  return `/payment-actions/${paymentActionId}`;
}

export function formatBytes(value: number) {
  if (!value) return "0 Б";
  if (value < 1024) return `${value} Б`;
  if (value < 1024 * 1024) return `${Math.round(value / 1024)} КБ`;
  return `${(value / 1024 / 1024).toFixed(1)} МБ`;
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB", minimumFractionDigits: 0, maximumFractionDigits: 2 }).format(value || 0);
}
