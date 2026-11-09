import Link from "next/link";
import { ArrowRight } from "lucide-react";

export function DemoLoginLink() {
  return (
    <div className="auth-demo-entry">
      <Link href="/demo?start=1" prefetch={false} className="ui-button ui-button--outline ui-button--md">
        Демо-вход <ArrowRight size={16} aria-hidden="true" />
      </Link>
      <p>Без почты и пароля. Отдельное пространство с примерами задач.</p>
      <Link href="/tech">Сценарии и архитектура проекта</Link>
    </div>
  );
}
