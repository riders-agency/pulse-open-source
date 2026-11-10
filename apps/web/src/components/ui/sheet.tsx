import * as React from "react";
import { cn } from "../../lib/utils";

export function Sheet({
  open,
  children,
  className,
}: {
  open: boolean;
  children: React.ReactNode;
  className?: string;
}) {
  if (!open) return null;
  return <aside className={cn("ui-sheet", className)}>{children}</aside>;
}

export function SheetHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("ui-sheet-header", className)} {...props} />;
}

export function SheetTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h2 className={cn("ui-sheet-title", className)} {...props} />;
}

export function SheetDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("ui-sheet-description", className)} {...props} />;
}
