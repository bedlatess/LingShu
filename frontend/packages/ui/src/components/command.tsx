import * as React from "react";
import { Command as CommandPrimitive } from "cmdk";
import { Check, Search } from "lucide-react";

import { cn } from "../lib/cn";
import { Dialog } from "./dialog";

export function Command({
  open,
  onClose,
  items,
  title = "Command",
  placeholder = "Search pages or actions",
  emptyText = "No matches found"
}: {
  open: boolean;
  onClose: () => void;
  items: { label: string; hint?: string; onSelect: () => void }[];
  title?: string;
  placeholder?: string;
  emptyText?: string;
}) {
  return (
    <Dialog open={open} onClose={onClose} title={title} className="max-w-xl">
      <CommandPrimitive className="grid gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <CommandPrimitive.Input
            autoFocus
            placeholder={placeholder}
            className="flex h-10 w-full rounded-md border border-input bg-card px-9 py-2 text-sm text-foreground outline-none transition-colors placeholder:text-[var(--ink-faint)] focus-visible:border-[var(--clay)] focus-visible:ring-2 focus-visible:ring-ring/30"
          />
        </div>
        <CommandPrimitive.List className="max-h-80 overflow-y-auto">
          <CommandPrimitive.Empty className="py-6 text-center text-sm text-muted-foreground">{emptyText}</CommandPrimitive.Empty>
          <CommandPrimitive.Group className="grid gap-1">
            {items.map((item) => (
              <CommandPrimitive.Item
                key={`${item.label}-${item.hint ?? ""}`}
                value={`${item.label} ${item.hint ?? ""}`}
                onSelect={() => {
                  item.onSelect();
                  onClose();
                }}
                className={cn("flex cursor-pointer items-center justify-between rounded-md px-3 py-2 text-left text-sm outline-none aria-selected:bg-[var(--bg-subtle)]")}
              >
                <span className="text-foreground">{item.label}</span>
                {item.hint ? <span className="text-xs text-muted-foreground">{item.hint}</span> : <Check className="h-4 w-4 text-[var(--clay)]" />}
              </CommandPrimitive.Item>
            ))}
          </CommandPrimitive.Group>
        </CommandPrimitive.List>
      </CommandPrimitive>
    </Dialog>
  );
}
