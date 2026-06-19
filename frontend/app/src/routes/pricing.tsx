import React from "react";
import { Link } from "react-router-dom";
import { ArrowRight, LayoutGrid, List } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { PublicModel } from "@lingshu/shared";

import { Badge, Button, Card, CardContent, CardHeader, CardTitle, EmptyState, Input, PageHeader, Sheet, Skeleton, Tabs } from "@lingshu/ui";
import { createAPI } from "@lingshu/shared";
import { SiteNav } from "@/components/site-nav";
import { PublicFooter } from "@/components/public-footer";
import { formatMoney } from "@/lib/utils";
import { trBillingMode, trType } from "@/lib/i18n";

export function PricingPage() {
  const { t } = useTranslation("pricing");
  const [models, setModels] = React.useState<PublicModel[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [query, setQuery] = React.useState("");
  const [billingMode, setBillingMode] = React.useState("");
  const [group, setGroup] = React.useState("");
  const [view, setView] = React.useState("grid");
  const [selectedModel, setSelectedModel] = React.useState<PublicModel | null>(null);

  React.useEffect(() => {
    createAPI().publicModels().then((result) => setModels(result.items)).finally(() => setLoading(false));
  }, []);
  const groups = React.useMemo(() => Array.from(new Set(models.map((model) => model.group || t("filters.defaultGroup")))), [models, t]);
  const visibleModels = React.useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return models.filter((model) => {
      const modelGroup = model.group || t("filters.defaultGroup");
      const keywordOK = !keyword || [model.public_name, model.type, modelGroup, model.billing_mode].some((value) => String(value ?? "").toLowerCase().includes(keyword));
      const billingOK = !billingMode || model.billing_mode === billingMode;
      const groupOK = !group || modelGroup === group;
      return keywordOK && billingOK && groupOK;
    });
  }, [models, query, billingMode, group, t]);

  return (
    <main className="min-h-screen bg-background">
      <SiteNav />
      <section className="mx-auto max-w-7xl px-4 py-12 sm:px-6">
        <PageHeader
          eyebrow={t("eyebrow")}
          title={t("title")}
          description={t("description")}
          action={<Button asChild><Link to="/login">{t("enterConsole")}<ArrowRight className="h-4 w-4" /></Link></Button>}
        />
        <Card className="mb-6">
          <CardContent className="grid gap-3 p-4 lg:grid-cols-[1fr_180px_180px_auto]">
            <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("filters.search")} />
            <select className="h-10 rounded-md border border-border bg-card px-3 text-sm text-foreground" value={billingMode} onChange={(event) => setBillingMode(event.target.value)}>
              <option value="">{t("filters.allBilling")}</option>
              <option value="token">{t("filters.token")}</option>
              <option value="per_call">{t("filters.perCall")}</option>
            </select>
            <select className="h-10 rounded-md border border-border bg-card px-3 text-sm text-foreground" value={group} onChange={(event) => setGroup(event.target.value)}>
              <option value="">{t("filters.allGroups")}</option>
              {groups.map((item) => <option key={item} value={item}>{item}</option>)}
            </select>
            <Tabs tabs={[{ value: "grid", label: <LayoutGrid className="h-4 w-4" /> }, { value: "list", label: <List className="h-4 w-4" /> }]} value={view} onChange={setView} />
          </CardContent>
        </Card>
        {loading ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{Array.from({ length: 6 }).map((_, index) => <Skeleton key={index} className="h-44" />)}</div>
        ) : visibleModels.length ? (
          <div className={view === "grid" ? "grid gap-4 md:grid-cols-2 xl:grid-cols-3" : "grid gap-3"}>
            {visibleModels.map((model) => <ModelPriceCard key={model.id} model={model} compact={view === "list"} onSelect={() => setSelectedModel(model)} />)}
          </div>
        ) : (
          <EmptyState title={t("emptyTitle")} description={t("emptyDescription")} />
        )}
      </section>
      <Sheet open={Boolean(selectedModel)} title={selectedModel?.public_name} onClose={() => setSelectedModel(null)}>
        {selectedModel ? <ModelDetail model={selectedModel} /> : null}
      </Sheet>
      <PublicFooter />
    </main>
  );
}

function ModelPriceCard({ model, compact, onSelect }: { model: PublicModel; compact?: boolean; onSelect: () => void }) {
  const { t } = useTranslation("pricing");
  return (
    <Card asChild className="transition-colors hover:border-[var(--border-strong)]">
      <button type="button" onClick={onSelect} className="block w-full cursor-pointer text-left">
      <CardHeader className={compact ? "p-4" : undefined}>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardTitle>{model.public_name}</CardTitle>
            <div className="mt-3 flex flex-wrap gap-2">
              <Badge>{trType(model.type)}</Badge>
              <Badge variant="info">{trBillingMode(model.billing_mode)}</Badge>
              {model.group ? <Badge variant="muted">{model.group}</Badge> : null}
            </div>
          </div>
          <div className={compact ? "grid gap-2 sm:min-w-96 sm:grid-cols-2" : "grid gap-3 text-sm"}>
            {model.billing_mode === "per_call" ? (
              <Price label={t("perCall")} value={`${formatMoney(model.price_per_call)} ${model.currency}`} />
            ) : (
              <>
                <Price label={t("input")} value={`${formatMoney(model.input_price_per_1m)} / 1M tokens`} />
                <Price label={t("output")} value={`${formatMoney(model.output_price_per_1m)} / 1M tokens`} />
              </>
            )}
          </div>
        </div>
      </CardHeader>
      </button>
    </Card>
  );
}

function ModelDetail({ model }: { model: PublicModel }) {
  const { t } = useTranslation("pricing");
  return (
    <div className="grid gap-5">
      <div className="flex flex-wrap gap-2">
        <Badge>{trType(model.type)}</Badge>
        <Badge variant="info">{trBillingMode(model.billing_mode)}</Badge>
        {model.group ? <Badge variant="muted">{model.group}</Badge> : null}
      </div>

      <section className="grid gap-3">
        <h3 className="font-serif text-base font-semibold text-foreground">{t("detail.priceTitle")}</h3>
        {model.billing_mode === "per_call" ? (
          <Price label={t("perCall")} value={`${formatMoney(model.price_per_call)} ${model.currency}`} />
        ) : (
          <>
            <Price label={t("input")} value={`${formatMoney(model.input_price_per_1m)} / 1M tokens`} />
            <Price label={t("output")} value={`${formatMoney(model.output_price_per_1m)} / 1M tokens`} />
          </>
        )}
      </section>

      <section className="grid gap-3">
        <h3 className="font-serif text-base font-semibold text-foreground">{t("detail.capabilityTitle")}</h3>
        <DetailItem label={t("detail.modelId")} value={model.id} mono />
        <DetailItem label={t("detail.context")} value={t("detail.contextUnavailable")} />
        <DetailItem label={t("detail.endpoints")} value={supportedEndpoints(model).join(", ")} mono />
      </section>
    </div>
  );
}

function DetailItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-md border border-border bg-[var(--bg-subtle)] p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={mono ? "mt-1 break-all font-mono text-xs text-foreground" : "mt-1 text-sm text-foreground"}>{value}</p>
    </div>
  );
}

function supportedEndpoints(model: PublicModel) {
  if (model.type === "embedding") return ["/v1/embeddings"];
  if (model.type === "image") return ["/v1/images/generations"];
  return ["/v1/chat/completions", "/v1/messages"];
}

function Price({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border bg-[var(--bg-subtle)] px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <strong className="font-mono text-xs">{value}</strong>
    </div>
  );
}
