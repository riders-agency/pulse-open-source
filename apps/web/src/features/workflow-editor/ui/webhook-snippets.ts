import type { ActionConfig } from "../model/types";

export type WebhookSnippet = {
  key: string;
  label: string;
  code: string;
};

export function buildWebhookSnippets({
  absoluteHookUrl,
  authType,
  samplePayload,
}: {
  absoluteHookUrl: string;
  authType: ActionConfig["auth_type"];
  samplePayload: string;
}): WebhookSnippet[] {
  const authCurlLine = authType === "none" ? "" : "  -H 'Authorization: Bearer PROJECT_SECRET_HERE' \\\n";
  const authHeaderLine = authType === "none" ? "" : `      "Authorization": "Bearer PROJECT_SECRET_HERE",\n`;
  const pythonAuthLine = authType === "none" ? "" : `    "Authorization": "Bearer PROJECT_SECRET_HERE",\n`;
  const phpAuthLine = authType === "none" ? "" : `    "Authorization: Bearer PROJECT_SECRET_HERE",\n`;
  return [
    {
      key: "curl",
      label: "curl",
      code: `curl -X POST '${absoluteHookUrl}' \\
  -H 'Content-Type: application/json' \\
  -H 'Idempotency-Key: order-1001' \\
${authCurlLine}  --data-binary @- <<'JSON'
${samplePayload}
JSON`,
    },
    {
      key: "prompt",
      label: "Vibe prompt",
      code: `Подключи отправку события в Riders Pulse.

Webhook URL: ${absoluteHookUrl}
Method: POST
Headers:
- Content-Type: application/json
- Idempotency-Key: стабильный id события, например order-1001
${authType === "none" ? "- Authorization не нужен для этого dev endpoint" : "- Authorization: Bearer PROJECT_SECRET_HERE"}

Body JSON:
${samplePayload}

Требования:
- отправляй событие после успешного бизнес-действия;
- повторяй запрос при timeout/5xx, не повторяй 4xx auth/validation;
- логируй HTTP status и body ответа;
- не хардкодь секрет в коде, возьми его из env/secret storage.`,
    },
    {
      key: "javascript",
      label: "JS fetch",
      code: `const response = await fetch("${absoluteHookUrl}", {
  method: "POST",
  headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": "order-1001",
${authHeaderLine}  },
  body: ${JSON.stringify(samplePayload)},
});

if (!response.ok) {
  throw new Error(\`Riders Pulse webhook failed: \${response.status} \${await response.text()}\`);
}`,
    },
    {
      key: "python",
      label: "Python",
      code: `import requests

response = requests.post(
    "${absoluteHookUrl}",
    headers={
        "Content-Type": "application/json",
        "Idempotency-Key": "order-1001",
${pythonAuthLine}    },
    data=${JSON.stringify(samplePayload)},
    timeout=30,
)
response.raise_for_status()`,
    },
    {
      key: "php",
      label: "PHP",
      code: `<?php
$payload = <<<'JSON'
${samplePayload}
JSON;

$ch = curl_init("${absoluteHookUrl}");
curl_setopt_array($ch, [
    CURLOPT_POST => true,
    CURLOPT_HTTPHEADER => [
        "Content-Type: application/json",
        "Idempotency-Key: order-1001",
${phpAuthLine}    ],
    CURLOPT_POSTFIELDS => $payload,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_TIMEOUT => 30,
]);

$body = curl_exec($ch);
$status = curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
if ($body === false || $status >= 400) {
    throw new RuntimeException("Riders Pulse webhook failed: " . curl_error($ch) . " HTTP " . $status);
}
curl_close($ch);`,
    },
  ];
}
