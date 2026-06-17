import React, { useEffect, useMemo, useState } from "react";
import ReactDOM from "react-dom/client";
import { AuditOutlined, DashboardOutlined, KeyOutlined, SettingOutlined, TeamOutlined } from "@ant-design/icons";
import { Alert, Button, Card, ConfigProvider, Form, Input, Layout, Modal, Select, Space, Table, Tabs, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { FormInstance } from "antd";
import {
  createAPI,
  designTokens,
  type APIKey,
  type AdminDashboard,
  type Announcement,
  type AuditLog,
  type Channel,
  type GatewayLog,
  type LedgerRecord,
  type ModelConfig,
  type RedeemCode,
  type SystemSetting,
  type User
} from "@lingshu/shared";
import "antd/dist/reset.css";

const { Header, Sider, Content } = Layout;

const modelDefaults = {
  public_name: "",
  type: "chat",
  group: "",
  billing_mode: "token",
  input_price_per_1k: "0",
  output_price_per_1k: "0",
  price_per_call: "0",
  rate_multiplier: "1.200",
  status: "enabled",
  sort_order: 0
};

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
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [auditCount, setAuditCount] = useState<number | null>(null);
  const [createdKey, setCreatedKey] = useState("");
  const [balanceTarget, setBalanceTarget] = useState<User | null>(null);
  const [editingModel, setEditingModel] = useState<ModelConfig | null>(null);
  const [balanceForm] = Form.useForm<{ amount: string; remark: string }>();
  const [modelForm] = Form.useForm<Omit<ModelConfig, "id">>();
  const [settingsForm] = Form.useForm<Record<string, string>>();
  const api = useMemo(() => createAPI(token), [token]);

  async function refresh() {
    if (!token) return;
    const [current, userList, audit, keyList, modelList, channelList, announcementList, redeemList, dash, logList, ledgerList, settingList, auditLogList] = await Promise.all([
      api.me(),
      api.listUsers(),
      api.auditCount(),
      api.listAPIKeys(),
      api.listModels(),
      api.listChannels(),
      api.listAnnouncements(),
      api.listRedeemCodes(),
      api.adminDashboard(),
      api.adminLogs(),
      api.adminLedger(),
      api.listSettings(),
      api.listAuditLogs()
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
    setAuditLogs(auditLogList.items);
    settingsForm.setFieldsValue(Object.fromEntries(settingList.items.map((item) => [item.key, item.value])));
  }

  useEffect(() => {
    refresh().catch((err) => setError(err.message));
  }, [token]);

  async function handleLogin(values: { login: string; password: string }) {
    setError("");
    const result = await createAPI().login(values.login, values.password);
    if (result.user.role !== "admin") {
      setError("当前账号不是管理员");
      return;
    }
    localStorage.setItem("lingshu_admin_token", result.token);
    setToken(result.token);
    setMe(result.user);
  }

  async function handleAdjustBalance(values: { amount: string; remark: string }) {
    if (!balanceTarget) return;
    await api.adjustUserBalance(balanceTarget.id, values);
    setBalanceTarget(null);
    balanceForm.resetFields();
    await refresh();
  }

  async function handleCreateModel(values: Omit<ModelConfig, "id">) {
    await api.createModel(normalizeModelPayload(values));
    await refresh();
  }

  async function handleUpdateModel(values: Omit<ModelConfig, "id">) {
    if (!editingModel) return;
    await api.updateModel(editingModel.id, normalizeModelPayload(values));
    setEditingModel(null);
    modelForm.resetFields();
    await refresh();
  }

  async function handlePatchSettings(values: Record<string, string>) {
    const items = settings.map((item) => ({ key: item.key, value: String(values[item.key] ?? "") }));
    const result = await api.patchSettings(items);
    setSettings(result.items);
    await refresh();
  }

  const userColumns: ColumnsType<User> = [
    { title: "用户名", dataIndex: "username" },
    { title: "邮箱", dataIndex: "email" },
    { title: "角色", dataIndex: "role" },
    { title: "状态", dataIndex: "status" },
    { title: "余额", dataIndex: "balance" },
    {
      title: "操作",
      render: (_, user) => <Button onClick={() => setBalanceTarget(user)}>充值/扣费</Button>
    }
  ];
  const modelColumns: ColumnsType<ModelConfig> = [
    { title: "模型", dataIndex: "public_name" },
    { title: "类型", dataIndex: "type" },
    { title: "计费", dataIndex: "billing_mode" },
    { title: "输入基准/1K", dataIndex: "input_price_per_1k" },
    { title: "输出基准/1K", dataIndex: "output_price_per_1k" },
    { title: "单次基准", dataIndex: "price_per_call" },
    { title: "倍率", dataIndex: "rate_multiplier" },
    { title: "状态", dataIndex: "status" },
    {
      title: "操作",
      render: (_, model) => (
        <Button
          onClick={() => {
            setEditingModel(model);
            modelForm.setFieldsValue(model);
          }}
        >
          编辑
        </Button>
      )
    }
  ];
  const channelColumns: ColumnsType<Channel> = [
    { title: "名称", dataIndex: "name" },
    { title: "类型", dataIndex: "provider_type" },
    { title: "Base URL", dataIndex: "base_url" },
    { title: "权重", dataIndex: "weight" },
    { title: "状态", dataIndex: "status" },
    { title: "健康", dataIndex: "health" }
  ];
  const keyColumns: ColumnsType<APIKey> = [
    { title: "名称", dataIndex: "name" },
    { title: "用户ID", dataIndex: "user_id" },
    { title: "Key", dataIndex: "mask" },
    { title: "状态", dataIndex: "status" }
  ];
  const logColumns: ColumnsType<GatewayLog> = [
    { title: "请求ID", dataIndex: "request_id" },
    { title: "状态", dataIndex: "status" },
    { title: "HTTP", dataIndex: "http_status" },
    { title: "Tokens", dataIndex: "total_tokens" },
    { title: "成本", dataIndex: "base_cost" },
    { title: "扣费", dataIndex: "charge" }
  ];
  const ledgerColumns: ColumnsType<LedgerRecord> = [
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
          <MenuPlaceholder />
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
              <Summary dashboard={dashboard} auditCount={auditCount} me={me} />
              <Tabs
                items={[
                  { key: "users", label: "用户", children: <UsersPane users={users} userColumns={userColumns} api={api} refresh={refresh} /> },
                  { key: "keys", label: "API Key", children: <KeysPane apiKeys={apiKeys} keyColumns={keyColumns} api={api} refresh={refresh} users={users} createdKey={createdKey} setCreatedKey={setCreatedKey} /> },
                  { key: "models", label: "模型", children: <ModelsPane models={models} modelColumns={modelColumns} onCreate={handleCreateModel} /> },
                  { key: "channels", label: "渠道", children: <ChannelsPane channels={channels} channelColumns={channelColumns} api={api} refresh={refresh} models={models} /> },
                  { key: "announcements", label: "公告", children: <AnnouncementsPane announcements={announcements} api={api} refresh={refresh} /> },
                  { key: "redeem", label: "兑换码", children: <RedeemPane redeemCodes={redeemCodes} api={api} refresh={refresh} createdCodes={createdCodes} setCreatedCodes={setCreatedCodes} /> },
                  { key: "reports", label: "报表", children: <ReportsPane dashboard={dashboard} logs={logs} ledger={ledger} logColumns={logColumns} ledgerColumns={ledgerColumns} /> },
                  { key: "settings", label: "系统设置", children: <SettingsPane settings={settings} form={settingsForm} onSave={handlePatchSettings} /> },
                  { key: "audit", label: "审计日志", children: <Card title="审计日志"><Table rowKey="id" columns={auditColumns} dataSource={auditLogs} /></Card> }
                ]}
              />
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
        <ModelForm form={modelForm} onFinish={handleUpdateModel} />
      </Modal>
    </Theme>
  );
}

function MenuPlaceholder() {
  return (
    <div style={{ color: "rgba(255,255,255,.72)", padding: "0 16px", display: "grid", gap: 12 }}>
      <Space><DashboardOutlined />Dashboard</Space>
      <Space><TeamOutlined />Users</Space>
      <Space><KeyOutlined />Keys</Space>
      <Space><SettingOutlined />Settings</Space>
      <Space><AuditOutlined />Audit</Space>
    </div>
  );
}

function Summary({ dashboard, auditCount, me }: { dashboard: AdminDashboard | null; auditCount: number | null; me: User }) {
  return (
    <Space wrap>
      <Card><Typography.Text type="secondary">管理员</Typography.Text><Typography.Title level={4}>{me.username}</Typography.Title></Card>
      <Card><Typography.Text type="secondary">今日请求</Typography.Text><Typography.Title level={4}>{dashboard?.today_requests ?? 0}</Typography.Title></Card>
      <Card><Typography.Text type="secondary">今日扣费</Typography.Text><Typography.Title level={4}>{dashboard?.today_charge ?? "0"}</Typography.Title></Card>
      <Card><Typography.Text type="secondary">毛利</Typography.Text><Typography.Title level={4}>{dashboard?.gross_profit ?? "0"}</Typography.Title></Card>
      <Card><Typography.Text type="secondary">审计日志</Typography.Text><Typography.Title level={4}>{auditCount ?? "--"}</Typography.Title></Card>
    </Space>
  );
}

function UsersPane({ users, userColumns, api, refresh }: { users: User[]; userColumns: ColumnsType<User>; api: ReturnType<typeof createAPI>; refresh: () => Promise<void> }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建用户">
        <Form layout="inline" onFinish={async (values) => { await api.createUser(values); await refresh(); }}>
          <Form.Item name="username" rules={[{ required: true }]}><Input placeholder="用户名" /></Form.Item>
          <Form.Item name="email"><Input placeholder="邮箱" /></Form.Item>
          <Form.Item name="password" rules={[{ required: true, min: 8 }]}><Input.Password placeholder="初始密码" /></Form.Item>
          <Form.Item name="role" initialValue="user"><Select style={{ width: 120 }} options={[{ value: "user" }, { value: "admin" }]} /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="用户列表"><Table rowKey="id" columns={userColumns} dataSource={users} /></Card>
    </Space>
  );
}

function KeysPane({ apiKeys, keyColumns, api, refresh, users, createdKey, setCreatedKey }: { apiKeys: APIKey[]; keyColumns: ColumnsType<APIKey>; api: ReturnType<typeof createAPI>; refresh: () => Promise<void>; users: User[]; createdKey: string; setCreatedKey: (key: string) => void }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {createdKey && <Alert type="success" message={`新 Key 仅显示一次：${createdKey}`} />}
      <Card title="创建 API Key">
        <Form layout="inline" onFinish={async (values) => { const result = await api.createAPIKey(values); setCreatedKey(result.plaintext); await refresh(); }}>
          <Form.Item name="user_id" rules={[{ required: true }]}>
            <Select showSearch style={{ width: 280 }} placeholder="选择用户" options={users.map((user) => ({ value: user.id, label: `${user.username} (${user.id})` }))} />
          </Form.Item>
          <Form.Item name="name" rules={[{ required: true }]}><Input placeholder="名称" /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="Key 列表"><Table rowKey="id" columns={keyColumns} dataSource={apiKeys} /></Card>
    </Space>
  );
}

function ModelsPane({ models, modelColumns, onCreate }: { models: ModelConfig[]; modelColumns: ColumnsType<ModelConfig>; onCreate: (values: Omit<ModelConfig, "id">) => Promise<void> }) {
  const [form] = Form.useForm<Omit<ModelConfig, "id">>();
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建模型"><ModelForm form={form} onFinish={async (values) => { await onCreate(values); form.resetFields(); }} /></Card>
      <Card title="模型列表"><Table rowKey="id" columns={modelColumns} dataSource={models} /></Card>
    </Space>
  );
}

function ModelForm({ form, onFinish }: { form: FormInstance<Omit<ModelConfig, "id">>; onFinish: (values: Omit<ModelConfig, "id">) => Promise<void> }) {
  return (
    <Form form={form} layout="vertical" onFinish={onFinish} initialValues={modelDefaults}>
      <Space wrap align="start">
        <Form.Item name="public_name" label="模型名" rules={[{ required: true }]}><Input style={{ width: 220 }} /></Form.Item>
        <Form.Item name="type" label="类型"><Select style={{ width: 140 }} options={["chat", "embedding", "image", "video"].map((value) => ({ value }))} /></Form.Item>
        <Form.Item name="billing_mode" label="计费"><Select style={{ width: 140 }} options={["token", "per_call"].map((value) => ({ value }))} /></Form.Item>
        <Form.Item name="group" label="分组"><Input style={{ width: 140 }} /></Form.Item>
        <Form.Item name="input_price_per_1k" label="输入基准/1K"><Input style={{ width: 140 }} /></Form.Item>
        <Form.Item name="output_price_per_1k" label="输出基准/1K"><Input style={{ width: 140 }} /></Form.Item>
        <Form.Item name="price_per_call" label="单次基准"><Input style={{ width: 140 }} /></Form.Item>
        <Form.Item name="rate_multiplier" label="倍率" rules={[{ required: true }]}><Input style={{ width: 120 }} /></Form.Item>
        <Form.Item name="status" label="状态"><Select style={{ width: 120 }} options={["enabled", "disabled"].map((value) => ({ value }))} /></Form.Item>
        <Form.Item name="sort_order" label="排序"><Input style={{ width: 100 }} /></Form.Item>
      </Space>
      <Button type="primary" htmlType="submit">保存</Button>
    </Form>
  );
}

function ChannelsPane({ channels, channelColumns, api, refresh, models }: { channels: Channel[]; channelColumns: ColumnsType<Channel>; api: ReturnType<typeof createAPI>; refresh: () => Promise<void>; models: ModelConfig[] }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建 OpenAI 兼容渠道">
        <Form layout="inline" onFinish={async (values) => { await api.createChannel({ ...values, weight: Number(values.weight ?? 1), timeout_seconds: 120, rpm_limit: 60, concurrency_limit: 5, fail_threshold: 5 }); await refresh(); }} initialValues={{ provider_type: "openai", status: "enabled", weight: 1 }}>
          <Form.Item name="name" rules={[{ required: true }]}><Input placeholder="名称" /></Form.Item>
          <Form.Item name="base_url" rules={[{ required: true }]}><Input placeholder="Base URL" style={{ width: 260 }} /></Form.Item>
          <Form.Item name="api_key" rules={[{ required: true }]}><Input.Password placeholder="上游 API Key" /></Form.Item>
          <Form.Item name="weight"><Input placeholder="权重" style={{ width: 90 }} /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="渠道列表"><Table rowKey="id" columns={channelColumns} dataSource={channels} /></Card>
      <Card title="绑定渠道模型">
        <Form layout="inline" onFinish={async (values) => { await api.bindChannelModel(values); await refresh(); }}>
          <Form.Item name="channel_id" rules={[{ required: true }]}><Select style={{ width: 260 }} options={channels.map((channel) => ({ value: channel.id, label: channel.name }))} /></Form.Item>
          <Form.Item name="model_id" rules={[{ required: true }]}><Select style={{ width: 260 }} options={models.map((model) => ({ value: model.id, label: model.public_name }))} /></Form.Item>
          <Form.Item name="upstream_model_name" rules={[{ required: true }]}><Input placeholder="上游模型名" /></Form.Item>
          <Button type="primary" htmlType="submit">绑定</Button>
        </Form>
      </Card>
    </Space>
  );
}

function AnnouncementsPane({ announcements, api, refresh }: { announcements: Announcement[]; api: ReturnType<typeof createAPI>; refresh: () => Promise<void> }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="发布公告">
        <Form layout="inline" onFinish={async (values) => { await api.createAnnouncement({ ...values, priority: Number(values.priority ?? 0), pinned: Boolean(values.pinned) }); await refresh(); }} initialValues={{ status: "online", priority: 0, pinned: false }}>
          <Form.Item name="title" rules={[{ required: true }]}><Input placeholder="标题" /></Form.Item>
          <Form.Item name="content" rules={[{ required: true }]}><Input placeholder="内容" style={{ width: 360 }} /></Form.Item>
          <Form.Item name="status"><Select style={{ width: 120 }} options={[{ value: "online" }, { value: "offline" }]} /></Form.Item>
          <Button type="primary" htmlType="submit">发布</Button>
        </Form>
      </Card>
      <Card title="公告列表"><Table rowKey="id" dataSource={announcements} columns={[{ title: "标题", dataIndex: "title" }, { title: "状态", dataIndex: "status" }, { title: "优先级", dataIndex: "priority" }, { title: "置顶", dataIndex: "pinned", render: (value) => (value ? "是" : "否") }]} /></Card>
    </Space>
  );
}

function RedeemPane({ redeemCodes, api, refresh, createdCodes, setCreatedCodes }: { redeemCodes: RedeemCode[]; api: ReturnType<typeof createAPI>; refresh: () => Promise<void>; createdCodes: string[]; setCreatedCodes: (codes: string[]) => void }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {createdCodes.length > 0 && <Alert type="success" message={`新兑换码仅显示一次：${createdCodes.join(", ")}`} />}
      <Card title="生成兑换码">
        <Form layout="inline" onFinish={async (values) => { const result = await api.createRedeemCodes({ ...values, count: Number(values.count ?? 1), max_uses: Number(values.max_uses ?? 1) }); setCreatedCodes(result.items.map((item) => item.code ?? "").filter(Boolean)); await refresh(); }} initialValues={{ count: 1, max_uses: 1 }}>
          <Form.Item name="amount" rules={[{ required: true }]}><Input placeholder="面额" /></Form.Item>
          <Form.Item name="count"><Input placeholder="数量" /></Form.Item>
          <Form.Item name="batch_name"><Input placeholder="批次" /></Form.Item>
          <Button type="primary" htmlType="submit">生成</Button>
        </Form>
      </Card>
      <Card title="兑换码列表"><Table rowKey="id" dataSource={redeemCodes} columns={[{ title: "前缀", dataIndex: "code_prefix" }, { title: "批次", dataIndex: "batch_name" }, { title: "面额", dataIndex: "amount" }, { title: "状态", dataIndex: "status" }, { title: "使用", render: (_, item) => `${item.used_count}/${item.max_uses}` }]} /></Card>
    </Space>
  );
}

function ReportsPane({ dashboard, logs, ledger, logColumns, ledgerColumns }: { dashboard: AdminDashboard | null; logs: GatewayLog[]; ledger: LedgerRecord[]; logColumns: ColumnsType<GatewayLog>; ledgerColumns: ColumnsType<LedgerRecord> }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Space wrap>
        <Card><Typography.Text type="secondary">活跃用户</Typography.Text><Typography.Title level={4}>{dashboard?.active_users ?? 0}</Typography.Title></Card>
        <Card><Typography.Text type="secondary">余额池</Typography.Text><Typography.Title level={4}>{dashboard?.balance_total ?? "0"}</Typography.Title></Card>
        <Card><Typography.Text type="secondary">成本</Typography.Text><Typography.Title level={4}>{dashboard?.today_base_cost ?? "0"}</Typography.Title></Card>
      </Space>
      <Card title="全站调用日志"><Table rowKey="request_id" columns={logColumns} dataSource={logs} /></Card>
      <Card title="全站账本"><Table rowKey={(item) => `${item.type}-${item.created_at}-${item.amount}`} columns={ledgerColumns} dataSource={ledger} /></Card>
    </Space>
  );
}

function SettingsPane({ settings, form, onSave }: { settings: SystemSetting[]; form: FormInstance<Record<string, string>>; onSave: (values: Record<string, string>) => Promise<void> }) {
  return (
    <Card title="系统设置">
      <Form form={form} layout="vertical" onFinish={onSave}>
        {settings.map((item) => (
          <Form.Item key={item.key} name={item.key} label={`${item.key} - ${item.description}`}>
            <Input />
          </Form.Item>
        ))}
        <Button type="primary" htmlType="submit">保存设置</Button>
      </Form>
    </Card>
  );
}

function normalizeModelPayload(values: Omit<ModelConfig, "id">): Omit<ModelConfig, "id"> {
  return {
    ...modelDefaults,
    ...values,
    sort_order: Number(values.sort_order ?? 0)
  };
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
    <App />
  </React.StrictMode>
);
