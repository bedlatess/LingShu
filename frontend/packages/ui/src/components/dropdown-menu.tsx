import * as React from "react";
import * as DropdownPrimitive from "@radix-ui/react-dropdown-menu";

import { cn } from "../lib/cn";

export function DropdownMenu({ trigger, children, align = "right" }: { trigger: React.ReactNode; children: React.ReactNode; align?: "left" | "right" }) {
  return (
    <DropdownPrimitive.Root>
      <DropdownPrimitive.Trigger asChild>
        <button type="button" className="inline-flex cursor-pointer">{trigger}</button>
      </DropdownPrimitive.Trigger>
      <DropdownPrimitive.Portal>
        <DropdownPrimitive.Content
          align={align === "right" ? "end" : "start"}
          sideOffset={8}
          className="z-50 min-w-40 rounded-md border border-border bg-card p-1 text-foreground shadow-[var(--shadow-md)] outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0"
        >
          {children}
        </DropdownPrimitive.Content>
      </DropdownPrimitive.Portal>
    </DropdownPrimitive.Root>
  );
}

export function DropdownItem({ children, onClick, destructive }: { children: React.ReactNode; onClick?: () => void; destructive?: boolean }) {
  return (
    <DropdownPrimitive.Item
      onSelect={onClick}
      className={cn("flex cursor-pointer select-none items-center gap-2 rounded-[4px] px-3 py-2 text-sm outline-none transition-colors focus:bg-[var(--bg-subtle)] data-[disabled]:pointer-events-none data-[disabled]:opacity-50", destructive ? "text-[var(--danger)]" : "text-foreground")}
    >
      {children}
    </DropdownPrimitive.Item>
  );
}

export const DropdownSeparator = DropdownPrimitive.Separator;
