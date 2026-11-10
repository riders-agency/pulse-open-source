import { useEffect } from "react";

export function useAuthGateEnvironment({
  hydrated,
  setHydrated,
  theme,
}: {
  hydrated: boolean;
  setHydrated: (value: boolean) => void;
  theme: string;
}) {
  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  useEffect(() => {
    if (hydrated) return;
    const timer = window.setTimeout(() => setHydrated(true), 500);
    return () => window.clearTimeout(timer);
  }, [hydrated, setHydrated]);
}
