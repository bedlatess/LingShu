export { cn } from "@lingshu/ui";

export function formatMoney(value?: string | number) {
  const numberValue = Number(value ?? 0);
  if (!Number.isFinite(numberValue)) return "0.00";
  return new Intl.NumberFormat("zh-CN", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 4
  }).format(numberValue);
}

export function formatCompact(value?: string | number) {
  const numberValue = Number(value ?? 0);
  if (!Number.isFinite(numberValue)) return "0";
  return new Intl.NumberFormat("en", { notation: "compact" }).format(numberValue);
}
