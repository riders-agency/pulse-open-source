import { CheckCircle2, Download, FileText, FileUp } from "lucide-react";

import type { PaymentDocument } from "../../../lib/api";
import { paymentDocumentHref } from "../api/paymentsApi";
import { formatBytes, type DocumentSlot } from "../model/paymentsModel";

export function PaymentDocumentSlotCard({
  slot,
  document,
  uploading,
  onFile,
}: {
  slot: DocumentSlot;
  document?: PaymentDocument;
  uploading: boolean;
  onFile: (file?: File) => void;
}) {
  return (
    <article className={`payment-document-slot ${document ? "is-done" : ""}`}>
      <div className="payment-document-icon">
        {document ? <CheckCircle2 size={18} aria-hidden /> : <FileText size={18} aria-hidden />}
      </div>
      <div className="payment-document-main">
        <strong>{slot.title}</strong>
        <span>{document ? document.file_name : slot.hint}</span>
        {document ? <small>{formatBytes(document.size_bytes)}</small> : null}
      </div>
      <div className="payment-document-actions">
        {document?.download_url ? (
          <a className="action-link" href={paymentDocumentHref(document.download_url)} target="_blank" rel="noreferrer">
            <Download size={14} aria-hidden /> Скачать
          </a>
        ) : null}
        <label className={`document-upload-button ${uploading ? "is-busy" : ""}`}>
          <FileUp size={14} aria-hidden />
          {uploading ? "Загрузка..." : document ? "Заменить" : "Загрузить"}
          <input
            type="file"
            hidden
            onChange={(event) => {
              onFile(event.target.files?.[0]);
              event.currentTarget.value = "";
            }}
          />
        </label>
      </div>
    </article>
  );
}
