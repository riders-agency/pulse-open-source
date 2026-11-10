import * as React from "react";
import { cn } from "../../lib/utils";

export function Table({ className, ...props }: React.TableHTMLAttributes<HTMLTableElement>) {
  return (
    <div className="ui-table-wrap">
      <table className={cn("ui-table", className)} {...props} />
    </div>
  );
}
