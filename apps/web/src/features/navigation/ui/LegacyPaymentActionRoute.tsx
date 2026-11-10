"use client";

import { usePulseStore } from "../../../store/pulse-store";
import { ClientRedirectPage } from "./ClientRedirectPage";

export default function LegacyPaymentActionRoute() {
  const activePaymentActionId = usePulseStore((s) => s.activePaymentActionId);

  return (
    <ClientRedirectPage
      label="Переход к оплате"
      to={activePaymentActionId ? `/payment-actions/${activePaymentActionId}` : "/project-payments"}
    />
  );
}
