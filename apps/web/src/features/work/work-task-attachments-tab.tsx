"use client";

import { Download, FileUp, Paperclip } from "lucide-react";

import { Button } from "../../components/ui/button";
import type { WorkTaskAttachment, WorkTaskDetail } from "../../lib/api";
import { formatDateTime } from "../../lib/dates";

export function WorkTaskAttachmentsTab({
  canWrite,
  downloadAttachment,
  setError,
  task,
  uploadAttachment,
  uploading,
}: {
  canWrite: boolean;
  downloadAttachment: (attachment: WorkTaskAttachment) => Promise<void>;
  setError: (value: string) => void;
  task: WorkTaskDetail;
  uploadAttachment: (file?: File) => Promise<void>;
  uploading: boolean;
}) {
  return (
    <section className="work-task-tab-grid">
      {canWrite ? (
        <label className={`work-attachment-drop ${uploading ? "is-busy" : ""}`}>
          <FileUp size={22} />
          <strong>{uploading ? "Загружаем..." : "Загрузить вложение"}</strong>
          <span>Файлы до 25 МБ. Вложения видят только внутренние участники задачи.</span>
          <input
            type="file"
            hidden
            onChange={(event) => {
              const file = event.target.files?.[0];
              event.currentTarget.value = "";
              void uploadAttachment(file).catch((err) => setError(err.message));
            }}
          />
        </label>
      ) : null}
      <div className="work-attachment-list">
        {(task.attachments || []).map((attachment) => (
          <article key={attachment.task_attachment_id}>
            <Paperclip size={16} />
            <div>
              <strong>{attachment.file_name}</strong>
              <span>{formatBytes(attachment.size_bytes)} · {attachment.uploaded_by_name || "Система"} · {formatDateTime(attachment.created_at)}</span>
            </div>
            <Button variant="secondary" size="sm" type="button" onClick={() => void downloadAttachment(attachment).catch((err) => setError(err.message))}>
              <Download size={15} /> Скачать
            </Button>
          </article>
        ))}
        {(task.attachments || []).length === 0 ? <div className="work-empty">Вложений пока нет.</div> : null}
      </div>
    </section>
  );
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} Б`;
  if (value < 1024 * 1024) return `${Math.round(value / 1024)} КБ`;
  return `${(value / 1024 / 1024).toFixed(1)} МБ`;
}
