import type { RefObject } from "react";

type Point = { x?: number; y?: number };

export function CobrowseOverlay({
  activeTab,
  click,
  clickRef,
  cursor,
  cursorRef,
  scrollbarThumbRef,
  showBlocked,
  touch,
  touchRef,
  widgetOpen,
}: {
  activeTab?: string;
  click: Point;
  clickRef: RefObject<HTMLSpanElement | null>;
  cursor: Point;
  cursorRef: RefObject<HTMLSpanElement | null>;
  scrollbarThumbRef: RefObject<HTMLSpanElement | null>;
  showBlocked: boolean;
  touch: Point;
  touchRef: RefObject<HTMLSpanElement | null>;
  widgetOpen: boolean;
}) {
  return (
    <>
      <div className="cobrowse-overlay">
        {cursor.x != null ? <span className="remote-cursor" ref={cursorRef} /> : null}
        {touch.x != null ? <span className="remote-touch" ref={touchRef} /> : null}
        {click.x != null ? <span className="remote-click" ref={clickRef} /> : null}
        <span className="remote-scrollbar"><span ref={scrollbarThumbRef} /></span>
      </div>
      {widgetOpen ? (
        <div className="remote-widget-state" data-widget-state="open">
          <strong>Консультация открыта</strong>
          <span>{activeTab ?? "chat"}</span>
        </div>
      ) : null}
      {showBlocked ? (
        <div className="cobrowse-blocked">
          <strong>Предпросмотр не загрузился</strong>
          <span>Движения курсора, прокрутка и события продолжают обновляться.</span>
        </div>
      ) : null}
    </>
  );
}
