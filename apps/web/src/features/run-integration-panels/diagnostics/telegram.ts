import type { ActionDiagnosticsWire } from "../../../contracts/workflow-step-output";
import type { ProviderDiagnosticsAnalyzer } from "./types";

function report(patch: ActionDiagnosticsWire): ActionDiagnosticsWire {
  return { provider: "telegram", retryable: false, ...patch };
}

export const analyzeTelegram: ProviderDiagnosticsAnalyzer = ({ statusCode, text, failed }) => {
  if (!failed) {
    return report({
      status: "success",
      reason_code: "telegram_accepted",
      summary: "Telegram подтвердил отправку сообщения.",
      human_reason: "Telegram API вернул успешный ответ.",
    });
  }
  if (statusCode === 401) {
    return report({
      status: "failed",
      reason_code: "telegram_invalid_token",
      summary: "Telegram не подтвердил отправку.",
      human_reason: "Telegram отклонил токен бота.",
      next_action: "Проверьте bot_token в интеграции Telegram.",
    });
  }
  if (statusCode === 429) {
    return report({
      status: "retryable",
      reason_code: "telegram_rate_limited",
      summary: "Telegram ограничил частоту отправки.",
      human_reason: "Сработал rate limit Telegram.",
      next_action: "Подождите и повторите действие.",
      retryable: true,
    });
  }
  if (statusCode === 403) {
    return report({
      status: "failed",
      reason_code: text.includes("blocked") ? "telegram_bot_blocked" : "telegram_forbidden",
      summary: "Telegram запретил отправку.",
      human_reason: "Бот не может писать в этот чат.",
      next_action: "Проверьте права бота и chat_id.",
    });
  }
  if (statusCode === 400 && text.includes("chat not found")) {
    return report({
      status: "failed",
      reason_code: "telegram_chat_not_found",
      summary: "Telegram не нашел chat_id.",
      human_reason: "Указанный чат не найден.",
      next_action: "Проверьте chat_id или начните диалог с ботом.",
    });
  }
  if (statusCode === 400 && (text.includes("parse") || text.includes("entity"))) {
    return report({
      status: "failed",
      reason_code: "telegram_bad_parse_mode",
      summary: "Telegram не разобрал сообщение.",
      human_reason: "Ошибка разметки Markdown/HTML.",
      next_action: "Исправьте parse_mode или шаблон.",
    });
  }
  if (statusCode >= 500) {
    return report({
      status: "retryable",
      reason_code: "telegram_provider_unavailable",
      summary: "Telegram временно недоступен.",
      human_reason: "Провайдер вернул серверную ошибку.",
      next_action: "Повторите действие позже.",
      retryable: true,
    });
  }
  return report({
    status: "failed",
    reason_code: "telegram_provider_error",
    summary: "Telegram вернул ошибку.",
    human_reason: text,
    next_action: "Сверьте chat_id, текст, parse_mode и сырой ответ.",
  });
};
