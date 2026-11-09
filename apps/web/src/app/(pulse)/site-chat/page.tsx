import { Suspense } from "react";

import { SiteChatConsole } from "../../../features/site-chat-console/SiteChatConsole";

export default function SiteChatPage() {
  return (
    <Suspense fallback={<main className="page"><div className="empty-state">Загружаем Live Assist...</div></main>}>
      <SiteChatConsole />
    </Suspense>
  );
}
