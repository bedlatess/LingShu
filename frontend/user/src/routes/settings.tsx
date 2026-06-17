import { ShieldCheck, UserRound } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";

export function SettingsPage() {
  const { user } = useAuth();
  return (
    <div className="page-grid">
      <PageHeader eyebrow="Settings" title="账户设置" description="查看当前用户状态和余额。密码修改仍使用后端 /api/auth/change-password。" />
      <div className="grid gap-5 lg:grid-cols-2">
        <Card className="glass">
          <CardHeader><CardTitle className="flex items-center gap-2"><UserRound className="h-4 w-4 text-primary" />账户</CardTitle></CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Row label="用户名" value={user?.username ?? "-"} />
            <Row label="邮箱" value={user?.email || "-"} />
            <Row label="角色" value={user?.role ?? "-"} />
            <Row label="状态" value={user?.status ?? "-"} />
          </CardContent>
        </Card>
        <Card className="glass">
          <CardHeader><CardTitle className="flex items-center gap-2"><ShieldCheck className="h-4 w-4 text-primary" />余额安全</CardTitle></CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Row label="当前余额" value={formatMoney(user?.balance)} />
            <Row label="充值方式" value="兑换码 + 管理员手动充值" />
            <Row label="扣费公式" value="base_cost × rate_multiplier" />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-white/10 bg-white/[0.035] px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
