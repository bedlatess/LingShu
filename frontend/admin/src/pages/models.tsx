import { useEffect, useState } from "react";
import { Alert, Card, Form, Space, Table } from "antd";
import type { ColumnsType } from "antd/es/table";
import { Link, useParams } from "react-router-dom";
import { createAPI, type ModelConfig, type ModelDetail } from "@lingshu/shared";

import { errText, metricCards, type Pager, tablePagination } from "./admin-page-utils";
import { ModelForm } from "./model-form";

type AdminAPI = ReturnType<typeof createAPI>;

export function ModelsPage({ models, modelColumns, onCreate, pager, setPager }: { models: ModelConfig[]; modelColumns: ColumnsType<ModelConfig>; onCreate: (values: Omit<ModelConfig, "id">) => Promise<void>; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  const [form] = Form.useForm<Omit<ModelConfig, "id">>();
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="创建模型"><ModelForm form={form} onFinish={async (values) => { await onCreate(values); form.resetFields(); }} /></Card>
      <Card title="模型列表"><Table rowKey="id" columns={modelColumns} dataSource={models} pagination={tablePagination(pager, setPager)} /></Card>
    </Space>
  );
}

export function ModelDetailPage({ api }: { api: AdminAPI }) {
  const { id } = useParams();
  const [detail, setDetail] = useState<ModelDetail | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!id) return;
    api.getModelDetail(id).then(setDetail).catch((err) => setError(errText(err)));
  }, [id]);
  if (!id) return <Alert type="error" message="缺少模型 ID" />;
  if (error) return <Alert type="error" message={error} />;
  if (!detail) return <Card title="模型详情">加载中...</Card>;
  const columns: ColumnsType<ModelDetail["channels"][number]> = [
    { title: "渠道", dataIndex: "channel_name", render: (_, item) => <Link to={`/channels/${item.channel_id}`}>{item.channel_name}</Link> },
    { title: "供应商", dataIndex: "provider_type" },
    { title: "上游模型名", dataIndex: "upstream_model_name" },
    { title: "状态", dataIndex: "status" },
    { title: "Base URL", dataIndex: "base_url" }
  ];
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Link to="/models">返回模型列表</Link>
      <Card title={`模型详情：${detail.model.public_name}`}>
        {metricCards([
          { label: "计费方式", value: detail.model.billing_mode },
          { label: "类型", value: detail.model.type },
          { label: "状态", value: detail.model.status },
          { label: "累计请求", value: detail.stats.requests },
          { label: "累计扣费", value: detail.stats.charge },
          { label: "毛利", value: detail.stats.gross_profit }
        ])}
      </Card>
      <Card title="绑定渠道"><Table rowKey="id" columns={columns} dataSource={detail.channels} /></Card>
    </Space>
  );
}
