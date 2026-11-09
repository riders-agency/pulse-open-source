import Link from "next/link";
import { CheckCircle2, FileUp } from "lucide-react";
import type { ReactNode } from "react";

import { Badge } from "../../../components/ui/badge";
import type { PaymentAction, PaymentDocument } from "../../../lib/api";
import { paymentStatusLabel, statusTone } from "../../../lib/client-labels";
import { formatDate } from "../../../lib/dates";
import { formatMoney, hasDocument, itemSummary, missingDocumentLabel, paymentActionPath } from "../model/paymentsModel";

export function PaymentFocusCard({
  icon,
  title,
  value,
  hint,
  tone,
}: {
  icon: ReactNode;
  title: string;
  value: string;
  hint: string;
  tone: "money" | "invoice" | "docs";
}) {
  return (
    <article className={`payment-focus-card payment-focus-card--${tone}`}>
      <div>{icon}</div>
      <span>{title}</span>
      <strong>{value}</strong>
      <small>{hint}</small>
    </article>
  );
}

export function PaymentActionCard({
  action,
  onClient,
  onPayment,
  uploadingDocumentKey,
  onDocument,
}: {
  action: PaymentAction;
  onClient: () => void;
  onPayment: () => void;
  uploadingDocumentKey: string;
  onDocument: (kind: string, file?: File) => void;
}) {
  const missing = missingDocumentLabel(action);
  return (
    <article className="payment-card">
      <div className="payment-card-top">
        <div className="payment-card-main">
          <Link className="payment-client" href={`/clients/${action.client_id}`} onClick={onClient}>
            {action.client_name}
          </Link>
          <strong>{formatMoney(action.total_amount)}</strong>
          <span>{formatDate(action.due_date)} · {itemSummary(action)}</span>
        </div>
        <Badge tone={statusTone(action.status)}>{paymentStatusLabel(action.status)}</Badge>
      </div>
      <DocumentSlots
        documents={action.documents ?? []}
        paid={action.status === "paid"}
        paymentActionId={action.payment_action_id}
        uploadingDocumentKey={uploadingDocumentKey}
        onDocument={onDocument}
      />
      <div className="payment-card-footer">
        {missing ? (
          <span className="payment-next-step">
            <FileUp size={16} aria-hidden />
            <span>Нужен документ: {missing}</span>
          </span>
        ) : (
          <span className="payment-next-step is-done">
            <CheckCircle2 size={16} aria-hidden />
            <span>Документы в порядке</span>
          </span>
        )}
        <Link className="action-link" href={paymentActionPath(action.payment_action_id)} onClick={onPayment}>
          {action.status === "paid" ? "Открыть оплату" : "Счет оплачен"}
        </Link>
      </div>
    </article>
  );
}

export function DocumentSlots({
  documents,
  paid,
  paymentActionId,
  uploadingDocumentKey,
  onDocument,
}: {
  documents: PaymentDocument[];
  paid: boolean;
  paymentActionId: string;
  uploadingDocumentKey: string;
  onDocument: (kind: string, file?: File) => void;
}) {
  const slots = [
    { kind: "invoice", label: "Счет" },
    { kind: "receipt", label: "Платежка" },
    { kind: "act", label: "Акт/УПД" },
  ];
  return (
    <div className="payment-doc-slots" aria-label="Документы оплаты">
      {slots.map((slot) => {
        const done = hasDocument(documents, slot.kind) || (slot.kind === "receipt" && paid);
        const uploading = uploadingDocumentKey === `${paymentActionId}:${slot.kind}`;
        return (
          <label className={`payment-doc-slot payment-doc-slot--upload ${done ? "is-done" : ""} ${uploading ? "is-busy" : ""}`} key={slot.kind}>
            {done ? <CheckCircle2 size={14} aria-hidden /> : <FileUp size={14} aria-hidden />}
            <span>{uploading ? "Загрузка..." : slot.label}</span>
            <input
              type="file"
              hidden
              onChange={(event) => {
                onDocument(slot.kind, event.target.files?.[0]);
                event.currentTarget.value = "";
              }}
            />
          </label>
        );
      })}
    </div>
  );
}
