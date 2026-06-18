import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva("inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium", {
  variants: {
    variant: {
      default: "border-transparent bg-[var(--bg-subtle)] text-foreground",
      clay: "border-transparent bg-[var(--clay-soft)] text-[var(--clay-hover)]",
      success: "border-transparent bg-[#E6EDE5] text-[#3D6B3B]",
      danger: "border-transparent bg-[var(--clay-soft)] text-[var(--danger)]",
      outline: "border-border text-muted-foreground"
    }
  },
  defaultVariants: {
    variant: "default"
  }
});

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />;
}
