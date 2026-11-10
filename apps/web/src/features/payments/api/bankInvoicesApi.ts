import { apiFetch, parseResponse } from "../../../lib/api/http";

import type { components } from "../../../lib/api/generated/openapi-types";

export type BankInvoiceInput = components["schemas"]["BankInvoiceInput"];
export type BankInvoiceDraft = components["schemas"]["BankInvoiceDraft"];
export type BankInvoiceWorkspace = components["schemas"]["BankInvoiceWorkspace"];

export async function readBankInvoice(id: string, token: string, signal: AbortSignal) {
  const path = `/api/v1/payment-actions/${id}/bank-invoice`;
  return parseResponse<BankInvoiceWorkspace>(await apiFetch(path, { token, signal }), { inlineError: true });
}

export async function saveBankInvoice(id: string, token: string, input: BankInvoiceInput, key: string, signal: AbortSignal) {
  const path = `/api/v1/payment-actions/${id}/bank-invoice`;
  return parseResponse<BankInvoiceDraft>(await apiFetch(path, {
    method: "PUT", token, json: input, signal, headers: { "Idempotency-Key": key },
  }), { inlineError: true });
}
