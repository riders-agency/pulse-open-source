import type { Metadata } from "next";
import { DemoWelcome } from "../../../features/demo/DemoWelcome";
import { DemoEmbed } from "../../../features/demo/DemoEmbed";
import { externalDemoUrl, publicDemoEnabled } from "../../../lib/config/demo";

export const dynamic = "force-dynamic";
export const metadata: Metadata = { title: "Попробовать Pulse — живое демо", robots: { index: false, follow: true } };
export default async function DemoPage({ searchParams }: { searchParams: Promise<{ start?: string }> }) {
  const autoStart = (await searchParams).start === "1";
  const enabled = publicDemoEnabled();
  const destination = enabled ? null : externalDemoUrl(autoStart);
  if (destination) return <DemoEmbed src={destination} />;
  return <DemoWelcome enabled={enabled} autoStart={autoStart} />;
}
