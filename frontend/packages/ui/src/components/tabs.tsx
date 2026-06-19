import * as React from "react";
import { cn } from "../lib/cn";

export function Tabs({ tabs, value, onChange, className }: { tabs: { value: string; label: React.ReactNode }[]; value: string; onChange: (value: string) => void; className?: string }) {
  return (
    <div className={cn("inline-flex rounded-md border border-border bg-[var(--bg-subtle)] p-1", className)}>
      {tabs.map((tab) => (
        <button
          type="button"
          key={tab.value}
          onClick={() => onChange(tab.value)}
          className={cn("cursor-pointer rounded-[4px] px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground", value === tab.value && "bg-card text-foreground shadow-[var(--shadow-xs)]")}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}

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
      className={cn("cursor-pointer rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30", active && "bg-card text-foreground shadow-[var(--shadow-xs)]", className)}
      {...props}
    />
  );
}
