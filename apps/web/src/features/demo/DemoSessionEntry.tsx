"use client";
import { createContext, useContext, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { startDemo } from "./startDemo";
import { usePulseStore } from "../../store/pulse-store";
const DemoEntryContext = createContext("/project-workflows?demo=1");
export const useDemoEntryHref = () => useContext(DemoEntryContext);
export function DemoSessionEntry({ href, children }: { href: string; children: React.ReactNode }) {
  const router = useRouter();
  const autoStart = useSearchParams().get("demo") === "1";
  const hydrated = usePulseStore((s) => s.hydrated);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!hydrated || !autoStart) return;
    let active = true;
    void startDemo().then(() => { if (active) router.replace("/project-workflows"); }).catch((err: unknown) => {
      if (active) setError(err instanceof Error ? err.message : "Не удалось открыть демо-компанию");
    });
    return () => { active = false; };
  }, [href, hydrated, router, autoStart]);
  return <DemoEntryContext.Provider value={href}>{error && <div role="alert" className="form-error">{error}</div>}{children}</DemoEntryContext.Provider>;
}
