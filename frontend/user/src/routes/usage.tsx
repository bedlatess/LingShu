import React from "react";
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts";
import type { UserDailyStat, UserGatewayLog, UserLedgerRecord, UserModelStat } from "@lingshu/shared/user-types";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { MeasuredChart } from "@/components/measured-chart";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { zhLedgerType, zhStatus } from "@/lib/i18n";

const tabs = [
  ["daily", "每日统计"],
  ["models", "按模型"],
  ["ledger", "扣费记录"],
  ["logs", "请求日志"]
] as const;

const CHART_COLORS = {
  clay: "#C6613F",
  clayFill: "#C6613F2A",
  grid: "#D8D4CA",
  axis: "#87867F",
  tooltipBg: "#FAF9F5",
  tooltipBorder: "#D8D4CA"
};

export function UsagePage() {
  const { api } = useAuth();
  const [active, setActive] = React.useState<(typeof tabs)[number][0]>("daily");
  const [daily, setDaily] = React.useState<UserDailyStat[]>([]);
  const [models, setModels] = React.useState<UserModelStat[]>([]);
  const [ledger, setLedger] = React.useState<UserLedgerRecord[]>([]);
  const [logs, setLogs] = React.useState<UserGatewayLog[]>([]);

  React.useEffect(() => {
    Promise.all([api.userDailyStats(), api.userModelStats(), api.userLedger(), api.userLogs()]).then(([dailyResult, modelResult, ledgerResult, logResult]) => {
      setDaily(dailyResult.items);
      setModels(modelResult.items);
      setLedger(ledgerResult.items);
      setLogs(logResult.items);
    });
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="用量统计" title="用量和消费明细" description="按每日、模型、记录、请求四个视角查看账户使用情况。" />
      <TabsList>
        {tabs.map(([value, label]) => <TabsTrigger key={value} value={value} activeValue={active} onSelect={(v) => setActive(v as typeof active)}>{label}</TabsTrigger>)}
      </TabsList>

      {active === "daily" && (
        <Card>
          <CardHeader><CardTitle>每日消费</CardTitle></CardHeader>
          <CardContent>{daily.length ? <Chart data={daily} /> : <EmptyState title="暂无每日统计" description="产生调用后会按天展示请求数和消费走势。" />}</CardContent>
        </Card>
      )}
      {active === "models" && (
        <Card>
          <CardHeader><CardTitle>按模型统计</CardTitle></CardHeader>
          <CardContent>{models.length ? <ModelBars data={models} /> : <EmptyState title="暂无模型统计" description="模型消费分布会在这里出现。" />}</CardContent>
        </Card>
      )}
      {active === "ledger" && (
        <RecordList
          items={ledger.map((item) => ({
            id: `${item.type}-${item.created_at}`,
            title: zhLedgerType(item.type),
            meta: [formatDateTime(item.created_at), item.remark, `余额 ${formatMoney(item.balance_before)} → ${formatMoney(item.balance_after)}`].filter(Boolean).join(" · "),
            value: formatMoney(item.amount)
          }))}
          empty="暂无记录"
        />
      )}
      {active === "logs" && (
        <RecordList
          items={logs.map((item) => ({
            id: item.request_id,
            title: `${zhStatus(item.status)} · ${item.model_id || "未知模型"}`,
            meta: [formatDateTime(item.created_at), `请求 ${shortID(item.request_id)}`, `${item.total_tokens} 个 token`, `状态码 ${item.http_status}`].join(" · "),
            value: formatMoney(item.charge)
          }))}
          empty="暂无请求日志"
        />
      )}
    </div>
  );
}

function Chart({ data }: { data: UserDailyStat[] }) {
  return (
    <MeasuredChart>
      {({ width, height }) => (
        <AreaChart data={data} width={width} height={height}>
          <CartesianGrid stroke={CHART_COLORS.grid} vertical={false} />
          <XAxis dataKey="day" stroke={CHART_COLORS.axis} tickLine={false} axisLine={false} />
          <YAxis stroke={CHART_COLORS.axis} tickLine={false} axisLine={false} />
          <Tooltip contentStyle={{ background: CHART_COLORS.tooltipBg, border: `1px solid ${CHART_COLORS.tooltipBorder}`, borderRadius: 6, color: "#141413" }} />
          <Area type="monotone" dataKey="charge" stroke={CHART_COLORS.clay} fill={CHART_COLORS.clayFill} strokeWidth={2} />
        </AreaChart>
      )}
    </MeasuredChart>
  );
}

function ModelBars({ data }: { data: UserModelStat[] }) {
  return (
    <MeasuredChart>
      {({ width, height }) => (
        <BarChart data={data} width={width} height={height}>
          <CartesianGrid stroke={CHART_COLORS.grid} vertical={false} />
          <XAxis dataKey="model_id" stroke={CHART_COLORS.axis} tickLine={false} axisLine={false} />
          <YAxis stroke={CHART_COLORS.axis} tickLine={false} axisLine={false} />
          <Tooltip contentStyle={{ background: CHART_COLORS.tooltipBg, border: `1px solid ${CHART_COLORS.tooltipBorder}`, borderRadius: 6, color: "#141413" }} />
          <Bar dataKey="charge" fill={CHART_COLORS.clay} radius={[6, 6, 0, 0]} />
        </BarChart>
      )}
    </MeasuredChart>
  );
}

function RecordList({ items, empty }: { items: Array<{ id: string; title: string; meta: string; value: string }>; empty: string }) {
  return (
    <Card>
      <CardContent className="grid gap-3 p-5">
        {items.length ? items.map((item) => (
          <div key={item.id} className="flex items-center justify-between rounded-lg border border-border bg-[var(--bg-subtle)] p-3">
            <div><p className="text-sm font-medium">{item.title}</p><p className="text-xs text-muted-foreground">{item.meta}</p></div>
            <strong className="font-serif text-sm">{item.value}</strong>
          </div>
        )) : <EmptyState title={empty} description="数据会在对应业务发生后自动出现。" />}
      </CardContent>
    </Card>
  );
}

function formatDateTime(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString("zh-CN", { hour12: false });
}

function shortID(value: string) {
  return value.length > 16 ? `${value.slice(0, 8)}...${value.slice(-6)}` : value;
}
