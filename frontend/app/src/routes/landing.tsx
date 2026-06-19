import React from "react";
import { Link } from "react-router-dom";
import { motion, useReducedMotion, type MotionProps } from "framer-motion";
import { ArrowRight, Gauge, KeyRound, Layers3, RadioTower, ShieldCheck, WalletCards } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { PublicModel } from "@lingshu/shared";
import { createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle, EmptyState, Skeleton } from "@lingshu/ui";

import { SiteNav } from "@/components/site-nav";
import { PublicFooter } from "@/components/public-footer";
import { trBillingMode, trType } from "@/lib/i18n";
import { formatMoney } from "@/lib/utils";

const features = [
  { icon: Layers3, key: "multiModel" },
  { icon: WalletCards, key: "balance" },
  { icon: RadioTower, key: "failover" },
  { icon: ShieldCheck, key: "privateOps" }
] as const;

const quickStats = [
  { key: "openaiCompatible", value: "/v1" },
  { key: "realtimeBilling", value: "usage" },
  { key: "operatorManaged", value: "admin" }
] as const;

const easeOut = [0.22, 1, 0.36, 1] as const;

function reveal(reduceMotion: boolean | null, delay = 0): MotionProps {
  if (reduceMotion) {
    return { initial: false };
  }
  return {
    initial: { opacity: 0, y: 24 },
    whileInView: { opacity: 1, y: 0 },
    viewport: { once: true, margin: "-80px" },
    transition: { duration: 0.45, delay, ease: easeOut }
  };
}

function pressable(reduceMotion: boolean | null): MotionProps {
  if (reduceMotion) {
    return {};
  }
  return {
    whileHover: { y: -4 },
    whileTap: { scale: 0.985 },
    transition: { duration: 0.22, ease: easeOut }
  };
}

function stagger(reduceMotion: boolean | null): MotionProps {
  if (reduceMotion) {
    return { initial: false };
  }
  return {
    initial: "hidden",
    whileInView: "show",
    viewport: { once: true, margin: "-80px" },
    variants: {
      hidden: {},
      show: { transition: { staggerChildren: 0.06 } }
    }
  };
}

const childReveal = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0, transition: { duration: 0.42, ease: easeOut } }
};

export function LandingPage() {
  const { t } = useTranslation("pricing");
  const reduceMotion = useReducedMotion();
  const [models, setModels] = React.useState<PublicModel[]>([]);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    createAPI().publicModels().then((result) => setModels(result.items.slice(0, 6))).finally(() => setLoading(false));
  }, []);

  return (
    <main className="min-h-screen bg-background">
      <SiteNav />
      <section className="mx-auto grid min-h-[calc(100vh-4rem)] max-w-7xl gap-10 px-4 py-16 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:items-center lg:py-20">
        <motion.div className="space-y-8" {...reveal(reduceMotion)}>
          <Badge variant="info">{t("landing.badge")}</Badge>
          <div className="space-y-5">
            <h1 className="max-w-4xl font-serif text-5xl font-semibold leading-[1.04] tracking-[-0.02em] text-foreground sm:text-6xl lg:text-7xl">{t("landing.heroTitle")}</h1>
            <p className="max-w-2xl text-lg leading-8 text-muted-foreground">{t("landing.heroDescription")}</p>
          </div>
          <div className="flex flex-col gap-3 sm:flex-row">
            <motion.div {...pressable(reduceMotion)}><Button asChild size="lg"><Link to="/login">{t("landing.primaryCta")}<ArrowRight className="h-4 w-4" /></Link></Button></motion.div>
            <motion.div {...pressable(reduceMotion)}><Button asChild variant="secondary" size="lg"><Link to="/pricing">{t("landing.secondaryCta")}</Link></Button></motion.div>
          </div>
          <div className="grid max-w-2xl gap-3 sm:grid-cols-3">
            {quickStats.map((item) => (
              <div key={item.key} className="rounded-lg border border-border bg-card p-4">
                <strong className="font-serif text-2xl text-foreground">{item.value}</strong>
                <p className="mt-2 text-xs leading-5 text-muted-foreground">{t(`landing.stats.${item.key}`)}</p>
              </div>
            ))}
          </div>
        </motion.div>

        <motion.div className="rounded-xl border border-border bg-card p-4 shadow-[var(--shadow-xs)]" {...reveal(reduceMotion, 0.08)}>
          <div className="rounded-lg border border-border bg-[var(--bg-subtle)] p-5">
            <div className="mb-5 flex items-center justify-between">
              <div>
                <p className="text-xs uppercase tracking-[0.16em] text-[var(--clay)]">{t("landing.previewEyebrow")}</p>
                <h2 className="mt-2 font-serif text-2xl font-semibold">{t("landing.previewTitle")}</h2>
              </div>
              <Gauge className="h-5 w-5 text-[var(--clay)]" />
            </div>
            <div className="grid gap-3">
              {[
                [t("landing.flow.client"), t("landing.flow.clientHint")],
                [t("landing.flow.gateway"), t("landing.flow.gatewayHint")],
                [t("landing.flow.upstream"), t("landing.flow.upstreamHint")]
              ].map(([title, hint], index) => (
                <div key={title} className="flex items-start gap-3 rounded-md border border-border bg-card px-4 py-3">
                  <span className="mt-0.5 grid h-6 w-6 shrink-0 place-items-center rounded-md bg-[var(--clay)] text-xs font-semibold text-white">{index + 1}</span>
                  <span>
                    <strong className="block text-sm text-foreground">{title}</strong>
                    <span className="text-xs leading-5 text-muted-foreground">{hint}</span>
                  </span>
                </div>
              ))}
            </div>
          </div>
        </motion.div>
      </section>

      <motion.section className="mx-auto max-w-7xl px-4 py-12 sm:px-6" {...stagger(reduceMotion)}>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {features.map(({ icon: Icon, key }) => (
            <motion.div key={key} variants={childReveal} {...pressable(reduceMotion)}>
            <Card>
              <CardContent className="p-5">
                <div className="mb-5 grid h-10 w-10 place-items-center rounded-md border border-border bg-[var(--bg-subtle)] text-[var(--clay)]">
                  <Icon className="h-5 w-5" />
                </div>
                <h3 className="font-serif text-lg font-semibold">{t(`landing.features.${key}.title`)}</h3>
                <p className="mt-3 text-sm leading-6 text-muted-foreground">{t(`landing.features.${key}.description`)}</p>
              </CardContent>
            </Card>
            </motion.div>
          ))}
        </div>
      </motion.section>

      <motion.section className="mx-auto grid max-w-7xl gap-6 px-4 py-12 sm:px-6 lg:grid-cols-[0.8fr_1.2fr]" {...reveal(reduceMotion)}>
        <div className="space-y-4">
          <Badge>{t("landing.pricingEyebrow")}</Badge>
          <h2 className="font-serif text-4xl font-semibold tracking-[-0.02em]">{t("landing.pricingTitle")}</h2>
          <p className="text-base leading-7 text-muted-foreground">{t("landing.pricingDescription")}</p>
            <motion.div className="inline-flex" {...pressable(reduceMotion)}><Button asChild variant="secondary"><Link to="/pricing">{t("landing.viewAllPrices")}<ArrowRight className="h-4 w-4" /></Link></Button></motion.div>
        </div>
        {loading ? (
          <div className="grid gap-4 md:grid-cols-2">{Array.from({ length: 4 }).map((_, index) => <Skeleton key={index} className="h-40" />)}</div>
        ) : models.length ? (
          <div className="grid gap-4 md:grid-cols-2">
            {models.map((model) => <motion.div key={model.id} {...pressable(reduceMotion)}><ModelPreview model={model} /></motion.div>)}
          </div>
        ) : (
          <EmptyState title={t("emptyTitle")} description={t("emptyDescription")} icon={<KeyRound className="h-5 w-5" />} />
        )}
      </motion.section>

      <motion.section className="mx-auto max-w-7xl px-4 py-12 pb-20 sm:px-6" {...reveal(reduceMotion)}>
        <div className="rounded-xl border border-border bg-card p-8 md:flex md:items-center md:justify-between md:gap-8">
          <div className="max-w-2xl">
            <h2 className="font-serif text-3xl font-semibold">{t("landing.ctaTitle")}</h2>
            <p className="mt-3 text-sm leading-6 text-muted-foreground">{t("landing.ctaDescription")}</p>
          </div>
          <motion.div className="mt-6 md:mt-0" {...pressable(reduceMotion)}><Button asChild><Link to="/login">{t("landing.primaryCta")}<ArrowRight className="h-4 w-4" /></Link></Button></motion.div>
        </div>
      </motion.section>
      <PublicFooter />
    </main>
  );
}

function ModelPreview({ model }: { model: PublicModel }) {
  const { t } = useTranslation("pricing");
  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <CardTitle>{model.public_name}</CardTitle>
          <Badge>{trType(model.type)}</Badge>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 text-sm">
        <Badge variant="info">{trBillingMode(model.billing_mode)}</Badge>
        {model.billing_mode === "per_call" ? (
          <PriceLine label={t("perCall")} value={`${formatMoney(model.price_per_call)} ${model.currency}`} />
        ) : (
          <>
            <PriceLine label={t("input")} value={`${formatMoney(model.input_price_per_1m)} / 1M tokens`} />
            <PriceLine label={t("output")} value={`${formatMoney(model.output_price_per_1m)} / 1M tokens`} />
          </>
        )}
      </CardContent>
    </Card>
  );
}

function PriceLine({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border bg-[var(--bg-subtle)] px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <strong className="font-mono text-xs">{value}</strong>
    </div>
  );
}
