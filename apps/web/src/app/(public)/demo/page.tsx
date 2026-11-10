import { redirect } from "next/navigation";
import { demoEntryHref } from "../../../lib/config/demo";
export const dynamic = "force-dynamic";
// Compatibility for old shared links; there is no separate demo UI or iframe.
export default function LegacyDemoLink() { redirect(demoEntryHref()); }
