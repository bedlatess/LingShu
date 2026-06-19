import React from "react";
import { useTranslation } from "react-i18next";
import { Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts";
import type { UserDailyStat, UserGatewayLog, UserLedgerRecord, UserModelStat } from "@lingshu/shared/user-types";

import { Button, Card, CardContent, CardHeader, CardTitle, DataTable, Dialog, EmptyState, Input, PageHeader, TabsList, TabsTrigger, Tag, toast } from "@lingshu/ui";
import { MeasuredChart } from "@/components/measured-chart";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { trLedgerType, trStatus } from "@/lib/i18n";

type UsageTab = "daily" | "models" | "ledger" | "logs";

export function UsagePage() {
  const { t, i18n } = useTranslation("usage");
  const { api } = useAuth();
  const tabs = React.useMemo(() => ([
    ["daily", t("tabs.daily")],
    ["models", t("tabs.models")],
    ["ledger", t("tabs.ledger")],
    ["logs", t("tabs.logs")]
  ] as const), [t]);
  const [active, setActive] = React.useState<UsageTab>("daily");
  const [daily, setDaily] = React.useState<UserDailyStat[]>([]);
  const [models, setModels] = React.useState<UserModelStat[]>([]);
  const [ledger, setLedger] = React.useState<UserLedgerRecord[]>([]);
  const [logs, setLogs] = React.useState<UserGatewayLog[]>([]);
  const [query, setQuery] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [fromDate, setFromDate] = React.useState("");
  const [toDate, setToDate] = React.useState("");
  const [selectedLog, setSelectedLog] = React.useState<UserGatewayLog | null>(null);
  const timezone = React.useMemo(() => Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC", []);

  React.useEffect(() => {
    let mounted = true;
    async function loadAll() {
      const [dailyResult, modelResult, ledgerResult, logResult] = await Promise.all([api.userDailyStats(), api.userModelStats(), api.userLedger(), api.userLogs()]);
      if (!mounted) return;
      setDaily(dailyResult.items);
      setModels(modelResult.items);
      setLedger(ledgerResult.items);
      setLogs(logResult.items);
    }
    void loadAll();
    return () => {
      mounted = false;
    };
  }, [api]);

  React.useEffect(() => {
    let mounted = true;
    const timer = window.setInterval(() => {
      api.userLogs().then((result) => {
        if (mounted) setLogs(result.items);
      }).catch(() => {
        // Keep the last successful snapshot; the global toast layer handles explicit writes.
      });
    }, 5000);
    return () => {
      mounted = false;
      window.clearInterval(timer);
    };
  }, [api]);

  const dateRange = React.useMemo(() => ({ from: fromDate, to: toDate }), [fromDate, toDate]);
  const visibleDaily = React.useMemo(() => filterDailyByDate(daily, dateRange), [daily, dateRange]);
  const visibleLedger = React.useMemo(() => filteredLedger(ledger, query, dateRange), [ledger, query, dateRange]);
  const visibleLogs = React.useMemo(() => filteredLogs(logs, query, status, dateRange), [logs, query, status, dateRange]);
  const inFlightLogs = React.useMemo(() => logs.filter(isInFlightLog), [logs]);

  async function downloadUsageCSV() {
    try {
      const blob = await api.downloadUserUsageCSV();
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = "usage.csv";
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      URL.revokeObjectURL(url);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "导出失败");
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} action={<Button variant="secondary" onClick={() => void downloadUsageCSV()}>{t("exportCSV", "导出 CSV")}</Button>} />
      <TabsList>
        {tabs.map(([value, label]) => <TabsTrigger key={value} value={value} activeValue={active} onSelect={(next) => setActive(next as UsageTab)}>{label}</TabsTrigger>)}
      </TabsList>
      <Card>
        <CardContent className="grid gap-3 p-4 md:grid-cols-[1fr_180px_180px_160px_160px]">
          <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("filters.search")} />
          <select className="h-10 rounded-md border border-border bg-card px-3 text-sm text-foreground" value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">{t("filters.allStatus")}</option>
            <option value="success">{t("filters.success")}</option>
            <option value="failed">{t("filters.failed")}</option>
          </select>
          <Input type="date" value={fromDate} onChange={(event) => setFromDate(event.target.value)} aria-label={t("filters.from")} />
          <Input type="date" value={toDate} onChange={(event) => setToDate(event.target.value)} aria-label={t("filters.to")} />
          <div className="rounded-md border border-border bg-[var(--bg-subtle)] px-3 py-2 text-sm text-muted-foreground">
            {t("filters.timezone", { timezone })}
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>{t("inFlight.title", { count: inFlightLogs.length })}</CardTitle>
        </CardHeader>
        <CardContent>
          {inFlightLogs.length ? (
            <div className="grid gap-2">
              {inFlightLogs.slice(0, 5).map((item) => (
                <button
                  key={item.request_id}
                  type="button"
                  onClick={() => setSelectedLog(item)}
                  className="grid gap-1 rounded-md border border-border bg-[var(--bg-subtle)] p-3 text-left text-sm transition-colors hover:border-[var(--border-strong)]"
                >
                  <span className="font-mono text-xs text-foreground">{shortID(item.request_id)}</span>
                  <span className="text-muted-foreground">{item.model_id} · {formatDate(item.created_at, i18n.resolvedLanguage)}</span>
                </button>
              ))}
            </div>
          ) : (
            <EmptyState title={t("inFlight.emptyTitle")} description={t("inFlight.emptyDescription")} />
          )}
          <p className="mt-3 text-xs text-muted-foreground">{t("inFlight.pollingHint")}</p>
        </CardContent>
      </Card>
      {active === "daily" ? (
        <Card>
          <CardHeader><CardTitle>{t("dailyTitle")}</CardTitle></CardHeader>
          <CardContent>{visibleDaily.length ? <UsageBars data={visibleDaily.map((item) => ({ label: item.day, value: Number(item.charge) }))} /> : <EmptyState title={t("dailyEmptyTitle")} description={t("dailyEmptyDescription")} />}</CardContent>
        </Card>
      ) : null}
      {active === "models" ? (
        <Card>
          <CardHeader><CardTitle>{t("modelsTitle")}</CardTitle></CardHeader>
          <CardContent>{models.length ? <UsageBars data={models.map((item) => ({ label: item.model_id, value: Number(item.charge) }))} /> : <EmptyState title={t("modelsEmptyTitle")} description={t("modelsEmptyDescription")} />}</CardContent>
        </Card>
      ) : null}
      {active === "ledger" ? (
        <DataTable
          data={visibleLedger}
          rowKey={(row, index) => `${row.type}-${row.created_at}-${index}`}
          columns={[
            { key: "type", title: t("ledgerTable.type"), render: (row) => trLedgerType(row.type) },
            { key: "amount", title: t("ledgerTable.amount"), render: (row) => formatMoney(row.amount) },
            { key: "balance_after", title: t("ledgerTable.balanceAfter"), render: (row) => formatMoney(row.balance_after) },
            { key: "remark", title: t("ledgerTable.remark") },
            { key: "created_at", title: t("ledgerTable.createdAt"), render: (row) => formatDate(row.created_at, i18n.resolvedLanguage) }
          ]}
        />
      ) : null}
      {active === "logs" ? (
        <DataTable
          data={visibleLogs}
          rowKey={(row) => row.request_id}
          onRowClick={(row) => setSelectedLog(row)}
          columns={[
            { key: "request_id", title: t("logsTable.requestId"), render: (row) => shortID(row.request_id) },
            { key: "model_id", title: t("logsTable.model") },
            { key: "status", title: t("logsTable.status"), render: (row) => <Tag variant={row.status === "success" ? "success" : "danger"}>{trStatus(row.status)}</Tag> },
            { key: "total_tokens", title: t("logsTable.tokens") },
            { key: "charge", title: t("logsTable.charge"), render: (row) => formatMoney(row.charge) },
            { key: "created_at", title: t("logsTable.createdAt"), render: (row) => formatDate(row.created_at, i18n.resolvedLanguage) }
          ]}
        />
      ) : null}
      <Dialog open={Boolean(selectedLog)} title={t("detail.title")} onClose={() => setSelectedLog(null)}>
        {selectedLog ? (
          <div className="grid gap-3 text-sm">
            <DetailRow label={t("logsTable.requestId")} value={selectedLog.request_id} mono />
            <DetailRow label={t("logsTable.model")} value={selectedLog.model_id} />
            <DetailRow label={t("logsTable.status")} value={`${trStatus(selectedLog.status)} / HTTP ${selectedLog.http_status || "-"}`} />
            <DetailRow label={t("logsTable.tokens")} value={String(selectedLog.total_tokens)} />
            <DetailRow label={t("logsTable.charge")} value={formatMoney(selectedLog.charge)} />
            <DetailRow label={t("logsTable.createdAt")} value={formatDate(selectedLog.created_at, i18n.resolvedLanguage)} />
          </div>
        ) : null}
      </Dialog>
    </div>
  );
}

function filteredLedger(items: UserLedgerRecord[], query: string, range: DateRange) {
  const keyword = query.trim().toLowerCase();
  return items.filter((item) => {
    const keywordOK = !keyword || [item.type, item.remark, item.amount, item.balance_after].some((value) => String(value ?? "").toLowerCase().includes(keyword));
    return keywordOK && isWithinRange(item.created_at, range);
  });
}

function filteredLogs(items: UserGatewayLog[], query: string, status: string, range: DateRange) {
  const keyword = query.trim().toLowerCase();
  return items.filter((item) => {
    const statusOK = !status || item.status === status;
    const keywordOK = !keyword || [item.request_id, item.model_id, item.status, item.charge].some((value) => String(value ?? "").toLowerCase().includes(keyword));
    return statusOK && keywordOK && isWithinRange(item.created_at, range);
  });
}

function DetailRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-md border border-border bg-[var(--bg-subtle)] p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={mono ? "mt-1 break-all font-mono text-xs text-foreground" : "mt-1 text-foreground"}>{value}</p>
    </div>
  );
}

function UsageBars({ data }: { data: { label: string; value: number }[] }) {
  return (
    <MeasuredChart>
      {({ width, height }) => (
        <BarChart data={data} width={width} height={height}>
          <CartesianGrid stroke="#D8D4CA" vertical={false} />
          <XAxis dataKey="label" stroke="#87867F" tickLine={false} axisLine={false} />
          <YAxis stroke="#87867F" tickLine={false} axisLine={false} />
          <Tooltip contentStyle={{ background: "#FAF9F5", border: "1px solid #D8D4CA", borderRadius: 6 }} />
          <Bar dataKey="value" fill="#C6613F" radius={[6, 6, 0, 0]} />
        </BarChart>
      )}
    </MeasuredChart>
  );
}

type DateRange = { from: string; to: string };

function filterDailyByDate(items: UserDailyStat[], range: DateRange) {
  return items.filter((item) => isDayWithinRange(item.day, range));
}

function isInFlightLog(item: UserGatewayLog) {
  return item.status !== "success" && item.http_status === 0;
}

function isWithinRange(value: string | undefined, range: DateRange) {
  if (!value) return true;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return true;
  const from = range.from ? new Date(`${range.from}T00:00:00`) : null;
  const to = range.to ? new Date(`${range.to}T23:59:59.999`) : null;
  if (from && date < from) return false;
  if (to && date > to) return false;
  return true;
}

function isDayWithinRange(day: string, range: DateRange) {
  if (!range.from && !range.to) return true;
  if (range.from && day < range.from) return false;
  if (range.to && day > range.to) return false;
  return true;
}

function formatDate(value?: string, language?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString(language === "zh" ? "zh-CN" : "en-US", { hour12: false });
}

function shortID(value: string) {
  return value.length > 16 ? `${value.slice(0, 8)}...${value.slice(-6)}` : value;
}
