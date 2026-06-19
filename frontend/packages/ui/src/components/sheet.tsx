import * as React from "react";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { X } from "lucide-react";

import { cn } from "../lib/cn";
import { Button } from "./button";

export function Sheet({ open, title, children, footer, onClose, className, side = "right" }: { open: boolean; title?: React.ReactNode; children: React.ReactNode; footer?: React.ReactNode; onClose: () => void; className?: string; side?: "right" | "left" }) {
  const reduceMotion = useReducedMotion();
  const x = side === "right" ? 24 : -24;
  return (
    <DialogPrimitive.Root open={open} onOpenChange={(nextOpen) => { if (!nextOpen) onClose(); }}>
      <AnimatePresence>
        {open ? (
          <DialogPrimitive.Portal forceMount>
            <DialogPrimitive.Overlay asChild forceMount>
              <motion.div
                className="fixed inset-0 z-50 bg-black/35"
                initial={reduceMotion ? false : { opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15, ease: "easeOut" }}
              />
            </DialogPrimitive.Overlay>
            <DialogPrimitive.Content asChild forceMount>
              <motion.div
                className={cn(
                  "fixed top-0 z-50 flex h-dvh w-full max-w-xl flex-col border-border bg-card text-card-foreground shadow-[var(--shadow-lg)] outline-none",
                  side === "right" ? "right-0 border-l" : "left-0 border-r",
                  className
                )}
                initial={reduceMotion ? false : { opacity: 0, x }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x }}
                transition={{ duration: 0.18, ease: "easeOut" }}
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
              </motion.div>
            </DialogPrimitive.Content>
          </DialogPrimitive.Portal>
        ) : null}
      </AnimatePresence>
    </DialogPrimitive.Root>
  );
}

export const Drawer = Sheet;
