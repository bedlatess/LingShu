import React from "react";
import { Card, message, Space, Typography } from "antd";
import { designTokens } from "@lingshu/shared";

export type Pager = { page: number; limit: number; total: number };

export const modelDefaults = {
  public_name: "",
  type: "chat",
  group: "",
  billing_mode: "token",
  input_price_per_1k: "0",
  output_price_per_1k: "0",
  price_per_call: "0",
  rate_multiplier: "1.200",
  status: "enabled",
  sort_order: 0
};

export const providerOptions = [
  { value: "openai", label: "OpenAI 兼容" },
  { value: "anthropic", label: "Anthropic Claude" }
];

export function errText(err: unknown) {
  return err instanceof Error ? err.message : "未知错误";
}

const pendingWrites = new Set<string>();

export async function runWrite(action: () => Promise<void>, failPrefix = "操作失败") {
  if (pendingWrites.has(failPrefix)) return;
  pendingWrites.add(failPrefix);
  try {
    await action();
  } catch (err) {
    message.error({ content: `${failPrefix}: ${errText(err)}`, key: `write-${failPrefix}` });
    throw err;
  } finally {
    pendingWrites.delete(failPrefix);
  }
}

export function exportCSV<T extends object>(filename: string, rows: T[]) {
  if (rows.length === 0) {
    message.info("当前没有可导出的数据");
    return;
  }
  const headers = Object.keys(rows[0] as Record<string, unknown>);
  const escapeCell = (value: unknown) => `"${String(value ?? "").replace(/"/g, '""')}"`;
  const csv = [
    headers.join(","),
    ...rows.map((row) => {
      const record = row as Record<string, unknown>;
      return headers.map((header) => escapeCell(record[header])).join(",");
    })
  ].join("\n");
  const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

export async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // Fall through to textarea fallback.
  }
  try {
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
  } catch {
    return false;
  }
}

export function fmtMoney(v?: string | number) {
  const n = Number(v ?? 0);
  return Number.isFinite(n) ? n.toLocaleString("zh-CN", { minimumFractionDigits: 2, maximumFractionDigits: 6 }) : "0";
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
    <div style={{ display: "flex", alignItems: "flex-end", gap: 8, height: 140 }}>
      {data.map((d) => (
        <div key={d.label} style={{ flex: 1, minWidth: 28, textAlign: "center" }}>
          <div style={{ height: `${(d.value / max) * 110}px`, background: designTokens.colors.clay, borderRadius: 4 }} title={`${d.value}`} />
          <div style={{ fontSize: 12, color: designTokens.colors.inkFaint, marginTop: 4 }}>{d.label.length > 5 ? d.label.slice(5) : d.label}</div>
        </div>
      ))}
    </div>
  );
}

export function metricCards(items: { label: string; value: React.ReactNode }[]) {
  return (
    <Space wrap>
      {items.map((item) => (
        <Card key={item.label}>
          <Typography.Text type="secondary">{item.label}</Typography.Text>
          <Typography.Title level={5} style={{ fontFamily: designTokens.font.serif }}>{item.value}</Typography.Title>
        </Card>
      ))}
    </Space>
  );
}

export function tablePagination(pager: Pager, setPager: React.Dispatch<React.SetStateAction<Pager>>) {
  return {
    current: pager.page,
    pageSize: pager.limit,
    total: pager.total,
    showSizeChanger: true,
    onChange: (page: number, pageSize: number) => setPager((prev) => ({ ...prev, page, limit: pageSize }))
  };
}

export function normalizeModelPayload(values: { sort_order?: number | string } & Record<string, unknown>) {
  return {
    ...modelDefaults,
    ...values,
    sort_order: Number(values.sort_order ?? 0)
  };
}
