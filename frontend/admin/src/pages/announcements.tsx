import { Button, Card, Form, Input, Modal, Select, Space, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { createAPI, type Announcement } from "@lingshu/shared";

import { type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function AnnouncementsPage({ announcements, api, refresh, pager, setPager }: { announcements: Announcement[]; api: AdminAPI; refresh: () => Promise<void>; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  const columns: ColumnsType<Announcement> = [
    { title: "标题", dataIndex: "title" },
    { title: "状态", dataIndex: "status" },
    { title: "优先级", dataIndex: "priority" },
    { title: "置顶", dataIndex: "pinned", render: (value) => (value ? "是" : "否") },
    {
      title: "操作",
      render: (_, item) => (
        <Button
          danger
          onClick={() => Modal.confirm({
            title: "确认删除公告？",
            content: item.title,
            okText: "确认删除",
            cancelText: "取消",
            onOk: () =>
              runWrite(async () => {
                await api.deleteAnnouncement(item.id);
                message.success("公告已删除");
                await refresh();
              }, "删除公告失败")
          })}
        >
          删除
        </Button>
      )
    }
  ];
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="发布公告">
        <Form layout="inline" onFinish={(values) => runWrite(async () => { await api.createAnnouncement({ ...values, priority: Number(values.priority ?? 0), pinned: Boolean(values.pinned) }); message.success("公告已发布"); await refresh(); }, "发布公告失败").catch(() => undefined)} initialValues={{ status: "online", priority: 0, pinned: false }}>
          <Form.Item name="title" rules={[{ required: true }]}><Input placeholder="标题" /></Form.Item>
          <Form.Item name="content" rules={[{ required: true }]}><Input placeholder="内容" style={{ width: 360 }} /></Form.Item>
          <Form.Item name="status"><Select style={{ width: 120 }} options={[{ value: "online" }, { value: "offline" }]} /></Form.Item>
          <Button type="primary" htmlType="submit">发布</Button>
        </Form>
      </Card>
      <Card title="公告列表"><Table rowKey="id" dataSource={announcements} columns={columns} pagination={tablePagination(pager, setPager)} /></Card>
    </Space>
  );
}
