import type { SiteChatWSEnvelope, SiteCobrowseFrame, VisitorStateSnapshot } from "../../../lib/api";

export function payloadNumber(payload: Record<string, unknown> | undefined, key: string) {
  const value = payload?.[key];
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
}

export function payloadString(payload: Record<string, unknown> | undefined, key: string) {
  const value = payload?.[key];
  return typeof value === "string" ? value : undefined;
}

export function eventPayloads(message: SiteChatWSEnvelope) {
  if (message.type !== "event.batch") return [{ type: message.type, payload: message.payload ?? {} }];
  const raw = message.payload?.events;
  if (!Array.isArray(raw)) return [];
  return raw.flatMap((item) => {
    if (!item || typeof item !== "object") return [];
    const event = item as { event_type?: unknown; payload?: unknown };
    if (typeof event.event_type !== "string") return [];
    return [{ type: event.event_type, payload: event.payload && typeof event.payload === "object" ? event.payload as Record<string, unknown> : {} }];
  });
}

export function frameKindFromEvent(type: string): SiteCobrowseFrame["kind"] | undefined {
  if (type === "monitoring.snapshot") return "snapshot";
  if (type === "monitoring.incremental") return "mutation";
  if (type === "cursor.move" || type === "cursor") return "cursor";
  if (type === "touch.move" || type === "touch") return "touch";
  if (type === "scroll") return "scroll";
  if (type === "viewport") return "viewport";
  if (type === "page.navigate" || type === "page") return "navigation";
  if (type === "click") return "click";
  if (type === "diagnostic") return "diagnostic";
  return undefined;
}

export function sequenceFromPayload(payload: Record<string, unknown>, fallback: number) {
  const sequence = payloadNumber(payload, "sequence");
  return sequence !== undefined ? sequence : fallback;
}

export function mirrorFromPayload(payload: Record<string, unknown> | undefined): VisitorStateSnapshot["mirror"] | undefined {
  const mirror = payload?.mirror;
  if (!mirror || typeof mirror !== "object") return undefined;
  return mirror as VisitorStateSnapshot["mirror"];
}

export function scrollContainersFromPayload(payload: Record<string, unknown> | undefined): NonNullable<VisitorStateSnapshot["mirror"]>["scroll_containers"] | undefined {
  const value = payload?.scroll_containers;
  return Array.isArray(value) ? value as NonNullable<VisitorStateSnapshot["mirror"]>["scroll_containers"] : undefined;
}

function viewportContextFromPayload(payload: Record<string, unknown> | undefined): VisitorStateSnapshot["viewport_context"] | undefined {
  const value = payload?.viewport_context;
  return value && typeof value === "object" ? value as VisitorStateSnapshot["viewport_context"] : undefined;
}

export function mergeLiveSnapshotEvent(snapshot: VisitorStateSnapshot, type: string, payload: Record<string, unknown>) {
  const next: VisitorStateSnapshot = {
    ...snapshot,
    activity: {
      ...snapshot.activity,
      connection_status: "online",
      is_active: true,
      last_seen_at: new Date().toISOString(),
    },
  };

  const url = payloadString(payload, "url");
  const title = payloadString(payload, "title");
  if (url) {
    next.current_url = url;
  }
  if (title) {
    next.current_title = title;
  }

  if (["page", "page.navigate", "viewport", "scroll", "monitoring.snapshot", "monitoring.incremental"].includes(type)) {
    const viewportWidth = payloadNumber(payload, "viewport_width") ?? payloadNumber(payload, "width");
    const viewportHeight = payloadNumber(payload, "viewport_height") ?? payloadNumber(payload, "height");
    if (viewportWidth && viewportHeight) {
      next.viewport = {
        ...snapshot.viewport,
        width: viewportWidth,
        height: viewportHeight,
        layout_width: payloadNumber(payload, "layout_width"),
        layout_height: payloadNumber(payload, "layout_height"),
        scale: payloadNumber(payload, "viewport_scale"),
        offset_x: payloadNumber(payload, "viewport_offset_x"),
        offset_y: payloadNumber(payload, "viewport_offset_y"),
      };
    }
    const documentWidth = payloadNumber(payload, "document_width");
    const documentHeight = payloadNumber(payload, "document_height");
    if (documentWidth && documentHeight) {
      next.document = { width: documentWidth, height: documentHeight };
    }
    const scrollX = payloadNumber(payload, "scroll_x") ?? payloadNumber(payload, "x");
    const scrollY = payloadNumber(payload, "scroll_y") ?? payloadNumber(payload, "y");
    if (scrollX !== undefined || scrollY !== undefined) {
      next.scroll = {
        ...snapshot.scroll,
        x: scrollX ?? snapshot.scroll?.x ?? 0,
        y: scrollY ?? snapshot.scroll?.y ?? 0,
        document_width: documentWidth ?? snapshot.scroll?.document_width ?? snapshot.document?.width,
        document_height: documentHeight ?? snapshot.scroll?.document_height ?? snapshot.document?.height,
      };
    }
  }

  const mirror = mirrorFromPayload(payload);
  const scrollContainers = scrollContainersFromPayload(payload);
  const viewportContext = viewportContextFromPayload(payload);
  if (viewportContext) {
    next.viewport_context = viewportContext;
  }
  if (mirror?.html) {
    next.mirror = {
      ...snapshot.mirror,
      ...mirror,
      scroll_containers: mirror.scroll_containers ?? scrollContainers ?? snapshot.mirror?.scroll_containers,
    };
  } else if (scrollContainers) {
    next.mirror = {
      ...snapshot.mirror,
      scroll_containers: scrollContainers,
    };
  }

  if (type === "cursor" || type === "cursor.move") {
    const x = payloadNumber(payload, "x");
    const y = payloadNumber(payload, "y");
    if (x !== undefined && y !== undefined) next.cursor = { x, y };
  }
  if (type === "touch" || type === "touch.move") {
    const x = payloadNumber(payload, "x");
    const y = payloadNumber(payload, "y");
    if (x !== undefined && y !== undefined) next.touch = { x, y };
  }
  if (type === "click") {
    const x = payloadNumber(payload, "x");
    const y = payloadNumber(payload, "y");
    if (x !== undefined && y !== undefined) next.last_click = { x, y };
  }
  if (type === "widget_state") {
    next.widget_state = {
      open: Boolean(payload.open),
      active_tab: payloadString(payload, "active_tab") || "chat",
      url: payloadString(payload, "url"),
      title: payloadString(payload, "title"),
    };
    next.signals = {
      ...next.signals,
      chat_open: Boolean(payload.open),
    };
  }

  return next;
}
