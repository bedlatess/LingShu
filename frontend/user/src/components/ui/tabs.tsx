import type * as React from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { cn } from "@/lib/utils";

export const Tabs = TabsPrimitive.Root;

export function TabsTrigger({ value, activeValue, onSelect, className, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { value: string; activeValue: string; onSelect: (value: string) => void }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(value)}
      className={cn(
        "rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        activeValue === value && "bg-card text-foreground shadow-[0_1px_2px_rgba(20,20,19,0.04)]",
        className
      )}
      {...props}
    />
  );
}

export function TabsList({ className, ...props }: React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>) {
  return <TabsPrimitive.List className={cn("inline-flex rounded-md border border-border bg-secondary p-1", className)} {...props} />;
}

export function TabsContent({ className, ...props }: React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>) {
  return <TabsPrimitive.Content className={cn("focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", className)} {...props} />;
}
