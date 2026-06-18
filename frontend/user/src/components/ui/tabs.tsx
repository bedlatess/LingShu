import type * as React from "react";
import { cn } from "@/lib/utils";

export function TabsList({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div role="tablist" className={cn("inline-flex rounded-md border border-border bg-secondary p-1", className)} {...props} />;
}

export function TabsTrigger({ value, activeValue, onSelect, className, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { value: string; activeValue: string; onSelect: (value: string) => void }) {
  const active = activeValue === value;
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={() => onSelect(value)}
      className={cn(
        "rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        active && "bg-card text-foreground shadow-[0_1px_2px_rgba(20,20,19,0.04)]",
        className
      )}
      {...props}
    />
  );
}
