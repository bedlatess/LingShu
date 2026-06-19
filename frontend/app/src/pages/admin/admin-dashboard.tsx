import React from "react";
import { useTranslation } from "react-i18next";
import { Activity, CircleDollarSign, KeyRound, RadioTower, Users, Waypoints } from "lucide-react";
import type { AdminDashboard, ReportRow, User, createAPI } from "@lingshu/shared";
import { Card, CardContent, CardHeader, CardTitle, DataTable, PageHeader, StatCard } from "@lingshu/ui";
import { fmtMoney, MiniBars } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function AdminDashboardPage({ api, me }: { api: AdminAPI; me: User }) {
  const { t } = useTranslation("admin");
  const [dashboard, setDashboard] = React.useState<AdminDashboard | null>(null);
  const [daily, setDaily] = React.useState<ReportRow[]>([]);
  const [models, setModels] = React.useState<ReportRow[]>([]);

  React.useEffect(() => {
    Promise.all([api.adminDashboard(), api.adminReportDaily("", ""), api.adminReportByModel("", "")]).then(([dash, dailyRows, modelRows]) => {
      setDashboard(dash);
      setDaily(dailyRows.items.slice(0, 7).reverse());
      setModels(modelRows.items.slice(0, 5));
    }).catch(() => undefined);
  }, [api]);

  const successRate = dashboard?.today_requests ? `${(((dashboard.today_successes ?? 0) / dashboard.today_requests) * 100).toFixed(1)}%` : "0%";

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("dashboard.eyebrow")} title={t("dashboard.title")} description={t("dashboard.description", { name: me.username })} />
      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-6">
        <StatCard label={t("dashboard.todayRequests")} value={dashboard?.today_requests ?? 0} hint={t("dashboard.successRate", { rate: successRate })} icon={Activity} />
        <StatCard label={t("dashboard.todayCharge")} value={fmtMoney(dashboard?.today_charge)} hint={t("dashboard.customerRevenue")} icon={CircleDollarSign} />
        <StatCard label={t("dashboard.todayCost")} value={fmtMoney(dashboard?.today_base_cost)} hint={t("dashboard.upstreamBaseCost")} icon={CircleDollarSign} />
        <StatCard label={t("dashboard.grossProfit")} value={fmtMoney(dashboard?.gross_profit)} hint={t("dashboard.profitFormula")} icon={CircleDollarSign} />
        <StatCard label={t("dashboard.users")} value={`${dashboard?.active_users ?? 0}/${dashboard?.total_users ?? 0}`} hint={t("dashboard.activeTotal")} icon={Users} />
        <StatCard label={t("dashboard.resources")} value={`${dashboard?.enabled_models ?? 0}/${dashboard?.healthy_channels ?? 0}`} hint={t("dashboard.modelsHealthyChannels")} icon={Waypoints} />
      </section>
      <section className="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        <Card>
          <CardHeader><CardTitle>{t("dashboard.last7DaysCharge")}</CardTitle></CardHeader>
          <CardContent><MiniBars data={daily.map((item) => ({ label: item.label, value: Number(item.charge) }))} /></CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>{t("dashboard.topModels")}</CardTitle></CardHeader>
          <CardContent>
            <DataTable
              data={models}
              rowKey={(row) => row.key}
              columns={[
                { key: "label", title: t("common.model") },
                { key: "requests", title: t("common.request") },
                { key: "base_cost", title: t("common.cost"), render: (row) => fmtMoney(row.base_cost) },
                { key: "charge", title: t("common.charge"), render: (row) => fmtMoney(row.charge) },
                { key: "gross_profit", title: t("common.profit"), render: (row) => fmtMoney(row.gross_profit) }
              ]}
            />
          </CardContent>
        </Card>
      </section>
      <section className="grid gap-4 md:grid-cols-3">
        <StatCard label={t("dashboard.balancePool")} value={fmtMoney(dashboard?.balance_total)} hint={t("dashboard.balanceTotal")} icon={CircleDollarSign} />
        <StatCard label="API Key" value={dashboard?.active_api_keys ?? 0} hint={t("dashboard.activeKeys")} icon={KeyRound} />
        <StatCard label={t("dashboard.channelHealth")} value={`${dashboard?.healthy_channels ?? 0}/${dashboard?.total_channels ?? 0}`} hint={t("dashboard.healthyTotal")} icon={RadioTower} />
      </section>
    </div>
  );
}
