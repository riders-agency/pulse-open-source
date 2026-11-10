import type { SiteCobrowseFrame, VisitorStateSnapshot } from "../../../lib/api";
import { clamp } from "./format";
import { mirrorFromPayload, scrollContainersFromPayload } from "./live-snapshot";

function positiveNumber(value: unknown) {
  const next = typeof value === "number" ? value : typeof value === "string" ? Number(value) : NaN;
  return Number.isFinite(next) && next > 0 ? next : undefined;
}

export function saneViewportSize(value: unknown, fallback: number, screenSize?: number) {
  const raw = positiveNumber(value) ?? fallback;
  const screen = positiveNumber(screenSize);
  const max = Math.max(fallback, screen ?? 0, 4096);
  if (screen && raw > screen * 1.35) return Math.round(clamp(screen, 320, max));
  return Math.round(clamp(raw, 320, max));
}

export function latestMirror(snapshot: VisitorStateSnapshot, frames: SiteCobrowseFrame[]): VisitorStateSnapshot["mirror"] | undefined {
  let scrollContainers = snapshot.mirror?.scroll_containers;
  for (let index = frames.length - 1; index >= 0; index -= 1) {
    const payload = frames[index]?.payload_json as Record<string, unknown> | undefined;
    const frameScrollContainers = scrollContainersFromPayload(payload);
    if (frameScrollContainers) scrollContainers = frameScrollContainers;
    const mirror = mirrorFromPayload(payload);
    if (mirror?.html) {
      return {
        ...snapshot.mirror,
        ...mirror,
        scroll_containers: mirror.scroll_containers ?? scrollContainers,
      };
    }
  }
  return snapshot.mirror ? { ...snapshot.mirror, scroll_containers: scrollContainers } : undefined;
}

export function mirrorScrollSignature(mirror: VisitorStateSnapshot["mirror"] | undefined) {
  return (mirror?.scroll_containers ?? [])
    .map((item) => `${item.path}:${item.left ?? 0}:${item.top ?? 0}`)
    .join("|");
}

export function applyMirrorScrollState(
  frame: HTMLIFrameElement | null,
  mirror: VisitorStateSnapshot["mirror"] | undefined,
  scrollX = 0,
  scrollY = 0,
  documentWidth = 0,
  documentHeight = 0,
) {
  if (!frame) return;
  const viewportWidth = Math.max(frame.clientWidth, 1);
  const viewportHeight = Math.max(frame.clientHeight, 1);
  const fullWidth = Math.max(documentWidth, viewportWidth, 1);
  const fullHeight = Math.max(documentHeight, viewportHeight, 1);
  const scrollContainers = mirror?.scroll_containers ?? [];
  try {
    frame.contentWindow?.postMessage({
      type: "rp:webvisor:apply-state",
      scrollX,
      scrollY,
      viewportWidth,
      viewportHeight,
      documentWidth: fullWidth,
      documentHeight: fullHeight,
      scrollContainers,
    }, "*");
  } catch {
    // Cross-origin or sandboxed frames can still receive later load-time messages.
  }
  let doc: Document | null = null;
  let win: Window | null = null;
  try {
    doc = frame.contentDocument;
    win = frame.contentWindow;
  } catch {
    return;
  }
  if (!doc) return;
  const html = doc.documentElement;
  const body = doc.body;
  if (html) {
    html.style.overflow = "auto";
    html.style.width = `${viewportWidth}px`;
    html.style.minWidth = `${fullWidth}px`;
    html.style.minHeight = `${fullHeight}px`;
    html.style.scrollbarWidth = "none";
    html.dataset.rpAppliedScrollX = String(Math.round(scrollX));
    html.dataset.rpAppliedScrollY = String(Math.round(scrollY));
  }
  if (body) {
    body.style.overflow = "visible";
    body.style.width = `${viewportWidth}px`;
    body.style.minWidth = `${fullWidth}px`;
    body.style.minHeight = `${fullHeight}px`;
    body.style.scrollbarWidth = "none";
    body.dataset.rpAppliedScrollX = String(Math.round(scrollX));
    body.dataset.rpAppliedScrollY = String(Math.round(scrollY));
  }
  try {
    win?.scrollTo({ left: scrollX, top: scrollY, behavior: "auto" });
  } catch {
    win?.scrollTo(scrollX, scrollY);
  }
  if (!scrollContainers.length) return;
  scrollContainers.forEach((item) => {
    if (!item.path) return;
    const node = doc.querySelector(`[data-rp-path="${String(item.path).replace(/"/g, '\\"')}"]`) as HTMLElement | null;
    if (!node) return;
    node.scrollLeft = Number(item.left ?? 0);
    node.scrollTop = Number(item.top ?? 0);
  });
}
