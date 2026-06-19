import React from "react";
import { Activity, Boxes, ChevronDown, Copy, CreditCard, KeyRound, Terminal, WalletCards } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Area, AreaChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts";
import type { UserDailyStat } from "@lingshu/shared";

import { Button, Card, CardContent, CardHeader, CardTitle, EmptyState, PageHeader, Progress, StatCard, Tabs, cn, toast } from "@lingshu/ui";
import { MeasuredChart } from "@/components/measured-chart";
import { useAuth } from "@/providers/auth";
import { useSiteInfo } from "@/providers/site-info";
import { formatMoney } from "@/lib/utils";
import { copyText } from "@/lib/clipboard";

export function DashboardPage() {
  const { t } = useTranslation("dashboard");
  const { api } = useAuth();
  const { siteInfo } = useSiteInfo();
  const navigate = useNavigate();
  const [dashboard, setDashboard] = React.useState<Awaited<ReturnType<typeof api.userDashboard>> | null>(null);
  const [daily, setDaily] = React.useState<UserDailyStat[]>([]);
  const [configTab, setConfigTab] = React.useState("claude");
  const [devOpen, setDevOpen] = React.useState(false);

  React.useEffect(() => {
    Promise.all([api.userDashboard(), api.userDailyStats()])
      .then(([dash, stats]) => {
        setDashboard(dash);
        setDaily(stats.items);
      })
      .catch(() => undefined);
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} />
      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label={t("stats.balance")} value={formatMoney(dashboard?.balance)} hint={t("stats.balanceHint")} icon={WalletCards} />
        <StatCard label={t("stats.todayCharge")} value={formatMoney(dashboard?.today_charge)} hint={t("stats.todayChargeHint", { count: dashboard?.today_requests ?? 0 })} icon={Activity} />
        <StatCard label={t("stats.monthCharge")} value={formatMoney(dashboard?.month_charge)} hint={t("stats.monthChargeHint")} icon={CreditCard} />
        <StatCard label={t("stats.models")} value={dashboard?.available_models ?? 0} hint={t("stats.modelsHint")} icon={Boxes} />
      </section>
      <Card>
        <CardHeader>
          <CardTitle>{t("trendTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          {daily.length ? <Trend data={daily} /> : <EmptyState title={t("trendEmptyTitle")} description={t("trendEmptyDescription")} />}
        </CardContent>
      </Card>
      <section className="grid gap-4 xl:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>{t("quota.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <Progress value={quotaPercent(dashboard?.total_charge, dashboard?.total_recharge)} />
            <div className="grid gap-3 text-sm sm:grid-cols-3">
              <Metric label={t("quota.granted")} value={formatMoney(dashboard?.total_recharge)} />
              <Metric label={t("quota.used")} value={formatMoney(dashboard?.total_charge)} />
              <Metric label={t("quota.remaining")} value={formatMoney(dashboard?.balance)} />
            </div>
            <p className="text-xs leading-5 text-muted-foreground">{t("quota.description")}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <div className="flex items-start justify-between gap-3">
              <div>
                <CardTitle>{t("developer.title")}</CardTitle>
                <p className="mt-1 text-xs leading-5 text-muted-foreground">{t("developer.description")}</p>
              </div>
              <Button variant="secondary" size="sm" onClick={() => navigate("/api-keys")}>
                <KeyRound className="h-4 w-4" />{t("developer.manageKeys")}
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            <button
              type="button"
              onClick={() => setDevOpen((prev) => !prev)}
              className="flex w-full items-center justify-between gap-2 rounded-md border border-border bg-[var(--bg-subtle)] px-3 py-2 text-sm text-foreground transition-colors hover:border-[var(--border-strong)]"
              aria-expanded={devOpen}
            >
              <span className="inline-flex items-center gap-2"><Terminal className="h-4 w-4 text-[var(--clay)]" />{t("developer.toggle")}</span>
              <ChevronDown className={cn("h-4 w-4 transition-transform duration-200", devOpen && "rotate-180")} />
            </button>
            <div className={cn("grid transition-[grid-template-rows] duration-200 ease-out", devOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]")}>
              <div className="overflow-hidden">
                <div className="space-y-4 pt-1">
                  <Tabs tabs={[{ value: "claude", label: t("quickConfig.claude") }, { value: "codex", label: t("quickConfig.codex") }]} value={configTab} onChange={setConfigTab} />
                  <ConfigSnippet value={configTab === "claude" ? claudeSnippet(apiBaseURL(siteInfo?.api_base_url)) : codexSnippet(apiBaseURL(siteInfo?.api_base_url))} copiedText={t("quickConfig.copied")} copyLabel={t("quickConfig.copy")} />
                  <p className="text-xs leading-5 text-muted-foreground">{t("quickConfig.description")}</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-border bg-[var(--bg-subtle)] p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <strong className="mt-1 block font-serif text-lg text-foreground">{value}</strong>
    </div>
  );
}

function ConfigSnippet({ value, copiedText, copyLabel }: { value: string; copiedText: string; copyLabel: string }) {
  return (
    <div className="rounded-md border border-border bg-[var(--bg-subtle)]">
      <div className="flex items-center justify-between gap-3 border-b border-border px-3 py-2">
        <span className="inline-flex items-center gap-2 text-xs text-muted-foreground"><Terminal className="h-3.5 w-3.5" /> base_url + API Key</span>
        <Button
          variant="secondary"
          size="sm"
          onClick={async () => {
            if (await copyText(value)) toast.success(copiedText);
          }}
        >
          <Copy className="h-4 w-4" />{copyLabel}
        </Button>
      </div>
      <pre className="overflow-x-auto p-4 text-xs leading-6"><code>{value}</code></pre>
    </div>
  );
}

function quotaPercent(used?: string, granted?: string) {
  const usedValue = Number(used ?? 0);
  const grantedValue = Number(granted ?? 0);
  if (!Number.isFinite(usedValue) || !Number.isFinite(grantedValue) || grantedValue <= 0) return 0;
  return Math.max(0, Math.min(100, (usedValue / grantedValue) * 100));
}

function apiBaseURL(configured?: string) {
  const value = configured?.trim();
  if (value) return value.replace(/\/$/, "");
  return `${window.location.origin}/v1`;
}

function claudeSnippet(baseURL: string) {
  return [
    `export ANTHROPIC_BASE_URL="${baseURL}"`,
    `export ANTHROPIC_API_KEY="ls-your-api-key"`,
    "",
    "claude"
  ].join("\n");
}

function codexSnippet(baseURL: string) {
  return [
    `export OPENAI_BASE_URL="${baseURL}/v1"`,
    `export OPENAI_API_KEY="ls-your-api-key"`,
    "",
    "codex"
  ].join("\n");
}

function Trend({ data }: { data: UserDailyStat[] }) {
  return (
    <MeasuredChart>
      {({ width, height }) => (
        <AreaChart width={width} height={height} data={data}>
          <CartesianGrid stroke="#D8D4CA" vertical={false} />
          <XAxis dataKey="day" stroke="#87867F" tickLine={false} axisLine={false} />
          <YAxis stroke="#87867F" tickLine={false} axisLine={false} />
          <Tooltip contentStyle={{ background: "#FAF9F5", border: "1px solid #D8D4CA", borderRadius: 6 }} />
          <Area type="monotone" dataKey="charge" stroke="#C6613F" fill="#C6613F2A" strokeWidth={2} />
        </AreaChart>
      )}
    </MeasuredChart>
  );
}
