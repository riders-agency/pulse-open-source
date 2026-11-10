import type { Client, ClientService } from "../../../lib/api";
import { advanceDate, dateBefore, todayDateInput } from "../../../lib/dates";

export type EditForm = {
  client_type: Client["client_type"];
  status: string;
  name: string;
  short_name: string;
  full_name: string;
  inn: string;
  kpp: string;
  ogrn: string;
  ogrnip: string;
  legal_address: string;
  actual_address: string;
  okved: string;
  comment: string;
};

export type ContactForm = {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  telegram: string;
};

export type PaymentForm = {
  client_service_id: string;
  amount: string;
  period_start: string;
  period_end: string;
  next_due_date: string;
  paid_at: string;
  payment_method: string;
  comment: string;
};

export type ResponsibleForm = {
  display_name: string;
  role: string;
  company_legal_entity_id: string;
  is_primary: boolean;
};

export function formFromClient(client: Client): EditForm {
  return {
    client_type: client.client_type,
    status: client.status || "active",
    name: client.name || "",
    short_name: client.short_name || "",
    full_name: client.full_name || "",
    inn: client.inn || "",
    kpp: client.kpp || "",
    ogrn: client.ogrn || "",
    ogrnip: client.ogrnip || "",
    legal_address: client.legal_address || "",
    actual_address: client.actual_address || "",
    okved: client.okved || "",
    comment: client.comment || "",
  };
}

export function emptyContactForm(): ContactForm {
  return { first_name: "", last_name: "", email: "", phone: "", telegram: "" };
}

export function emptyResponsibleForm(): ResponsibleForm {
  return { display_name: "", role: "account_manager", company_legal_entity_id: "", is_primary: false };
}

export function emptyPaymentForm(): PaymentForm {
  return {
    client_service_id: "",
    amount: "",
    period_start: "",
    period_end: "",
    next_due_date: "",
    paid_at: todayDateInput(),
    payment_method: "legal_invoice",
    comment: "",
  };
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB", maximumFractionDigits: 0 }).format(value || 0);
}

export function shortID(value?: string) {
  if (!value) return "—";
  return value.slice(0, 8);
}

export function paymentCycleFields(service: ClientService, periodStart: string) {
  const start = periodStart || service.next_due_date || todayDateInput();
  const next = advanceDate(start, service.billing_period);
  return {
    period_start: start,
    period_end: next ? dateBefore(next) : start,
    next_due_date: next || "",
  };
}

export function manualPaymentIdempotencyKey(clientId: string, clientServiceId: string, periodStart: string, amount: number) {
  return ["manual-payment", clientId, clientServiceId, periodStart, String(amount)].join("-");
}
