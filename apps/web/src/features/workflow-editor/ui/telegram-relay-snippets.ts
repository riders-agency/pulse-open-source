import { buildTelegramRelayPromptSnippet, telegramRelayDocURL } from "./telegram-relay-prompt-snippet";
import { buildTelegramRelayPollingSnippets } from "./telegram-relay-polling-snippets";

export { telegramRelayDocURL };

export type TelegramRelayPollingSnippetSettings = {
  timeout: number;
  limit: number;
  dropPending: boolean;
};

export const defaultTelegramRelayPollingSnippetSettings: TelegramRelayPollingSnippetSettings = {
  timeout: 50,
  limit: 100,
  dropPending: false,
};

export type TelegramRelaySnippet = {
  key: string;
  label: string;
  code: string;
};

export function buildTelegramRelaySnippets({
  relayKey,
  relayUrl,
  secretValue,
  allowedUpdates,
  polling,
}: {
  relayKey: string;
  relayUrl: string;
  secretValue: string;
  allowedUpdates: string[];
  polling: TelegramRelayPollingSnippetSettings;
}): TelegramRelaySnippet[] {
  const allowedUpdatesJSON = JSON.stringify(allowedUpdates, null, 2);
  const allowedUpdatesInline = JSON.stringify(allowedUpdates);
  return [
    {
      key: "curl",
      label: "curl",
      code: `curl -X POST "https://api.telegram.org/bot$BOT_TOKEN/setWebhook" \\
  -H "Content-Type: application/json" \\
  --data '{
    "url": "${relayUrl}",
    "secret_token": "${secretValue}",
    "allowed_updates": ${allowedUpdatesInline}
  }'`,
    },
    {
      ...buildTelegramRelayPromptSnippet({ relayKey, relayUrl, secretValue, allowedUpdatesInline, polling }),
    },
    {
      key: "python",
      label: "Python webhook",
      code: `import asyncio
import os
from telegram import Bot

BOT_TOKEN = os.environ["TELEGRAM_BOT_TOKEN"]
WEBHOOK_SECRET = os.environ["TELEGRAM_WEBHOOK_SECRET"]
WEBHOOK_URL = os.environ.get("RIDERS_PULSE_RELAY_URL", "${relayUrl}")
ALLOWED_UPDATES = ${allowedUpdatesJSON}

async def main():
    bot = Bot(BOT_TOKEN)
    result = await bot.set_webhook(
        url=WEBHOOK_URL,
        secret_token=WEBHOOK_SECRET,
        allowed_updates=ALLOWED_UPDATES,
        drop_pending_updates=False,
    )
    print({"ok": result, "webhook_url": WEBHOOK_URL})

asyncio.run(main())`,
    },
    ...buildTelegramRelayPollingSnippets({ relayUrl, secretValue, allowedUpdatesJSON, allowedUpdatesInline, polling }),
    {
      key: "php",
      label: "PHP direct",
      code: `<?php
$botToken = getenv('TELEGRAM_BOT_TOKEN');
$webhookSecret = getenv('TELEGRAM_WEBHOOK_SECRET');
$relayUrl = getenv('RIDERS_PULSE_RELAY_URL') ?: '${relayUrl}';
$allowedUpdates = ${allowedUpdatesInline};

if (!$botToken || !$webhookSecret) {
    fwrite(STDERR, "Set TELEGRAM_BOT_TOKEN and TELEGRAM_WEBHOOK_SECRET\\n");
    exit(1);
}

$payload = json_encode([
    'url' => $relayUrl,
    'secret_token' => $webhookSecret,
    'allowed_updates' => $allowedUpdates,
    'drop_pending_updates' => false,
], JSON_UNESCAPED_SLASHES);

$ch = curl_init("https://api.telegram.org/bot{$botToken}/setWebhook");
curl_setopt_array($ch, [
    CURLOPT_POST => true,
    CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
    CURLOPT_POSTFIELDS => $payload,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_TIMEOUT => 20,
]);
$response = curl_exec($ch);
$status = curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
$error = curl_error($ch);
curl_close($ch);

if ($error || $status < 200 || $status >= 300) {
    fwrite(STDERR, "setWebhook failed: {$status} {$error} {$response}\\n");
    exit(1);
}

echo $response . PHP_EOL;
`,
    },
    {
      key: "test-post",
      label: "test POST",
      code: `curl -X POST "${relayUrl}" \\
  -H "Content-Type: application/json" \\
  -H "X-Telegram-Bot-Api-Secret-Token: ${secretValue}" \\
  --data '{
    "update_id": 1000001,
    "message": {
      "message_id": 51,
      "date": 1760000000,
      "chat": {"id": 123456789, "type": "private", "username": "client"},
      "from": {"id": 123456789, "is_bot": false, "first_name": "Ivan"},
      "text": "/start order_1001",
      "entities": [{"offset": 0, "length": 6, "type": "bot_command"}]
    }
  }'`,
    },
  ];
}
