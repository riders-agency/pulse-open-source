"use client";

import { usePulseStore } from "../../../store/pulse-store";
import { ClientRedirectPage } from "./ClientRedirectPage";

export default function LegacyClientRoute() {
  const activeClientId = usePulseStore((s) => s.activeClientId);

  return <ClientRedirectPage label="Переход к клиенту" to={activeClientId ? `/clients/${activeClientId}` : "/clients"} />;
}
