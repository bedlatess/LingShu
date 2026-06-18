import { useEffect, useState } from "react";
import { Alert, Button, Card, Form, Input, Select, Space, Table, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { Link, useParams } from "react-router-dom";
import { createAPI, type APIKey, type GatewayLog, type LedgerRecord, type User } from "@lingshu/shared";

import { errText, exportCSV, type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function UsersPage({ users, userColumns, api, refresh, pager, setPager }: { users: User[]; userColumns: ColumnsType<User>; api: AdminAPI; refresh: () => Promise<void>; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建用户">
        <Form layout="inline" onFinish={(values) => runWrite(async () => { await api.createUser(values); message.success("用户已创建"); await refresh(); }, "创建用户失败").catch(() => undefined)}>
          <Form.Item name="username" rules={[{ required: true }]}><Input placeholder="用户名" /></Form.Item>
          <Form.Item name="email"><Input placeholder="邮箱" /></Form.Item>
          <Form.Item name="password" rules={[{ required: true, min: 8 }]}><Input.Password placeholder="初始密码" /></Form.Item>
          <Form.Item name="role" initialValue="user"><Select style={{ width: 120 }} options={[{ value: "user" }, { value: "admin" }]} /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="用户列表"><Table rowKey="id" columns={userColumns} dataSource={users} pagination={tablePagination(pager, setPager)} /></Card>
    </Space>
  );
}

export function UserDetailPage({ api }: { api: AdminAPI }) {
  const { id } = useParams();
  const [user, setUser] = useState<User | null>(null);
  const [summary, setSummary] = useState<{ total_charge: string; total_recharge: string } | null>(null);
  const [apiKeys, setAPIKeys] = useState<APIKey[]>([]);
  const [logs, setLogs] = useState<GatewayLog[]>([]);
  const [ledger, setLedger] = useState<LedgerRecord[]>([]);
  const [apiKeysPager, setAPIKeysPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [logsPager, setLogsPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [ledgerPager, setLedgerPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [logStatus, setLogStatus] = useState<string>("all");
  const [logModel, setLogModel] = useState<string>("");
  const [ledgerType, setLedgerType] = useState<string>("all");
  const [dateFrom, setDateFrom] = useState<string>("");
  const [dateTo, setDateTo] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function refreshDetail() {
    if (!id) return;
    setLoading(true);
    setError("");
    try {
      const [userItem, keyList, logList, ledgerList, summaryItem] = await Promise.all([
        api.getUser(id),
        api.adminUserAPIKeys(id, apiKeysPager.page, apiKeysPager.limit),
        api.adminUserLogs(id, logsPager.page, logsPager.limit, { status: logStatus, model: logModel, from: dateFrom, to: dateTo }),
        api.adminUserLedger(id, ledgerPager.page, ledgerPager.limit, { type: ledgerType, from: dateFrom, to: dateTo }),
        api.adminUserSummary(id)
      ]);
      setUser(userItem);
      setAPIKeys(keyList.items);
      setLogs(logList.items);
      setLedger(ledgerList.items);
      setSummary(summaryItem);
      setAPIKeysPager((prev) => ({ ...prev, total: keyList.total }));
      setLogsPager((prev) => ({ ...prev, total: logList.total }));
      setLedgerPager((prev) => ({ ...prev, total: ledgerList.total }));
    } catch (err) {
      setError(errText(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    refreshDetail();
  }, [id]);

  useEffect(() => {
    if (id) {
      refreshDetail();
    }
  }, [logStatus, logModel, ledgerType, dateFrom, dateTo, apiKeysPager.page, apiKeysPager.limit, logsPager.page, logsPager.limit, ledgerPager.page, ledgerPager.limit]);

  if (!id) return <Alert type="error" message="缺少用户 ID" />;
  if (loading) return <Card title="用户详情">加载中...</Card>;
  if (error) return <Alert type="error" message={error} />;
  if (!user) return <Alert type="warning" message="未找到用户" />;

  const inDateRange = (value: string) => {
    if (!value) return true;
    const time = new Date(value).getTime();
    if (dateFrom && time < new Date(dateFrom).getTime()) return false;
    if (dateTo && time > new Date(`${dateTo}T23:59:59`).getTime()) return false;
    return true;
  };
  const filteredLogs = logs.filter((item) => inDateRange(item.created_at));
  const filteredLedger = ledger.filter((item) => inDateRange(item.created_at));
  const modelOptions = Array.from(new Set(logs.map((item) => item.model_id).filter(Boolean))).map((value) => ({ value, label: value }));
  const ledgerTypeOptions = Array.from(new Set(ledger.map((item) => item.type).filter(Boolean))).map((value) => ({ value, label: value }));
  const resetFilters = () => {
    setDateFrom("");
    setDateTo("");
    setLogStatus("all");
    setLogModel("");
    setLedgerType("all");
  };

  const sharedFilterBar = (
    <Space wrap style={{ marginBottom: 16 }}>
      <Input type="date" aria-label="开始日期" value={dateFrom} onChange={(event) => setDateFrom(event.target.value)} />
      <Input type="date" aria-label="结束日期" value={dateTo} onChange={(event) => setDateTo(event.target.value)} />
      <Button onClick={resetFilters}>重置筛选</Button>
    </Space>
  );

  const detailKeyColumns: ColumnsType<APIKey> = [
    { title: "名称", dataIndex: "name" },
    { title: "密钥", dataIndex: "mask" },
    { title: "状态", dataIndex: "status" },
    { title: "创建时间", dataIndex: "created_at" },
    {
      title: "操作",
      render: (_, key) => (
        <Space>
          {key.status === "active" && <Button onClick={() => runWrite(async () => { await api.disableAPIKey(key.id); message.success("密钥已停用"); await refreshDetail(); }, "停用密钥失败").catch(() => undefined)}>停用</Button>}
          <Button danger onClick={() => runWrite(async () => { await api.deleteAPIKey(key.id); message.success("密钥已删除"); await refreshDetail(); }, "删除密钥失败").catch(() => undefined)}>删除</Button>
        </Space>
      )
    }
  ];
  const detailLedgerColumns: ColumnsType<LedgerRecord> = [
    { title: "类型", dataIndex: "type" },
    { title: "金额", dataIndex: "amount" },
    { title: "前余额", dataIndex: "balance_before" },
    { title: "后余额", dataIndex: "balance_after" },
    { title: "成本", dataIndex: "base_cost" },
    { title: "倍率", dataIndex: "rate_multiplier" },
    { title: "备注", dataIndex: "remark" },
    { title: "时间", dataIndex: "created_at" }
  ];
  const detailLogColumns: ColumnsType<GatewayLog> = [
    { title: "请求ID", dataIndex: "request_id" },
    { title: "模型ID", dataIndex: "model_id" },
    { title: "状态", dataIndex: "status" },
    { title: "HTTP", dataIndex: "http_status" },
    { title: "Token 数", dataIndex: "total_tokens" },
    { title: "成本", dataIndex: "base_cost" },
    { title: "扣费", dataIndex: "charge" },
    { title: "时间", dataIndex: "created_at" }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Link to="/users">返回用户列表</Link>
      <Card title={`用户详情：${user.username}`}>
        <Space wrap>
          <Card><Typography.Text type="secondary">用户 ID</Typography.Text><Typography.Paragraph copyable>{user.id}</Typography.Paragraph></Card>
          <Card><Typography.Text type="secondary">角色</Typography.Text><Typography.Title level={5}>{user.role}</Typography.Title></Card>
          <Card><Typography.Text type="secondary">状态</Typography.Text><Typography.Title level={5}>{user.status}</Typography.Title></Card>
          <Card><Typography.Text type="secondary">余额</Typography.Text><Typography.Title level={5}>{user.balance}</Typography.Title></Card>
          <Card><Typography.Text type="secondary">API 密钥</Typography.Text><Typography.Title level={5}>{apiKeys.length}</Typography.Title></Card>
          <Card><Typography.Text type="secondary">创建时间</Typography.Text><Typography.Title level={5}>{new Date(user.created_at).toLocaleString()}</Typography.Title></Card>
          <Card><Typography.Text type="secondary">累计消费</Typography.Text><Typography.Title level={5}>{summary?.total_charge ?? "0"}</Typography.Title></Card>
          <Card><Typography.Text type="secondary">累计充值</Typography.Text><Typography.Title level={5}>{summary?.total_recharge ?? "0"}</Typography.Title></Card>
        </Space>
      </Card>
      <Card title="API 密钥"><Table rowKey="id" columns={detailKeyColumns} dataSource={apiKeys} pagination={tablePagination(apiKeysPager, setAPIKeysPager)} /></Card>
      <Card
        title="最近账本"
        extra={<Button onClick={() => exportCSV(`user-${id}-ledger-filtered.csv`, filteredLedger)}>导出当前筛选 CSV</Button>}
      >
        {sharedFilterBar}
        <Space wrap style={{ marginBottom: 16 }}>
          <Select style={{ width: 180 }} value={ledgerType} onChange={setLedgerType} options={[{ value: "all", label: "全部类型" }, ...ledgerTypeOptions]} />
        </Space>
        <Table rowKey={(item) => `${item.type}-${item.created_at}-${item.amount}`} columns={detailLedgerColumns} dataSource={filteredLedger} pagination={tablePagination(ledgerPager, setLedgerPager)} />
      </Card>
      <Card
        title="最近请求"
        extra={<Button onClick={() => exportCSV(`user-${id}-logs-filtered.csv`, filteredLogs)}>导出当前筛选 CSV</Button>}
      >
        {sharedFilterBar}
        <Space wrap style={{ marginBottom: 16 }}>
          <Select style={{ width: 160 }} value={logStatus} onChange={setLogStatus} options={[{ value: "all", label: "全部状态" }, { value: "success", label: "成功" }, { value: "failed", label: "失败" }, { value: "partial", label: "部分成功" }]} />
          <Select allowClear showSearch style={{ width: 240 }} value={logModel || undefined} placeholder="筛选模型" onChange={(value) => setLogModel(value ?? "")} options={modelOptions} />
        </Space>
        <Table rowKey="request_id" columns={detailLogColumns} dataSource={filteredLogs} pagination={tablePagination(logsPager, setLogsPager)} />
      </Card>
    </Space>
  );
}
