import * as React from "react";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { X } from "lucide-react";

import { cn } from "../lib/cn";
import { Button } from "./button";

export function Sheet({ open, title, children, footer, onClose, className, side = "right" }: { open: boolean; title?: React.ReactNode; children: React.ReactNode; footer?: React.ReactNode; onClose: () => void; className?: string; side?: "right" | "left" }) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={(nextOpen) => { if (!nextOpen) onClose(); }}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/35 data-[state=closed]:animate-out data-[state=open]:animate-in data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <DialogPrimitive.Content
          className={cn(
            "fixed top-0 z-50 flex h-dvh w-full max-w-xl flex-col border-border bg-card text-card-foreground shadow-[var(--shadow-lg)] outline-none data-[state=closed]:animate-out data-[state=open]:animate-in data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
            side === "right" ? "right-0 border-l" : "left-0 border-r",
            className
          )}
        >
          <div className="flex items-center justify-between border-b border-border p-4">
            <DialogPrimitive.Title className="font-serif text-lg font-semibold">{title}</DialogPrimitive.Title>
            <DialogPrimitive.Close asChild>
              <Button variant="ghost" size="icon" aria-label="Close">
                <X className="h-4 w-4" />
              </Button>
            </DialogPrimitive.Close>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto p-5">{children}</div>
          {footer ? <div className="flex justify-end gap-2 border-t border-border p-4">{footer}</div> : null}
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

export const Drawer = Sheet;
