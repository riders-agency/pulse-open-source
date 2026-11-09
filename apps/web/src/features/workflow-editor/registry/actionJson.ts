import { jsonText } from "../../../lib/api";

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

export function parseJSONText(text: string, fallback: unknown = {}) {
  const trimmed = text.trim();
  if (!trimmed) return fallback;
  return JSON.parse(trimmed) as unknown;
}

export function extractFirstJSONObject(text: string) {
  const source = text
    .trim()
    .replace(/^\uFEFF/, "")
    .replace(/^```(?:json)?\s*/i, "")
    .replace(/\s*```$/i, "")
    .trim();
  if (!source) return "";

  let start = -1;
  let stack: string[] = [];
  let inString = false;
  let escaped = false;
  for (let i = 0; i < source.length; i += 1) {
    const ch = source[i];
    if (start < 0) {
      if (ch === "{" || ch === "[") {
        start = i;
        stack = [ch === "{" ? "}" : "]"];
      }
      continue;
    }
    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (ch === "\\") {
        escaped = true;
      } else if (ch === "\"") {
        inString = false;
      }
      continue;
    }
    if (ch === "\"") {
      inString = true;
      continue;
    }
    if (ch === "{" || ch === "[") {
      stack.push(ch === "{" ? "}" : "]");
      continue;
    }
    if (ch === "}" || ch === "]") {
      if (stack[stack.length - 1] !== ch) break;
      stack.pop();
      if (stack.length === 0) return source.slice(start, i + 1);
    }
  }
  return source;
}

export function parseScenarioPatchText(text: string) {
  const raw = extractFirstJSONObject(text);
  const parsed = JSON.parse(raw) as unknown;
  if (Array.isArray(parsed)) {
    return { format: "riders-pulse-workflow-patch-v1", schema_version: 1, operations: parsed };
  }
  if (isRecord(parsed) && typeof parsed.op === "string") {
    return { format: "riders-pulse-workflow-patch-v1", schema_version: 1, operations: [parsed] };
  }
  return parsed;
}

export function objectHasKeys(value: unknown) {
  return !!value && typeof value === "object" && !Array.isArray(value) && Object.keys(value).length > 0;
}

export function prettyInput(value: unknown, fallbackText: string) {
  if (objectHasKeys(value)) return jsonText(value);
  try {
    return jsonText(JSON.parse(fallbackText));
  } catch {
    return fallbackText;
  }
}

export function templatePayloadFromInput(value: unknown) {
  if (!value || typeof value !== "object" || Array.isArray(value)) return value;
  const object = value as Record<string, unknown>;
  if (object.webhook && typeof object.webhook === "object") return object.webhook;
  if (object.telegram_relay && typeof object.telegram_relay === "object") return object.telegram_relay;
  if (object.payload && typeof object.payload === "object") return object.payload;
  const { _pulse, ...payload } = object;
  void _pulse;
  return payload;
}

export function collectTemplatePaths(value: unknown, prefix = "", acc: string[] = []) {
  if (acc.length >= 8) return acc;
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    if (prefix) acc.push(prefix);
    return acc;
  }
  const entries = Object.entries(value as Record<string, unknown>).filter(([key]) => key !== "_pulse");
  if (!entries.length && prefix) acc.push(prefix);
  for (const [key, child] of entries) {
    const next = prefix ? `${prefix}.${key}` : key;
    if (child && typeof child === "object" && !Array.isArray(child)) {
      collectTemplatePaths(child, next, acc);
    } else {
      acc.push(next);
    }
    if (acc.length >= 8) break;
  }
  return acc;
}

export function templateVariablesFromText(slug: string, inputText: string) {
  let parsed: unknown = {};
  try {
    parsed = parseJSONText(inputText, {});
  } catch {
    parsed = {};
  }
  const payload = templatePayloadFromInput(parsed);
  const paths = collectTemplatePaths(payload);
  const variables = paths.map((path) => `{{ ${slug}.${path} }}`);
  return [...variables, "{{ system.link }}", "{{ system.id }}"];
}

export function emailDefaultsFromVariables(variables: string[]) {
  const external = variables.filter((item) => !item.includes("system."));
  const email = external.find((item) => /email|mail|e-mail/i.test(item)) ?? "";
  const id = external.find((item) => /(^|\.)id(\s|_|\})|_id/i.test(item)) ?? "";
  const bodyLines = external.slice(0, 4).map((item) => `${item}: ${item}`);
  return {
    to: email,
    subject: id ? `Событие ${id}` : "Событие {{ system.id }}",
    body: [...bodyLines, "", "Запуск: {{ system.link }}"].join("\n").trim(),
  };
}
