import type { ActionKind } from "../model/types";

export const palette: Array<{ group: string; items: Array<{ type: ActionKind; label: string; icon: string }> }> = [
  { group: "Триггеры", items: [{ type: "webhook.trigger", label: "Входящий вебхук", icon: "webhook" }, { type: "telegram.relay.trigger", label: "Telegram-ретранслятор", icon: "telegram" }, { type: "schedule.trigger", label: "Периодический запуск", icon: "schedule" }, { type: "manual.trigger", label: "Ручной запуск", icon: "manual" }] },
  { group: "Клиенты", items: [{ type: "clients.manual_payment.trigger", label: "Счет оплачен", icon: "manual" }, { type: "clients.payment_due.trigger", label: "Счета к оплате", icon: "payment" }] },
  { group: "Логика", items: [{ type: "condition.branch", label: "Условие", icon: "condition" }, { type: "person.fio.normalize", label: "ФИО: нормализация", icon: "condition" }, { type: "person.phone.normalize", label: "Телефон: нормализация", icon: "condition" }, { type: "stop", label: "Стоп", icon: "stop" }] },
  { group: "AI", items: [{ type: "ai.apinet.normalize_text", label: "APINET: нормализация текста", icon: "sparkles" }] },
  { group: "Управление", items: [{ type: "delay.wait", label: "Задержка", icon: "delay" }, { type: "bot.set_state", label: "Состояние бота", icon: "bot" }, { type: "http.request", label: "HTTP запрос", icon: "http" }] },
  { group: "Интеграции", items: [{ type: "uon.create_lead", label: "U-ON: создать обращение", icon: "uon" }, { type: "yclients.certificate.issue", label: "YClients: выпустить сертификат", icon: "certificate" }] },
  { group: "Уведомления", items: [{ type: "max.send_message", label: "Отправить в MAX", icon: "max" }, { type: "telegram.send_message", label: "Отправить в Telegram", icon: "telegram" }, { type: "telegram.edit_message_text", label: "Telegram: редактировать сообщение", icon: "telegram" }, { type: "telegram.answer_callback_query", label: "Telegram: callback answer", icon: "telegram" }, { type: "telegram.send_document", label: "Telegram: документ", icon: "telegram" }, { type: "email.send", label: "Отправить Email", icon: "email" }] },
];
