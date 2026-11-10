import type { ActionConfig, ActionKind, ActionNodeType, RetryPolicy } from "../model/types";

import { allTelegramUpdateTypes, defaultSample, telegramRelaySample } from "./actionSamples";
import { providerActionDefault } from "./actionProviderDefaults";

export function defaultConfig(type: ActionKind): ActionConfig {
  const retry_policy: RetryPolicy = { max_attempts: 3, strategy: "exponential", initial_interval: "5s", timeout: "30s" };
  if (type === "webhook.trigger") {
    return {
      auth_type: "project_secret",
      endpoint_key: "orders_paid",
      sample_payload_json: defaultSample,
      json_fields: [
        { path: "order.id", type: "string", required: true, description: "Номер заказа", example: "1001" },
        { path: "order.total", type: "number", required: true, description: "Сумма заказа", example: "12500" },
        { path: "order.customer.email", type: "string", required: false, description: "Email клиента", example: "client@example.com" },
      ],
    };
  }
  if (type === "schedule.trigger") {
    return {
      enabled: true,
      schedule_key: "periodic",
      interval_seconds: 3600,
      timezone: "Europe/Moscow",
      active_from: "",
      active_until: "",
    };
  }
  if (type === "clients.payment_due.trigger") {
    return {
      enabled: true,
      timezone: "Europe/Moscow",
      remind_every_days: 1,
      days_before_due: 7,
      active_from: "",
      active_until: "",
    };
  }
  if (type === "clients.manual_payment.trigger") {
    return {};
  }
  if (type === "clients.payment_cycle.collect") {
    return {
      days_before_due: 30,
    };
  }
  if (type === "telegram.relay.trigger") {
    return {
      action_dependency_id: "",
      relay_key: "support_bot",
      allowed_updates: allTelegramUpdateTypes,
      sample_payload_json: telegramRelaySample,
    };
  }
  if (type === "bot.set_state") {
    return {
      state_key: "waiting_for_reply",
      context_json: "{\n  \"last_command\": \"{{ telegram_relay.command }}\",\n  \"chat_id\": \"{{ telegram_relay.chat.id }}\"\n}",
      merge_context: true,
    };
  }
  if (type === "person.fio.normalize") {
    return {
      input: "{{ webhook.name }}",
      output_fields: {},
    };
  }
  if (type === "person.phone.normalize") {
    return {
      input: "{{ webhook.phone }}",
      output_fields: {},
    };
  }
  const providerDefault = providerActionDefault(type, retry_policy);
  if (providerDefault) return providerDefault;
  return { retry_policy };
}

export function normalizeEmailConfig(config: ActionConfig): ActionConfig {
  const retry_policy = config.retry_policy ?? defaultConfig("email.send").retry_policy!;
  return {
    ...config,
    retry_policy,
    smtp_action_dependency_id: config.smtp_action_dependency_id ?? "",
    to: config.to ?? "",
    subject: config.subject ?? "Событие {{ system.id }}",
    body: config.body ?? "Запуск: {{ system.link }}",
  };
}

export function normalizeEmailNode(node: ActionNodeType): ActionNodeType {
  if (node.data?.action_type_key !== "email.send") return node;
  return {
    ...node,
    data: {
      ...node.data,
      config: normalizeEmailConfig(node.data.config),
    },
  };
}
