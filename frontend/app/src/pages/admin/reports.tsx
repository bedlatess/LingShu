import React from "react";
import { useTranslation } from "react-i18next";
import type { GatewayLog, LedgerRecord, ReportRow, createAPI } from "@lingshu/shared";
import { Button, Card, CardContent, DataTable, PageHeader, TabsList, TabsTrigger } from "@lingshu/ui";
import { exportCSV, fmtMoney, formatDateMinute, MiniBars } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;
type ReportTab = "daily" | "user" | "model" | "channel" | "logs" | "ledger";

export function ReportsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [tab, setTab] = React.useState<ReportTab>("daily");
  const [rows, setRows] = React.useState<ReportRow[]>([]);
  const [logs, setLogs] = React.useState<GatewayLog[]>([]);
  const [ledger, setLedger] = React.useState<LedgerRecord[]>([]);

  async function refresh() {
    const [daily, byUser, byModel, byChannel, logList, ledgerList] = await Promise.all([api.adminReportDaily("", ""), api.adminReportByUser("", ""), api.adminReportByModel("", ""), api.adminReportByChannel("", ""), api.adminLogs(1, 50), api.adminLedger(1, 50)]);
    const map = { daily: daily.items, user: byUser.items, model: byModel.items, channel: byChannel.items };
    setRows(map[tab === "logs" || tab === "ledger" ? "daily" : tab]);
    setLogs(logList.items);
    setLedger(ledgerList.items);
  }

  React.useEffect(() => { refresh(); }, [api, tab]);

  const reportColumns = [
    { key: "label", title: t("reports.dimension") },
    { key: "requests", title: t("common.request") },
    { key: "base_cost", title: t("common.cost"), render: (row: ReportRow) => fmtMoney(row.base_cost) },
    { key: "charge", title: t("common.charge"), render: (row: ReportRow) => fmtMoney(row.charge) },
    { key: "gross_profit", title: t("common.profit"), render: (row: ReportRow) => fmtMoney(row.gross_profit) }
  ];

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("reports.eyebrow")} title={t("reports.title")} description={t("reports.description")} action={<Button variant="secondary" onClick={() => exportCSV(`report-${tab}.csv`, (tab === "logs" ? logs : tab === "ledger" ? ledger : rows) as object[])}>{t("common.exportCSV")}</Button>} />
      <TabsList>{(["daily", "user", "model", "channel", "logs", "ledger"] as ReportTab[]).map((value) => <TabsTrigger key={value} value={value} activeValue={tab} onSelect={(next) => setTab(next as ReportTab)}>{label(value, t)}</TabsTrigger>)}</TabsList>
      {tab === "logs" ? (
        <DataTable
          data={logs}
          rowKey={(row) => row.request_id}
          columns={[
            { key: "request_id", title: t("common.request") },
            { key: "model_id", title: t("common.model") },
            { key: "base_cost", title: t("common.cost"), render: (row) => fmtMoney(row.base_cost) },
            { key: "charge", title: t("common.charge"), render: (row) => fmtMoney(row.charge) },
            { key: "created_at", title: t("common.time"), render: (row) => formatDateMinute(row.created_at) }
          ]}
        />
      ) : tab === "ledger" ? (
        <DataTable
          data={ledger}
          rowKey={(row, i) => `${row.created_at}-${i}`}
          columns={[
            { key: "user_id", title: t("common.user") },
            { key: "type", title: t("common.type") },
            { key: "amount", title: t("common.amount"), render: (row) => fmtMoney(row.amount) },
            { key: "base_cost", title: t("common.cost"), render: (row) => fmtMoney(row.base_cost) },
            { key: "rate_multiplier", title: t("common.multiplier") },
            { key: "created_at", title: t("common.time"), render: (row) => formatDateMinute(row.created_at) }
          ]}
        />
      ) : (
        <>
          <Card><CardContent className="p-5"><MiniBars data={rows.slice(0, 12).map((item) => ({ label: item.label, value: Number(item.charge) }))} /></CardContent></Card>
          <DataTable data={rows} rowKey={(row) => row.key} columns={reportColumns} />
        </>
      )}
    </div>
  );
}

function label(tab: ReportTab, t: (key: string) => string) {
  return t(`reports.tabs.${tab}`);
}
