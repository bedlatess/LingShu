import React from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardHeader, CardTitle, Input, PageHeader, toast } from "@lingshu/ui";
import { useAuth } from "@/providers/auth";
import { trStatus } from "@/lib/i18n";

export function SettingsPage() {
  const { t } = useTranslation("settings");
  const { api, user } = useAuth();
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setLoading(true);
    try {
      await api.changePassword({ old_password: oldPassword, new_password: newPassword });
      setOldPassword("");
      setNewPassword("");
      toast.success(t("success"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("failed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} />
      <Card>
        <CardHeader><CardTitle>{t("accountTitle")}</CardTitle></CardHeader>
        <CardContent className="grid gap-3 text-sm">
          <Info label={t("username")} value={user?.username ?? "-"} />
          <Info label={t("email")} value={user?.email || "-"} />
          <Info label={t("status")} value={trStatus(user?.status)} />
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>{t("passwordTitle")}</CardTitle></CardHeader>
        <CardContent>
          <form className="grid max-w-md gap-4" onSubmit={submit}>
            <label className="grid gap-2 text-sm">{t("oldPassword")}<Input type="password" value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} required /></label>
            <label className="grid gap-2 text-sm">{t("newPassword")}<Input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} minLength={8} required /></label>
            <Button type="submit" disabled={loading}>{loading ? t("saving") : t("savePassword")}</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border bg-[var(--bg-subtle)] px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
