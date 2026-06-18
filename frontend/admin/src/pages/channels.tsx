import { useEffect, useState } from "react";
import { Alert, Button, Card, Form, Input, Select, Space, Table, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { Link, useParams } from "react-router-dom";
import { createAPI, type Channel, type ChannelDetail, type ModelConfig } from "@lingshu/shared";

import { errText, metricCards, providerOptions, type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function ChannelsPage({ channels, channelColumns, api, refresh, models, pager, setPager }: { channels: Channel[]; channelColumns: ColumnsType<Channel>; api: AdminAPI; refresh: () => Promise<void>; models: ModelConfig[]; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建上游渠道">
        <Form layout="inline" onFinish={(values) => runWrite(async () => { await api.createChannel({ ...values, weight: Number(values.weight ?? 1), timeout_seconds: 120, rpm_limit: 60, concurrency_limit: 5, fail_threshold: 5 }); message.success("渠道已创建"); await refresh(); }, "创建渠道失败").catch(() => undefined)} initialValues={{ provider_type: "openai", status: "enabled", weight: 1 }}>
          <Form.Item name="name" rules={[{ required: true }]}><Input placeholder="名称" /></Form.Item>
          <Form.Item name="provider_type" rules={[{ required: true }]}><Select style={{ width: 180 }} options={providerOptions} /></Form.Item>
          <Form.Item name="base_url" rules={[{ required: true }]}><Input placeholder="Base URL" style={{ width: 260 }} /></Form.Item>
          <Form.Item name="api_key" rules={[{ required: true }]}><Input.Password placeholder="上游 API Key" /></Form.Item>
          <Form.Item name="weight"><Input placeholder="权重" style={{ width: 90 }} /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="渠道列表"><Table rowKey="id" columns={channelColumns} dataSource={channels} pagination={tablePagination(pager, setPager)} /></Card>
      <Card title="绑定渠道模型">
        <Form layout="inline" onFinish={(values) => runWrite(async () => { await api.bindChannelModel(values); message.success("渠道模型已绑定"); await refresh(); }, "绑定渠道模型失败").catch(() => undefined)}>
          <Form.Item name="channel_id" rules={[{ required: true }]}><Select style={{ width: 260 }} options={channels.map((channel) => ({ value: channel.id, label: channel.name }))} /></Form.Item>
          <Form.Item name="model_id" rules={[{ required: true }]}><Select style={{ width: 260 }} options={models.map((model) => ({ value: model.id, label: model.public_name }))} /></Form.Item>
          <Form.Item name="upstream_model_name" rules={[{ required: true }]}><Input placeholder="上游模型名" /></Form.Item>
          <Button type="primary" htmlType="submit">绑定</Button>
        </Form>
      </Card>
    </Space>
  );
}

export function ChannelDetailPage({ api }: { api: AdminAPI }) {
  const { id } = useParams();
  const [detail, setDetail] = useState<ChannelDetail | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) return;
    api.getChannelDetail(id).then(setDetail).catch((err) => setError(errText(err)));
  }, [id]);

  if (!id) return <Alert type="error" message="缺少渠道 ID" />;
  if (error) return <Alert type="error" message={error} />;
  if (!detail) return <Card title="渠道详情">加载中...</Card>;

  const columns: ColumnsType<ChannelDetail["models"][number]> = [
    { title: "模型", dataIndex: "model_name", render: (_, item) => <Link to={`/models/${item.model_id}`}>{item.model_name}</Link> },
    { title: "上游模型名", dataIndex: "upstream_model_name" },
    { title: "状态", dataIndex: "status" },
    { title: "绑定时间", dataIndex: "created_at" }
  ];
  const successRate = detail.stats.requests > 0 ? `${((detail.stats.successes / detail.stats.requests) * 100).toFixed(2)}%` : "0%";

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Link to="/channels">返回渠道列表</Link>
      <Card title={`渠道详情：${detail.channel.name}`}>
        {metricCards([
          { label: "供应商", value: detail.channel.provider_type },
          { label: "状态", value: detail.channel.status },
          { label: "健康状态", value: detail.channel.health },
          { label: "权重", value: detail.channel.weight },
          { label: "累计请求", value: detail.stats.requests },
          { label: "成功率", value: successRate },
          { label: "平均延迟(ms)", value: detail.stats.average_latency }
        ])}
        <Typography.Paragraph copyable style={{ marginTop: 16 }}>
          {detail.channel.base_url}
        </Typography.Paragraph>
      </Card>
      <Card title="绑定模型"><Table rowKey="id" columns={columns} dataSource={detail.models} /></Card>
    </Space>
  );
}
