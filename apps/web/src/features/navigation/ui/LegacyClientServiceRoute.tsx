"use client";

import { usePulseStore } from "../../../store/pulse-store";
import { ClientRedirectPage } from "./ClientRedirectPage";

export default function LegacyClientServiceRoute() {
  const activeClientId = usePulseStore((s) => s.activeClientId);
  const activeClientServiceId = usePulseStore((s) => s.activeClientServiceId);
  const to = activeClientServiceId
    ? `/client-services/${activeClientServiceId}`
    : activeClientId
      ? `/clients/${activeClientId}`
      : "/clients";

  return <ClientRedirectPage label="Переход к услуге клиента" to={to} />;
}
