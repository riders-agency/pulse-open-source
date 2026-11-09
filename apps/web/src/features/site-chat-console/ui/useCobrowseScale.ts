import { useEffect, useState, type RefObject } from "react";

export function useCobrowseScale(frameRef: RefObject<HTMLDivElement | null>, width: number, height: number) {
  const [scale, setScale] = useState(1);

  useEffect(() => {
    const node = frameRef.current;
    if (!node) return;
    const resize = () => {
      const rect = node.getBoundingClientRect();
      setScale(Math.max(0.08, Math.min(rect.width / width, rect.height / height, 1.4)));
    };
    resize();
    const observer = new ResizeObserver(resize);
    observer.observe(node);
    return () => observer.disconnect();
  }, [frameRef, height, width]);

  return scale;
}
