import * as React from "react";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { X } from "lucide-react";

import { cn } from "../lib/cn";

export function Dialog({ open, title, children, footer, onClose, className }: { open: boolean; title?: React.ReactNode; children: React.ReactNode; footer?: React.ReactNode; onClose: () => void; className?: string }) {
  const reduceMotion = useReducedMotion();
  return (
    <DialogPrimitive.Root open={open} onOpenChange={(nextOpen) => { if (!nextOpen) onClose(); }}>
      <AnimatePresence>
        {open ? (
          <DialogPrimitive.Portal forceMount>
            <DialogPrimitive.Overlay asChild forceMount>
              <motion.div
                className="fixed inset-0 z-50 bg-[rgba(20,20,19,0.28)] backdrop-blur-[2px]"
                initial={reduceMotion ? false : { opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.18, ease: "easeOut" }}
              />
            </DialogPrimitive.Overlay>
            <DialogPrimitive.Content asChild forceMount>
              <motion.div
                className={cn(
                  "fixed left-1/2 top-1/2 z-50 w-[calc(100vw-2rem)] max-w-lg rounded-xl border border-border bg-card text-card-foreground shadow-[0_8px_40px_rgba(20,20,19,0.12),0_2px_8px_rgba(20,20,19,0.06)] outline-none",
                  className
                )}
                style={{ translate: "-50% -50%" }}
                initial={reduceMotion ? false : { opacity: 0, scale: 0.96, y: -6 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.96, y: -6 }}
                transition={{ duration: 0.18, ease: [0.22, 1, 0.36, 1] }}
              >
                <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
                  <DialogPrimitive.Title className="font-serif text-lg font-semibold tracking-tight">{title}</DialogPrimitive.Title>
                  <DialogPrimitive.Close asChild>
                    <button
                      type="button"
                      aria-label="Close"
                      className="grid size-8 place-items-center rounded-md text-muted-foreground transition-colors hover:bg-[var(--bg-subtle)] hover:text-foreground"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </DialogPrimitive.Close>
                </div>
                <div className="px-5 py-4">{children}</div>
                {footer ? <div className="flex flex-row-reverse gap-2 border-t border-border px-5 py-3.5">{footer}</div> : null}
              </motion.div>
            </DialogPrimitive.Content>
          </DialogPrimitive.Portal>
        ) : null}
      </AnimatePresence>
    </DialogPrimitive.Root>
  );
}

export const DialogRoot = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;
export const DialogClose = DialogPrimitive.Close;
