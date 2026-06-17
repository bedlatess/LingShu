import * as React from "react";
import { cn } from "@/lib/utils";

export function Tabs({ value, onValueChange, children }: { value: string; onValueChange: (value: string) => void; children: React.ReactNode }) {
  return <div data-value={value} data-on-value-change={onValueChange}>{children}</div>;
}

export function TabsList({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("inline-flex rounded-lg border border-border bg-muted p-1", className)} {...props} />;
}

export function TabsTrigger({ value, activeValue, onSelect, className, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { value: string; activeValue: string; onSelect: (value: string) => void }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(value)}
      className={cn("rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground", activeValue === value && "bg-background text-foreground shadow-sm", className)}
      {...props}
    />
  );
}
