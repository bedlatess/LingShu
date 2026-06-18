import React from "react";
import { Link } from "react-router-dom";
import { ArrowRight, BadgeDollarSign, Boxes, Cpu, Sparkles } from "lucide-react";
import { createAPI } from "@lingshu/shared";
import type { PublicModel, PublicSiteInfo } from "@lingshu/shared";

import { SiteNav } from "@/components/site-nav";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const api = createAPI();

export function PricingPage() {
  const [models, setModels] = React.useState<PublicModel[]>([]);
  const [siteInfo, setSiteInfo] = React.useState<PublicSiteInfo>({ site_name: "LingShu", registration_enabled: false, contact_info: "", login_url: "/login" });
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    document.title = `${siteInfo.site_name} - AI API 价格表`;
  }, [siteInfo.site_name]);

  React.useEffect(() => {
    let mounted = true;
    Promise.all([api.publicModels(), api.siteInfo()])
      .then(([modelResult, info]) => {
        if (!mounted) return;
        setModels(modelResult.items);
        setSiteInfo(info);
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const grouped = React.useMemo(() => groupModels(models), [models]);

  return (
    <main className="min-h-screen bg-background">
      <SiteNav siteName={siteInfo.site_name} />
      <section className="mx-auto grid w-full max-w-7xl gap-10 px-4 py-14 sm:px-6 lg:px-8">
        <div className="grid gap-8 lg:grid-cols-[1fr_360px] lg:items-end">
          <div>
            <Badge variant="clay" className="mb-5">公开价格</Badge>
            <h1 className="max-w-3xl font-serif text-4xl font-semibold leading-tight text-foreground sm:text-6xl">{siteInfo.site_name} - AI API 价格表</h1>
            <p className="mt-5 max-w-2xl text-base leading-7 text-muted-foreground">
              统一接入 OpenAI / Anthropic 兼容模型，按实际用量透明计费。你只需要平台 API Key 和模型名，就能在现有 SDK 中切换到灵枢网关。
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Button asChild>
                <Link to="/login">
                  {siteInfo.registration_enabled ? "免费注册" : "立即接入"} <ArrowRight className="h-4 w-4" />
                </Link>
              </Button>
              <Button asChild variant="outline">
                <a href="/api.md">查看文档</a>
              </Button>
            </div>
          </div>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Sparkles className="h-5 w-5 text-primary" /> 接入概览
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 text-sm text-muted-foreground">
              <div className="flex items-center justify-between rounded-lg border border-border bg-[var(--bg-subtle)] p-3">
                <span>可用模型</span>
                <strong className="font-serif text-foreground">{models.length}</strong>
              </div>
              <div className="flex items-center justify-between rounded-lg border border-border bg-[var(--bg-subtle)] p-3">
                <span>计费货币</span>
                <strong className="font-serif text-foreground">USD</strong>
              </div>
              {siteInfo.contact_info ? <p className="leading-6">联系方式：{siteInfo.contact_info}</p> : null}
            </CardContent>
          </Card>
        </div>

        {loading ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {Array.from({ length: 6 }).map((_, index) => (
              <Skeleton key={index} className="h-48 rounded-lg" />
            ))}
          </div>
        ) : (
          <div className="grid gap-8">
            {grouped.map(([group, items]) => (
              <section key={group} className="grid gap-4">
                <div className="flex items-center gap-3">
                  <Boxes className="h-5 w-5 text-primary" />
                  <h2 className="font-serif text-xl font-semibold">{group}</h2>
                </div>
                <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                  {items.map((model) => (
                    <ModelPriceCard key={model.id} model={model} />
                  ))}
                </div>
              </section>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

function ModelPriceCard({ model }: { model: PublicModel }) {
  return (
    <Card className="group transition-colors hover:border-[var(--border-strong)]">
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div>
            <CardTitle className="leading-6">{model.public_name}</CardTitle>
            <p className="mt-2 flex items-center gap-2 text-sm text-muted-foreground">
              <Cpu className="h-4 w-4" /> {model.type}
            </p>
          </div>
          <Badge>{billingModeText(model.billing_mode)}</Badge>
        </div>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div className="grid grid-cols-2 gap-3">
          <PriceCell label="输入 / 1M" value={model.input_price_per_1m} currency={model.currency} />
          <PriceCell label="输出 / 1M" value={model.output_price_per_1m} currency={model.currency} />
        </div>
        {model.billing_mode === "per_call" && model.price_per_call ? <PriceCell label="单次调用" value={model.price_per_call} currency={model.currency} /> : null}
        <Button asChild variant="secondary" className="justify-between">
          <Link to="/login">
            立即接入 <ArrowRight className="h-4 w-4 transition group-hover:translate-x-0.5" />
          </Link>
        </Button>
      </CardContent>
    </Card>
  );
}

function PriceCell({ label, value, currency }: { label: string; value: string; currency: string }) {
  return (
    <div className="rounded-lg border border-border bg-[var(--bg-subtle)] p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-2 flex items-baseline gap-1 font-serif text-lg font-semibold text-foreground">
        <BadgeDollarSign className="h-4 w-4 text-primary" />
        {Number(value || 0).toFixed(6)}
        <span className="text-xs font-normal text-muted-foreground">{currency}</span>
      </p>
    </div>
  );
}

function groupModels(models: PublicModel[]) {
  const groups = new Map<string, PublicModel[]>();
  for (const model of models) {
    const group = model.group?.trim() || "通用";
    groups.set(group, [...(groups.get(group) ?? []), model]);
  }
  return Array.from(groups.entries());
}

function billingModeText(value: string) {
  return value === "per_call" ? "按次" : "按量";
}
