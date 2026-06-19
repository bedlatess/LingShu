import { i18n } from "../i18n";

export function trStatus(value?: string) {
  if (!value) return "-";
  return i18n.t(`common:status.${value}`, value);
}

export function trBillingMode(value?: string) {
  if (!value) return "-";
  return i18n.t(`common:billingMode.${value}`, value);
}

export function trType(value?: string) {
  if (!value) return "-";
  return i18n.t(`common:type.${value}`, value);
}

export function trLedgerType(value?: string) {
  if (!value) return "-";
  return i18n.t(`common:ledgerType.${value}`, value);
}
