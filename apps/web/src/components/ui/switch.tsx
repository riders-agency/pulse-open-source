import * as React from "react";
import { cn } from "../../lib/utils";

export function Switch({
  checked,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { checked?: boolean }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      className={cn("ui-switch", checked ? "is-checked" : "", className)}
      {...props}
    >
      <span />
    </button>
  );
}
