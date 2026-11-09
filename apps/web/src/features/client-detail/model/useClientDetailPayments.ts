import { useState } from "react";

import type { ClientService } from "../../../lib/api";
import { todayDateInput } from "../../../lib/dates";
import { recordClientManualPayment } from "../api/clientDetailApi";
import {
  manualPaymentIdempotencyKey,
  paymentCycleFields,
  type PaymentForm,
} from "./clientDetailModel";
import { validatePayment } from "./clientDetailStateHelpers";

export function useClientDetailPayments({
  activeProjectId,
  clientId,
  load,
  paymentForm,
  services,
  setError,
  setPaymentForm,
  token,
}: {
  activeProjectId: string;
  clientId?: string;
  load: () => void;
  paymentForm: PaymentForm;
  services: ClientService[];
  setError: (value: string) => void;
  setPaymentForm: (update: (current: PaymentForm) => PaymentForm) => void;
  token?: string | null;
}) {
  const [paymentBusy, setPaymentBusy] = useState(false);

  const selectPaymentService = (service: ClientService) => {
    setPaymentForm((current) => ({
      ...current,
      client_service_id: service.client_service_id,
      amount: String(service.amount || ""),
      ...paymentCycleFields(service, service.next_due_date || current.period_start || todayDateInput()),
    }));
  };

  const updatePaymentPeriodStart = (periodStart: string) => {
    const service = services.find((item) => item.client_service_id === paymentForm.client_service_id);
    setPaymentForm((current) => ({ ...current, ...(service ? paymentCycleFields(service, periodStart) : { period_start: periodStart }) }));
  };

  const recordPayment = async () => {
    const validation = validatePayment(clientId, services, paymentForm);
    if (validation) {
      setError(validation);
      return;
    }
    const amount = Number(paymentForm.amount.replace(",", "."));
    const idempotencyKey = manualPaymentIdempotencyKey(clientId, paymentForm.client_service_id, paymentForm.period_start, amount);
    setPaymentBusy(true);
    setError("");
    try {
      if (!token) return;
      await recordClientManualPayment(clientId, activeProjectId, token, {
        client_service_id: paymentForm.client_service_id,
        amount,
        currency: "RUB",
        period_start: paymentForm.period_start,
        period_end: paymentForm.period_end || undefined,
        paid_at: paymentForm.paid_at,
        payment_method: paymentForm.payment_method,
        next_due_date: paymentForm.next_due_date || undefined,
        comment: paymentForm.comment.trim(),
        idempotency_key: idempotencyKey,
      });
      setPaymentForm((current) => ({ ...current, comment: "" }));
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось внести оплату");
    } finally {
      setPaymentBusy(false);
    }
  };

  return { paymentBusy, recordPayment, selectPaymentService, updatePaymentPeriodStart };
}
