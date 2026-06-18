import { Button, Form, Input, Select, Space } from "antd";
import type { FormInstance } from "antd";
import { perKToM, perMToK, type ModelConfig } from "@lingshu/shared";

import { modelDefaults } from "./admin-page-utils";

export function ModelForm({ form, onFinish }: { form: FormInstance<Omit<ModelConfig, "id">>; onFinish: (values: Omit<ModelConfig, "id">) => Promise<void> }) {
  async function handleFinish(values: Omit<ModelConfig, "id">) {
    await onFinish({
      ...values,
      input_price_per_1k: perMToK(values.input_price_per_1k),
      output_price_per_1k: perMToK(values.output_price_per_1k)
    });
  }

  return (
    <Form form={form} layout="vertical" onFinish={handleFinish} initialValues={{ ...modelDefaults, input_price_per_1k: perKToM(modelDefaults.input_price_per_1k), output_price_per_1k: perKToM(modelDefaults.output_price_per_1k) }}>
      <Space wrap align="start">
        <Form.Item name="public_name" label="模型名" rules={[{ required: true }]}><Input style={{ width: 220 }} /></Form.Item>
        <Form.Item name="type" label="类型"><Select style={{ width: 140 }} options={["chat", "embedding", "image", "video"].map((value) => ({ value }))} /></Form.Item>
        <Form.Item name="billing_mode" label="计费"><Select style={{ width: 140 }} options={["token", "per_call"].map((value) => ({ value }))} /></Form.Item>
        <Form.Item name="group" label="分组"><Input style={{ width: 140 }} /></Form.Item>
        <Form.Item name="input_price_per_1k" label="输入价 / 1M token"><Input style={{ width: 150 }} /></Form.Item>
        <Form.Item name="output_price_per_1k" label="输出价 / 1M token"><Input style={{ width: 150 }} /></Form.Item>
        <Form.Item name="price_per_call" label="单次基准"><Input style={{ width: 140 }} /></Form.Item>
        <Form.Item name="rate_multiplier" label="倍率" rules={[{ required: true }]}><Input style={{ width: 120 }} /></Form.Item>
        <Form.Item name="status" label="状态"><Select style={{ width: 120 }} options={["enabled", "disabled"].map((value) => ({ value }))} /></Form.Item>
        <Form.Item name="sort_order" label="排序"><Input style={{ width: 100 }} /></Form.Item>
      </Space>
      <Button type="primary" htmlType="submit">保存</Button>
    </Form>
  );
}
