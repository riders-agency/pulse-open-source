import type { Metadata } from "next";
import { OfficeDemo } from "../../../../features/tech/OfficeDemo";
import { officeDemoEndpoint } from "../../../../lib/config/officeDemo";

export const dynamic = "force-dynamic";
export const metadata: Metadata = { title: "ИИ-отделы в работе — Pulse", description: "Почта, классификация, поручения отделам, обновление учебного сайта и черновик счёта. Исполняемый граф LangGraph и подтверждение человека." };
export default function OfficePage() { return <OfficeDemo endpoint={officeDemoEndpoint()} />; }
