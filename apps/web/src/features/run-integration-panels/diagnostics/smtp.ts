import type { ActionDiagnosticsWire } from "../../../contracts/workflow-step-output";
import { smtpEnvelope } from "./helpers";
import type { ProviderDiagnosticsAnalyzer } from "./types";

function report(patch: ActionDiagnosticsWire): ActionDiagnosticsWire {
  return { provider: "smtp", retryable: false, ...patch };
}

export const analyzeSMTP: ProviderDiagnosticsAnalyzer = ({ provider, text, failed }) => {
  if (!failed) {
    return report({
      status: "success",
      reason_code: "smtp_accepted",
      summary: "SMTP сервер принял письмо.",
      human_reason: "SMTP сессия завершилась успешно.",
    });
  }
  const smtp = smtpEnvelope(provider);
  const stage = typeof smtp?.failed_stage === "string" ? smtp.failed_stage : "";
  const smtpCode = typeof smtp?.smtp_code === "string" ? smtp.smtp_code : "";
  if (stage === "auth" || smtpCode === "535" || smtpCode === "534") {
    return report({
      status: "failed",
      reason_code: "smtp_auth_failed",
      summary: "SMTP отклонил логин или пароль.",
      human_reason: "Авторизация SMTP не прошла.",
      next_action: "Проверьте username/password, пароль приложения и encryption.",
    });
  }
  if (stage === "rcpt_to" || ["550", "551", "553"].includes(smtpCode)) {
    return report({
      status: "failed",
      reason_code: "smtp_recipient_rejected",
      summary: "SMTP отклонил получателя.",
      human_reason: "Адрес получателя не принят сервером.",
      next_action: "Проверьте email получателя и ограничения домена.",
    });
  }
  if (stage === "mail_from") {
    return report({
      status: "failed",
      reason_code: "smtp_sender_rejected",
      summary: "SMTP отклонил отправителя.",
      human_reason: "from_email не разрешен сервером.",
      next_action: "Проверьте from_email в интеграции.",
    });
  }
  if (stage === "starttls") {
    return report({
      status: "failed",
      reason_code: "smtp_tls_failed",
      summary: "Ошибка TLS SMTP.",
      human_reason: "Не удалось установить защищенное соединение.",
      next_action: "Проверьте host/port/encryption.",
    });
  }
  if (stage === "tcp_dial" || stage === "smtp_client" || smtpCode.startsWith("4")) {
    return report({
      status: "retryable",
      reason_code: "smtp_connection_failed",
      summary: "SMTP временно недоступен.",
      human_reason: "Не удалось подключиться или сервер вернул временную ошибку.",
      next_action: "Проверьте доступность SMTP и повторите.",
      retryable: true,
    });
  }
  return report({
    status: "failed",
    reason_code: "smtp_provider_error",
    summary: "SMTP вернул ошибку.",
    human_reason: text,
    next_action: "Смотрите последний ответ сервера и этап SMTP.",
  });
};
