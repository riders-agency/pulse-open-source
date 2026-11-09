import { useCallback, useEffect, useRef, type RefObject } from "react";

export function useAIEmployeeScroll(chatLogRef: RefObject<HTMLDivElement | null>, open: boolean) {
  const follow = useRef(true);
  useEffect(() => {
    const element = chatLogRef.current;
    follow.current = true;
    if (!element) return;
    const onScroll = () => { follow.current = element.scrollHeight - element.scrollTop - element.clientHeight < 80; };
    element.addEventListener("scroll", onScroll, { passive: true });
    return () => element.removeEventListener("scroll", onScroll);
  }, [chatLogRef, open]);
  return useCallback((behavior: ScrollBehavior = "smooth", force = false) => {
    if (force) follow.current = true;
    if (!follow.current) return;
    requestAnimationFrame(() => {
      const element = chatLogRef.current;
      if (element && follow.current) element.scrollTo({ top: element.scrollHeight, behavior });
    });
  }, [chatLogRef]);
}
