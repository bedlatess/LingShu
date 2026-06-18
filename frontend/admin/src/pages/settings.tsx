import { useState } from "react";
import { Alert, Button, Card, Form, Input, Modal, Space, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { FormInstance } from "antd";
import type { CleanupHistoryEntry, CleanupResult, SystemSetting } from "@lingshu/shared";

export function SettingsPage({
  settings,
  form,
  onSave,
  cleanupHistory = [],
  onRunCleanup,
  onRefreshCleanupHistory
}: {
  settings: SystemSetting[];
  form: FormInstance<Record<string, string>>;
  onSave: (values: Record<string, string>) => Promise<void>;
  cleanupHistory?: CleanupHistoryEntry[];
  onRunCleanup?: () => Promise<CleanupResult[]>;
  onRefreshCleanupHistory?: () => Promise<void>;
}) {
  const [cleaning, setCleaning] = useState(false);
  const latest = cleanupHistory[0];
  const latestSummary = latest?.results.map((item) => `${item.table}: ${item.deleted}`).join(" / ") || "暂无清理记录";
  const columns: ColumnsType<CleanupHistoryEntry> = [
    { title: "开始时间", dataIndex: "started_at" },
    { title: "结束时间", dataIndex: "ended_at" },
    {
      title: "结果",
      render: (_, item) => item.results.map((result) => `${result.table}: ${result.deleted}${result.err ? ` (${result.err})` : ""}`).join(" / ")
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="系统设置">
        <Form form={form} layout="vertical" onFinish={onSave}>
          {settings.map((item) => (
            <Form.Item key={item.key} name={item.key} label={item.description || item.key} extra={item.key}>
              <Input />
            </Form.Item>
          ))}
          <Button type="primary" htmlType="submit">保存设置</Button>
        </Form>
      </Card>

      <Card title="系统清理">
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Alert message="清理 30 天前的请求日志、90 天前的审计日志、过期公告与失效兑换码。账本数据永不清理。" type="info" showIcon />
          <div>上次清理时间：{latest?.started_at ?? "暂无"}</div>
          <div>删除汇总：{latestSummary}</div>
          <Space>
            <Button
              type="primary"
              loading={cleaning}
              onClick={() =>
                Modal.confirm({
                  title: "确认立即清理？",
                  content: "清理会删除过期请求明细、审计日志、已下线过期公告和过期未使用兑换码，不会删除账本。",
                  okText: "执行",
                  cancelText: "取消",
                  onOk: async () => {
                    if (!onRunCleanup) return;
                    setCleaning(true);
                    try {
                      const results = await onRunCleanup();
                      message.success(`清理完成：${results.map((item) => `${item.table} -${item.deleted}`).join(" / ")}`);
                    } finally {
                      setCleaning(false);
                    }
                  }
                })
              }
            >
              立即清理
            </Button>
            <Button onClick={onRefreshCleanupHistory}>刷新历史</Button>
          </Space>
          <Table rowKey="id" size="small" columns={columns} dataSource={cleanupHistory.slice(0, 10)} pagination={false} />
        </Space>
      </Card>
    </Space>
  );
}
