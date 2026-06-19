import React from "react";
import { toast } from "@lingshu/ui";
import { i18n } from "../../i18n";

export type Pager = { page: number; limit: number; total: number };

export const modelDefaults = {
  public_name: "",
  type: "chat",
  group: "",
  billing_mode: "token",
  input_price_per_1k: "0",
  output_price_per_1k: "0",
  cache_creation_price_per_1k: "0",
  cache_read_price_per_1k: "0",
  price_per_call: "0",
  rate_multiplier: "1.200",
  status: "enabled",
  sort_order: 0
};

export const providerOptions = [
  { value: "openai", label: i18n.t("admin:providers.openai") },
  { value: "anthropic", label: i18n.t("admin:providers.anthropic") }
];

export function errText(err: unknown) {
  return err instanceof Error ? err.message : i18n.t("admin:common.unknownError");
}

const pendingWrites = new Set<string>();

export async function runWrite(action: () => Promise<void>, failPrefix = i18n.t("admin:common.actionFailed")) {
  if (pendingWrites.has(failPrefix)) return;
  pendingWrites.add(failPrefix);
  try {
    await action();
  } catch (err) {
    toast.error(`${failPrefix}: ${errText(err)}`);
    throw err;
  } finally {
    pendingWrites.delete(failPrefix);
  }
}

export function exportCSV<T extends object>(filename: string, rows: T[]) {
  if (rows.length === 0) {
    toast.info(i18n.t("admin:common.noExportData"));
    return;
  }
  const headers = Object.keys(rows[0] as Record<string, unknown>);
  const escapeCell = (value: unknown) => `"${String(value ?? "").replace(/"/g, '""')}"`;
  const csv = [headers.join(","), ...rows.map((row) => {
    const record = row as Record<string, unknown>;
    return headers.map((header) => escapeCell(record[header])).join(",");
  })].join("\n");
  const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

export async function downloadBlob(filename: string, load: () => Promise<Blob>) {
  try {
    const blob = await load();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
  } catch (err) {
    toast.error(`${i18n.t("admin:common.exportFailed")}: ${errText(err)}`);
  }
}

export async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // Use textarea fallback below.
  }
  const ta = document.createElement("textarea");
  ta.value = text;
  ta.style.position = "fixed";
  ta.style.opacity = "0";
  document.body.appendChild(ta);
  ta.focus();
  ta.select();
  const ok = document.execCommand("copy");
  document.body.removeChild(ta);
  return ok;
}

export function fmtMoney(v?: string | number) {
  const n = Number(v ?? 0);
  return Number.isFinite(n) ? n.toLocaleString(i18n.resolvedLanguage === "zh" ? "zh-CN" : "en-US", { minimumFractionDigits: 2, maximumFractionDigits: 6 }) : "0";
}

export function formatDateMinute(value?: string | Date | null, empty = "-") {
  if (!value) return empty;
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return empty;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function MiniBars({ data }: { data: { label: string; value: number }[] }) {
  const max = Math.max(1, ...data.map((d) => d.value));
  return (
    <div className="flex h-36 items-end gap-2">
      {data.map((d) => (
        <div key={d.label} className="min-w-7 flex-1 text-center">
          <div className="rounded-t bg-[var(--clay)]" style={{ height: `${Math.max(4, (d.value / max) * 120)}px` }} title={`${d.value}`} />
          <div className="mt-1 truncate text-[11px] text-muted-foreground">{d.label.length > 8 ? d.label.slice(-5) : d.label}</div>
        </div>
      ))}
    </div>
  );
}

export function normalizeModelPayload(values: { sort_order?: number | string } & Record<string, unknown>) {
  return { ...modelDefaults, ...values, sort_order: Number(values.sort_order ?? 0) };
}

export function statusVariant(status?: string) {
  if (status === "active" || status === "enabled" || status === "success" || status === "healthy") return "success" as const;
  if (status === "disabled" || status === "banned" || status === "failed" || status === "unhealthy") return "danger" as const;
  return "muted" as const;
}

export function SimpleForm({ children, onSubmit, className }: { children: React.ReactNode; onSubmit: (event: React.FormEvent<HTMLFormElement>) => void; className?: string }) {
  return <form className={className ?? "grid gap-4"} onSubmit={onSubmit}>{children}</form>;
}
