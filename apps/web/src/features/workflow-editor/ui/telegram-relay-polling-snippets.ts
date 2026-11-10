export function buildTelegramRelayPollingSnippets({
  relayUrl,
  secretValue,
  allowedUpdatesJSON,
  allowedUpdatesInline,
  polling,
}: {
  relayUrl: string;
  secretValue: string;
  allowedUpdatesJSON: string;
  allowedUpdatesInline: string;
  polling: { timeout: number; limit: number; dropPending: boolean };
}) {
  return [
    {
      key: "python-polling",
      label: "Python polling",
      code: `import asyncio
import json
import os
import httpx
from telegram import Bot

BOT_TOKEN = os.environ["TELEGRAM_BOT_TOKEN"]
RELAY_SECRET = os.environ["TELEGRAM_WEBHOOK_SECRET"]
RELAY_URL = os.environ.get("RIDERS_PULSE_RELAY_URL", "${relayUrl}")
ALLOWED_UPDATES = ${allowedUpdatesJSON}
TIMEOUT = int(os.environ.get("TELEGRAM_POLLING_TIMEOUT", "${polling.timeout}"))
LIMIT = int(os.environ.get("TELEGRAM_POLLING_LIMIT", "${polling.limit}"))
DROP_PENDING = os.environ.get("TELEGRAM_POLLING_DROP_PENDING", "${polling.dropPending}") == "true"

async def main():
    bot = Bot(BOT_TOKEN)
    await bot.delete_webhook(drop_pending_updates=DROP_PENDING)
    offset = None
    async with httpx.AsyncClient(timeout=TIMEOUT + 15) as client:
        while True:
            updates = await bot.get_updates(
                offset=offset,
                timeout=TIMEOUT,
                limit=LIMIT,
                allowed_updates=ALLOWED_UPDATES,
            )
            for update in updates:
                payload = update.to_dict()
                response = await client.post(
                    RELAY_URL,
                    headers={
                        "Content-Type": "application/json",
                        "X-Relay-Secret": RELAY_SECRET,
                    },
                    content=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
                )
                if response.status_code != 202:
                    raise RuntimeError(f"relay failed: {response.status_code} {response.text}")
                offset = update.update_id + 1

asyncio.run(main())`,
    },
    {
      key: "http-polling",
      label: "HTTP polling",
      code: `# Любой язык без SDK: Telegram getUpdates -> Riders Pulse relay.
# 1. POST https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/deleteWebhook
#    JSON: {"drop_pending_updates": ${polling.dropPending}}
# 2. loop:
#    POST https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getUpdates
#    JSON: {"offset": offset, "timeout": ${polling.timeout}, "limit": ${polling.limit}, "allowed_updates": ${allowedUpdatesInline}}
# 3. for each result[] Update:
#    POST ${relayUrl}
#    Headers:
#      Content-Type: application/json
#      X-Relay-Secret: ${secretValue}
#    Body: raw Update JSON
# 4. if relay status == 202:
#      offset = update.update_id + 1
#    else:
#      retry same update, do not advance offset
#
# No WebSocket is required. Telegram getUpdates is HTTP long polling.`,
    },
  ];
}
