import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

/**
 * Входящие вебхуки должны попадать в Go API (через nginx -> :9810), а не обрабатываться Next (:9811).
 * Если запрос случайно пришел напрямую на Node, явно сообщаем, что ingress настроен неверно.
 */
export function proxy(_req: NextRequest) {
  return NextResponse.json(
    {
      type: "about:blank",
      title: "webhook routed to frontend",
      status: 503,
      detail:
        "POST /hooks/* обслуживает Riders Pulse API. Настройте прокси с /hooks на backend (обычно 127.0.0.1:9810).",
    },
    { status: 503 },
  );
}

export const config = {
  matcher: ["/hooks", "/hooks/:path*"],
};
