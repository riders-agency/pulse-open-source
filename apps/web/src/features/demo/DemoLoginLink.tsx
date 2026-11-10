"use client";
import { useDemoEntryHref } from "./DemoSessionEntry";
import Link from "next/link";
import { ArrowRight } from "lucide-react";

export function DemoLoginLink() {
  const href = useDemoEntryHref();
  return (
    <div className="auth-demo-entry">
      <Link href={href} prefetch={false} className="ui-button ui-button--outline ui-button--md">
        Демо-вход <ArrowRight size={16} aria-hidden="true" />
      </Link>
      <p>Без пароля. Демо-компания с задачами и работающим сценарием ИИ-отделов.</p>
      <Link href="/tech">Сценарии и архитектура проекта</Link>
    </div>
  );
}
