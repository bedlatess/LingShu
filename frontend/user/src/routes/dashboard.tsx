import React from "react";
import { Activity, Boxes, Clock3, CreditCard, WalletCards } from "lucide-react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { StatCard } from "@/components/stat-card";
import { LoadingGrid } from "@/components/loading-grid";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { formatCompact, formatMoney } from "@/lib/utils";
import type { UserDashboard, DailyStat, GatewayLog } from "@lingshu/shared";

export function DashboardPage() {
  const { api } = useAuth();
  const [dashboard, setDashboard] = React.useState<UserDashboard | null>(null);
  const [daily, setDaily] = React.useState<DailyStat[]>([]);
  const [logs, setLogs] = React.useState<GatewayLog[]>([]);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    Promise.all([api.userDashboard(), api.userDailyStats(), api.userLogs()])
      .then(([dash, dailyResult, logResult]) => {
        setDashboard(dash);
        setDaily(dailyResult.items);
        setLogs(logResult.items);
      })
      .finally(() => setLoading(false));
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Overview" title="你的 AI API 运行仪表盘" description="余额、预占、模型和扣费趋势集中在一个安静的工作台里。" />
      {loading || !dashboard ? (
        <LoadingGrid />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          <StatCard icon={WalletCards} label="余额" value={formatMoney(dashboard.balance)} hint="Postgres 持久余额" />
          <StatCard icon={CreditCard} label="今日消费" value={formatMoney(dashboard.today_charge)} hint={`${dashboard.today_requests} requests today`} tone="violet" />
          <StatCard icon={Activity} label="本月消费" value={formatMoney(dashboard.month_charge)} hint="按本月 gateway_requests 聚合" tone="blue" />
          <StatCard icon={Clock3} label="请求中预占" value={formatMoney(dashboard.frozen)} hint="Redis frozen reservation" tone="amber" />
          <StatCard icon={Boxes} label="可用模型" value={String(dashboard.available_models)} hint="已启用模型数" />
        </div>
      )}

      <div className="grid gap-5 xl:grid-cols-[1.4fr_0.8fr]">
        <Card className="glass">
          <CardHeader>
            <CardTitle>消费趋势</CardTitle>
          </CardHeader>
          <CardContent>
            {daily.length === 0 ? (
              <EmptyState title="还没有趋势数据" description="完成第一次调用后，这里会显示每日请求和扣费走势。" />
            ) : (
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={daily}>
                    <defs>
                      <linearGradient id="charge" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#2dd4bf" stopOpacity={0.38} />
                        <stop offset="95%" stopColor="#2dd4bf" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid stroke="rgba(255,255,255,.08)" vertical={false} />
                    <XAxis dataKey="day" stroke="rgba(255,255,255,.45)" tickLine={false} axisLine={false} />
                    <YAxis stroke="rgba(255,255,255,.45)" tickLine={false} axisLine={false} />
                    <Tooltip contentStyle={{ background: "#0f172a", border: "1px solid rgba(255,255,255,.12)", borderRadius: 8 }} />
                    <Area type="monotone" dataKey="charge" stroke="#2dd4bf" fill="url(#charge)" strokeWidth={2} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="glass">
          <CardHeader>
            <CardTitle>最近调用</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3">
            {logs.length === 0 ? (
              <EmptyState title="暂无调用" description="创建 API Key 并调用 /v1/chat/completions 后会出现记录。" />
            ) : (
              logs.slice(0, 6).map((log) => (
                <div key={log.request_id} className="flex items-center justify-between rounded-lg border border-white/10 bg-white/[0.035] p-3">
                  <div>
                    <p className="text-sm font-medium">{log.status}</p>
                    <p className="text-xs text-muted-foreground">{formatCompact(log.total_tokens)} tokens</p>
                  </div>
                  <strong className="text-sm">{formatMoney(log.charge)}</strong>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
