import * as React from "react";
import { cn } from "../../lib/utils";

export function Badge({
  className,
  tone = "default",
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & { tone?: "default" | "success" | "warning" | "danger" | "info" | "muted" }) {
  return <span className={cn("ui-badge", `ui-badge--${tone}`, className)} {...props} />;
}
