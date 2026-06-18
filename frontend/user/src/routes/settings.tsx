import React from "react";
import { KeyRound, ShieldCheck, UserRound } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/page-header";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { zhStatus } from "@/lib/i18n";

export function SettingsPage() {
  const { user, api } = useAuth();
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [submitting, setSubmitting] = React.useState(false);

  async function changePassword(event: React.FormEvent) {
    event.preventDefault();
    if (newPassword.length < 6) {
      toast.error("新密码至少 6 位");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("两次输入的新密码不一致");
      return;
    }
    setSubmitting(true);
    try {
      await api.changePassword({ old_password: oldPassword, new_password: newPassword });
      toast.success("密码已修改");
      setOldPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      const message = err instanceof Error ? err.message : "修改失败";
      toast.error(`修改失败：${message === "invalid credentials" ? "原密码错误" : message}`);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow="账户设置" title="账户设置" description="查看当前账户信息并管理登录密码。" />
      <div className="grid gap-5 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><UserRound className="h-4 w-4 text-primary" />账户</CardTitle></CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Row label="用户名" value={user?.username ?? "-"} />
            <Row label="邮箱" value={user?.email || "-"} />
            <Row label="角色" value={user?.role ?? "-"} />
            <Row label="状态" value={zhStatus(user?.status)} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><ShieldCheck className="h-4 w-4 text-primary" />余额安全</CardTitle></CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Row label="当前余额" value={formatMoney(user?.balance)} />
            <Row label="安全提示" value="请妥善保管 API 密钥" />
          </CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader><CardTitle className="flex items-center gap-2"><KeyRound className="h-4 w-4 text-primary" />修改密码</CardTitle></CardHeader>
        <CardContent>
          <form className="grid max-w-md gap-4" onSubmit={changePassword}>
            <div className="grid gap-1.5">
              <label className="text-sm text-muted-foreground">原密码</label>
              <Input type="password" value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} placeholder="请输入原密码" required />
            </div>
            <div className="grid gap-1.5">
              <label className="text-sm text-muted-foreground">新密码</label>
              <Input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} placeholder="至少 6 位" required />
            </div>
            <div className="grid gap-1.5">
              <label className="text-sm text-muted-foreground">确认新密码</label>
              <Input type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} placeholder="再次输入新密码" required />
            </div>
            <Button type="submit" disabled={submitting}>{submitting ? "提交中..." : "确认修改"}</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-[var(--bg-subtle)] px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <strong className="text-foreground">{value}</strong>
    </div>
  );
}
