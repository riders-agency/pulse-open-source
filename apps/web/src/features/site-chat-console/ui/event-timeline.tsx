import type { SiteChatMessage, SiteCobrowseFrame } from "../../../lib/api";
import { formatDate } from "../../../lib/utils";

export function EventTimeline({ frames, messages, compact = false }: { frames: SiteCobrowseFrame[]; messages: SiteChatMessage[]; compact?: boolean }) {
  const items = [
    ...frames.slice(-120).map((frame) => ({ id: frame.frame_id, at: frame.occurred_at, label: frame.kind, body: JSON.stringify(frame.payload_json).slice(0, 240) })),
    ...messages.slice(-60).map((message) => ({ id: message.message_id, at: message.created_at, label: message.author_kind, body: message.body })),
  ].sort((a, b) => Date.parse(b.at) - Date.parse(a.at));
  return (
    <div className={`live-event-timeline ${compact ? "is-compact" : ""}`}>
      {items.length === 0 ? <div className="empty-state">Событий пока нет.</div> : null}
      {items.map((item) => (
        <article key={item.id}>
          <strong>{item.label}</strong>
          <span>{formatDate(item.at)}</span>
          <p>{item.body}</p>
        </article>
      ))}
    </div>
  );
}
