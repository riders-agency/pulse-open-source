import { clamp } from "../model/format";

export type CobrowseStyleNodes = {
  shell: HTMLElement | null;
  viewport: HTMLElement | null;
  page: HTMLElement | null;
  iframe: HTMLElement | null;
  cursor: HTMLElement | null;
  touch: HTMLElement | null;
  click: HTMLElement | null;
  scrollbarThumb: HTMLElement | null;
};

export type CobrowseStyleState = {
  width: number;
  height: number;
  scale: number;
  cursor: { x?: unknown; y?: unknown };
  touch: { x?: unknown; y?: unknown };
  click: { x?: unknown; y?: unknown };
  scrollbarTop: number;
  scrollbarHeight: number;
};

export function applyCobrowseStyleVars(nodes: CobrowseStyleNodes, state: CobrowseStyleState) {
  setPx(nodes.shell, "--remote-shell-width", state.width * state.scale);
  setPx(nodes.shell, "--remote-shell-height", state.height * state.scale);
  setPx(nodes.viewport, "--remote-viewport-width", state.width);
  setPx(nodes.viewport, "--remote-viewport-height", state.height);
  nodes.viewport?.style.setProperty("--remote-viewport-scale", String(state.scale));
  setPx(nodes.page, "--remote-preview-width", state.width);
  setPx(nodes.page, "--remote-preview-height", state.height);
  nodes.page?.style.setProperty("--remote-preview-transform", "translate3d(0, 0, 0)");
  setPx(nodes.iframe, "--remote-preview-width", state.width);
  setPx(nodes.iframe, "--remote-preview-height", state.height);
  applyPointerVars(nodes.cursor, state.cursor, state.width, state.height);
  applyPointerVars(nodes.touch, state.touch, state.width, state.height);
  applyPointerVars(nodes.click, state.click, state.width, state.height);
  setPx(nodes.scrollbarThumb, "--remote-scrollbar-top", state.scrollbarTop);
  setPx(nodes.scrollbarThumb, "--remote-scrollbar-height", state.scrollbarHeight);
}

function applyPointerVars(node: HTMLElement | null, point: { x?: unknown; y?: unknown }, width: number, height: number) {
  setPx(node, "--remote-pointer-x", clamp(Number(point.x), 0, width));
  setPx(node, "--remote-pointer-y", clamp(Number(point.y), 0, height));
}

function setPx(node: HTMLElement | null, name: string, value: number) {
  node?.style.setProperty(name, `${value}px`);
}
