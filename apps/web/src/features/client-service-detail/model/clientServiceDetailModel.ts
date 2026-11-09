import type { ClientService } from "../../../lib/api";
import { invoiceScenarioFromMetadata, invoiceScenarioOptions } from "../../../lib/client-labels";
import { toDateInput } from "../../../lib/dates";

export type ServiceForm = {
  project_id: string;
  service_type_key: string;
  name: string;
  status: string;
  amount: string;
  billing_period: string;
  next_due_date: string;
  invoice_scenario_key: string;
};

export type ClientServiceCreateForm = Omit<ServiceForm, "status">;

export const emptyClientServiceCreateForm: ClientServiceCreateForm = {
  project_id: "",
  service_type_key: "support",
  name: "",
  amount: "",
  billing_period: "monthly",
  next_due_date: "",
  invoice_scenario_key: String(invoiceScenarioOptions[0].key),
};

export function formFromService(service: ClientService): ServiceForm {
  return {
    project_id: service.project_id,
    service_type_key: service.service_type_key,
    name: service.name || "",
    status: service.status || "active",
    amount: String(service.amount ?? ""),
    billing_period: service.billing_period || "monthly",
    next_due_date: toDateInput(service.next_due_date),
    invoice_scenario_key: invoiceScenarioFromMetadata(service.metadata_json),
  };
}

export function clientServiceCreatePayload(form: ClientServiceCreateForm) {
  return {
    project_id: form.project_id,
    service_type_key: form.service_type_key,
    name: form.name.trim(),
    amount: form.amount ? Number(form.amount.replace(",", ".")) : undefined,
    currency: "RUB",
    billing_period: form.billing_period,
    next_due_date: form.next_due_date,
    metadata_json: {
      invoice_scenario: {
        key: form.invoice_scenario_key,
        source: "service",
      },
    },
  };
}

export function serviceFormPayload(form: ServiceForm, metadata: Record<string, unknown> = {}) {
  return {
    project_id: form.project_id,
    service_type_key: form.service_type_key,
    name: form.name.trim(),
    status: form.status,
    amount: form.amount ? Number(form.amount.replace(",", ".")) : undefined,
    currency: "RUB",
    billing_period: form.billing_period,
    next_due_date: form.next_due_date,
    metadata_json: {
      ...metadata,
      invoice_scenario: {
        key: form.invoice_scenario_key,
        source: "service",
      },
    },
  };
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB", maximumFractionDigits: 0 }).format(value || 0);
}
