import React from "react";
import { Activity, Boxes, Clock3, CreditCard, WalletCards } from "lucide-react";
import { Area, AreaChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { StatCard } from "@/components/stat-card";
import { LoadingGrid } from "@/components/loading-grid";
import { EmptyState } from "@/components/empty-state";
import { MeasuredChart } from "@/components/measured-chart";
import { useAuth } from "@/providers/auth";
import { formatCompact, formatMoney } from "@/lib/utils";
import type { UserDashboard, UserDailyStat, UserGatewayLog } from "@lingshu/shared/user-types";
import { zhStatus } from "@/lib/i18n";

const CHART_COLORS = {
  clay: "#C6613F",
  clayFill: "#C6613F2A",
  grid: "#D8D4CA",
  axis: "#87867F",
  tooltipBg: "#FAF9F5",
  tooltipBorder: "#D8D4CA"
};

export function DashboardPage() {
  const { api } = useAuth();
  const [dashboard, setDashboard] = React.useState<UserDashboard | null>(null);
  const [daily, setDaily] = React.useState<UserDailyStat[]>([]);
  const [logs, setLogs] = React.useState<UserGatewayLog[]>([]);
  const [loading, setLoading] = React.useState(true);

  const refresh = React.useCallback(() => {
    return Promise.all([api.userDashboard(), api.userDailyStats(), api.userLogs()])
      .then(([dash, dailyResult, logResult]) => {
        setDashboard(dash);
        setDaily(dailyResult.items);
        setLogs(logResult.items);
      });
  }, [api]);

  React.useEffect(() => {
    refresh().finally(() => setLoading(false));
  }, [refresh]);

  React.useEffect(() => {
    const onFocus = () => refresh().catch(() => undefined);
    const onBalanceChanged = () => refresh().catch(() => undefined);
    window.addEventListener("focus", onFocus);
    window.addEventListener("lingshu:balance-changed", onBalanceChanged);
    return () => {
      window.removeEventListener("focus", onFocus);
      window.removeEventListener("lingshu:balance-changed", onBalanceChanged);
    };
  }, [refresh]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="概览" title="你的 AI API 运行仪表盘" description="余额、预扣、模型和消费趋势集中在一个清晰的工作台里。" />
      {loading || !dashboard ? (
        <LoadingGrid />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          <StatCard icon={WalletCards} label="余额" value={formatMoney(dashboard.balance)} hint="账户当前可用余额" />
          <StatCard icon={CreditCard} label="今日消费" value={formatMoney(dashboard.today_charge)} hint={`今日 ${dashboard.today_requests} 次请求`} />
          <StatCard icon={Activity} label="本月消费" value={formatMoney(dashboard.month_charge)} hint="本月累计请求消费" />
          <StatCard icon={Clock3} label="请求中预扣" value={formatMoney(dashboard.frozen)} hint="正在处理的请求金额" />
          <StatCard icon={Boxes} label="可用模型" value={String(dashboard.available_models)} hint="已启用模型数" />
        </div>
      )}

      <div className="grid gap-5 xl:grid-cols-[1.4fr_0.8fr]">
        <Card>
          <CardHeader>
            <CardTitle>消费趋势</CardTitle>
          </CardHeader>
          <CardContent>
            {daily.length === 0 ? (
              <EmptyState title="还没有趋势数据" description="完成第一次调用后，这里会显示每日请求和扣费走势。" />
            ) : (
              <MeasuredChart>
                {({ width, height }) => (
                  <AreaChart data={daily} width={width} height={height}>
                    <defs>
                      <linearGradient id="charge" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor={CHART_COLORS.clay} stopOpacity={0.24} />
                        <stop offset="95%" stopColor={CHART_COLORS.clay} stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid stroke={CHART_COLORS.grid} vertical={false} />
                    <XAxis dataKey="day" stroke={CHART_COLORS.axis} tickLine={false} axisLine={false} />
                    <YAxis stroke={CHART_COLORS.axis} tickLine={false} axisLine={false} />
                    <Tooltip contentStyle={{ background: CHART_COLORS.tooltipBg, border: `1px solid ${CHART_COLORS.tooltipBorder}`, borderRadius: 6, color: "#141413" }} />
                    <Area type="monotone" dataKey="charge" stroke={CHART_COLORS.clay} fill="url(#charge)" strokeWidth={2} />
                  </AreaChart>
                )}
              </MeasuredChart>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>最近调用</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3">
            {logs.length === 0 ? (
              <EmptyState title="暂无调用" description="创建平台密钥并发起调用后会出现记录。" />
            ) : (
              logs.slice(0, 6).map((log) => (
                <div key={log.request_id} className="flex items-center justify-between rounded-lg border border-border bg-[var(--bg-subtle)] p-3">
                  <div>
                    <p className="text-sm font-medium">{zhStatus(log.status)}</p>
                    <p className="text-xs text-muted-foreground">{formatCompact(log.total_tokens)} 个 token</p>
                  </div>
                  <strong className="font-serif text-sm">{formatMoney(log.charge)}</strong>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
