export function statusTone(status?: string): "success" | "warning" | "danger" | "info" | "muted" {
  if (status === "succeeded") return "success";
  if (status === "failed") return "danger";
  if (status === "retrying" || status === "waiting") return "warning";
  if (status === "running") return "info";
  return "muted";
}

export function statusLabel(status?: string) {
  const labels: Record<string, string> = {
    pending: "В очереди",
    not_started: "Не начато",
    running: "Выполняется",
    succeeded: "Готово",
    failed: "Ошибка",
    cancelled: "Отменено",
    waiting: "Ждёт решения",
    retrying: "Повтор",
    skipped: "Пропущено",
  };
  return status ? labels[status] ?? status : "не запускалось";
}

export function actionTypeLabel(type?: string) {
  const labels: Record<string, string> = {
    "mailbox.incoming.trigger": "Входящая почта",
    "office.classify": "Классификация письма",
    "office.route": "Передать в отделы",
    "office.website.prepare": "ИИ-отдел: страница сайта",
    "office.invoice.prepare": "Бухгалтерия: счёт",
    "office.approve": "Подтверждение человеком",
    "office.apply": "Сохранить результат",
    "webhook.trigger": "Входящий вебхук",
    "telegram.relay.trigger": "Telegram-ретранслятор",
    "schedule.trigger": "Периодический запуск",
    "manual.trigger": "Ручной запуск",
    "clients.payment_due.trigger": "Счета к оплате",
    "clients.payment_cycle.collect": "Отчет по оплатам",
    "condition.branch": "Условие",
    "stop": "Стоп",
    "delay.wait": "Задержка",
    "bot.set_state": "Состояние бота",
    "http.request": "HTTP запрос",
    "max.send_message": "Отправить в MAX",
    "telegram.send_message": "Отправить в Telegram",
    "uon.create_lead": "U-ON: создать обращение",
    "email.send": "Отправить Email",
  };
  return type ? labels[type] ?? type : "Действие";
}

export function sourceTypeLabel(source?: string) {
  const labels: Record<string, string> = {
    module_trigger: "Входящее событие",
    webhook: "Вебхук",
    manual: "Ручной запуск",
    schedule: "Расписание",
    desktop_message: "Рабочий стол",
    replay: "Повтор",
    test: "Тест",
    telegram_relay: "Telegram-ретранслятор",
    client_payment_due: "Оплаты клиентов",
  };
  return source ? labels[source] ?? source : "—";
}
