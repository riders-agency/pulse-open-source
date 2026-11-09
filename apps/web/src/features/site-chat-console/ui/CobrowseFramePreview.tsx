import type { Dispatch, RefObject, SetStateAction } from "react";

import type { VisitorStateSnapshot } from "../../../lib/api";
import { applyMirrorScrollState } from "../model/cobrowse";
import type { IframePreviewStatus } from "../model/types";

export function CobrowseFramePreview({
  documentHeight,
  documentWidth,
  iframeRef,
  livePreviewURL,
  mirror,
  pageRef,
  scrollX,
  scrollY,
  setIframeStatus,
  useLivePreview,
  useMirrorPreview,
}: {
  documentHeight: number;
  documentWidth: number;
  iframeRef: RefObject<HTMLIFrameElement | null>;
  livePreviewURL: string;
  mirror: VisitorStateSnapshot["mirror"] | undefined;
  pageRef: RefObject<HTMLDivElement | null>;
  scrollX: number;
  scrollY: number;
  setIframeStatus: Dispatch<SetStateAction<IframePreviewStatus>>;
  useLivePreview: boolean;
  useMirrorPreview: boolean;
}) {
  const onLoad = () => {
    setIframeStatus("loaded");
    applyMirrorScrollState(iframeRef.current, mirror, scrollX, scrollY, documentWidth, documentHeight);
  };

  if (useLivePreview) {
    return (
      <div className="remote-page" ref={pageRef}>
        <iframe
          ref={iframeRef}
          title="Visitor live page"
          name="riders-pulse-webvisor-preview"
          src={livePreviewURL}
          sandbox="allow-same-origin allow-scripts allow-forms allow-popups"
          scrolling="no"
          className="remote-live-frame"
          data-rp-live-frame="1"
          onLoad={onLoad}
        />
      </div>
    );
  }

  if (useMirrorPreview) {
    return (
      <div className="remote-page" ref={pageRef}>
        <iframe
          ref={iframeRef}
          title="Visitor DOM mirror"
          srcDoc={mirror?.html ?? ""}
          sandbox="allow-scripts"
          scrolling="no"
          name="riders-pulse-webvisor-preview"
          className="remote-mirror-frame"
          data-rp-mirror-frame="1"
          onLoad={onLoad}
        />
      </div>
    );
  }

  return <div className="cobrowse-fallback">Нет URL активной страницы.</div>;
}
