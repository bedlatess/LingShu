import React from "react";
import { useTranslation } from "react-i18next";
import type { CleanupHistoryEntry, SystemSetting, createAPI } from "@lingshu/shared";
import { Button, Card, CardContent, CardHeader, CardTitle, DataTable, Input, PageHeader, Select, Tabs, toast } from "@lingshu/ui";
import { formatDateMinute, runWrite } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function SettingsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [settings, setSettings] = React.useState<SystemSetting[]>([]);
  const [values, setValues] = React.useState<Record<string, string>>({});
  const [history, setHistory] = React.useState<CleanupHistoryEntry[]>([]);
  const [tab, setTab] = React.useState("general");

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

  async function saveAlert(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      await api.patchSettings(alertKeys.map((key) => ({ key, value: values[key] ?? defaultAlertValue(key) })));
      toast.success(t("settings.saveSuccess"));
      await refresh();
    }, t("settings.saveFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("settings.eyebrow")} title={t("settings.title")} description={t("settings.description")} />
      <Tabs
        value={tab}
        onChange={setTab}
        tabs={[
          { value: "general", label: t("settings.tabs.general") },
          { value: "alerts", label: t("settings.tabs.alerts") }
        ]}
      />
      {tab === "alerts" ? (
        <Card>
          <CardHeader><CardTitle>{t("settings.alerts.title")}</CardTitle></CardHeader>
          <CardContent>
            <form className="grid gap-4" onSubmit={saveAlert}>
              <label className="grid gap-2 text-sm">
                <span>{t("settings.alerts.enabled")}</span>
                <Select value={values.alert_enabled ?? "false"} onChange={(e) => setValues({ ...values, alert_enabled: e.target.value })}>
                  <option value="true">{t("common.enabled")}</option>
                  <option value="false">{t("common.disabled")}</option>
                </Select>
              </label>
              <div className="grid gap-4 md:grid-cols-2">
                <SettingInput label={t("settings.alerts.channelFailureThreshold")} value={values.alert_channel_failure_threshold ?? "5"} onChange={(value) => setValues({ ...values, alert_channel_failure_threshold: value })} />
                <SettingInput label={t("settings.alerts.gateway5xxRateThreshold")} value={values.alert_gateway_5xx_rate_threshold ?? "0.20"} onChange={(value) => setValues({ ...values, alert_gateway_5xx_rate_threshold: value })} />
                <SettingInput label={t("settings.alerts.upstreamErrorRateThreshold")} value={values.alert_upstream_error_rate_threshold ?? "0.20"} onChange={(value) => setValues({ ...values, alert_upstream_error_rate_threshold: value })} />
                <SettingInput label={t("settings.alerts.lowBalanceThreshold")} value={values.alert_low_balance_threshold ?? "5"} onChange={(value) => setValues({ ...values, alert_low_balance_threshold: value })} />
              </div>
              <SettingInput label={t("settings.alerts.emailRecipients")} value={values.alert_email_recipients ?? ""} onChange={(value) => setValues({ ...values, alert_email_recipients: value })} />
              <SettingInput label={t("settings.alerts.webhookURL")} value={values.alert_webhook_url ?? ""} onChange={(value) => setValues({ ...values, alert_webhook_url: value })} />
              <label className="grid gap-2 text-sm">
                <span>{t("settings.alerts.webhookProvider")}</span>
                <Select value={values.alert_webhook_provider ?? "generic"} onChange={(e) => setValues({ ...values, alert_webhook_provider: e.target.value })}>
                  <option value="generic">Generic JSON</option>
                  <option value="wechat">企业微信</option>
                  <option value="feishu">飞书</option>
                  <option value="dingtalk">钉钉</option>
                  <option value="discord">Discord</option>
                </Select>
              </label>
              <Button type="submit">{t("settings.saveSettings")}</Button>
            </form>
          </CardContent>
        </Card>
      ) : (
        <>
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
        </>
      )}
    </div>
  );
}

const alertKeys = [
  "alert_enabled",
  "alert_channel_failure_threshold",
  "alert_gateway_5xx_rate_threshold",
  "alert_upstream_error_rate_threshold",
  "alert_low_balance_threshold",
  "alert_email_recipients",
  "alert_webhook_url",
  "alert_webhook_provider"
];

function defaultAlertValue(key: string) {
  if (key === "alert_enabled") return "false";
  if (key === "alert_webhook_provider") return "generic";
  if (key === "alert_channel_failure_threshold" || key === "alert_low_balance_threshold") return "5";
  if (key.endsWith("_threshold")) return "0.20";
  return "";
}

function SettingInput({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="grid gap-2 text-sm">
      <span>{label}</span>
      <Input value={value} onChange={(event) => onChange(event.target.value)} />
    </label>
  );
}
