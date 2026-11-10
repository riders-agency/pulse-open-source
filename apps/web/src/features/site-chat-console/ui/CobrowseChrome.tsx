import { Badge } from "../../../components/ui/badge";
import { shortURL } from "../model/format";

export function CobrowseToolbar({
  documentHeight,
  documentWidth,
  height,
  scrollY,
  url,
  width,
}: {
  documentHeight: number;
  documentWidth: number;
  height: number;
  scrollY: number;
  url: string;
  width: number;
}) {
  return (
    <header className="cobrowse-toolbar">
      <strong>{shortURL(url)}</strong>
      <span>{width}×{height} · page {Math.round(documentWidth)}×{Math.round(documentHeight)} · scroll {Math.round(scrollY)}px</span>
    </header>
  );
}

export function CobrowseFooter({
  browser,
  deviceType,
  framesCount,
  useLivePreview,
  useMirrorPreview,
  widgetOpen,
}: {
  browser?: string;
  deviceType?: string;
  framesCount: number;
  useLivePreview: boolean;
  useMirrorPreview: boolean;
  widgetOpen: boolean;
}) {
  return (
    <footer className="cobrowse-footer">
      <Badge>{framesCount} frames</Badge>
      {useLivePreview ? <Badge tone="success">Страница в реальном времени</Badge> : null}
      {useMirrorPreview ? <Badge tone="success">Копия страницы</Badge> : null}
      {widgetOpen ? <Badge tone="success">Чат открыт</Badge> : null}
      <Badge>{browser ?? "browser"}</Badge>
      <Badge>{deviceType ?? "device"}</Badge>
    </footer>
  );
}
