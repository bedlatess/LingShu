import React from "react";
import { useTranslation } from "react-i18next";
import type { OpsDashboard, OpsStatusBucket, OpsTrendPoint, createAPI } from "@lingshu/shared";
import { Badge, Card, CardContent, CardDescription, CardHeader, CardTitle, DataTable, EmptyState, PageHeader, Progress, StatCard } from "@lingshu/ui";
import { Activity, BarChart3, Clock, Gauge, RadioTower, Repeat, TimerReset, Zap } from "lucide-react";
import { formatDateMinute, MiniBars, statusVariant } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

function formatCompact(value: number) {
  return new Intl.NumberFormat(undefined, { notation: "compact", maximumFractionDigits: 1 }).format(value);
}

function TrendLineChart({
  data,
  series,
  emptyTitle,
  emptyDescription,
  unit = "",
  height = 180
}: {
  data: OpsTrendPoint[];
  series: { key: keyof OpsTrendPoint; label: string; color: string }[];
  emptyTitle: string;
  emptyDescription: string;
  unit?: string;
  height?: number;
}) {
  const width = 720;
  const padding = { top: 18, right: 20, bottom: 28, left: 44 };
  const numericValues = data.flatMap((item) => series.map((entry) => Number(item[entry.key] ?? 0))).filter(Number.isFinite);
  const max = Math.max(1, ...numericValues);
  const plotWidth = width - padding.left - padding.right;
  const plotHeight = height - padding.top - padding.bottom;
  const xFor = (index: number) => padding.left + (data.length <= 1 ? plotWidth / 2 : (index / (data.length - 1)) * plotWidth);
  const yFor = (value: number) => padding.top + plotHeight - (value / max) * plotHeight;

  if (data.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} icon={<BarChart3 className="h-5 w-5" />} />;
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        {series.map((entry) => (
          <span key={String(entry.key)} className="inline-flex items-center gap-2 text-xs text-muted-foreground">
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: entry.color }} />
            {entry.label}
          </span>
        ))}
      </div>
      <div className="overflow-hidden rounded-md border border-border bg-[var(--bg-subtle)] p-3">
        <svg className="h-auto w-full" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={series.map((entry) => entry.label).join(" / ")}>
          {[0, 0.5, 1].map((ratio) => {
            const y = padding.top + plotHeight - ratio * plotHeight;
            return (
              <g key={ratio}>
                <line x1={padding.left} x2={width - padding.right} y1={y} y2={y} stroke="var(--border)" strokeDasharray="4 6" />
                <text x={padding.left - 10} y={y + 4} textAnchor="end" className="fill-muted-foreground text-[11px]">
                  {formatCompact(max * ratio)}{unit}
                </text>
              </g>
            );
          })}
          {series.map((entry) => {
            const points = data.map((item, index) => `${xFor(index)},${yFor(Number(item[entry.key] ?? 0))}`).join(" ");
            return <polyline key={String(entry.key)} points={points} fill="none" stroke={entry.color} strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" />;
          })}
          {data.map((item, index) => (
            <text key={item.bucket} x={xFor(index)} y={height - 7} textAnchor={index === 0 ? "start" : index === data.length - 1 ? "end" : "middle"} className="fill-muted-foreground text-[10px]">
              {item.bucket.slice(-5)}
            </text>
          ))}
        </svg>
      </div>
    </div>
  );
}

function StatusDistribution({ data, emptyTitle, emptyDescription }: { data: OpsStatusBucket[]; emptyTitle: string; emptyDescription: string }) {
  const total = data.reduce((sum, item) => sum + item.count, 0);
  if (total === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} icon={<Gauge className="h-5 w-5" />} />;
  }
  return (
    <div className="space-y-4">
      {data.map((item) => {
        const percent = (item.count / total) * 100;
        const variant = item.status.startsWith("2") ? "success" as const : item.status.startsWith("4") ? "warning" as const : item.status.startsWith("5") ? "danger" as const : "muted" as const;
        return (
          <div key={item.status} className="space-y-2">
            <div className="flex items-center justify-between gap-3 text-sm">
              <Badge variant={variant}>{item.status}</Badge>
              <span className="font-mono text-xs text-muted-foreground">{item.count} / {percent.toFixed(1)}%</span>
            </div>
            <Progress value={percent} />
          </div>
        );
      })}
    </div>
  );
}

export function OpsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [data, setData] = React.useState<OpsDashboard | null>(null);
  React.useEffect(() => { api.adminOps().then(setData); }, [api]);
  const summary = data?.summary;
  const trends = data?.trends ?? [];
  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("ops.eyebrow")} title={t("ops.title")} description={t("ops.description")} />
      {data?.alerts?.length ? (
        <Card>
          <CardHeader>
            <CardTitle>{t("ops.activeAlerts")}</CardTitle>
            <CardDescription>{t("ops.activeAlertsDescription")}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3">
            {data.alerts.map((alert) => (
              <div key={alert.id} className="rounded-md border border-[var(--danger)]/40 bg-[var(--danger)]/5 p-3">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant={alert.severity === "critical" ? "danger" : alert.severity === "warning" ? "warning" : "muted"}>{alert.severity}</Badge>
                  <span className="font-medium text-foreground">{alert.title}</span>
                  <span className="font-mono text-xs text-muted-foreground">{alert.rule_key}</span>
                </div>
                <p className="mt-2 text-sm text-muted-foreground">{alert.message}</p>
              </div>
            ))}
          </CardContent>
        </Card>
      ) : null}
      <section className="grid gap-4 md:grid-cols-3 xl:grid-cols-6">
        <StatCard label="RPM" value={summary?.rpm ?? 0} hint={t("ops.lastMinute")} icon={Zap} />
        <StatCard label="TPM" value={summary?.tpm ?? 0} hint={t("ops.lastMinute")} icon={Activity} />
        <StatCard label={t("ops.requests24h")} value={summary?.requests_24h ?? 0} hint={t("ops.errorRate", { rate: summary?.error_rate_24h ?? "0" })} icon={Repeat} />
        <StatCard label={t("ops.p50")} value={`${summary?.p50_latency_ms ?? 0}ms`} hint={t("ops.gatewayRecorded")} icon={Clock} />
        <StatCard label={t("ops.p95")} value={`${summary?.p95_latency_ms ?? 0}ms`} hint={t("ops.gatewayRecorded")} icon={TimerReset} />
        <StatCard label={t("ops.switches")} value={summary?.channel_switches ?? 0} hint={t("ops.failover")} icon={RadioTower} />
      </section>
      <section className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <CardHeader>
            <CardTitle>{t("ops.trafficTrend")}</CardTitle>
            <CardDescription>{t("ops.trafficTrendDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <TrendLineChart
              data={trends}
              emptyTitle={t("ops.emptyTrendTitle")}
              emptyDescription={t("ops.emptyTrendDescription")}
              series={[
                { key: "requests", label: t("ops.requests"), color: "var(--clay)" },
                { key: "failures", label: t("ops.failures"), color: "var(--danger)" }
              ]}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("ops.statusDistribution")}</CardTitle>
            <CardDescription>{t("ops.statusDistributionDescription")}</CardDescription>
          </CardHeader>
          <CardContent><StatusDistribution data={data?.statuses ?? []} emptyTitle={t("ops.emptyStatusTitle")} emptyDescription={t("ops.emptyStatusDescription")} /></CardContent>
        </Card>
      </section>
      <section className="grid gap-4 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t("ops.chargeTrend")}</CardTitle>
            <CardDescription>{t("ops.chargeTrendDescription")}</CardDescription>
          </CardHeader>
          <CardContent><MiniBars data={trends.map((item) => ({ label: item.bucket, value: Number(item.charge) }))} /></CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("ops.latencyTrend")}</CardTitle>
            <CardDescription>{t("ops.latencyTrendDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <TrendLineChart
              data={trends}
              emptyTitle={t("ops.emptyTrendTitle")}
              emptyDescription={t("ops.emptyTrendDescription")}
              unit="ms"
              series={[
                { key: "avg_latency_ms", label: t("ops.averageLatency"), color: "var(--info)" },
                { key: "p95_latency_ms", label: t("ops.p95"), color: "var(--warning)" }
              ]}
            />
          </CardContent>
        </Card>
      </section>
      <DataTable
        data={data?.channels ?? []}
        rowKey={(row) => row.id}
        columns={[
          { key: "name", title: t("common.channel") },
          { key: "provider_type", title: t("channels.table.provider") },
          { key: "health", title: t("channels.table.health"), render: (row) => <Badge variant={statusVariant(row.health)}>{row.health}</Badge> },
          { key: "error_rate_24h", title: t("ops.table.errorRate24h"), render: (row) => <div className="min-w-32"><Progress value={Number(row.error_rate_24h)} /></div> },
          { key: "avg_latency_ms", title: t("ops.table.averageLatency"), render: (row) => `${row.avg_latency_ms}ms` },
          { key: "last_success_at", title: t("ops.table.lastSuccess"), render: (row) => formatDateMinute(row.last_success_at) }
        ]}
      />
    </div>
  );
}
