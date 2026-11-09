import type { ActionDiagnosticsWire } from "../../../contracts/workflow-step-output";
import type { ProviderDiagnosticsAnalyzer } from "./types";

function report(patch: ActionDiagnosticsWire): ActionDiagnosticsWire {
  return { provider: "max", retryable: false, ...patch };
}

export const analyzeMax: ProviderDiagnosticsAnalyzer = ({ statusCode, text, failed }) => {
  if (!failed) {
    return report({
      status: "success",
      reason_code: "max_accepted",
      summary: "MAX принял сообщение.",
      human_reason: "MAX API вернул успешный ответ.",
    });
  }
  if (statusCode === 401 || statusCode === 403) {
    return report({
      status: "failed",
      reason_code: "max_auth_failed",
      summary: "MAX отклонил запрос.",
      human_reason: "Токен или права интеграции не подходят.",
      next_action: "Проверьте токен MAX: заголовок должен быть Authorization: <token> без Bearer.",
    });
  }
  if (statusCode === 429) {
    return report({
      status: "retryable",
      reason_code: "max_rate_limited",
      summary: "MAX ограничил частоту запросов.",
      human_reason: "Сработал rate limit MAX.",
      next_action: "Подождите и повторите действие.",
      retryable: true,
    });
  }
  if (statusCode === 404 || text.includes("chat") || text.includes("recipient")) {
    return report({
      status: "failed",
      reason_code: "max_recipient_not_found",
      summary: "MAX не нашел чат или получателя.",
      human_reason: "Получатель сообщения не найден.",
      next_action: "Проверьте chat_id или user_id и доступ бота к получателю.",
    });
  }
  if (statusCode >= 500) {
    return report({
      status: "retryable",
      reason_code: "max_provider_unavailable",
      summary: "MAX временно недоступен.",
      human_reason: "Провайдер вернул серверную ошибку.",
      next_action: "Повторите действие позже.",
      retryable: true,
    });
  }
  return report({
    status: "failed",
    reason_code: "max_provider_error",
    summary: "MAX вернул ошибку.",
    human_reason: text,
    next_action: "Проверьте настройки действия и сырой ответ MAX.",
  });
};
