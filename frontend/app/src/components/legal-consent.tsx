import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

export function LegalConsent({ checked, onCheckedChange, className }: { checked: boolean; onCheckedChange: (checked: boolean) => void; className?: string }) {
  const { t } = useTranslation("auth");
  return (
    <label className={cn("flex items-start gap-3 rounded-md border border-border bg-[var(--bg-subtle)] p-3 text-xs leading-5 text-muted-foreground", className)}>
      <input
        type="checkbox"
        checked={checked}
        onChange={(event) => onCheckedChange(event.target.checked)}
        className="mt-1 h-4 w-4 rounded border-border accent-[var(--clay)]"
      />
      <span>
        {t("agreePrefix")}
        <Link className="mx-1 text-[var(--clay)] hover:underline" to="/legal/tos">{t("terms")}</Link>
        {t("and")}
        <Link className="mx-1 text-[var(--clay)] hover:underline" to="/legal/privacy">{t("privacy")}</Link>
      </span>
    </label>
  );
}
