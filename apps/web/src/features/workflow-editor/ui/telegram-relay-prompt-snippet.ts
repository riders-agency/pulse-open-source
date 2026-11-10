import { aiWorkflowDocsURL } from "../registry/actions";
import type { TelegramRelayPollingSnippetSettings, TelegramRelaySnippet } from "./telegram-relay-snippets";

export const telegramRelayDocURL = "https://pulse.riders.agency/api/v1/docs/integrations/telegram-relay-trigger.md";

export function buildTelegramRelayPromptSnippet({
  relayKey,
  relayUrl,
  secretValue,
  allowedUpdatesInline,
  polling,
}: {
  relayKey: string;
  relayUrl: string;
  secretValue: string;
  allowedUpdatesInline: string;
  polling: TelegramRelayPollingSnippetSettings;
}): TelegramRelaySnippet {
  return {
    key: "prompt",
    label: "Vibe prompt",
    code: `Сделай универсальное подключение Telegram Bot API к Riders Pulse relay в нужном языке/стеке.

Обязательно прочитай документацию Riders Pulse:
- Main contract: ${aiWorkflowDocsURL}
- Telegram relay integration: ${telegramRelayDocURL}

Telegram Bot API docs: https://core.telegram.org/bots/api

Нужно реализовать один из режимов:
1. Webhook mode: вызвать Bot API setWebhook и отправлять Telegram Update JSON напрямую в Riders Pulse relay.
2. Polling mode: вызвать Bot API deleteWebhook, затем циклом getUpdates через HTTP long polling, каждый raw Update JSON отправлять POST в Riders Pulse relay, после HTTP 202 двигать offset на update_id + 1.

WebSocket не нужен. getUpdates у Telegram это HTTP long polling. Webhook и getUpdates одновременно не работают, поэтому polling mode сначала удаляет webhook.

Можно использовать любой язык программирования. Если в выбранном языке есть Telegram SDK, используй его для Bot API. Если SDK нет, делай прямые HTTPS POST к https://api.telegram.org/bot<TOKEN>/<METHOD>.

ENV keys:
- TELEGRAM_BOT_TOKEN=bot token из BotFather
- TELEGRAM_WEBHOOK_SECRET=${secretValue}
- RIDERS_PULSE_RELAY_URL=${relayUrl}
- TELEGRAM_ALLOWED_UPDATES=${allowedUpdatesInline}
- TELEGRAM_POLLING_TIMEOUT=${polling.timeout}
- TELEGRAM_POLLING_LIMIT=${polling.limit}
- TELEGRAM_POLLING_DROP_PENDING=${polling.dropPending}

Важно:
- ${relayKey} в конце URL это route key, не пароль.
- TELEGRAM_WEBHOOK_SECRET это отдельный auth header. Он нужен, чтобы любой человек с URL не мог дернуть сценарий.
- Если TELEGRAM_WEBHOOK_SECRET здесь placeholder, нажми в Riders Pulse кнопку "Сгенерировать webhook secret" и скопируй prompt заново.

HTTP auth keys, которые принимает relay:
- X-Telegram-Bot-Api-Secret-Token: TELEGRAM_WEBHOOK_SECRET
- X-Relay-Secret: TELEGRAM_WEBHOOK_SECRET
- Authorization: Bearer TELEGRAM_WEBHOOK_SECRET

Webhook requirements:
- setWebhook JSON body: {"url": RIDERS_PULSE_RELAY_URL, "secret_token": TELEGRAM_WEBHOOK_SECRET, "allowed_updates": TELEGRAM_ALLOWED_UPDATES, "drop_pending_updates": false}

Polling requirements:
- deleteWebhook JSON body: {"drop_pending_updates": TELEGRAM_POLLING_DROP_PENDING}
- getUpdates JSON body: {"offset": offset, "timeout": TELEGRAM_POLLING_TIMEOUT, "limit": TELEGRAM_POLLING_LIMIT, "allowed_updates": TELEGRAM_ALLOWED_UPDATES}
- relay POST headers: {"Content-Type":"application/json","X-Relay-Secret": TELEGRAM_WEBHOOK_SECRET}
- relay должен получать raw Telegram Update JSON без потери полей.
- offset обновлять только после успешного relay HTTP 202.

Проверка:
- отправь /start боту
- relay должен вернуть HTTP 202
- Riders Pulse должен сохранить raw update, normalized update, command, chat/user/message ids и запустить workflow, если есть опубликованный telegram.relay.trigger
- команды любого вида должны сохраняться; callback data вида "/command payload" тоже считать command

Template keys для дальнейших действий:
- {{ telegram_relay.update_id }}
- {{ telegram_relay.update_type }}
- {{ telegram_relay.command }}
- {{ telegram_relay.text }}
- {{ telegram_relay.callback_data }}
- {{ telegram_relay.chat.id }}
- {{ telegram_relay.from.id }}
- {{ telegram_relay.raw_update }}
- {{ ${relayKey}.message.text }}
- {{ ${relayKey}.command }}
- {{ telegram.update_id }}
- {{ bot_session.state_key }}
- {{ bot_session.context_json }}
- {{ system.id }}
- {{ system.link }}

Не логируй TELEGRAM_BOT_TOKEN и TELEGRAM_WEBHOOK_SECRET. Все секреты брать только из env/secret storage. Логировать можно update_id, update_type, command, chat.id, from.id и relay HTTP status.`,
  };
}
