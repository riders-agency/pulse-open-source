export function clientTypeLabel(type?: string) {
  if (type === "individual") return "Физлицо";
  if (type === "individual_entrepreneur") return "ИП";
  return "Юрлицо";
}

export function legalEntityTypeLabel(type?: string) {
  if (type === "individual_entrepreneur") return "ИП";
  return "Юрлицо";
}

export function statusLabel(status?: string) {
  const labels: Record<string, string> = {
    active: "Активно",
    archived: "В архиве",
    deleted: "Удалено",
    disabled: "Выключено",
    error: "Ошибка",
    pending: "Ожидает",
    sent: "Отправлено",
    snoozed: "Отложено",
    paid: "Оплачено",
    cancelled: "Отменено",
    needs_decision: "Отправлено",
    paused: "Пауза",
    completed: "Завершено",
    draft: "Черновик",
  };
  return labels[status ?? ""] ?? "Не указано";
}

export function paymentStatusLabel(status?: string) {
  const labels: Record<string, string> = {
    pending: "Ждет счет",
    scheduled: "Ждет счет",
    sent: "Счет подготовлен",
    snoozed: "Отложен",
    needs_decision: "Счет подготовлен",
    paid: "Счет оплачен",
    cancelled: "Отменен",
  };
  return labels[status ?? ""] ?? statusLabel(status);
}

export function paymentSourceLabel(source?: string) {
  const labels: Record<string, string> = {
    manual: "ручная отметка",
    api_key: "API-ключ",
    bank: "банк",
    email: "входящее письмо",
    incoming_email: "входящее письмо",
    kontur_elba: "Контур Эльба",
    integration: "интеграция",
    other_integration: "интеграция",
    workflow: "сценарий",
  };
  return labels[source ?? ""] ?? "не указан";
}

export function paymentMethodLabel(method?: string) {
  const labels: Record<string, string> = {
    cash: "наличные",
    legal_invoice: "по счету с юрлица",
    bank_transfer: "банковский перевод",
    card: "карта",
    other: "другое",
  };
  return labels[method ?? "legal_invoice"] ?? "не указан";
}

export function statusTone(status?: string): "success" | "warning" | "danger" | "info" | "default" {
  if (status === "active" || status === "paid" || status === "completed") return "success";
  if (status === "pending" || status === "sent" || status === "snoozed" || status === "needs_decision" || status === "paused") return "warning";
  if (status === "deleted" || status === "cancelled" || status === "error") return "danger";
  if (status === "archived" || status === "disabled") return "info";
  return "default";
}

export function billingPeriodLabel(period?: string) {
  const labels: Record<string, string> = {
    once: "Разово",
    weekly: "Еженедельно",
    monthly: "Ежемесячно",
    quarterly: "Ежеквартально",
    semiannual: "Раз в полгода",
    yearly: "Ежегодно",
  };
  return labels[period ?? ""] ?? "Не указано";
}

export const invoiceScenarioOptions = [
  {
    key: "bank_invoice_responsible_email",
    label: "Банк -> письмо ответственному",
    description: "Счет выставляется через банк, затем Pulse пишет основному контакту заказчика.",
  },
  {
    key: "email_telegram_invoice",
    label: "Почта + Telegram",
    description: "Счет появляется из почты или Telegram и прикрепляется к оплате.",
  },
  {
    key: "kontur_elba_invoice",
    label: "Контур Эльба",
    description: "Счет готовится через интеграцию с Контур Эльбой.",
  },
] as const;

export function invoiceScenarioLabel(key?: string) {
  if (key === "mixed_invoice_scenarios") return "Несколько сценариев";
  return invoiceScenarioOptions.find((item) => item.key === key)?.label ?? invoiceScenarioOptions[0].label;
}

export function invoiceScenarioFromMetadata(metadata?: Record<string, unknown>) {
  const raw = metadata?.invoice_scenario;
  if (raw && typeof raw === "object" && !Array.isArray(raw)) {
    const key = (raw as Record<string, unknown>).key;
    if (typeof key === "string" && key.trim()) return key;
  }
  const legacy = metadata?.invoice_scenario_key;
  if (typeof legacy === "string" && legacy.trim()) return legacy;
  return invoiceScenarioOptions[0].key;
}

export function responsibleRoleLabel(role?: string) {
  const labels: Record<string, string> = {
    account_manager: "Аккаунт",
    technical_lead: "Техника",
    finance: "Финансы",
    legal: "Документы",
    owner: "Владелец",
    other: "Другое",
  };
  return labels[role ?? ""] ?? "Ответственный";
}
