import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatMoney(value?: string | number) {
  const numberValue = Number(value ?? 0);
  if (!Number.isFinite(numberValue)) return "0.000000";
  return numberValue.toFixed(6);
}

export function formatCompact(value?: string | number) {
  const numberValue = Number(value ?? 0);
  if (!Number.isFinite(numberValue)) return "0";
  return new Intl.NumberFormat("en", { notation: "compact" }).format(numberValue);
}
