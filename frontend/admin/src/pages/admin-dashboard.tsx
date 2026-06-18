import type { AdminDashboard, User } from "@lingshu/shared";

import { metricCards } from "./admin-page-utils";

export function AdminDashboardPage({ dashboard, auditCount, me }: { dashboard: AdminDashboard | null; auditCount: number | null; me: User }) {
  return metricCards([
    { label: "管理员", value: me.username },
    { label: "今日请求", value: dashboard?.today_requests ?? 0 },
    { label: "今日扣费", value: dashboard?.today_charge ?? "0" },
    { label: "毛利", value: dashboard?.gross_profit ?? "0" },
    { label: "审计日志", value: auditCount ?? "--" }
  ]);
}
