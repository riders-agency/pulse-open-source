import { redirect } from "next/navigation";
import { demoEntryHref } from "../../../../lib/config/demo";
export const dynamic = "force-dynamic";
export default function OfficePage() { redirect(demoEntryHref()); }
