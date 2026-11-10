import { ApiError, apiBase, type SiteChatCallSession, type SiteVisitorSession } from "../../../lib/api";

export function shortURL(value?: string) {
  if (!value) return "Нет страницы";
  try {
    const url = new URL(value);
    return `${url.hostname}${url.pathname}`;
  } catch {
    return value;
  }
}

export function operatorPreviewURL(value: string) {
  if (!value) return "";
  try {
    const url = new URL(value);
    url.searchParams.set("rp_operator_preview", "1");
    url.searchParams.set("rp_webvisor", "1");
    return url.toString();
  } catch {
    return value;
  }
}

export function attachmentURL(value?: string) {
  if (!value) return "";
  if (/^https?:\/\//i.test(value)) return value;
  return `${apiBase}${value}`;
}

export function formatBytes(bytes?: number) {
  const size = Number(bytes || 0);
  if (size < 1024) return `${size} Б`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} КБ`;
  return `${(size / 1024 / 1024).toFixed(size > 10 * 1024 * 1024 ? 0 : 1)} МБ`;
}

export function truncateFileName(name: string, max = 28) {
  if (name.length <= max) return name;
  const dot = name.lastIndexOf(".");
  const ext = dot > 0 ? name.slice(dot) : "";
  const stem = dot > 0 ? name.slice(0, dot) : name;
  return `${stem.slice(0, Math.max(8, max - ext.length - 3))}...${ext}`;
}

export function formatDuration(totalSeconds?: number) {
  const seconds = Math.max(0, Math.round(Number(totalSeconds || 0)));
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  if (hours > 0) return `${hours}ч ${minutes % 60}м`;
  if (minutes > 0) return `${minutes}м ${seconds % 60}с`;
  return `${seconds}с`;
}

export function formatDurationMinutes(totalSeconds?: number) {
  const seconds = Math.max(0, Math.round(Number(totalSeconds || 0)));
  if (seconds < 60) return "менее 1 мин";
  return `${Math.floor(seconds / 60)} мин`;
}

export function isOnline(session: SiteVisitorSession) {
  if (session.status === "ended") return false;
  const activity = session.visitor_state_snapshot?.activity;
  if (activity?.connection_status === "offline") return false;
  if (activity?.connection_status === "online") return true;
  const ts = Date.parse(session.last_activity_at || "");
  return Number.isFinite(ts) && Date.now() - ts < 30000;
}

export function visitorTitle(session?: SiteVisitorSession | null) {
  if (!session) return "Выберите посетителя";
  const name = session.visitor_name?.trim();
  const contact = session.visitor_contact?.trim();
  if (name && contact) return `${name} · ${contact}`;
  if (name) return name;
  if (contact) return contact;
  return session.current_title || shortURL(session.current_url) || `Посетитель ${session.visitor_id.slice(0, 8)}`;
}

export function visitorContactLine(session?: SiteVisitorSession | null) {
  if (!session) return "";
  return [session.visitor_name, session.visitor_contact].map((value) => value?.trim()).filter(Boolean).join(" · ");
}

export function callActionErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    if (error.status === 409 || error.code === "visitor_offline") {
      return "Посетитель уже ушел со страницы. Звонок можно начать только пока посетитель на сайте.";
    }
    return error.detail || error.message || "Звонок не запустился.";
  }
  return "Браузер не дал доступ к микрофону оператора.";
}

export function callLabel(call?: SiteChatCallSession | null) {
  if (!call) return "Нет звонка";
  const direction = call.direction === "visitor_to_operator" ? "Входящий" : "Исходящий";
  return `${direction} · ${callStatusLabel(call.status)}`;
}

export function clamp(value: number, min: number, max: number) {
  if (!Number.isFinite(value)) return min;
  return Math.max(min, Math.min(max, value));
}

export function callStatusLabel(status: string) {
  return ({ ringing: "Ожидаем ответа", accepted: "Разговор", ended: "Завершён", rejected: "Отклонён", cancelled: "Отменён", missed: "Пропущен", failed: "Ошибка соединения" } as Record<string, string>)[status] || "Подключение";
}
