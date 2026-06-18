import { useEffect, useMemo, useState } from "react";
import { Alert, Button, Card, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { Link, useParams } from "react-router-dom";
import {
  createAPI,
  perKToM,
  perMToK,
  type Channel,
  type ChannelDetail,
  type ChannelModelImportInput,
  type ModelConfig,
  type ProviderModel
} from "@lingshu/shared";

import { errText, metricCards, providerOptions, type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

type SyncRow = ProviderModel & {
  key: string;
  public_name: string;
  billing_mode: string;
  input_price_per_m: string;
  output_price_per_m: string;
  price_per_call: string;
  rate_multiplier: string;
  state: "new" | "exists" | "bound";
};

export function ChannelsPage({
  channels,
  channelColumns,
  api,
  refresh,
  pager,
  setPager
}: {
  channels: Channel[];
  channelColumns: ColumnsType<Channel>;
  api: AdminAPI;
  refresh: () => Promise<void>;
  models: ModelConfig[];
  pager: Pager;
  setPager: React.Dispatch<React.SetStateAction<Pager>>;
}) {
  const [form] = Form.useForm();
  const [keyword, setKeyword] = useState("");
  const [provider, setProvider] = useState<string>();
  const [status, setStatus] = useState<string>();
  const [health, setHealth] = useState<string>();

  const filtered = useMemo(() => {
    const q = keyword.trim().toLowerCase();
    return channels.filter((channel) => {
      if (q && !channel.name.toLowerCase().includes(q)) return false;
      if (provider && channel.provider_type !== provider) return false;
      if (status && channel.status !== status) return false;
      if (health && channel.health !== health) return false;
      return true;
    });
  }, [channels, health, keyword, provider, status]);

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建上游渠道">
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) =>
            runWrite(async () => {
              await api.createChannel({
                ...values,
                weight: Number(values.weight ?? 1),
                timeout_seconds: Number(values.timeout_seconds ?? 120),
                rpm_limit: Number(values.rpm_limit ?? 60),
                concurrency_limit: Number(values.concurrency_limit ?? 5),
                fail_threshold: Number(values.fail_threshold ?? 5)
              });
              message.success("渠道已创建");
              form.resetFields();
              await refresh();
            }, "创建渠道失败").catch(() => undefined)
          }
          initialValues={{ provider_type: "openai", status: "enabled", weight: 1, timeout_seconds: 120, rpm_limit: 60, concurrency_limit: 5, fail_threshold: 5 }}
        >
          <Space wrap align="start">
            <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input style={{ width: 220 }} /></Form.Item>
            <Form.Item name="provider_type" label="供应商" rules={[{ required: true }]}><Select style={{ width: 180 }} options={providerOptions} /></Form.Item>
            <Form.Item name="base_url" label="Base URL" rules={[{ required: true }]}><Input style={{ width: 320 }} /></Form.Item>
            <Form.Item name="api_key" label="上游 API Key" rules={[{ required: true }]}><Input.Password style={{ width: 260 }} /></Form.Item>
            <Form.Item name="weight" label="权重"><InputNumber min={1} style={{ width: 110 }} /></Form.Item>
            <Form.Item name="timeout_seconds" label="超时(秒)"><InputNumber min={1} style={{ width: 110 }} /></Form.Item>
            <Form.Item name="rpm_limit" label="RPM"><InputNumber min={1} style={{ width: 110 }} /></Form.Item>
            <Form.Item name="concurrency_limit" label="并发"><InputNumber min={1} style={{ width: 110 }} /></Form.Item>
            <Form.Item name="fail_threshold" label="失败阈值"><InputNumber min={1} style={{ width: 110 }} /></Form.Item>
            <Form.Item name="status" label="状态"><Select style={{ width: 120 }} options={[{ value: "enabled", label: "启用" }, { value: "disabled", label: "停用" }]} /></Form.Item>
          </Space>
          <Button type="primary" htmlType="submit">创建渠道</Button>
        </Form>
      </Card>

      <Card title="渠道列表">
        <Space wrap style={{ marginBottom: 12 }}>
          <Input.Search placeholder="搜索渠道名称" allowClear onSearch={setKeyword} onChange={(event) => setKeyword(event.target.value)} style={{ width: 240 }} />
          <Select allowClear placeholder="供应商" style={{ width: 180 }} options={providerOptions} value={provider} onChange={setProvider} />
          <Select allowClear placeholder="状态" style={{ width: 140 }} options={[{ value: "enabled", label: "启用" }, { value: "disabled", label: "停用" }]} value={status} onChange={setStatus} />
          <Select allowClear placeholder="健康" style={{ width: 140 }} options={[{ value: "healthy", label: "健康" }, { value: "unhealthy", label: "异常" }]} value={health} onChange={setHealth} />
        </Space>
        <Table rowKey="id" columns={channelColumns} dataSource={filtered} pagination={tablePagination({ ...pager, total: filtered.length || pager.total }, setPager)} />
      </Card>
    </Space>
  );
}

export function ChannelDetailPage({ api }: { api: AdminAPI }) {
  const { id } = useParams();
  const [detail, setDetail] = useState<ChannelDetail | null>(null);
  const [error, setError] = useState("");
  const [syncOpen, setSyncOpen] = useState(false);
  const [syncLoading, setSyncLoading] = useState(false);
  const [importing, setImporting] = useState(false);
  const [syncRows, setSyncRows] = useState<SyncRow[]>([]);
  const [selectedSyncKeys, setSelectedSyncKeys] = useState<React.Key[]>([]);
  const [selectedBindingKeys, setSelectedBindingKeys] = useState<React.Key[]>([]);
  const [bulkRate, setBulkRate] = useState("1.200");
  const [bulkInput, setBulkInput] = useState("0");
  const [bulkOutput, setBulkOutput] = useState("0");

  async function loadDetail() {
    if (!id) return;
    setError("");
    try {
      setDetail(await api.getChannelDetail(id));
    } catch (err) {
      setError(errText(err));
    }
  }

  useEffect(() => {
    loadDetail();
  }, [id]);

  async function openSync() {
    if (!id || !detail) return;
    setSyncOpen(true);
    setSyncLoading(true);
    try {
      const result = await api.syncChannelModels(id);
      const boundNames = new Set(result.existing_bindings.map((item) => item.upstream_model_name));
      const existingPublic = new Set(result.existing_bindings.map((item) => item.model_name));
      const rows = result.upstream_models.map((model) => {
        const state: SyncRow["state"] = boundNames.has(model.id) ? "bound" : existingPublic.has(model.id) ? "exists" : "new";
        return {
          ...model,
          key: model.id,
          public_name: model.id,
          billing_mode: model.type === "image" ? "per_call" : "token",
          input_price_per_m: "0",
          output_price_per_m: "0",
          price_per_call: "0",
          rate_multiplier: "1.200",
          state
        };
      });
      setSyncRows(rows);
      setSelectedSyncKeys(rows.filter((row) => row.state !== "bound").slice(0, 5).map((row) => row.key));
    } catch (err) {
      message.error(`同步失败: ${errText(err)}`);
    } finally {
      setSyncLoading(false);
    }
  }

  function patchSyncRow(key: string, patch: Partial<SyncRow>) {
    setSyncRows((rows) => rows.map((row) => (row.key === key ? { ...row, ...patch } : row)));
  }

  function applyBulk() {
    setSyncRows((rows) =>
      rows.map((row) =>
        selectedSyncKeys.includes(row.key)
          ? { ...row, rate_multiplier: bulkRate, input_price_per_m: bulkInput, output_price_per_m: bulkOutput }
          : row
      )
    );
  }

  async function importSelected() {
    if (!id) return;
    const rows = syncRows.filter((row) => selectedSyncKeys.includes(row.key));
    if (rows.length === 0) {
      message.info("请选择要导入的模型");
      return;
    }
    setImporting(true);
    try {
      const payload: ChannelModelImportInput[] = rows.map((row) => ({
        upstream_name: row.id,
        public_name: row.public_name || row.id,
        type: row.type || "chat",
        billing_mode: row.billing_mode,
        input_price_per_1k: perMToK(row.input_price_per_m),
        output_price_per_1k: perMToK(row.output_price_per_m),
        price_per_call: row.price_per_call,
        rate_multiplier: row.rate_multiplier,
        status: "enabled"
      }));
      const result = await api.importChannelModels(id, { strategy: "create_or_bind", models: payload });
      message.success(`导入完成：${result.items.length} 个模型已创建或绑定`);
      setSyncOpen(false);
      await loadDetail();
    } catch (err) {
      message.error(`导入失败: ${errText(err)}`);
    } finally {
      setImporting(false);
    }
  }

  async function unbindSelected(keys: React.Key[]) {
    if (!id || keys.length === 0) return;
    await runWrite(async () => {
      for (const modelID of keys) {
        await api.unbindChannelModel(id, String(modelID));
      }
      message.success("已批量解绑");
      await loadDetail();
    }, "批量解绑失败").catch(() => undefined);
  }

  if (!id) return <Alert type="error" message="缺少渠道 ID" />;
  if (error) return <Alert type="error" message={error} />;
  if (!detail) return <Card title="渠道详情">加载中...</Card>;

  const columns: ColumnsType<ChannelDetail["models"][number]> = [
    { title: "模型", dataIndex: "model_name", render: (_, item) => <Link to={`/models/${item.model_id}`}>{item.model_name}</Link> },
    { title: "上游模型名", dataIndex: "upstream_model_name" },
    { title: "状态", dataIndex: "status" },
    { title: "绑定时间", dataIndex: "created_at" }
  ];
  const syncColumns: ColumnsType<SyncRow> = [
    { title: "上游模型名", dataIndex: "id", width: 230 },
    { title: "public_name", dataIndex: "public_name", render: (_, row) => <Input value={row.public_name} onChange={(event) => patchSyncRow(row.key, { public_name: event.target.value })} /> },
    { title: "类型", dataIndex: "type", render: (_, row) => <Select value={row.type} onChange={(value) => patchSyncRow(row.key, { type: value })} options={["chat", "embedding", "image", "video"].map((value) => ({ value }))} style={{ width: 120 }} /> },
    { title: "计费", dataIndex: "billing_mode", render: (_, row) => <Select value={row.billing_mode} onChange={(value) => patchSyncRow(row.key, { billing_mode: value })} options={["token", "per_call"].map((value) => ({ value }))} style={{ width: 110 }} /> },
    { title: "输入价/1M", dataIndex: "input_price_per_m", render: (_, row) => <Input value={row.input_price_per_m} onChange={(event) => patchSyncRow(row.key, { input_price_per_m: event.target.value })} style={{ width: 110 }} /> },
    { title: "输出价/1M", dataIndex: "output_price_per_m", render: (_, row) => <Input value={row.output_price_per_m} onChange={(event) => patchSyncRow(row.key, { output_price_per_m: event.target.value })} style={{ width: 110 }} /> },
    { title: "倍率", dataIndex: "rate_multiplier", render: (_, row) => <Input value={row.rate_multiplier} onChange={(event) => patchSyncRow(row.key, { rate_multiplier: event.target.value })} style={{ width: 90 }} /> },
    {
      title: "状态",
      dataIndex: "state",
      render: (state) => state === "bound" ? <Tag color="green">已绑定</Tag> : state === "exists" ? <Tag color="blue">已存在仅绑定</Tag> : <Tag>新建</Tag>
    }
  ];
  const successRate = detail.stats.requests > 0 ? `${((detail.stats.successes / detail.stats.requests) * 100).toFixed(2)}%` : "0%";

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Link to="/channels">返回渠道列表</Link>
      <Card
        title={`渠道详情：${detail.channel.name}`}
        extra={<Button type="primary" onClick={openSync}>同步上游模型</Button>}
      >
        {metricCards([
          { label: "供应商", value: detail.channel.provider_type },
          { label: "状态", value: detail.channel.status },
          { label: "健康状态", value: detail.channel.health },
          { label: "权重", value: detail.channel.weight },
          { label: "已绑模型", value: detail.models.length },
          { label: "最近延迟(ms)", value: detail.channel.last_latency_ms ?? 0 },
          { label: "累计请求", value: detail.stats.requests },
          { label: "成功率", value: successRate },
          { label: "平均延迟(ms)", value: detail.stats.average_latency }
        ])}
        <Typography.Paragraph copyable style={{ marginTop: 16 }}>
          {detail.channel.base_url}
        </Typography.Paragraph>
      </Card>
      <Card title="最近测试">
        <Space wrap>
          <Tag color={detail.channel.health === "healthy" ? "green" : "red"}>{detail.channel.health}</Tag>
          <Typography.Text>最近成功：{detail.channel.last_success_at ?? "暂无"}</Typography.Text>
          <Typography.Text>最近延迟：{detail.channel.last_latency_ms ?? 0} ms</Typography.Text>
        </Space>
      </Card>
      <Card
        title="绑定列表"
        extra={<Button danger onClick={() => unbindSelected(selectedBindingKeys)}>批量解绑选中</Button>}
      >
        <Table
          rowKey="model_id"
          columns={columns}
          dataSource={detail.models}
          rowSelection={{ selectedRowKeys: selectedBindingKeys, onChange: setSelectedBindingKeys }}
        />
      </Card>
      <Modal
        title="同步上游模型"
        open={syncOpen}
        onCancel={() => setSyncOpen(false)}
        onOk={importSelected}
        okText="确认导入"
        confirmLoading={importing}
        width={1280}
      >
        <Space wrap style={{ marginBottom: 12 }}>
          <Input placeholder="批量倍率" value={bulkRate} onChange={(event) => setBulkRate(event.target.value)} style={{ width: 120 }} />
          <Input placeholder="批量输入价/1M" value={bulkInput} onChange={(event) => setBulkInput(event.target.value)} style={{ width: 150 }} />
          <Input placeholder="批量输出价/1M" value={bulkOutput} onChange={(event) => setBulkOutput(event.target.value)} style={{ width: 150 }} />
          <Button onClick={applyBulk}>应用到选中</Button>
        </Space>
        <Table
          size="small"
          rowKey="key"
          loading={syncLoading}
          columns={syncColumns}
          dataSource={syncRows}
          rowSelection={{ selectedRowKeys: selectedSyncKeys, onChange: setSelectedSyncKeys, getCheckboxProps: (row) => ({ disabled: row.state === "bound" }) }}
          pagination={{ pageSize: 8 }}
          scroll={{ x: 1180 }}
        />
      </Modal>
    </Space>
  );
}
