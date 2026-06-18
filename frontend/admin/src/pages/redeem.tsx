import { Alert, Button, Card, Form, Input, Modal, Space, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { createAPI, type RedeemCode } from "@lingshu/shared";

import { type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function RedeemPage({ redeemCodes, api, refresh, createdCodes, setCreatedCodes, pager, setPager }: { redeemCodes: RedeemCode[]; api: AdminAPI; refresh: () => Promise<void>; createdCodes: string[]; setCreatedCodes: (codes: string[]) => void; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  const columns: ColumnsType<RedeemCode> = [
    { title: "前缀", dataIndex: "code_prefix" },
    { title: "批次", dataIndex: "batch_name" },
    { title: "面额", dataIndex: "amount" },
    { title: "状态", dataIndex: "status" },
    { title: "使用", render: (_, item) => `${item.used_count}/${item.max_uses}` },
    {
      title: "操作",
      render: (_, item) => item.status === "active" ? (
        <Button
          danger
          onClick={() => Modal.confirm({
            title: "确认禁用兑换码？",
            content: `批次：${item.batch_name || item.code_prefix}`,
            okText: "确认禁用",
            cancelText: "取消",
            onOk: () =>
              runWrite(async () => {
                await api.disableRedeemCode(item.id);
                message.success("兑换码已禁用");
                await refresh();
              }, "禁用兑换码失败")
          })}
        >
          禁用
        </Button>
      ) : "-"
    }
  ];
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {createdCodes.length > 0 && <Alert type="success" message={`新兑换码仅显示一次：${createdCodes.join(", ")}`} />}
      <Card title="生成兑换码">
        <Form layout="inline" onFinish={(values) => runWrite(async () => { const result = await api.createRedeemCodes({ ...values, count: Number(values.count ?? 1), max_uses: Number(values.max_uses ?? 1) }); setCreatedCodes(result.items.map((item) => item.code ?? "").filter(Boolean)); message.success("兑换码已生成"); await refresh(); }, "生成兑换码失败").catch(() => undefined)} initialValues={{ count: 1, max_uses: 1 }}>
          <Form.Item name="amount" rules={[{ required: true }]}><Input placeholder="面额" /></Form.Item>
          <Form.Item name="count"><Input placeholder="数量" /></Form.Item>
          <Form.Item name="batch_name"><Input placeholder="批次" /></Form.Item>
          <Button type="primary" htmlType="submit">生成</Button>
        </Form>
      </Card>
      <Card title="兑换码列表"><Table rowKey="id" dataSource={redeemCodes} columns={columns} pagination={tablePagination(pager, setPager)} /></Card>
    </Space>
  );
}
