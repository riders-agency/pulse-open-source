import { AuthGate } from "../../features/auth/AuthGate";
import { DemoSessionEntry } from "../../features/demo/DemoSessionEntry";
import { demoEntryHref } from "../../lib/config/demo";
export const dynamic = "force-dynamic";
export default function PulseAppLayout({ children }: { children: React.ReactNode }) {
  return <DemoSessionEntry href={demoEntryHref()}><AuthGate>{children}</AuthGate></DemoSessionEntry>;
}
