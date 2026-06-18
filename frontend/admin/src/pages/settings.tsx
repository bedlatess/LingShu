import { Button, Card, Form, Input } from "antd";
import type { FormInstance } from "antd";
import type { SystemSetting } from "@lingshu/shared";

export function SettingsPage({ settings, form, onSave }: { settings: SystemSetting[]; form: FormInstance<Record<string, string>>; onSave: (values: Record<string, string>) => Promise<void> }) {
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
