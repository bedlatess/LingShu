import { useEffect, useState } from "react";
import { Button, Card, Input, Space, Table, Tabs, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { createAPI, type AdminDashboard, type GatewayLog, type LedgerRecord, type ReportRow } from "@lingshu/shared";

import { errText, exportCSV, type Pager, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function ReportsPage({ api, dashboard, logs, ledger, logColumns, ledgerColumns, logsPager, setLogsPager, ledgerPager, setLedgerPager }: { api: AdminAPI; dashboard: AdminDashboard | null; logs: GatewayLog[]; ledger: LedgerRecord[]; logColumns: ColumnsType<GatewayLog>; ledgerColumns: ColumnsType<LedgerRecord>; logsPager: Pager; setLogsPager: React.Dispatch<React.SetStateAction<Pager>>; ledgerPager: Pager; setLedgerPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [daily, setDaily] = useState<ReportRow[]>([]);
  const [byUser, setByUser] = useState<ReportRow[]>([]);
  const [byModel, setByModel] = useState<ReportRow[]>([]);
  const [byChannel, setByChannel] = useState<ReportRow[]>([]);
  const [loading, setLoading] = useState(false);

  async function loadReports() {
    setLoading(true);
    try {
      const [dailyResult, userResult, modelResult, channelResult] = await Promise.all([
        api.adminReportDaily(from, to),
        api.adminReportByUser(from, to),
        api.adminReportByModel(from, to),
        api.adminReportByChannel(from, to)
      ]);
      setDaily(dailyResult.items);
      setByUser(userResult.items);
      setByModel(modelResult.items);
      setByChannel(channelResult.items);
    } catch (err) {
      message.error(`加载报表失败: ${errText(err)}`);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadReports();
  }, []);

  const reportColumns: ColumnsType<ReportRow> = [
    { title: "维度", dataIndex: "label" },
    { title: "请求数", dataIndex: "requests" },
    { title: "成功", dataIndex: "successes" },
    { title: "失败", dataIndex: "failures" },
    { title: "成本", dataIndex: "base_cost" },
    { title: "扣费", dataIndex: "charge" },
    { title: "毛利", dataIndex: "gross_profit" }
  ];
  const reportTable = (key: string, rows: ReportRow[]) => (
    <Card
      title="聚合报表"
      extra={<Button onClick={() => exportCSV(`report-${key}.csv`, rows)}>导出 CSV</Button>}
    >
      <Table rowKey="key" loading={loading} columns={reportColumns} dataSource={rows} />
    </Card>
  );

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Space wrap>
        <Card><Typography.Text type="secondary">活跃用户</Typography.Text><Typography.Title level={4}>{dashboard?.active_users ?? 0}</Typography.Title></Card>
        <Card><Typography.Text type="secondary">余额池</Typography.Text><Typography.Title level={4}>{dashboard?.balance_total ?? "0"}</Typography.Title></Card>
        <Card><Typography.Text type="secondary">成本</Typography.Text><Typography.Title level={4}>{dashboard?.today_base_cost ?? "0"}</Typography.Title></Card>
      </Space>
      <Card>
        <Space wrap>
          <Input type="date" aria-label="开始日期" value={from} onChange={(event) => setFrom(event.target.value)} />
          <Input type="date" aria-label="结束日期" value={to} onChange={(event) => setTo(event.target.value)} />
          <Button type="primary" onClick={loadReports}>查询</Button>
          <Button onClick={() => { setFrom(""); setTo(""); }}>重置日期</Button>
        </Space>
      </Card>
      <Tabs
        items={[
          { key: "daily", label: "按日", children: reportTable("daily", daily) },
          { key: "user", label: "按用户", children: reportTable("by-user", byUser) },
          { key: "model", label: "按模型", children: reportTable("by-model", byModel) },
          { key: "channel", label: "按渠道", children: reportTable("by-channel", byChannel) },
          { key: "logs", label: "请求日志", children: <Card title="全站调用日志"><Table rowKey="request_id" columns={logColumns} dataSource={logs} pagination={tablePagination(logsPager, setLogsPager)} /></Card> },
          { key: "ledger", label: "账本", children: <Card title="全站账本"><Table rowKey={(item) => `${item.type}-${item.created_at}-${item.amount}`} columns={ledgerColumns} dataSource={ledger} pagination={tablePagination(ledgerPager, setLedgerPager)} /></Card> }
        ]}
      />
    </Space>
  );
}
