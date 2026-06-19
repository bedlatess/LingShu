import * as React from "react";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";

export function Popover({ trigger, children, align = "center" }: { trigger: React.ReactNode; children: React.ReactNode; align?: "start" | "center" | "end" }) {
  const [open, setOpen] = React.useState(false);
  const reduceMotion = useReducedMotion();
  return (
    <PopoverPrimitive.Root open={open} onOpenChange={setOpen}>
      <PopoverPrimitive.Trigger asChild>
        <button type="button" className="inline-flex cursor-pointer">{trigger}</button>
      </PopoverPrimitive.Trigger>
      <AnimatePresence>
        {open ? (
          <PopoverPrimitive.Portal forceMount>
            <PopoverPrimitive.Content asChild forceMount align={align} sideOffset={8}>
              <motion.div
                className="z-50 w-72 rounded-md border border-border bg-card p-4 text-foreground shadow-[var(--shadow-md)] outline-none"
                initial={reduceMotion ? false : { opacity: 0, scale: 0.96 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.96 }}
                transition={{ duration: 0.15, ease: "easeOut" }}
              >
                {children}
              </motion.div>
            </PopoverPrimitive.Content>
          </PopoverPrimitive.Portal>
        ) : null}
      </AnimatePresence>
    </PopoverPrimitive.Root>
  );
}

export const PopoverRoot = PopoverPrimitive.Root;
export const PopoverTrigger = PopoverPrimitive.Trigger;
export const PopoverContent = PopoverPrimitive.Content;
