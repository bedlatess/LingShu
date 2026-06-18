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
        <Card className="glass">
          <CardHeader><CardTitle>每日消费</CardTitle></CardHeader>
          <CardContent>{daily.length ? <Chart data={daily} /> : <EmptyState title="暂无每日统计" description="产生调用后会按天展示请求数和消费走势。" />}</CardContent>
        </Card>
      )}
      {active === "models" && (
        <Card className="glass">
          <CardHeader><CardTitle>按模型统计</CardTitle></CardHeader>
          <CardContent>{models.length ? <ModelBars data={models} /> : <EmptyState title="暂无模型统计" description="模型消费分布会在这里出现。" />}</CardContent>
        </Card>
      )}
      {active === "ledger" && <RecordList items={ledger.map((item) => ({ id: `${item.type}-${item.created_at}`, title: zhLedgerType(item.type), meta: item.remark, value: item.amount }))} empty="暂无记录" />}
      {active === "logs" && <RecordList items={logs.map((item) => ({ id: item.request_id, title: zhStatus(item.status), meta: `${item.total_tokens} 个 token · 状态码 ${item.http_status}`, value: formatMoney(item.charge) }))} empty="暂无请求日志" />}
    </div>
  );
}

function Chart({ data }: { data: UserDailyStat[] }) {
  return (
    <MeasuredChart>
      {({ width, height }) => (
        <AreaChart data={data} width={width} height={height}>
          <CartesianGrid stroke="rgba(255,255,255,.08)" vertical={false} />
          <XAxis dataKey="day" stroke="rgba(255,255,255,.45)" tickLine={false} axisLine={false} />
          <YAxis stroke="rgba(255,255,255,.45)" tickLine={false} axisLine={false} />
          <Tooltip contentStyle={{ background: "#0f172a", border: "1px solid rgba(255,255,255,.12)", borderRadius: 8 }} />
          <Area type="monotone" dataKey="charge" stroke="#2dd4bf" fill="#2dd4bf33" strokeWidth={2} />
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
          <CartesianGrid stroke="rgba(255,255,255,.08)" vertical={false} />
          <XAxis dataKey="model_id" stroke="rgba(255,255,255,.45)" tickLine={false} axisLine={false} />
          <YAxis stroke="rgba(255,255,255,.45)" tickLine={false} axisLine={false} />
          <Tooltip contentStyle={{ background: "#0f172a", border: "1px solid rgba(255,255,255,.12)", borderRadius: 8 }} />
          <Bar dataKey="charge" fill="#a78bfa" radius={[8, 8, 0, 0]} />
        </BarChart>
      )}
    </MeasuredChart>
  );
}

function RecordList({ items, empty }: { items: Array<{ id: string; title: string; meta: string; value: string }>; empty: string }) {
  return (
    <Card className="glass">
      <CardContent className="grid gap-3 p-5">
        {items.length ? items.map((item) => (
          <div key={item.id} className="flex items-center justify-between rounded-lg border border-white/10 bg-white/[0.035] p-3">
            <div><p className="text-sm font-medium">{item.title}</p><p className="text-xs text-muted-foreground">{item.meta}</p></div>
            <strong className="text-sm">{item.value}</strong>
          </div>
        )) : <EmptyState title={empty} description="数据会在对应业务发生后自动出现。" />}
      </CardContent>
    </Card>
  );
}
