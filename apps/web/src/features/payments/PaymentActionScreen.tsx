"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import type { PaymentAction } from "../../lib/api";
import { confirmPaymentAction, getPaymentAction, uploadPaymentActionDocument } from "./api/paymentsApi";
import {
  documentSlots,
  formatMoney,
  latestDocument,
  nextMissingDocument,
} from "./model/paymentsModel";
import { PaymentDocumentSlotCard } from "./ui/PaymentDocumentSlotCard";
import { PaymentConfirmCard } from "./ui/PaymentConfirmCard";
import { PaymentItemsCard } from "./ui/PaymentItemsCard";
import { BankInvoiceCard } from "./ui/BankInvoiceCard";
import { paymentStatusLabel, statusTone } from "../../lib/client-labels";
import { addDaysDate, formatDate, todayDateInput } from "../../lib/dates";
import { usePulseStore } from "../../store/pulse-store";

export default function PaymentActionPage() {
  const params = useParams<{ paymentActionId?: string }>();
  const token = usePulseStore((s) => s.token);
  const activePaymentActionId = usePulseStore((s) => s.activePaymentActionId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setActivePaymentAction = usePulseStore((s) => s.setActivePaymentAction);
  const [action, setAction] = useState<PaymentAction | null>(null);
  const [paidAt, setPaidAt] = useState(() => todayDateInput());
  const [amount, setAmount] = useState("");
  const [paymentMethod, setPaymentMethod] = useState("legal_invoice");
  const [nextDueDate, setNextDueDate] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [uploadingKind, setUploadingKind] = useState("");
  const paymentActionId = params.paymentActionId || activePaymentActionId;
  const loadVersion = useRef(0);

  const load = useCallback(() => {
    const version = ++loadVersion.current;
    if (!token) return;
    if (!paymentActionId) {
      setError("Не выбрана оплата. Откройте ее из очереди оплат.");
      return;
    }
    getPaymentAction(paymentActionId, token)
      .then((row) => {
        if (version !== loadVersion.current) return;
        setAction(row);
        setActiveProject(row.project_id);
        if (params.paymentActionId) setActivePaymentAction(row.payment_action_id);
        setAmount(String(row.total_amount || ""));
        setNextDueDate((current) => current || addDaysDate(row.period_end, 1));
      })
      .catch((err) => { if (version === loadVersion.current) setError(err.message); });
  }, [params.paymentActionId, paymentActionId, setActivePaymentAction, setActiveProject, token]);

  useEffect(() => { load(); return () => { loadVersion.current++; }; }, [load]);

  const uploadDocument = async (kind: string, file?: File) => {
    if (!token || !paymentActionId || !file) return;
    setUploadingKind(kind);
    setError("");
    try {
      const form = new FormData();
      form.append("file", file);
      form.append("kind", kind);
      await uploadPaymentActionDocument(paymentActionId, token, form);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить документ");
    } finally {
      setUploadingKind("");
    }
  };

  const confirm = async () => {
    if (!token || !paymentActionId) return;
    setError("");
    setBusy(true);
    try {
      await confirmPaymentAction(paymentActionId, token, {
        amount: amount ? Number(amount) : undefined,
        paid_at: paidAt,
        source: "manual",
        payment_method: paymentMethod,
        next_due_date: nextDueDate || undefined,
      });
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось отметить оплату");
    } finally {
      setBusy(false);
    }
  };

  const documents = action?.documents ?? [];

  return (
    <main className="page payment-detail-page" data-feature="payments">
      <div className="page-title">
        <div>
          <h1>{action?.client_name ?? "Оплата"}</h1>
          <p className="muted">{action ? `${action.project_name} · счет на ${formatMoney(action.total_amount)}` : "Оплата клиента"}</p>
        </div>
        {action ? <Link href={`/projects/${action.project_id}/payments`}><Button variant="secondary">К оплатам</Button></Link> : null}
      </div>
      {error ? <Badge tone="danger">{error}</Badge> : null}

      <section className="payment-focus payment-focus--compact" aria-label="Состояние оплаты">
        <div className="payment-focus-card payment-focus-card--money">
          <span>Статус</span>
          <strong><Badge tone={statusTone(action?.status ?? "")}>{paymentStatusLabel(action?.status)}</Badge></strong>
          <small>{action?.status === "paid" ? "Оплата закрыта" : "Деньги еще ждем"}</small>
        </div>
        <div className="payment-focus-card payment-focus-card--invoice">
          <span>Сумма</span>
          <strong>{formatMoney(action?.total_amount ?? 0)}</strong>
          <small>{formatDate(action?.due_date ?? "")}</small>
        </div>
        <div className="payment-focus-card payment-focus-card--docs">
          <span>Документы</span>
          <strong>{documents.length} / 3</strong>
          <small>{nextMissingDocument(documents, action?.status) || "Все основные документы есть"}</small>
        </div>
      </section>

      <section className="payment-detail-layout">
        <div className="payment-detail-main">
          <PaymentItemsCard action={action} />
          {action && action.payment_action_id === paymentActionId && token ? <BankInvoiceCard key={`${action.payment_action_id}:${token}`} id={action.payment_action_id} token={token} paymentStatus={action.status} /> : null}

          <Card>
            <CardHeader>
              <CardTitle>Документы</CardTitle>
              <CardDescription>Счет, подтверждение оплаты и закрывающие документы хранятся прямо в оплате.</CardDescription>
            </CardHeader>
            <CardContent className="payment-document-list">
              {documentSlots.map((slot) => (
                <PaymentDocumentSlotCard
                  key={slot.kind}
                  slot={slot}
                  document={latestDocument(documents, slot.kind)}
                  uploading={uploadingKind === slot.kind}
                  onFile={(file) => uploadDocument(slot.kind, file)}
                />
              ))}
            </CardContent>
          </Card>
        </div>

        <PaymentConfirmCard
          action={action}
          amount={amount}
          busy={busy}
          nextDueDate={nextDueDate}
          paidAt={paidAt}
          paymentMethod={paymentMethod}
          setAmount={setAmount}
          setNextDueDate={setNextDueDate}
          setPaidAt={setPaidAt}
          setPaymentMethod={setPaymentMethod}
          onConfirm={confirm}
        />
      </section>
    </main>
  );
}
