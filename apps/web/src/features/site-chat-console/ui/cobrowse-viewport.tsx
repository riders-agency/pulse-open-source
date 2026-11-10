import { useCallback, useEffect, useMemo, useRef } from "react";

import type { SiteCobrowseFrame, VisitorStateSnapshot } from "../../../lib/api";
import { applyMirrorScrollState, latestMirror, mirrorScrollSignature, saneViewportSize } from "../model/cobrowse";
import { clamp, operatorPreviewURL } from "../model/format";
import { CobrowseFramePreview } from "./CobrowseFramePreview";
import { CobrowseFooter, CobrowseToolbar } from "./CobrowseChrome";
import { CobrowseOverlay } from "./CobrowseOverlay";
import { applyCobrowseStyleVars } from "./cobrowse-style";
import { useCobrowsePreviewStatus } from "./useCobrowsePreviewStatus";
import { useCobrowseScale } from "./useCobrowseScale";

export function CobrowseViewport({ snapshot, frames, fullscreen }: { snapshot: VisitorStateSnapshot; frames: SiteCobrowseFrame[]; fullscreen: boolean }) {
  const viewport = snapshot.viewport ?? {};
  const cursor = snapshot.cursor ?? {};
  const touch = snapshot.touch ?? {};
  const click = snapshot.last_click ?? {};
  const scroll = snapshot.scroll ?? {};
  const screen = snapshot.analytics?.screen ?? {};
  const mirror = useMemo(() => latestMirror(snapshot, frames), [frames, snapshot]);
  const hasMirror = !!mirror?.html;
  const mirrorScrolls = mirrorScrollSignature(mirror);
  const width = saneViewportSize(viewport.width ?? viewport.layout_width, 1280, screen.width);
  const height = saneViewportSize(viewport.height ?? viewport.layout_height, 720, screen.height);
  const rawScrollX = Number(scroll.x ?? 0);
  const rawScrollY = Number(scroll.y ?? 0);
  const documentWidth = Math.max(Number(snapshot.document?.width ?? scroll.document_width ?? 0), width, rawScrollX + width);
  const documentHeight = Math.max(Number(snapshot.document?.height ?? scroll.document_height ?? 0), height, rawScrollY + height);
  const scrollX = clamp(rawScrollX, 0, Math.max(documentWidth - width, 0));
  const scrollY = clamp(rawScrollY, 0, Math.max(documentHeight - height, 0));
  const scrollbarHeight = clamp((height / Math.max(documentHeight, height)) * height, 24, height);
  const scrollbarTop = documentHeight > height ? clamp((scrollY / (documentHeight - height)) * (height - scrollbarHeight), 0, height - scrollbarHeight) : 0;
  const frameRef = useRef<HTMLDivElement | null>(null);
  const shellRef = useRef<HTMLDivElement | null>(null);
  const viewportRef = useRef<HTMLDivElement | null>(null);
  const pageRef = useRef<HTMLDivElement | null>(null);
  const iframeRef = useRef<HTMLIFrameElement | null>(null);
  const cursorRef = useRef<HTMLSpanElement | null>(null);
  const touchRef = useRef<HTMLSpanElement | null>(null);
  const clickRef = useRef<HTMLSpanElement | null>(null);
  const scrollbarThumbRef = useRef<HTMLSpanElement | null>(null);
  const scrollStateRef = useRef({ mirror, scrollX, scrollY, documentWidth, documentHeight });
  const scrollApplyRef = useRef({ timer: 0, lastAt: 0, lastX: Number.NaN, lastY: Number.NaN });
  const scale = useCobrowseScale(frameRef, width, height);
  const url = snapshot.current_url || "";
  const widgetOpen = snapshot.widget_state?.open;
  const livePreviewURL = useMemo(() => operatorPreviewURL(url), [url]);
  const useLivePreview = !!livePreviewURL;
  const useMirrorPreview = !useLivePreview && hasMirror;
  const { iframeStatus, previewAck, setIframeStatus } = useCobrowsePreviewStatus({ hasMirror, iframeRef, livePreviewURL, url });
  scrollStateRef.current = { mirror, scrollX, scrollY, documentWidth, documentHeight };

  const applyPreviewScroll = useCallback((force = false) => {
    const applyRef = scrollApplyRef.current;
    const now = performance.now();
    const minIntervalMs = 120;
    const run = () => {
      applyRef.timer = 0;
      const next = scrollStateRef.current;
      const moved = Math.abs(applyRef.lastX - next.scrollX) > 2 || Math.abs(applyRef.lastY - next.scrollY) > 2;
      if (!force && !moved) return;
      applyMirrorScrollState(iframeRef.current, next.mirror, next.scrollX, next.scrollY, next.documentWidth, next.documentHeight);
      applyRef.lastAt = performance.now();
      applyRef.lastX = next.scrollX;
      applyRef.lastY = next.scrollY;
    };

    if (force || now - applyRef.lastAt >= minIntervalMs) {
      if (applyRef.timer) {
        window.clearTimeout(applyRef.timer);
        applyRef.timer = 0;
      }
      window.requestAnimationFrame(run);
      return;
    }

    if (applyRef.timer) window.clearTimeout(applyRef.timer);
    applyRef.timer = window.setTimeout(() => window.requestAnimationFrame(run), minIntervalMs - (now - applyRef.lastAt));
  }, []);

  useEffect(() => {
    applyCobrowseStyleVars({
      shell: shellRef.current,
      viewport: viewportRef.current,
      page: pageRef.current,
      iframe: iframeRef.current,
      cursor: cursorRef.current,
      touch: touchRef.current,
      click: clickRef.current,
      scrollbarThumb: scrollbarThumbRef.current,
    }, {
      width,
      height,
      scale,
      cursor,
      touch,
      click,
      scrollbarTop,
      scrollbarHeight,
    });
  }, [click.x, click.y, cursor.x, cursor.y, height, scale, scrollbarHeight, scrollbarTop, touch.x, touch.y, width]);

  useEffect(() => {
    if (!iframeRef.current) return;
    applyPreviewScroll(false);
    const timers = [240, 700, 1500, 2800].map((delay) => window.setTimeout(() => applyPreviewScroll(true), delay));
    return () => timers.forEach((timer) => window.clearTimeout(timer));
  }, [applyPreviewScroll, documentHeight, documentWidth, iframeStatus, mirror?.captured_at, mirror?.html, mirrorScrolls, previewAck, scrollX, scrollY]);

  useEffect(() => {
    if (!useLivePreview) return;
    const interval = window.setInterval(() => {
      applyPreviewScroll(true);
    }, 900);
    return () => window.clearInterval(interval);
  }, [applyPreviewScroll, useLivePreview]);

  useEffect(() => {
    return () => {
      const timer = scrollApplyRef.current.timer;
      if (timer) window.clearTimeout(timer);
    };
  }, []);

  return (
    <section className={`cobrowse-panel ${fullscreen ? "is-fullscreen" : ""}`}>
      <CobrowseToolbar documentHeight={documentHeight} documentWidth={documentWidth} height={height} scrollY={scrollY} url={url} width={width} />
      <div className="cobrowse-stage" ref={frameRef}>
        <div className="remote-viewport-shell" ref={shellRef}>
          <div className="remote-viewport" ref={viewportRef}>
            <CobrowseFramePreview
              documentHeight={documentHeight}
              documentWidth={documentWidth}
              iframeRef={iframeRef}
              livePreviewURL={livePreviewURL}
              mirror={mirror}
              pageRef={pageRef}
              scrollX={scrollX}
              scrollY={scrollY}
              setIframeStatus={setIframeStatus}
              useLivePreview={useLivePreview}
              useMirrorPreview={useMirrorPreview}
            />
            <CobrowseOverlay
              activeTab={snapshot.widget_state?.active_tab}
              click={click}
              clickRef={clickRef}
              cursor={cursor}
              cursorRef={cursorRef}
              scrollbarThumbRef={scrollbarThumbRef}
              showBlocked={iframeStatus === "unavailable" && !hasMirror}
              touch={touch}
              touchRef={touchRef}
              widgetOpen={Boolean(widgetOpen)}
            />
          </div>
        </div>
      </div>
      <CobrowseFooter
        browser={snapshot.analytics?.browser}
        deviceType={snapshot.analytics?.device_type}
        framesCount={frames.length}
        useLivePreview={useLivePreview}
        useMirrorPreview={useMirrorPreview}
        widgetOpen={Boolean(widgetOpen)}
      />
    </section>
  );
}
