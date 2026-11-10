"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { usePulseStore } from "../../../store/pulse-store";

export default function LegacyNewClientServiceRoute() {
  const router = useRouter();
  const activeClientId = usePulseStore((s) => s.activeClientId);

  useEffect(() => {
    router.replace(activeClientId ? `/clients/${activeClientId}/services/new` : "/clients");
  }, [activeClientId, router]);

  return <main className="page" aria-label="Переход к новой услуге клиента" />;
}
