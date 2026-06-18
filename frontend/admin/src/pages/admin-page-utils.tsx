import React from "react";
import { Card, message, Space, Typography } from "antd";

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

export async function runWrite(action: () => Promise<void>, failPrefix = "操作失败") {
  try {
    await action();
  } catch (err) {
    message.error(`${failPrefix}: ${errText(err)}`);
    throw err;
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

export function metricCards(items: { label: string; value: React.ReactNode }[]) {
  return (
    <Space wrap>
      {items.map((item) => (
        <Card key={item.label}>
          <Typography.Text type="secondary">{item.label}</Typography.Text>
          <Typography.Title level={5}>{item.value}</Typography.Title>
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
