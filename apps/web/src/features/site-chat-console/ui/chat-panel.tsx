import { type KeyboardEvent as ReactKeyboardEvent, useEffect, useRef } from "react";
import { FileText, ImageIcon, Paperclip, Send, X } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { Textarea } from "../../../components/ui/input";
import type { SiteChatAttachment, SiteChatMessage } from "../../../lib/api";
import { formatDate } from "../../../lib/utils";
import { attachmentURL, formatBytes, truncateFileName } from "../model/format";
import type { PendingAttachment } from "../model/types";

export function ChatPanel({
  messages,
  draft,
  setDraft,
  pendingAttachments,
  onFiles,
  onRemoveAttachment,
  onSend,
  disabled,
}: {
  messages: SiteChatMessage[];
  draft: string;
  setDraft: (value: string) => void;
  pendingAttachments: PendingAttachment[];
  onFiles: (files: FileList | File[]) => void;
  onRemoveAttachment: (id: string) => void;
  onSend: () => void;
  disabled: boolean;
}) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const threadRef = useRef<HTMLDivElement | null>(null);
  const atBottom = useRef(true);
  const lastMessageID = messages.at(-1)?.message_id;
  useEffect(() => {
    if (atBottom.current && threadRef.current) threadRef.current.scrollTop = threadRef.current.scrollHeight;
  }, [lastMessageID]);
  const canSend = !disabled && (!!draft.trim() || pendingAttachments.some((item) => item.attachment));
  const onComposerKeyDown = (event: ReactKeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key !== "Enter" || event.shiftKey || event.ctrlKey || event.metaKey || event.nativeEvent.isComposing) return;
    event.preventDefault();
    if (canSend) onSend();
  };
  return (
    <section
      className="live-chat-panel"
      onDragOver={(event) => event.preventDefault()}
      onDrop={(event) => {
        event.preventDefault();
        if (!disabled && event.dataTransfer.files.length) onFiles(event.dataTransfer.files);
      }}
    >
      <div className="live-chat-thread" role="log" aria-label="Переписка с посетителем" aria-live="polite" ref={threadRef} onScroll={() => { const el = threadRef.current; if (el) atBottom.current = el.scrollHeight - el.scrollTop - el.clientHeight < 80; }}>
        {messages.length === 0 ? <div className="empty-state">Сообщений пока нет. Представьтесь и предложите помощь посетителю.</div> : null}
        {messages.map((message) => (
          <article className={`live-chat-bubble ${message.author_kind === "operator" ? "mine" : message.author_kind === "system" ? "system" : "their"}`} key={message.message_id}>
            <span>{message.author_kind === "system" ? "Событие" : message.author_name || (message.author_kind === "operator" ? "Оператор" : "Посетитель")} · {formatDate(message.created_at)}</span>
            {message.body ? <p>{message.body}</p> : null}
            <AttachmentGallery attachments={message.attachments ?? []} />
          </article>
        ))}
      </div>
      {pendingAttachments.length ? (
        <div className="pending-attachments">
          {pendingAttachments.map((item) => <PendingAttachmentChip key={item.id} item={item} onRemove={() => onRemoveAttachment(item.id)} />)}
        </div>
      ) : null}
      <footer className="live-chat-composer">
        <Textarea
          aria-label="Сообщение посетителю"
          disabled={disabled}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onPaste={(event) => {
            if (!disabled && event.clipboardData.files.length) onFiles(event.clipboardData.files);
          }}
          onKeyDown={onComposerKeyDown}
          rows={3}
          placeholder="Сообщение посетителю"
        />
        <input ref={fileInputRef} type="file" multiple hidden onChange={(event) => {
          if (event.target.files?.length) onFiles(event.target.files);
          event.currentTarget.value = "";
        }} />
        <Button variant="secondary" onClick={() => fileInputRef.current?.click()} disabled={disabled} aria-label="Прикрепить файл" title="Прикрепить файл"><Paperclip size={16} /></Button>
        <Button onClick={onSend} disabled={!canSend}><Send size={16} /> Отправить</Button>
        <small className="live-composer-hint">Enter — отправить · Shift + Enter — новая строка</small>
      </footer>
    </section>
  );
}

export function AttachmentGallery({ attachments }: { attachments: SiteChatAttachment[] }) {
  if (!attachments.length) return null;
  return (
    <div className="chat-attachments">
      {attachments.map((attachment) => {
        const url = attachmentURL(attachment.public_url);
        if (attachment.kind === "image") {
          return (
            <a className="chat-attachment image" href={url} target="_blank" rel="noreferrer" key={attachment.attachment_id}>
              <img src={url} alt={attachment.file_name} />
            </a>
          );
        }
        if (attachment.kind === "audio") {
          return (
            <div className="chat-attachment audio" key={attachment.attachment_id}>
              <span>{attachment.file_name}</span>
              <audio controls src={url} />
            </div>
          );
        }
        return (
          <a className="chat-attachment file" href={url} target="_blank" rel="noreferrer" key={attachment.attachment_id}>
            <FileText size={16} />
            <span>{truncateFileName(attachment.file_name)}</span>
            <small>{formatBytes(attachment.size_bytes)}</small>
          </a>
        );
      })}
    </div>
  );
}

function PendingAttachmentChip({ item, onRemove }: { item: PendingAttachment; onRemove: () => void }) {
  return (
    <button type="button" className={`pending-attachment ${item.error ? "is-error" : ""}`} onClick={onRemove}>
      {item.localUrl ? <img src={item.localUrl} alt="" /> : item.file.type.startsWith("image/") ? <ImageIcon size={18} /> : <FileText size={18} />}
      <span>{truncateFileName(item.file.name)}<small>{item.error || (item.uploading ? "Загрузка" : formatBytes(item.file.size))}</small></span>
      <X size={14} />
    </button>
  );
}
