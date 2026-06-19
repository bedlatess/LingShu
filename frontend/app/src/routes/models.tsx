import React from "react";
import { Search } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { UserModelConfig } from "@lingshu/shared";

import { Badge, Card, CardContent, EmptyState, Input, PageHeader } from "@lingshu/ui";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { trBillingMode, trType } from "@/lib/i18n";

export function ModelsPage() {
  const { t } = useTranslation("models");
  const { api } = useAuth();
  const [items, setItems] = React.useState<UserModelConfig[]>([]);
  const [query, setQuery] = React.useState("");

  React.useEffect(() => {
    api.userModels().then((result) => setItems(result.items));
  }, [api]);

  const filtered = items.filter((item) => `${item.public_name} ${item.type} ${item.group}`.toLowerCase().includes(query.toLowerCase()));

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} />
      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input className="pl-9" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("searchPlaceholder")} />
      </div>
      {filtered.length ? (
        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((model) => (
            <Card key={model.id} className="transition-colors hover:border-[var(--border-strong)]">
              <CardContent className="grid gap-4 p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h2 className="font-serif text-xl font-semibold">{model.public_name}</h2>
                    <p className="mt-1 text-sm text-muted-foreground">{model.group || t("defaultGroup")}</p>
                  </div>
                  <Badge variant={model.status === "enabled" ? "success" : "muted"}>{model.status === "enabled" ? t("enabled") : t("disabled")}</Badge>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Badge>{trType(model.type)}</Badge>
                  <Badge variant="info">{trBillingMode(model.billing_mode)}</Badge>
                </div>
                <div className="grid gap-2 rounded-md border border-border bg-[var(--bg-subtle)] p-3 text-sm">
                  {model.billing_mode === "per_call" ? (
                    <Row label={t("perCall")} value={formatMoney(model.call_unit_price)} />
                  ) : (
                    <>
                      <Row label={t("input")} value={`${formatMoney(model.input_unit_price)} / 1M tokens`} />
                      <Row label={t("output")} value={`${formatMoney(model.output_unit_price)} / 1M tokens`} />
                    </>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </section>
      ) : (
        <EmptyState title={t("emptyTitle")} description={t("emptyDescription")} />
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-muted-foreground">{label}</span>
      <strong className="font-mono text-xs">{value}</strong>
    </div>
  );
}
