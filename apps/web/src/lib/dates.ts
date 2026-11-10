type DateParts = {
  year: number;
  month: number;
  day: number;
};

const isoDatePattern = /^(\d{4})-(\d{2})-(\d{2})/;
const ruDatePattern = /^(\d{2})\.(\d{2})\.(\d{4})$/;

export const dateInputProps = {
  inputMode: "numeric" as const,
  placeholder: "дд.мм.гггг",
  pattern: "\\d{2}\\.\\d{2}\\.\\d{4}",
};

export function formatDate(value?: string | Date | null, fallback = "—") {
  const parts = parseDateParts(value);
  if (!parts) return fallback;
  return `${pad2(parts.day)}.${pad2(parts.month)}.${parts.year}`;
}

export function formatDateTime(value?: string | Date | null, fallback = "—") {
  const date = parseDateTime(value);
  if (!date) return formatDate(value, fallback);
  return `${formatDate(date, fallback)} ${pad2(date.getHours())}:${pad2(date.getMinutes())}`;
}

export function formatDateRange(start?: string | null, end?: string | null) {
  const a = formatDate(start, "");
  const b = formatDate(end, "");
  if (!a && !b) return "—";
  if (!b || a === b) return a;
  return `${a} — ${b}`;
}

export function todayDateInput() {
  return formatDate(new Date(), "");
}

export function toDateInput(value?: string | Date | null) {
  return formatDate(value, "");
}

export function toISODate(value?: string | Date | null) {
  const parts = parseDateParts(value);
  if (!parts) return "";
  return `${parts.year}-${pad2(parts.month)}-${pad2(parts.day)}`;
}

export function compareDate(a?: string | Date | null, b?: string | Date | null) {
  return toISODate(a).localeCompare(toISODate(b));
}

export function isBeforeDate(a?: string | Date | null, b?: string | Date | null) {
  const left = toISODate(a);
  const right = toISODate(b);
  return Boolean(left && right && left < right);
}

export function addDaysDate(value: string, days: number) {
  const date = utcDate(value);
  if (!date) return "";
  date.setUTCDate(date.getUTCDate() + days);
  return formatDateFromUTC(date);
}

export function advanceDate(value: string, billingPeriod: string) {
  const date = utcDate(value);
  if (!date || billingPeriod === "once") return "";
  if (billingPeriod === "weekly") date.setUTCDate(date.getUTCDate() + 7);
  else if (billingPeriod === "quarterly") date.setUTCMonth(date.getUTCMonth() + 3);
  else if (billingPeriod === "semiannual") date.setUTCMonth(date.getUTCMonth() + 6);
  else if (billingPeriod === "yearly") date.setUTCFullYear(date.getUTCFullYear() + 1);
  else date.setUTCMonth(date.getUTCMonth() + 1);
  return formatDateFromUTC(date);
}

export function dateBefore(value: string) {
  return addDaysDate(value, -1);
}

function parseDateParts(value?: string | Date | null): DateParts | null {
  if (!value) return null;
  if (value instanceof Date) {
    if (Number.isNaN(value.getTime())) return null;
    return {
      year: value.getFullYear(),
      month: value.getMonth() + 1,
      day: value.getDate(),
    };
  }
  const raw = value.trim();
  if (!raw) return null;

  let match = raw.match(ruDatePattern);
  if (match) {
    return validParts(Number(match[3]), Number(match[2]), Number(match[1]));
  }

  match = raw.match(isoDatePattern);
  if (match) {
    return validParts(Number(match[1]), Number(match[2]), Number(match[3]));
  }

  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return null;
  return {
    year: date.getFullYear(),
    month: date.getMonth() + 1,
    day: date.getDate(),
  };
}

function parseDateTime(value?: string | Date | null) {
  if (!value) return null;
  const date = value instanceof Date ? value : new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function utcDate(value: string) {
  const parts = parseDateParts(value);
  if (!parts) return null;
  return new Date(Date.UTC(parts.year, parts.month - 1, parts.day));
}

function validParts(year: number, month: number, day: number): DateParts | null {
  if (!Number.isInteger(year) || !Number.isInteger(month) || !Number.isInteger(day)) return null;
  const date = new Date(Date.UTC(year, month - 1, day));
  if (date.getUTCFullYear() !== year || date.getUTCMonth() + 1 !== month || date.getUTCDate() !== day) {
    return null;
  }
  return { year, month, day };
}

function formatDateFromUTC(date: Date) {
  return `${pad2(date.getUTCDate())}.${pad2(date.getUTCMonth() + 1)}.${date.getUTCFullYear()}`;
}

function pad2(value: number) {
  return String(value).padStart(2, "0");
}
