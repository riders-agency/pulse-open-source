import Link from "next/link";
import type { ReactNode } from "react";

export function SidebarLink({
  href,
  icon,
  active,
  children,
}: {
  href: string;
  icon: ReactNode;
  active?: boolean;
  children: ReactNode;
}) {
  return (
    <Link className={`sidebar-link ${active ? "is-active" : ""}`} href={href} aria-current={active ? "page" : undefined}>
      {icon}
      <span>{children}</span>
    </Link>
  );
}
