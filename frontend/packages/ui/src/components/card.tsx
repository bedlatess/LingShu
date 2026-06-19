import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cn } from "../lib/cn";

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  asChild?: boolean;
}

export const Card = React.forwardRef<HTMLDivElement, CardProps>(({ className, asChild = false, ...props }, ref) => {
  const Comp = asChild ? Slot : "div";
  return <Comp ref={ref} className={cn("rounded-lg border border-border bg-card text-card-foreground shadow-[var(--shadow-xs)] transition-shadow duration-200 ease-out hover:shadow-[var(--shadow-sm)]", className)} {...props} />;
});
Card.displayName = "Card";

export function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex flex-col gap-1.5 p-6", className)} {...props} />;
}

export function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn("font-serif text-lg font-semibold leading-tight", className)} {...props} />;
}

export function CardDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("text-sm leading-6 text-muted-foreground", className)} {...props} />;
}

export function CardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("p-6 pt-0", className)} {...props} />;
}

export function CardFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex items-center gap-3 border-t border-border p-4", className)} {...props} />;
}
