import { useMemo, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tabs, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { createAPI, designTokens, type Announcement } from "@lingshu/shared";

import { type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;
type AnnouncementFormValues = Pick<Announcement, "title" | "content" | "status" | "priority" | "pinned">;

const defaultAnnouncement: AnnouncementFormValues = {
  title: "",
  content: "",
  status: "online",
  priority: 0,
  pinned: false
};

function MarkdownPreview({ value }: { value?: string }) {
  return (
    <div style={{ minHeight: 220, padding: 16, border: `1px solid ${designTokens.colors.border}`, borderRadius: 6, background: designTokens.colors.surface, color: designTokens.colors.ink }}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{value || "暂无内容"}</ReactMarkdown>
    </div>
  );
}

export function AnnouncementsPage({
  announcements,
  api,
  refresh,
  pager,
  setPager
}: {
  announcements: Announcement[];
  api: AdminAPI;
  refresh: () => Promise<void>;
  pager: Pager;
  setPager: React.Dispatch<React.SetStateAction<Pager>>;
}) {
  const [form] = Form.useForm<AnnouncementFormValues>();
  const [editForm] = Form.useForm<AnnouncementFormValues>();
  const [editing, setEditing] = useState<Announcement | null>(null);
  const draftContent = Form.useWatch("content", form);
  const editContent = Form.useWatch("content", editForm);

  const statusOptions = useMemo(() => [
    { value: "online", label: "上线" },
    { value: "offline", label: "下线" }
  ], []);
  const pinnedOptions = useMemo(() => [
    { value: false, label: "否" },
    { value: true, label: "是" }
  ], []);

  async function create(values: AnnouncementFormValues) {
    await runWrite(async () => {
      await api.createAnnouncement({ ...values, priority: Number(values.priority ?? 0), pinned: Boolean(values.pinned) });
      message.success("公告已发布");
      form.resetFields();
      await refresh();
    }, "发布公告失败").catch(() => undefined);
  }

  async function update(values: AnnouncementFormValues) {
    if (!editing) return;
    await runWrite(async () => {
      await api.updateAnnouncement(editing.id, { ...values, priority: Number(values.priority ?? 0), pinned: Boolean(values.pinned) });
      message.success("公告已更新");
      setEditing(null);
      editForm.resetFields();
      await refresh();
    }, "更新公告失败").catch(() => undefined);
  }

  const columns: ColumnsType<Announcement> = [
    { title: "标题", dataIndex: "title" },
    { title: "状态", dataIndex: "status" },
    { title: "优先级", dataIndex: "priority" },
    { title: "置顶", dataIndex: "pinned", render: (value) => (value ? "是" : "否") },
    {
      title: "操作",
      render: (_, item) => (
        <Space>
          <Button
            onClick={() => {
              setEditing(item);
              editForm.setFieldsValue({
                title: item.title,
                content: item.content,
                status: item.status,
                priority: item.priority,
                pinned: item.pinned
              });
            }}
          >
            编辑
          </Button>
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
        </Space>
      )
    }
  ];

  const formContent = (
    <Form form={form} layout="vertical" onFinish={create} initialValues={defaultAnnouncement}>
      <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="content" label="内容" rules={[{ required: true }]}>
        <Input.TextArea rows={10} placeholder="支持 Markdown，例如 # 标题、- 列表、**粗体**、```代码块```、[链接](https://...)" />
      </Form.Item>
      <Space wrap>
        <Form.Item name="status" label="状态"><Select style={{ width: 140 }} options={statusOptions} /></Form.Item>
        <Form.Item name="priority" label="优先级"><Input style={{ width: 120 }} /></Form.Item>
        <Form.Item name="pinned" label="置顶"><Select style={{ width: 120 }} options={pinnedOptions} /></Form.Item>
      </Space>
      <Button type="primary" htmlType="submit">发布</Button>
    </Form>
  );

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="发布公告">
        <Tabs
          items={[
            { key: "edit", label: "编辑", children: formContent },
            { key: "preview", label: "预览", children: <MarkdownPreview value={draftContent} /> }
          ]}
        />
      </Card>
      <Card title="公告列表"><Table rowKey="id" dataSource={announcements} columns={columns} pagination={tablePagination(pager, setPager)} /></Card>
      <Modal title="编辑公告" open={Boolean(editing)} onCancel={() => setEditing(null)} onOk={() => editForm.submit()} width={860} destroyOnClose>
        <Tabs
          items={[
            {
              key: "edit",
              label: "编辑",
              children: (
                <Form form={editForm} layout="vertical" onFinish={update}>
                  <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item name="content" label="内容" rules={[{ required: true }]}><Input.TextArea rows={10} /></Form.Item>
                  <Space wrap>
                    <Form.Item name="status" label="状态"><Select style={{ width: 140 }} options={statusOptions} /></Form.Item>
                    <Form.Item name="priority" label="优先级"><Input style={{ width: 120 }} /></Form.Item>
                    <Form.Item name="pinned" label="置顶"><Select style={{ width: 120 }} options={pinnedOptions} /></Form.Item>
                  </Space>
                </Form>
              )
            },
            { key: "preview", label: "预览", children: <MarkdownPreview value={editContent} /> }
          ]}
        />
      </Modal>
    </Space>
  );
}
