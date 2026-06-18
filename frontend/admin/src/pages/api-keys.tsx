import { Alert, Button, Card, Form, Input, Select, Space, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { createAPI, type APIKey, type User } from "@lingshu/shared";

import { type Pager, runWrite, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function ApiKeysPage({ apiKeys, keyColumns, api, refresh, users, createdKey, setCreatedKey, pager, setPager }: { apiKeys: APIKey[]; keyColumns: ColumnsType<APIKey>; api: AdminAPI; refresh: () => Promise<void>; users: User[]; createdKey: string; setCreatedKey: (key: string) => void; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {createdKey && <Alert type="success" message={`新 Key 仅显示一次：${createdKey}`} />}
      <Card title="创建 API Key">
        <Form layout="inline" onFinish={(values) => runWrite(async () => { const result = await api.createAPIKey(values); setCreatedKey(result.plaintext); message.success("API 密钥已创建"); await refresh(); }, "创建 API 密钥失败").catch(() => undefined)}>
          <Form.Item name="user_id" rules={[{ required: true }]}>
            <Select showSearch style={{ width: 280 }} placeholder="选择用户" options={users.map((user) => ({ value: user.id, label: `${user.username} (${user.id})` }))} />
          </Form.Item>
          <Form.Item name="name" rules={[{ required: true }]}><Input placeholder="名称" /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="Key 列表"><Table rowKey="id" columns={keyColumns} dataSource={apiKeys} pagination={tablePagination(pager, setPager)} /></Card>
    </Space>
  );
}
