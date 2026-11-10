export function localDateISO(date = new Date()) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function dateLabel(value: unknown) {
  const raw = typeof value === "string" ? value : "";
  if (!raw) return "Не указан";
  const [year, month, day] = raw.split("-");
  if (year && month && day) return `${day}.${month}.${year}`;
  return raw;
}

export function dateInputValue(value: unknown) {
  const raw = typeof value === "string" ? value : "";
  if (!raw) return "";
  const iso = raw.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (iso) return `${iso[3]}.${iso[2]}.${iso[1]}`;
  return raw;
}

export function dateInputToISO(value: string) {
  const raw = value.trim();
  const ru = raw.match(/^(\d{1,2})\.(\d{1,2})\.(\d{4})$/);
  if (!ru) return raw;
  const day = ru[1].padStart(2, "0");
  const month = ru[2].padStart(2, "0");
  return `${ru[3]}-${month}-${day}`;
}

export function messageTimestampLabel(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
