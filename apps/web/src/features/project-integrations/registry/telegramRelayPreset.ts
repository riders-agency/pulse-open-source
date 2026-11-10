const allTelegramUpdateTypes = [
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

function randomRelayToken(length = 40) {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-";
  if (typeof crypto !== "undefined" && crypto.getRandomValues) {
    const values = crypto.getRandomValues(new Uint8Array(length));
    return Array.from(values, (value) => alphabet[value % alphabet.length]).join("");
  }
  return Array.from({ length }, () => alphabet[Math.floor(Math.random() * alphabet.length)]).join("");
}

export function telegramRelayPreset() {
  const relayKey = `tg_${randomRelayToken(18).toLowerCase()}`;
  return JSON.stringify(
    {
      relay_key: relayKey,
      relay_url: `https://pulse.riders.agency/api/v1/telegram/relay/${relayKey}`,
      webhook_secret: randomRelayToken(48),
      allowed_updates: allTelegramUpdateTypes,
      default_chat_id: "",
      country_relays: [],
      polling_enabled: false,
      polling_timeout_seconds: 50,
      polling_limit: 100,
      polling_delete_webhook: true,
      polling_drop_pending_updates: false,
    },
    null,
    2,
  );
}
