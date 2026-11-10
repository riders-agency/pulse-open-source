import * as React from "react";
import { cn } from "../../lib/utils";

export function Tabs({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("ui-tabs", className)} {...props} />;
}

export function TabButton({
  active,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { active?: boolean }) {
  return <button type="button" className={cn("ui-tab", active ? "is-active" : "", className)} {...props} />;
}
