import React from "react";
import { useTranslation } from "react-i18next";
import type { CleanupHistoryEntry, SystemSetting, createAPI } from "@lingshu/shared";
import { Button, Card, CardContent, CardHeader, CardTitle, DataTable, Input, PageHeader, toast } from "@lingshu/ui";
import { formatDateMinute, runWrite } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function SettingsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [settings, setSettings] = React.useState<SystemSetting[]>([]);
  const [values, setValues] = React.useState<Record<string, string>>({});
  const [history, setHistory] = React.useState<CleanupHistoryEntry[]>([]);

  async function refresh() {
    const [settingList, cleanup] = await Promise.all([api.listSettings(1, 100), api.cleanupHistory(10)]);
    setSettings(settingList.items);
    setValues(Object.fromEntries(settingList.items.map((item) => [item.key, item.value])));
    setHistory(cleanup.items);
  }

  React.useEffect(() => { refresh(); }, [api]);

  async function save(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      await api.patchSettings(settings.map((item) => ({ key: item.key, value: values[item.key] ?? "" })));
      toast.success(t("settings.saveSuccess"));
      await refresh();
    }, t("settings.saveFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("settings.eyebrow")} title={t("settings.title")} description={t("settings.description")} />
      <Card>
        <CardHeader><CardTitle>{t("settings.configItems")}</CardTitle></CardHeader>
        <CardContent>
          <form className="grid gap-4" onSubmit={save}>
            {settings.map((item) => (
              <label key={item.key} className="grid gap-2 text-sm">
                <span>{item.key}</span>
                <Input value={values[item.key] ?? ""} onChange={(e) => setValues({ ...values, [item.key]: e.target.value })} />
                <span className="text-xs text-muted-foreground">{item.description}</span>
              </label>
            ))}
            <Button type="submit">{t("settings.saveSettings")}</Button>
          </form>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>{t("settings.cleanup")}</CardTitle></CardHeader>
        <CardContent className="grid gap-4">
          <Button variant="secondary" onClick={() => runWrite(async () => { await api.runCleanup(); toast.success(t("settings.cleanupSuccess")); await refresh(); }, t("settings.cleanupFailed"))}>{t("settings.runCleanup")}</Button>
          <DataTable
            data={history}
            rowKey={(row) => row.id}
            columns={[
              { key: "started_at", title: t("settings.table.startedAt"), render: (row) => formatDateMinute(row.started_at) },
              { key: "ended_at", title: t("settings.table.endedAt"), render: (row) => formatDateMinute(row.ended_at) },
              { key: "results", title: t("settings.table.result"), render: (row) => row.results.map((r) => `${r.table}:${r.deleted}`).join(" / ") }
            ]}
          />
        </CardContent>
      </Card>
    </div>
  );
}
