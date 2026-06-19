import React from "react";
import { useTranslation } from "react-i18next";
import type { CleanupHistoryEntry, createAPI } from "@lingshu/shared";
import { Button, Card, CardContent, CardHeader, CardTitle, DataTable, Input, PageHeader, Select, Tabs, toast } from "@lingshu/ui";
import { formatDateMinute, runWrite } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function SettingsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [values, setValues] = React.useState<Record<string, string>>({});
  const [history, setHistory] = React.useState<CleanupHistoryEntry[]>([]);
  const [tab, setTab] = React.useState("general");

  async function refresh() {
    const [settingList, cleanup] = await Promise.all([api.listSettings(1, 100), api.cleanupHistory(10)]);
    setValues(Object.fromEntries(settingList.items.map((item) => [item.key, item.value])));
    setHistory(cleanup.items);
  }

  React.useEffect(() => { refresh(); }, [api]);

  async function saveGroup(event: React.FormEvent, keys: string[]) {
    event.preventDefault();
    await runWrite(async () => {
      await api.patchSettings(keys.map((key) => ({ key, value: values[key] ?? "" })));
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

  const updateValue = React.useCallback((key: string, value: string) => {
    setValues((current) => ({ ...current, [key]: value }));
  }, []);

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
          <CardHeader>
            <CardTitle>{t("settings.alerts.title")}</CardTitle>
            <p className="text-sm leading-6 text-muted-foreground">{t("settings.alerts.description")}</p>
          </CardHeader>
          <CardContent>
            <form className="grid gap-4" onSubmit={saveAlert}>
              <label className="grid gap-2 text-sm">
                <span>{t("settings.alerts.enabled")}</span>
                <Select value={values.alert_enabled ?? "false"} onChange={(event) => updateValue("alert_enabled", event.target.value)}>
                  <option value="true">{t("common.enabled")}</option>
                  <option value="false">{t("common.disabled")}</option>
                </Select>
                <span className="text-xs leading-5 text-muted-foreground">{t("settings.helpers.alert_enabled")}</span>
              </label>
              <div className="grid gap-4 md:grid-cols-2">
                <SettingInput label={t("settings.alerts.channelFailureThreshold")} helper={t("settings.helpers.alert_channel_failure_threshold")} value={values.alert_channel_failure_threshold ?? "5"} onChange={(value) => updateValue("alert_channel_failure_threshold", value)} />
                <SettingInput label={t("settings.alerts.gateway5xxRateThreshold")} helper={t("settings.helpers.alert_gateway_5xx_rate_threshold")} value={values.alert_gateway_5xx_rate_threshold ?? "0.20"} onChange={(value) => updateValue("alert_gateway_5xx_rate_threshold", value)} />
                <SettingInput label={t("settings.alerts.upstreamErrorRateThreshold")} helper={t("settings.helpers.alert_upstream_error_rate_threshold")} value={values.alert_upstream_error_rate_threshold ?? "0.20"} onChange={(value) => updateValue("alert_upstream_error_rate_threshold", value)} />
                <SettingInput label={t("settings.alerts.lowBalanceThreshold")} helper={t("settings.helpers.alert_low_balance_threshold")} value={values.alert_low_balance_threshold ?? "5"} onChange={(value) => updateValue("alert_low_balance_threshold", value)} />
              </div>
              <SettingInput label={t("settings.alerts.emailRecipients")} helper={t("settings.helpers.alert_email_recipients")} value={values.alert_email_recipients ?? ""} onChange={(value) => updateValue("alert_email_recipients", value)} />
              <SettingInput label={t("settings.alerts.webhookURL")} helper={t("settings.helpers.alert_webhook_url")} value={values.alert_webhook_url ?? ""} onChange={(value) => updateValue("alert_webhook_url", value)} />
              <label className="grid gap-2 text-sm">
                <span>{t("settings.alerts.webhookProvider")}</span>
                <Select value={values.alert_webhook_provider ?? "generic"} onChange={(event) => updateValue("alert_webhook_provider", event.target.value)}>
                  <option value="generic">Generic JSON</option>
                  <option value="wechat">{t("settings.alerts.providers.wechat")}</option>
                  <option value="feishu">{t("settings.alerts.providers.feishu")}</option>
                  <option value="dingtalk">{t("settings.alerts.providers.dingtalk")}</option>
                  <option value="discord">Discord</option>
                </Select>
                <span className="text-xs leading-5 text-muted-foreground">{t("settings.helpers.alert_webhook_provider")}</span>
              </label>
              <Button type="submit">{t("settings.saveSettings")}</Button>
            </form>
          </CardContent>
        </Card>
      ) : (
        <>
          <section className="grid gap-4 xl:grid-cols-3">
            <SettingsCard title={t("settings.groups.basic.title")} description={t("settings.groups.basic.description")} keys={basicKeys} values={values} onChange={updateValue} onSubmit={(event) => saveGroup(event, basicKeys)} />
            <SettingsCard title={t("settings.groups.security.title")} description={t("settings.groups.security.description")} keys={securityKeys} values={values} onChange={updateValue} onSubmit={(event) => saveGroup(event, securityKeys)} />
            <SettingsCard title={t("settings.groups.smtp.title")} description={t("settings.groups.smtp.description")} keys={smtpKeys} values={values} onChange={updateValue} onSubmit={(event) => saveGroup(event, smtpKeys)} />
          </section>
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

const basicKeys = ["site_name", "site_icp", "site_police_beian", "contact_email", "api_base_url", "registration_mode"];
const securityKeys = ["captcha_enabled", "captcha_provider", "trusted_proxy_enabled", "trusted_proxy_hops", "access_blacklist_auto_ttl_days"];
const smtpKeys = ["smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from"];
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

function SettingsCard({
  title,
  description,
  keys,
  values,
  onChange,
  onSubmit
}: {
  title: string;
  description: string;
  keys: string[];
  values: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onSubmit: (event: React.FormEvent) => void;
}) {
  const { t } = useTranslation("admin");
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <p className="text-sm leading-6 text-muted-foreground">{description}</p>
      </CardHeader>
      <CardContent>
        <form className="grid gap-4" onSubmit={onSubmit}>
          {keys.map((key) => <SettingField key={key} settingKey={key} value={values[key] ?? ""} onChange={(value) => onChange(key, value)} />)}
          <Button type="submit" variant="secondary">{t("settings.saveGroup")}</Button>
        </form>
      </CardContent>
    </Card>
  );
}

function SettingField({ settingKey, value, onChange }: { settingKey: string; value: string; onChange: (value: string) => void }) {
  const { t } = useTranslation("admin");
  if (booleanKeys.has(settingKey)) {
    return (
      <label className="grid gap-2 text-sm">
        <span>{t(`settings.labels.${settingKey}`)}</span>
        <Select value={value || "false"} onChange={(event) => onChange(event.target.value)}>
          <option value="true">{t("common.enabled")}</option>
          <option value="false">{t("common.disabled")}</option>
        </Select>
        <span className="text-xs leading-5 text-muted-foreground">{t(`settings.helpers.${settingKey}`)}</span>
      </label>
    );
  }
  if (settingKey === "registration_mode") {
    return (
      <label className="grid gap-2 text-sm">
        <span>{t("settings.labels.registration_mode")}</span>
        <Select value={value || "closed"} onChange={(event) => onChange(event.target.value)}>
          <option value="closed">{t("settings.registrationModes.closed")}</option>
          <option value="invite">{t("settings.registrationModes.invite")}</option>
          <option value="open">{t("settings.registrationModes.open")}</option>
        </Select>
        <span className="text-xs leading-5 text-muted-foreground">{t("settings.helpers.registration_mode")}</span>
      </label>
    );
  }
  if (settingKey === "captcha_provider") {
    return (
      <label className="grid gap-2 text-sm">
        <span>{t("settings.labels.captcha_provider")}</span>
        <Select value={value || "turnstile"} onChange={(event) => onChange(event.target.value)}>
          <option value="turnstile">Cloudflare Turnstile</option>
          <option value="hcaptcha">hCaptcha</option>
        </Select>
        <span className="text-xs leading-5 text-muted-foreground">{t("settings.helpers.captcha_provider")}</span>
      </label>
    );
  }
  return (
    <SettingInput
      label={t(`settings.labels.${settingKey}`)}
      helper={t(`settings.helpers.${settingKey}`)}
      type={passwordKeys.has(settingKey) ? "password" : numericKeys.has(settingKey) ? "number" : "text"}
      value={value}
      onChange={onChange}
    />
  );
}

const booleanKeys = new Set(["captcha_enabled", "trusted_proxy_enabled"]);
const numericKeys = new Set(["smtp_port", "trusted_proxy_hops", "access_blacklist_auto_ttl_days"]);
const passwordKeys = new Set(["smtp_pass"]);

function SettingInput({ label, helper, type = "text", value, onChange }: { label: string; helper?: string; type?: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="grid gap-2 text-sm">
      <span>{label}</span>
      <Input type={type} value={value} onChange={(event) => onChange(event.target.value)} />
      {helper ? <span className="text-xs leading-5 text-muted-foreground">{helper}</span> : null}
    </label>
  );
}
