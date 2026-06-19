import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "../lib/cn";

const buttonVariants = cva(
  "inline-flex min-h-9 cursor-pointer items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-[background-color,border-color,color,box-shadow,transform] duration-150 ease-out active:scale-[0.98] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--clay)]/30 focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 disabled:active:scale-100",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-[var(--clay-hover)]",
        secondary: "border border-border bg-card text-foreground hover:border-[var(--border-strong)] hover:bg-[var(--bg-subtle)]",
        ghost: "text-foreground hover:bg-[var(--bg-subtle)]",
        outline: "border border-input bg-transparent text-foreground hover:bg-[var(--bg-subtle)]",
        destructive: "bg-destructive text-destructive-foreground hover:bg-[#8f3328]",
        link: "h-auto min-h-0 p-0 text-[var(--clay)] underline-offset-4 hover:underline"
      },
      size: {
        default: "h-9 px-4",
        sm: "h-8 px-3 text-[13px]",
        lg: "h-10 px-5",
        icon: "h-9 w-9 p-0"
      }
    },
    defaultVariants: { variant: "default", size: "default" }
  }
);

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(({ className, variant, size, asChild = false, ...props }, ref) => {
  const Comp = asChild ? Slot : "button";
  return <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />;
});
Button.displayName = "Button";
