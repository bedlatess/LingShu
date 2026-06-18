import React, { Suspense, useEffect, useMemo, useState } from "react";
import ReactDOM from "react-dom/client";
import { AuditOutlined, DashboardOutlined, KeyOutlined, SettingOutlined, TeamOutlined } from "@ant-design/icons";
import { Alert, Button, Card, ConfigProvider, Drawer, Form, Input, Layout, Menu, Modal, Select, Space, Spin, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { BrowserRouter, Link, Navigate, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import {
  createAPI,
  designTokens,
  perKToM,
  type APIKey,
  type AdminDashboard,
  type Announcement,
  type AuditLog,
  type Channel,
  type CleanupHistoryEntry,
  type GatewayLog,
  type LedgerRecord,
  type ModelConfig,
  type RedeemCode,
  type SystemSetting,
  type User
} from "@lingshu/shared";
import { errText, normalizeModelPayload, providerOptions, type Pager, runWrite } from "./pages/admin-page-utils";
import { ModelForm } from "./pages/model-form";
import "antd/dist/reset.css";

const { Header, Sider, Content } = Layout;

const AdminDashboardPage = React.lazy(() => import("./pages/admin-dashboard").then((module) => ({ default: module.AdminDashboardPage })));
const UsersPage = React.lazy(() => import("./pages/users").then((module) => ({ default: module.UsersPage })));
const UserDetailPage = React.lazy(() => import("./pages/users").then((module) => ({ default: module.UserDetailPage })));
const ApiKeysPage = React.lazy(() => import("./pages/api-keys").then((module) => ({ default: module.ApiKeysPage })));
const ModelsPage = React.lazy(() => import("./pages/models").then((module) => ({ default: module.ModelsPage })));
const ModelDetailPage = React.lazy(() => import("./pages/models").then((module) => ({ default: module.ModelDetailPage })));
const ChannelsPage = React.lazy(() => import("./pages/channels").then((module) => ({ default: module.ChannelsPage })));
const ChannelDetailPage = React.lazy(() => import("./pages/channels").then((module) => ({ default: module.ChannelDetailPage })));
const AnnouncementsPage = React.lazy(() => import("./pages/announcements").then((module) => ({ default: module.AnnouncementsPage })));
const RedeemPage = React.lazy(() => import("./pages/redeem").then((module) => ({ default: module.RedeemPage })));
const ReportsPage = React.lazy(() => import("./pages/reports").then((module) => ({ default: module.ReportsPage })));
const SettingsPage = React.lazy(() => import("./pages/settings").then((module) => ({ default: module.SettingsPage })));
const AuditPage = React.lazy(() => import("./pages/audit").then((module) => ({ default: module.AuditPage })));

const adminMenuItems = [
  { key: "/dashboard", icon: <DashboardOutlined />, label: "概览" },
  { key: "/users", icon: <TeamOutlined />, label: "用户管理" },
  { key: "/api-keys", icon: <KeyOutlined />, label: "API 密钥" },
  { key: "/models", icon: <SettingOutlined />, label: "模型管理" },
  { key: "/channels", icon: <SettingOutlined />, label: "渠道管理" },
  { key: "/announcements", icon: <SettingOutlined />, label: "公告管理" },
  { key: "/redeem", icon: <KeyOutlined />, label: "兑换码" },
  { key: "/reports", icon: <DashboardOutlined />, label: "数据报表" },
  { key: "/settings", icon: <SettingOutlined />, label: "系统设置" },
  { key: "/audit", icon: <AuditOutlined />, label: "审计日志" }
];

function App() {
  const [token, setToken] = useState(() => localStorage.getItem("lingshu_admin_token") ?? "");
  const [me, setMe] = useState<User | null>(null);
  const [error, setError] = useState("");
  const [users, setUsers] = useState<User[]>([]);
  const [apiKeys, setAPIKeys] = useState<APIKey[]>([]);
  const [models, setModels] = useState<ModelConfig[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [redeemCodes, setRedeemCodes] = useState<RedeemCode[]>([]);
  const [createdCodes, setCreatedCodes] = useState<string[]>([]);
  const [dashboard, setDashboard] = useState<AdminDashboard | null>(null);
  const [logs, setLogs] = useState<GatewayLog[]>([]);
  const [ledger, setLedger] = useState<LedgerRecord[]>([]);
  const [settings, setSettings] = useState<SystemSetting[]>([]);
  const [cleanupHistory, setCleanupHistory] = useState<CleanupHistoryEntry[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [auditCount, setAuditCount] = useState<number | null>(null);
  const [usersPager, setUsersPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [keysPager, setKeysPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [modelsPager, setModelsPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [channelsPager, setChannelsPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [announcementsPager, setAnnouncementsPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [redeemPager, setRedeemPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [logsPager, setLogsPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [ledgerPager, setLedgerPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [settingsPager, setSettingsPager] = useState<Pager>({ page: 1, limit: 100, total: 0 });
  const [auditPager, setAuditPager] = useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [createdKey, setCreatedKey] = useState("");
  const [balanceTarget, setBalanceTarget] = useState<User | null>(null);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [resetPasswordTarget, setResetPasswordTarget] = useState<User | null>(null);
  const [editingModel, setEditingModel] = useState<ModelConfig | null>(null);
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null);
  const [balanceForm] = Form.useForm<{ amount: string; remark: string }>();
  const [userForm] = Form.useForm<{ username: string; email: string; status: string }>();
  const [passwordForm] = Form.useForm<{ password: string }>();
  const [modelForm] = Form.useForm<Omit<ModelConfig, "id">>();
  const [channelForm] = Form.useForm<Partial<Channel> & { api_key?: string }>();
  const [settingsForm] = Form.useForm<Record<string, string>>();
  const api = useMemo(() => createAPI(token), [token]);

  async function refresh() {
    if (!token) return;
    const [current, userList, audit, keyList, modelList, channelList, announcementList, redeemList, dash, logList, ledgerList, settingList, auditLogList, cleanupList] = await Promise.all([
      api.me(),
      api.listUsers(usersPager.page, usersPager.limit),
      api.auditCount(),
      api.listAPIKeys(keysPager.page, keysPager.limit),
      api.listModels(modelsPager.page, modelsPager.limit),
      api.listChannels(channelsPager.page, channelsPager.limit),
      api.listAnnouncements(announcementsPager.page, announcementsPager.limit),
      api.listRedeemCodes(redeemPager.page, redeemPager.limit),
      api.adminDashboard(),
      api.adminLogs(logsPager.page, logsPager.limit),
      api.adminLedger(ledgerPager.page, ledgerPager.limit),
      api.listSettings(settingsPager.page, settingsPager.limit),
      api.listAuditLogs(auditPager.page, auditPager.limit),
      api.cleanupHistory(10)
    ]);
    setMe(current);
    setUsers(userList.items);
    setAuditCount(audit.count);
    setAPIKeys(keyList.items);
    setModels(modelList.items);
    setChannels(channelList.items);
    setAnnouncements(announcementList.items);
    setRedeemCodes(redeemList.items);
    setDashboard(dash);
    setLogs(logList.items);
    setLedger(ledgerList.items);
    setSettings(settingList.items);
    setCleanupHistory(cleanupList.items);
    setAuditLogs(auditLogList.items);
    setUsersPager((prev) => ({ ...prev, total: userList.total }));
    setKeysPager((prev) => ({ ...prev, total: keyList.total }));
    setModelsPager((prev) => ({ ...prev, total: modelList.total }));
    setChannelsPager((prev) => ({ ...prev, total: channelList.total }));
    setAnnouncementsPager((prev) => ({ ...prev, total: announcementList.total }));
    setRedeemPager((prev) => ({ ...prev, total: redeemList.total }));
    setLogsPager((prev) => ({ ...prev, total: logList.total }));
    setLedgerPager((prev) => ({ ...prev, total: ledgerList.total }));
    setSettingsPager((prev) => ({ ...prev, total: settingList.total }));
    setAuditPager((prev) => ({ ...prev, total: auditLogList.total }));
    settingsForm.setFieldsValue(Object.fromEntries(settingList.items.map((item) => [item.key, item.value])));
  }

  useEffect(() => {
    const onUnauthorized = () => {
      localStorage.removeItem("lingshu_admin_token");
      setToken("");
      setMe(null);
      message.warning("登录已过期，请重新登录");
    };
    window.addEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
    refresh().catch((err) => setError(errText(err)));
    return () => window.removeEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
  }, [token, usersPager.page, usersPager.limit, keysPager.page, keysPager.limit, modelsPager.page, modelsPager.limit, channelsPager.page, channelsPager.limit, announcementsPager.page, announcementsPager.limit, redeemPager.page, redeemPager.limit, logsPager.page, logsPager.limit, ledgerPager.page, ledgerPager.limit, settingsPager.page, settingsPager.limit, auditPager.page, auditPager.limit]);

  async function handleLogin(values: { login: string; password: string }) {
    setError("");
    try {
      const result = await createAPI().login(values.login, values.password);
      if (result.user.role !== "admin") {
        setError("当前账号不是管理员");
        return;
      }
      localStorage.setItem("lingshu_admin_token", result.token);
      setToken(result.token);
      setMe(result.user);
      message.success("登录成功");
    } catch (err) {
      const text = errText(err);
      setError(text);
      message.error(`登录失败: ${text}`);
    }
  }

  async function handleAdjustBalance(values: { amount: string; remark: string }) {
    if (!balanceTarget) return;
    await runWrite(async () => {
      await api.adjustUserBalance(balanceTarget.id, values);
      message.success("余额已调整");
      setBalanceTarget(null);
      balanceForm.resetFields();
      await refresh();
    }, "调整余额失败");
  }

  async function handleUpdateUser(values: { username: string; email: string; status: string }) {
    if (!editingUser) return;
    await runWrite(async () => {
      await api.updateUser(editingUser.id, values);
      message.success("用户信息已更新");
      setEditingUser(null);
      userForm.resetFields();
      await refresh();
    }, "更新用户失败");
  }

  async function handleResetPassword(values: { password: string }) {
    if (!resetPasswordTarget) return;
    await runWrite(async () => {
      await api.resetUserPassword(resetPasswordTarget.id, values.password);
      message.success("密码已重置");
      setResetPasswordTarget(null);
      passwordForm.resetFields();
    }, "重置密码失败");
  }

  async function handleCreateModel(values: Omit<ModelConfig, "id">) {
    await runWrite(async () => {
      await api.createModel(normalizeModelPayload(values));
      message.success("模型已创建");
      await refresh();
    }, "创建模型失败");
  }

  async function handleUpdateModel(values: Omit<ModelConfig, "id">) {
    if (!editingModel) return;
    await runWrite(async () => {
      await api.updateModel(editingModel.id, normalizeModelPayload(values));
      message.success("模型已更新");
      setEditingModel(null);
      modelForm.resetFields();
      await refresh();
    }, "更新模型失败");
  }

  async function handleUpdateChannel(values: Partial<Channel> & { api_key?: string }) {
    if (!editingChannel) return;
    await runWrite(async () => {
      await api.updateChannel(editingChannel.id, {
        name: String(values.name ?? editingChannel.name),
        provider_type: String(values.provider_type ?? editingChannel.provider_type),
        base_url: String(values.base_url ?? editingChannel.base_url),
        api_key: values.api_key,
        status: String(values.status ?? editingChannel.status),
        weight: Number(values.weight ?? editingChannel.weight ?? 1)
      });
      message.success("渠道已更新");
      setEditingChannel(null);
      channelForm.resetFields();
      await refresh();
    }, "更新渠道失败");
  }

  async function handlePatchSettings(values: Record<string, string>) {
    await runWrite(async () => {
      const items = settings.map((item) => ({ key: item.key, value: String(values[item.key] ?? "") }));
      const result = await api.patchSettings(items);
      setSettings(result.items);
      message.success("系统设置已保存");
      await refresh();
    }, "保存设置失败");
  }

  async function handleRunCleanup() {
    const result = await api.runCleanup();
    const history = await api.cleanupHistory(10);
    setCleanupHistory(history.items);
    return result.items;
  }

  async function refreshCleanupHistory() {
    const history = await api.cleanupHistory(10);
    setCleanupHistory(history.items);
  }

  const userColumns: ColumnsType<User> = [
    { title: "用户名", dataIndex: "username", render: (_, user) => <Link to={`/users/${user.id}`}>{user.username}</Link> },
    { title: "邮箱", dataIndex: "email" },
    { title: "角色", dataIndex: "role" },
    { title: "状态", dataIndex: "status" },
    { title: "余额", dataIndex: "balance" },
    {
      title: "操作",
      render: (_, user) => (
        <Space>
          <Button onClick={() => { setEditingUser(user); userForm.setFieldsValue({ username: user.username, email: user.email, status: user.status }); }}>编辑</Button>
          <Button onClick={() => setBalanceTarget(user)}>充值/扣费</Button>
          <Button onClick={() => setResetPasswordTarget(user)}>重置密码</Button>
          {user.status === "active" && (
            <Button danger onClick={() => runWrite(async () => { await api.banUser(user.id); message.success("用户已封禁"); await refresh(); }, "封禁用户失败")}>封禁</Button>
          )}
        </Space>
      )
    }
  ];
  const modelColumns: ColumnsType<ModelConfig> = [
    { title: "模型", dataIndex: "public_name", render: (_, model) => <Link to={`/models/${model.id}`}>{model.public_name}</Link> },
    { title: "类型", dataIndex: "type" },
    { title: "计费", dataIndex: "billing_mode" },
    { title: "输入价/1M", dataIndex: "input_price_per_1k", render: (value) => perKToM(value) },
    { title: "输出价/1M", dataIndex: "output_price_per_1k", render: (value) => perKToM(value) },
    { title: "单次基准", dataIndex: "price_per_call" },
    { title: "倍率", dataIndex: "rate_multiplier" },
    { title: "状态", dataIndex: "status" },
    {
      title: "操作",
      render: (_, model) => (
        <Space>
          <Button
            onClick={() => {
              setEditingModel(model);
              modelForm.setFieldsValue({
                ...model,
                input_price_per_1k: perKToM(model.input_price_per_1k),
                output_price_per_1k: perKToM(model.output_price_per_1k)
              });
            }}
          >
            编辑
          </Button>
          <Button
            danger
            onClick={() => Modal.confirm({
              title: "确认删除模型？",
              content: `模型 ${model.public_name} 将被停用并保留历史记录。`,
              okText: "确认删除",
              cancelText: "取消",
              onOk: () =>
                runWrite(async () => {
                  await api.deleteModel(model.id);
                  message.success("模型已删除");
                  await refresh();
                }, "删除模型失败")
            })}
          >
            删除
          </Button>
        </Space>
      )
    }
  ];
  const channelColumns: ColumnsType<Channel> = [
    { title: "名称", dataIndex: "name", render: (_, channel) => <Link to={`/channels/${channel.id}`}>{channel.name}</Link> },
    { title: "类型", dataIndex: "provider_type" },
    { title: "Base URL", dataIndex: "base_url" },
    { title: "权重", dataIndex: "weight" },
    { title: "已绑模型", dataIndex: "bound_count", render: (_, channel) => <Link to={`/channels/${channel.id}`}>{channel.bound_count ?? 0}</Link> },
    { title: "最近延迟(ms)", dataIndex: "last_latency_ms", render: (value) => value ?? 0 },
    { title: "最近成功", dataIndex: "last_success_at", render: (value) => value ?? "-" },
    { title: "状态", dataIndex: "status" },
    { title: "健康", dataIndex: "health" },
    {
      title: "操作",
      render: (_, channel) => (
        <Space>
          <Button onClick={() => { setEditingChannel(channel); channelForm.setFieldsValue(channel); }}>编辑</Button>
          <Button onClick={() => runWrite(async () => { const result = await api.testChannel(channel.id, channel.base_url); if (result.ok) { message.success(`通过 ${result.latency_ms}ms`); } else { message.error(`${result.category}: ${result.message}`); } }, "测试渠道失败").catch(() => undefined)}>测试</Button>
          <Button
            danger
            onClick={() => Modal.confirm({
              title: "确认删除渠道？",
              content: `渠道 ${channel.name} 将被停用并保留历史记录。`,
              okText: "确认删除",
              cancelText: "取消",
              onOk: () =>
                runWrite(async () => {
                  await api.deleteChannel(channel.id);
                  message.success("渠道已删除");
                  await refresh();
                }, "删除渠道失败")
            })}
          >
            删除
          </Button>
        </Space>
      )
    }
  ];
  const keyColumns: ColumnsType<APIKey> = [
    { title: "名称", dataIndex: "name" },
    { title: "用户ID", dataIndex: "user_id", render: (value) => value ? <Link to={`/users/${value}`}>{value}</Link> : "-" },
    { title: "Key", dataIndex: "mask" },
    { title: "状态", dataIndex: "status" },
    {
      title: "操作",
      render: (_, key) => (
        <Space>
          {key.status === "active" && <Button onClick={() => runWrite(async () => { await api.disableAPIKey(key.id); message.success("密钥已停用"); await refresh(); }, "停用密钥失败").catch(() => undefined)}>停用</Button>}
          <Button
            danger
            onClick={() => Modal.confirm({
              title: "确认删除密钥？",
              content: "删除会停用密钥并保留历史调用记录。",
              okText: "确认删除",
              cancelText: "取消",
              onOk: () =>
                runWrite(async () => {
                  await api.deleteAPIKey(key.id);
                  message.success("密钥已删除");
                  await refresh();
                }, "删除密钥失败")
            })}
          >
            删除
          </Button>
        </Space>
      )
    }
  ];
  const logColumns: ColumnsType<GatewayLog> = [
    { title: "请求ID", dataIndex: "request_id" },
    { title: "用户ID", dataIndex: "user_id", render: (value) => value ? <Link to={`/users/${value}`}>{value}</Link> : "-" },
    { title: "状态", dataIndex: "status" },
    { title: "HTTP", dataIndex: "http_status" },
    { title: "Tokens", dataIndex: "total_tokens" },
    { title: "成本", dataIndex: "base_cost" },
    { title: "扣费", dataIndex: "charge" }
  ];
  const ledgerColumns: ColumnsType<LedgerRecord> = [
    { title: "用户ID", dataIndex: "user_id", render: (value) => value ? <Link to={`/users/${value}`}>{value}</Link> : "-" },
    { title: "类型", dataIndex: "type" },
    { title: "金额", dataIndex: "amount" },
    { title: "前余额", dataIndex: "balance_before" },
    { title: "后余额", dataIndex: "balance_after" },
    { title: "成本", dataIndex: "base_cost" },
    { title: "倍率", dataIndex: "rate_multiplier" },
    { title: "备注", dataIndex: "remark" }
  ];
  const auditColumns: ColumnsType<AuditLog> = [
    { title: "动作", dataIndex: "action" },
    { title: "对象", dataIndex: "target_type" },
    { title: "对象ID", dataIndex: "target_id" },
    { title: "操作者", dataIndex: "actor_id" },
    { title: "IP", dataIndex: "ip" },
    { title: "时间", dataIndex: "created_at" }
  ];

  if (!token || !me) {
    return (
      <Theme>
        <div style={{ minHeight: "100vh", display: "grid", placeItems: "center", background: "#f6f8fb" }}>
          <Card title="LingShu 管理端登录" style={{ width: 380 }}>
            {error && <Alert type="error" message={error} style={{ marginBottom: 16 }} />}
            <Form layout="vertical" onFinish={handleLogin}>
              <Form.Item name="login" label="用户名或邮箱" rules={[{ required: true }]}>
                <Input autoComplete="username" />
              </Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true }]}>
                <Input.Password autoComplete="current-password" />
              </Form.Item>
              <Button type="primary" htmlType="submit" block>
                登录
              </Button>
            </Form>
          </Card>
        </div>
      </Theme>
    );
  }

  return (
    <Theme>
      <Layout style={{ minHeight: "100vh" }}>
        <Sider width={220}>
          <div style={{ color: "white", fontWeight: 700, padding: 20 }}>LingShu Admin</div>
          <AdminMenu />
        </Sider>
        <Layout>
          <Header style={{ background: "#fff", borderBottom: "1px solid #f0f0f0" }}>
            <Space style={{ width: "100%", justifyContent: "space-between" }}>
              <Typography.Text strong>管理工作台</Typography.Text>
              <Button
                onClick={() => {
                  localStorage.removeItem("lingshu_admin_token");
                  setToken("");
                  setMe(null);
                }}
              >
                退出
              </Button>
            </Space>
          </Header>
          <Content style={{ padding: 24 }}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>
              {error && <Alert type="error" message={error} closable onClose={() => setError("")} />}
              <Suspense fallback={<Spin />}>
                <Routes>
                  <Route index element={<Navigate to="/dashboard" replace />} />
                  <Route path="/dashboard" element={<AdminDashboardPage dashboard={dashboard} auditCount={auditCount} me={me} api={api} />} />
                  <Route path="/users" element={<UsersPage users={users} userColumns={userColumns} api={api} refresh={refresh} pager={usersPager} setPager={setUsersPager} />} />
                  <Route path="/users/:id" element={<UserDetailPage api={api} />} />
                  <Route path="/api-keys" element={<ApiKeysPage apiKeys={apiKeys} keyColumns={keyColumns} api={api} refresh={refresh} users={users} createdKey={createdKey} setCreatedKey={setCreatedKey} pager={keysPager} setPager={setKeysPager} />} />
                  <Route path="/models" element={<ModelsPage models={models} modelColumns={modelColumns} onCreate={handleCreateModel} pager={modelsPager} setPager={setModelsPager} />} />
                  <Route path="/models/:id" element={<ModelDetailPage api={api} />} />
                  <Route path="/channels" element={<ChannelsPage channels={channels} channelColumns={channelColumns} api={api} refresh={refresh} models={models} pager={channelsPager} setPager={setChannelsPager} />} />
                  <Route path="/channels/:id" element={<ChannelDetailPage api={api} />} />
                  <Route path="/announcements" element={<AnnouncementsPage announcements={announcements} api={api} refresh={refresh} pager={announcementsPager} setPager={setAnnouncementsPager} />} />
                  <Route path="/redeem" element={<RedeemPage redeemCodes={redeemCodes} api={api} refresh={refresh} createdCodes={createdCodes} setCreatedCodes={setCreatedCodes} pager={redeemPager} setPager={setRedeemPager} />} />
                  <Route path="/reports" element={<ReportsPage api={api} dashboard={dashboard} logs={logs} ledger={ledger} logColumns={logColumns} ledgerColumns={ledgerColumns} logsPager={logsPager} setLogsPager={setLogsPager} ledgerPager={ledgerPager} setLedgerPager={setLedgerPager} />} />
                  <Route path="/settings" element={<SettingsPage settings={settings} form={settingsForm} onSave={handlePatchSettings} cleanupHistory={cleanupHistory} onRunCleanup={handleRunCleanup} onRefreshCleanupHistory={refreshCleanupHistory} />} />
                  <Route path="/audit" element={<AuditPage api={api} auditColumns={auditColumns} initialLogs={auditLogs} pager={auditPager} setPager={setAuditPager} />} />
                  <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
              </Suspense>
            </Space>
          </Content>
        </Layout>
      </Layout>
      <Modal title={balanceTarget ? `调整余额：${balanceTarget.username}` : "调整余额"} open={Boolean(balanceTarget)} onCancel={() => setBalanceTarget(null)} onOk={() => balanceForm.submit()} destroyOnClose>
        <Form form={balanceForm} layout="vertical" onFinish={handleAdjustBalance}>
          <Form.Item label="金额" name="amount" rules={[{ required: true, message: "请输入金额，扣费使用负数" }]}>
            <Input placeholder="例如 100 或 -10" />
          </Form.Item>
          <Form.Item label="备注" name="remark" rules={[{ required: true, message: "备注必填" }]}>
            <Input placeholder="手动充值/运营扣费原因" />
          </Form.Item>
        </Form>
      </Modal>
      <Modal title={editingModel ? `编辑模型：${editingModel.public_name}` : "编辑模型"} open={Boolean(editingModel)} onCancel={() => setEditingModel(null)} onOk={() => modelForm.submit()} destroyOnClose width={760}>
        <ModelForm form={modelForm} onFinish={handleUpdateModel} hideSubmit />
      </Modal>
      <Modal title={editingUser ? `编辑用户：${editingUser.username}` : "编辑用户"} open={Boolean(editingUser)} onCancel={() => setEditingUser(null)} onOk={() => userForm.submit()} destroyOnClose>
        <Form form={userForm} layout="vertical" onFinish={handleUpdateUser}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="email" label="邮箱"><Input /></Form.Item>
          <Form.Item name="status" label="状态"><Select options={[{ value: "active", label: "启用" }, { value: "banned", label: "封禁" }]} /></Form.Item>
        </Form>
      </Modal>
      <Modal title={resetPasswordTarget ? `重置密码：${resetPasswordTarget.username}` : "重置密码"} open={Boolean(resetPasswordTarget)} onCancel={() => setResetPasswordTarget(null)} onOk={() => passwordForm.submit()} destroyOnClose>
        <Form form={passwordForm} layout="vertical" onFinish={handleResetPassword}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 8 }]}><Input.Password /></Form.Item>
        </Form>
      </Modal>
      <Drawer
        title={editingChannel ? `编辑渠道：${editingChannel.name}` : "编辑渠道"}
        open={Boolean(editingChannel)}
        onClose={() => setEditingChannel(null)}
        destroyOnHidden
        width={520}
        extra={<Button type="primary" onClick={() => channelForm.submit()}>保存</Button>}
      >
        <Form form={channelForm} layout="vertical" onFinish={handleUpdateChannel}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="provider_type" label="供应商"><Select options={providerOptions} /></Form.Item>
          <Form.Item name="base_url" label="上游地址" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="api_key" label="新密钥"><Input.Password placeholder="留空则不修改" /></Form.Item>
          <Form.Item name="status" label="状态"><Select options={[{ value: "enabled", label: "启用" }, { value: "disabled", label: "停用" }]} /></Form.Item>
          <Form.Item name="weight" label="权重"><Input /></Form.Item>
        </Form>
      </Drawer>
    </Theme>
  );
}

function AdminMenu() {
  const navigate = useNavigate();
  const location = useLocation();
  const selectedKey = location.pathname.startsWith("/users/") ? "/users" : location.pathname === "/" ? "/dashboard" : location.pathname;
  return <Menu theme="dark" mode="inline" selectedKeys={[selectedKey]} items={adminMenuItems} onClick={({ key }) => navigate(key)} />;
}


function Theme({ children }: { children: React.ReactNode }) {
  return (
    <ConfigProvider theme={{ token: { colorPrimary: designTokens.colors.brand, borderRadius: 8 } }}>
      {children}
    </ConfigProvider>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>
);

