import * as React from "react";
import { ChevronDown } from "lucide-react";

import { cn } from "../lib/cn";

export function Form({ className, ...props }: React.FormHTMLAttributes<HTMLFormElement>) {
  return <form className={cn("grid gap-4", className)} {...props} />;
}

export const Label = React.forwardRef<HTMLLabelElement, React.LabelHTMLAttributes<HTMLLabelElement>>(({ className, ...props }, ref) => (
  <label ref={ref} className={cn("text-sm font-medium leading-none text-foreground", className)} {...props} />
));
Label.displayName = "Label";

export function Field({ label, hint, error, children, className }: { label?: string; hint?: string; error?: string; children: React.ReactNode; className?: string }) {
  const id = React.useId();
  const describedBy = error ? `${id}-error` : hint ? `${id}-hint` : undefined;
  return (
    <div className={cn("grid gap-2", className)}>
      {label ? <Label htmlFor={id}>{label}</Label> : null}
      {React.isValidElement(children) ? React.cloneElement(children as React.ReactElement<Record<string, unknown>>, { id, "aria-describedby": describedBy, "aria-invalid": Boolean(error) }) : children}
      {error ? <p id={`${id}-error`} role="alert" className="text-xs text-[var(--danger)]">{error}</p> : hint ? <p id={`${id}-hint`} className="text-xs leading-5 text-muted-foreground">{hint}</p> : null}
    </div>
  );
}

export const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(({ className, type, ...props }, ref) => (
  <input
    type={type}
    className={cn(
      "flex h-9 w-full rounded-md border border-input bg-card px-3 py-2 text-sm text-foreground transition-colors placeholder:text-[var(--ink-faint)] focus-visible:border-[var(--clay)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50",
      className
    )}
    ref={ref}
    {...props}
  />
));
Input.displayName = "Input";

export const Textarea = React.forwardRef<HTMLTextAreaElement, React.TextareaHTMLAttributes<HTMLTextAreaElement>>(({ className, ...props }, ref) => (
  <textarea
    className={cn(
      "min-h-24 w-full rounded-md border border-input bg-card px-3 py-2 text-sm leading-6 text-foreground transition-colors placeholder:text-[var(--ink-faint)] focus-visible:border-[var(--clay)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50",
      className
    )}
    ref={ref}
    {...props}
  />
));
Textarea.displayName = "Textarea";

export function Select({ className, children, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <span className="relative inline-flex w-full">
      <select
        className={cn("h-9 w-full appearance-none rounded-md border border-input bg-card px-3 py-2 pr-8 text-sm text-foreground transition-colors focus-visible:border-[var(--clay)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50", className)}
        {...props}
      >
        {children}
      </select>
      <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    </span>
  );
}

export function Switch({ checked, onCheckedChange, className, ...props }: { checked?: boolean; onCheckedChange?: (checked: boolean) => void; className?: string } & Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "onChange">) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      className={cn("inline-flex h-6 w-11 cursor-pointer items-center rounded-full border border-border bg-[var(--bg-subtle)] p-0.5 transition-colors data-[checked=true]:bg-[var(--clay)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50", className)}
      data-checked={checked}
      onClick={(event) => {
        props.onClick?.(event);
        if (!props.disabled) onCheckedChange?.(!checked);
      }}
      {...props}
    >
      <span className={cn("h-5 w-5 rounded-full bg-card shadow-[var(--shadow-sm)] transition-transform", checked && "translate-x-5")} />
    </button>
  );
}
