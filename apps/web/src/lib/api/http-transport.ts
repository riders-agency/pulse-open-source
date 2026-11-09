export type TransportPolicy = {
  retries?: number;
  timeoutMs?: number;
};

const DEFAULT_TIMEOUT_MS = 12_000;
const DEFAULT_RETRY_DELAY_MS = 300;

export async function transportFetch(url: string, init: RequestInit, policy: TransportPolicy) {
  const retries = Math.max(0, policy.retries ?? defaultRetries(init));
  const timeoutMs = policy.timeoutMs ?? DEFAULT_TIMEOUT_MS;
  let lastError: unknown;
  for (let attempt = 0; attempt <= retries; attempt += 1) {
    const request = withTimeout(init, timeoutMs);
    try {
      const response = await fetch(url, request.init);
      request.cleanup();
      if (!shouldRetryResponse(response, init) || attempt >= retries) return response;
      await delay(DEFAULT_RETRY_DELAY_MS * (attempt + 1));
    } catch (error) {
      const timedOut = request.timedOut();
      request.cleanup();
      lastError = error;
      if (attempt >= retries || !isRetryableFetchError(error, timedOut)) throw error;
      await delay(DEFAULT_RETRY_DELAY_MS * (attempt + 1));
    }
  }
  throw lastError instanceof Error ? lastError : new Error("API request failed");
}

export function retryPolicy(init: RequestInit, override: TransportPolicy): TransportPolicy {
  return {
    retries: override.retries ?? defaultRetries(init),
    timeoutMs: override.timeoutMs ?? DEFAULT_TIMEOUT_MS,
  };
}

function defaultRetries(init: RequestInit) {
  return isSafeMethod(init.method) ? 2 : 0;
}

function shouldRetryResponse(response: Response, init: RequestInit) {
  if (!isSafeMethod(init.method)) return false;
  return [408, 429, 502, 503, 504].includes(response.status);
}

function isRetryableFetchError(error: unknown, timedOut: boolean) {
  if (!(error instanceof Error)) return false;
  if (error.name === "AbortError") return timedOut;
  return error.name === "TimeoutError" || error instanceof TypeError;
}

function withTimeout(init: RequestInit, timeoutMs: number) {
  if (timeoutMs <= 0) return { init, cleanup: () => undefined, timedOut: () => false };
  const controller = new AbortController();
  let didTimeout = false;
  const timer = setTimeout(() => {
    didTimeout = true;
    controller.abort();
  }, timeoutMs);
  const merged = mergeSignals(init.signal, controller.signal);
  return {
    init: { ...init, signal: merged.signal },
    cleanup: () => {
      clearTimeout(timer);
      merged.cleanup();
    },
    timedOut: () => didTimeout,
  };
}

function mergeSignals(first: AbortSignal | null | undefined, second: AbortSignal) {
  if (!first) return { signal: second, cleanup: () => undefined };
  if (first.aborted) return { signal: first, cleanup: () => undefined };
  if (second.aborted) return { signal: second, cleanup: () => undefined };
  const controller = new AbortController();
  const abort = () => controller.abort();
  first.addEventListener("abort", abort, { once: true });
  second.addEventListener("abort", abort, { once: true });
  return {
    signal: controller.signal,
    cleanup: () => {
      first.removeEventListener("abort", abort);
      second.removeEventListener("abort", abort);
    },
  };
}

function isSafeMethod(method: string | undefined) {
  return ["GET", "HEAD", "OPTIONS"].includes((method ?? "GET").toUpperCase());
}

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
