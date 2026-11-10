import { useEffect, useState, type RefObject } from "react";

import type { IframePreviewStatus } from "../model/types";

export function useCobrowsePreviewStatus({
  hasMirror,
  iframeRef,
  livePreviewURL,
  url,
}: {
  hasMirror: boolean;
  iframeRef: RefObject<HTMLIFrameElement | null>;
  livePreviewURL: string;
  url: string;
}) {
  const [iframeStatus, setIframeStatus] = useState<IframePreviewStatus>("idle");
  const [previewAck, setPreviewAck] = useState(0);

  useEffect(() => {
    if (!url && !hasMirror) {
      setIframeStatus("idle");
      return;
    }
    if (!url && hasMirror) {
      setIframeStatus("loaded");
      return;
    }
    setIframeStatus("loading");
  }, [hasMirror, livePreviewURL, url]);

  useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (!iframeRef.current?.contentWindow || event.source !== iframeRef.current.contentWindow) return;
      const data = event.data as { source?: string; type?: string } | null;
      if (!data || data.source !== "riders-pulse-webvisor-preview") return;
      if (data.type === "rp:webvisor:mirror-ready") {
        setIframeStatus("loaded");
        setPreviewAck((value) => value + 1);
      }
    };
    window.addEventListener("message", onMessage);
    return () => window.removeEventListener("message", onMessage);
  }, [iframeRef]);

  return { iframeStatus, previewAck, setIframeStatus };
}
