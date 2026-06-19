import { AlertTriangle } from "lucide-react";

import { Button, Dialog } from "@lingshu/ui";

export type ConfirmIntent = "default" | "danger";

export function ConfirmDialog({
  open,
  title,
  description,
  confirmText,
  cancelText,
  intent = "default",
  loading = false,
  onConfirm,
  onCancel
}: {
  open: boolean;
  title: string;
  description: string;
  confirmText: string;
  cancelText: string;
  intent?: ConfirmIntent;
  loading?: boolean;
  onConfirm: () => void | Promise<void>;
  onCancel: () => void;
}) {
  return (
    <Dialog
      open={open}
      title={title}
      onClose={onCancel}
      footer={(
        <>
          <Button type="button" variant="secondary" onClick={onCancel} disabled={loading}>{cancelText}</Button>
          <Button type="button" variant={intent === "danger" ? "destructive" : "default"} onClick={onConfirm} disabled={loading}>{confirmText}</Button>
        </>
      )}
    >
      <div className="flex gap-3">
        <span className="mt-0.5 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border bg-[var(--bg-subtle)] text-[var(--danger)]">
          <AlertTriangle className="h-4 w-4" />
        </span>
        <p className="text-sm leading-6 text-muted-foreground">{description}</p>
      </div>
    </Dialog>
  );
}
