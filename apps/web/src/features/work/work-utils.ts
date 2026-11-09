import type { WorkAccess, WorkBoard, WorkTask } from "../../lib/api";
import { formatDate } from "../../lib/dates";

export type WorkEnvelope<T> = {
  access: WorkAccess;
  items?: T[];
  board?: WorkBoard;
  overview?: T;
  report?: T;
  task?: T;
  timer?: T | null;
};

export function minutesLabel(value?: number) {
  const total = Math.max(0, Number(value || 0));
  const hours = Math.floor(total / 60);
  const minutes = total % 60;
  if (!hours) return `${minutes}м`;
  if (!minutes) return `${hours}ч`;
  return `${hours}ч ${minutes}м`;
}

export function timerClock(startedAt?: string, elapsedMinutes = 0, now = Date.now()) {
  const startedMs = startedAt ? Date.parse(startedAt) : NaN;
  const elapsedMs = Number.isFinite(startedMs) ? Math.max(elapsedMinutes * 60_000, now - startedMs) : elapsedMinutes * 60_000;
  const totalSeconds = Math.max(0, Math.floor(elapsedMs / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

export function shortDate(value?: string) {
  return formatDate(value, "Без срока");
}

export function priorityTone(priority: string) {
  if (priority === "urgent" || priority === "high") return "danger";
  if (priority === "medium") return "warning";
  if (priority === "low") return "info";
  return "muted";
}

export function taskIsOverdue(task: WorkTask) {
  if (!task.due_date || task.completed_at) return false;
  const due = new Date(task.due_date);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return due < today;
}

export function visible(access: WorkAccess | undefined, module: string) {
  return Boolean(access?.visible_modules?.includes(module));
}

export function taskTypeLabel(value?: string) {
  return {
    task: "Задача",
    bug: "Баг",
    feature: "Фича",
    request: "Запрос",
    idea: "Идея",
    support: "Поддержка",
    content: "Контент",
    design: "Дизайн",
    legal: "Юр. вопрос",
    finance: "Финансы",
    hr: "HR",
    document: "Документ",
  }[value || ""] || value || "Задача";
}

export function priorityLabel(value?: string) {
  return {
    none: "Без приоритета",
    low: "Низкий",
    medium: "Средний",
    high: "Высокий",
    urgent: "Срочный",
  }[value || ""] || value || "Без приоритета";
}

export function timeSourceLabel(value?: string) {
  return {
    manual: "Вручную",
    timer: "Таймер",
    status_auto: "Автоматически",
    import: "Импорт",
  }[value || ""] || value || "Время";
}

export function visibilityLabel(value?: string) {
  return {
    internal: "Внутренний",
    client: "Клиентский",
  }[value || ""] || value || "Внутренний";
}

export function statusLabel(value?: string) {
  const key = (value || "").trim().toLowerCase();
  return {
    backlog: "Бэклог",
    todo: "К работе",
    new: "Новая",
    "new request": "Новый запрос",
    triage: "Разбор",
    "need client info": "Нужно уточнение клиента",
    estimate: "Оценка",
    "waiting approval": "Ждет согласования",
    planned: "Запланировано",
    "in progress": "В работе",
    "internal review": "Внутреннее ревью",
    "code review": "Код-ревью",
    review: "Ревью",
    qa: "Тестирование",
    "client acceptance": "Приемка клиента",
    done: "Готово",
    canceled: "Отменено",
    cancelled: "Отменено",
    "won't fix": "Не исправлять",
    reported: "Сообщено",
    confirmed: "Подтверждено",
    "ready for deploy": "Готово к выпуску",
    closed: "Закрыто",
  }[key] || value || "Без статуса";
}
