import * as React from "react";
import { Command as CommandIcon } from "lucide-react";
import { cva, type VariantProps } from "class-variance-authority";
import { Toaster as Sonner, toast } from "sonner";

import { cn } from "../lib/cn";
import { Button } from "./button";
import { Card, CardContent } from "./card";

export function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("animate-pulse rounded-md bg-[var(--bg-subtle)]", className)} {...props} />;
}

const badgeVariants = cva("inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium", {
  variants: {
    variant: {
      default: "border-border bg-card text-foreground",
      success: "border-[var(--success)]/30 bg-[var(--success-soft)] text-[var(--success)]",
      warning: "border-[var(--warning)]/30 bg-[var(--warning-soft)] text-[var(--warning)]",
      danger: "border-[var(--danger)]/30 bg-[var(--danger-soft)] text-[var(--danger)]",
      info: "border-[var(--info)]/30 bg-[var(--info-soft)] text-[var(--info)]",
      muted: "border-border bg-[var(--bg-subtle)] text-muted-foreground"
    }
  },
  defaultVariants: { variant: "default" }
});

export function Badge({ className, variant, ...props }: React.HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badgeVariants>) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />;
}

export const Tag = Badge;

export function Separator({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("h-px w-full bg-border", className)} {...props} />;
}

export function Progress({ value = 0, className }: { value?: number; className?: string }) {
  const safe = Math.max(0, Math.min(100, value));
  return (
    <div className={cn("h-2 overflow-hidden rounded-full bg-[var(--bg-subtle)]", className)} role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={safe}>
      <div className="h-full rounded-full bg-[var(--clay)] transition-all" style={{ width: `${safe}%` }} />
    </div>
  );
}

export function Avatar({ name, src, className }: { name?: string; src?: string; className?: string }) {
  const initials = (name ?? "LS").slice(0, 2).toUpperCase();
  return (
    <div className={cn("grid h-9 w-9 place-items-center overflow-hidden rounded-md border border-border bg-[var(--bg-subtle)] text-xs font-semibold text-foreground", className)}>
      {src ? <img src={src} alt={name ?? "avatar"} className="h-full w-full object-cover" /> : initials}
    </div>
  );
}

export function Alert({ title, children, variant = "info", className }: { title?: string; children?: React.ReactNode; variant?: "info" | "success" | "warning" | "danger"; className?: string }) {
  const map = {
    info: "border-[var(--info)]/30 bg-[var(--info-soft)] text-[var(--info)]",
    success: "border-[var(--success)]/30 bg-[var(--success-soft)] text-[var(--success)]",
    warning: "border-[var(--warning)]/30 bg-[var(--warning-soft)] text-[var(--warning)]",
    danger: "border-[var(--danger)]/30 bg-[var(--danger-soft)] text-[var(--danger)]"
  };
  return (
    <div className={cn("rounded-md border px-4 py-3 text-sm", map[variant], className)} role={variant === "danger" ? "alert" : "status"}>
      {title ? <p className="font-medium">{title}</p> : null}
      {children ? <div className={cn("leading-6", title && "mt-1")}>{children}</div> : null}
    </div>
  );
}

export function EmptyState({ title, description, action, onAction, icon }: { title: string; description: string; action?: string; onAction?: () => void; icon?: React.ReactNode }) {
  return (
    <div className="grid place-items-center rounded-lg border border-dashed border-border bg-card p-8 text-center">
      <div className="mb-4 grid h-12 w-12 place-items-center rounded-md border border-border bg-[var(--bg-subtle)] text-[var(--clay)]">{icon ?? <CommandIcon className="h-5 w-5" />}</div>
      <h3 className="font-serif text-base font-semibold text-foreground">{title}</h3>
      <p className="mt-2 max-w-sm text-sm leading-6 text-muted-foreground">{description}</p>
      {action && onAction ? (
        <Button className="mt-4" variant="secondary" onClick={onAction}>
          {action}
        </Button>
      ) : null}
    </div>
  );
}

export function PageHeader({ eyebrow, title, description, action }: { eyebrow?: string; title: string; description?: string; action?: React.ReactNode }) {
  return (
    <div className="mb-8 flex flex-col gap-5 border-b border-border pb-6 md:flex-row md:items-end md:justify-between">
      <div className="flex min-w-0 flex-col gap-3">
        {eyebrow ? <p className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--clay)]">{eyebrow}</p> : null}
        <h1 className="max-w-3xl font-serif text-3xl font-semibold leading-[1.1] tracking-[-0.02em] text-foreground sm:text-[2.5rem]">{title}</h1>
        {description ? <p className="max-w-2xl text-[15px] leading-7 text-muted-foreground">{description}</p> : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}

export function StatCard({ label, value, hint, icon: Icon, trend }: { label: string; value: React.ReactNode; hint?: string; icon?: React.ComponentType<{ className?: string }>; trend?: string }) {
  return (
    <Card className="group overflow-hidden transition-colors hover:border-[var(--border-strong)]">
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <p className="text-sm text-muted-foreground">{label}</p>
            <strong className="mt-2 block truncate font-serif text-2xl font-semibold tracking-[-0.02em] text-foreground">{value}</strong>
          </div>
          {Icon ? (
            <div className="grid h-10 w-10 shrink-0 place-items-center rounded-md border border-border bg-[var(--bg-subtle)] text-[var(--clay)]">
              <Icon className="h-5 w-5" />
            </div>
          ) : null}
        </div>
        {hint || trend ? <p className="mt-4 text-xs text-muted-foreground">{trend ? <span className="text-[var(--success)]">{trend} </span> : null}{hint}</p> : null}
      </CardContent>
    </Card>
  );
}

export function Toaster() {
  return <Sonner richColors position="top-center" toastOptions={{ duration: 3500 }} />;
}

export { toast };
