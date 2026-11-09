import type { ActionNodeType } from "../model/types";

export const defaultSample = `{
  "order": {
    "id": "1001",
    "total": 12500,
    "customer": {
      "email": "client@example.com"
    }
  }
}`;

export const telegramRelaySample = `{
  "_pulse": {
    "source_type": "telegram_relay",
    "telegram_relay_key": "support_bot"
  },
  "telegram_relay": {
    "relay_key": "support_bot",
    "update_id": 1000001,
    "update_type": "message",
    "command": "start",
    "text": "/start order_1001",
    "chat": { "id": 123456789, "type": "private", "username": "client" },
    "from": { "id": 123456789, "username": "client", "first_name": "Ivan" },
    "message": { "message_id": 51, "text": "/start order_1001" }
  }
}`;

export const allTelegramUpdateTypes = [
  "message",
  "edited_message",
  "channel_post",
  "edited_channel_post",
  "business_connection",
  "business_message",
  "edited_business_message",
  "deleted_business_messages",
  "guest_message",
  "message_reaction",
  "message_reaction_count",
  "inline_query",
  "chosen_inline_result",
  "callback_query",
  "shipping_query",
  "pre_checkout_query",
  "purchased_paid_media",
  "poll",
  "poll_answer",
  "my_chat_member",
  "chat_member",
  "chat_join_request",
  "chat_boost",
  "removed_chat_boost",
  "managed_bot",
];

export const aiWorkflowDocsURL = "https://pulse.riders.agency/api/v1/docs/integrations/workflow-ai-patch-contract.md";

export function paymentDueBatchSample(projectId: string) {
  return JSON.stringify(
    {
      _pulse: {
        source_type: "client_payment_due",
        batch: true,
        project_id: projectId,
        trigger_action_key: "payment_due",
      },
      project: { project_id: projectId },
      payment_cycle: { batch: true },
    },
    null,
    2,
  );
}

export function triggerSampleFrom(nodes: ActionNodeType[], projectId?: string) {
  const paymentDue = nodes.find((node) => node.data?.action_type_key === "clients.payment_due.trigger");
  if (paymentDue && projectId) return paymentDueBatchSample(projectId);
  const trigger = nodes.find((node) => node.data?.action_type_key === "webhook.trigger");
  const relay = nodes.find((node) => node.data?.action_type_key === "telegram.relay.trigger");
  if (!trigger && relay) return telegramRelaySample;
  if (trigger) return trigger.data.config.sample_payload_json || defaultSample;
  return "{}";
}

export function webhookSlugFromNodes(nodes: ActionNodeType[]) {
  const trigger = nodes.find((node) => node.data?.action_type_key === "webhook.trigger");
  const slug = trigger?.data.config.endpoint_key?.trim();
  if (slug) return slug;
  const relay = nodes.find((node) => node.data?.action_type_key === "telegram.relay.trigger");
  const relayKey = relay?.data.config.relay_key?.trim();
  return relayKey || "telegram_relay";
}
