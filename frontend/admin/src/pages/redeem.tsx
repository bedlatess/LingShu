import { useState } from "react";
import { Alert, Button, Card, Drawer, Form, Input, Modal, Space, Table, Tag, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { Link } from "react-router-dom";
import { createAPI, designTokens, type RedeemCode, type RedeemRecord } from "@lingshu/shared";

import { copyText, fmtMoney, type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function RedeemPage({ redeemCodes, api, refresh, createdCodes, setCreatedCodes, pager, setPager }: { redeemCodes: RedeemCode[]; api: AdminAPI; refresh: () => Promise<void>; createdCodes: string[]; setCreatedCodes: (codes: string[]) => void; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  const [recordsOpen, setRecordsOpen] = useState(false);
  const [records, setRecords] = useState<RedeemRecord[]>([]);
  const [recordsLoading, setRecordsLoading] = useState(false);

  async function copyCode(code: string) {
    const ok = await copyText(code);
    ok ? message.success("已复制兑换码") : message.error("复制失败，请手动选择复制");
  }

  async function openRecords(item: RedeemCode) {
    setRecordsOpen(true);
    setRecordsLoading(true);
    try {
      const result = await api.listRedeemRecords(item.id);
      setRecords(result.items);
    } catch (err) {
      message.error(`加载使用记录失败: ${err instanceof Error ? err.message : "未知错误"}`);
    } finally {
      setRecordsLoading(false);
    }
  }

  const columns: ColumnsType<RedeemCode> = [
    {
      title: "完整卡号",
      render: (_, item) => {
        const code = item.code || `${item.code_prefix}****`;
        return (
          <Space>
            <Typography.Text code>{code}</Typography.Text>
            {item.code ? <Button size="small" onClick={() => copyCode(item.code!)}>复制</Button> : null}
          </Space>
        );
      }
    },
    { title: "批次", dataIndex: "batch_name" },
    { title: "面额", dataIndex: "amount", render: (value) => fmtMoney(value) },
    { title: "状态", dataIndex: "status", render: (value) => <Tag color={value === "unused" ? designTokens.colors.success : value === "disabled" ? designTokens.colors.danger : "default"}>{value}</Tag> },
    { title: "使用", render: (_, item) => `${item.used_count}/${item.max_uses}` },
    { title: "有效期", dataIndex: "expires_at", render: (value) => value ? new Date(value).toLocaleString("zh-CN") : "永久" },
    {
      title: "操作",
      render: (_, item) => (
        <Space>
          <Button onClick={() => openRecords(item)}>使用记录</Button>
          {item.status === "unused" ? (
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
          ) : null}
        </Space>
      )
    }
  ];

  const recordColumns: ColumnsType<RedeemRecord> = [
    { title: "用户", dataIndex: "username", render: (value, item) => <Link to={`/users/${item.user_id}`}>{value}</Link> },
    { title: "入账金额", dataIndex: "amount", render: (value) => fmtMoney(value) },
    { title: "IP", dataIndex: "client_ip" },
    { title: "兑换时间", dataIndex: "created_at", render: (value) => new Date(value).toLocaleString("zh-CN") }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {createdCodes.length > 0 ? (
        <Alert
          type="success"
          message="新兑换码已生成"
          description={<Space direction="vertical">{createdCodes.map((code) => <Typography.Text code key={code} copyable={{ text: code }}>{code}</Typography.Text>)}</Space>}
        />
      ) : null}
      <Card title="生成兑换码">
        <Form
          layout="inline"
          onFinish={(values) => runWrite(async () => {
            const result = await api.createRedeemCodes({
              ...values,
              count: Number(values.count ?? 1),
              max_uses: Number(values.max_uses ?? 1),
              expires_at: values.expires_at || undefined
            });
            setCreatedCodes(result.items.map((item) => item.code ?? "").filter(Boolean));
            message.success("兑换码已生成");
            await refresh();
          }, "生成兑换码失败").catch(() => undefined)}
          initialValues={{ count: 1, max_uses: 1 }}
        >
          <Form.Item name="amount" rules={[{ required: true }]}><Input placeholder="面额" /></Form.Item>
          <Form.Item name="count"><Input placeholder="数量" /></Form.Item>
          <Form.Item name="batch_name"><Input placeholder="批次" /></Form.Item>
          <Form.Item name="max_uses"><Input placeholder="可用次数" /></Form.Item>
          <Form.Item name="expires_at"><Input type="date" placeholder="有效期" /></Form.Item>
          <Button type="primary" htmlType="submit">生成</Button>
        </Form>
      </Card>
      <Card title="兑换码列表">
        <Table rowKey="id" dataSource={redeemCodes} columns={columns} pagination={tablePagination(pager, setPager)} />
      </Card>
      <Drawer title="兑换使用记录" open={recordsOpen} onClose={() => setRecordsOpen(false)} width={720}>
        <Table rowKey="id" loading={recordsLoading} dataSource={records} columns={recordColumns} pagination={false} />
      </Drawer>
    </Space>
  );
}
