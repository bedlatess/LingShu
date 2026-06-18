import { useEffect, useState } from "react";
import { Card, Col, Row, Space, Table, Typography, message } from "antd";
import type { AdminDashboard, ReportRow, User, createAPI } from "@lingshu/shared";

import { errText, fmtMoney, MiniBars } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

function Metric({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <Card>
      <Typography.Text type="secondary">{label}</Typography.Text>
      <Typography.Title level={4} style={{ margin: "8px 0 0" }}>{value}</Typography.Title>
    </Card>
  );
}

export function AdminDashboardPage({ dashboard, auditCount, me, api }: { dashboard: AdminDashboard | null; auditCount: number | null; me: User; api: AdminAPI }) {
  const [daily, setDaily] = useState<ReportRow[]>([]);
  const [models, setModels] = useState<ReportRow[]>([]);
  const [channels, setChannels] = useState<ReportRow[]>([]);

  useEffect(() => {
    Promise.all([api.adminReportDaily("", ""), api.adminReportByModel("", ""), api.adminReportByChannel("", "")])
      .then(([dailyResult, modelResult, channelResult]) => {
        setDaily(dailyResult.items.slice(0, 7).reverse());
        setModels(modelResult.items.slice(0, 5));
        setChannels(channelResult.items.slice(0, 5));
      })
      .catch((err) => message.error(`加载概览趋势失败: ${errText(err)}`));
  }, [api]);

  const successRate = dashboard?.today_requests ? `${(((dashboard.today_successes ?? 0) / dashboard.today_requests) * 100).toFixed(1)}%` : "0%";
  const topColumns = [
    { title: "维度", dataIndex: "label" },
    { title: "请求", dataIndex: "requests" },
    { title: "扣费", dataIndex: "charge", render: (value: string) => fmtMoney(value) }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Typography.Title level={3} style={{ margin: 0 }}>运营概览</Typography.Title>
      <Typography.Text type="secondary">当前管理员：{me.username}</Typography.Text>

      <Card title="今日">
        <Row gutter={[16, 16]}>
          <Col xs={24} md={8} xl={4}><Metric label="今日请求" value={dashboard?.today_requests ?? 0} /></Col>
          <Col xs={24} md={8} xl={4}><Metric label="成功" value={dashboard?.today_successes ?? 0} /></Col>
          <Col xs={24} md={8} xl={4}><Metric label="失败" value={dashboard?.today_failures ?? 0} /></Col>
          <Col xs={24} md={8} xl={4}><Metric label="成功率" value={successRate} /></Col>
          <Col xs={24} md={8} xl={4}><Metric label="今日扣费" value={fmtMoney(dashboard?.today_charge)} /></Col>
          <Col xs={24} md={8} xl={4}><Metric label="今日成本" value={fmtMoney(dashboard?.today_base_cost)} /></Col>
        </Row>
      </Card>

      <Card title="全站资产">
        <Row gutter={[16, 16]}>
          <Col xs={24} md={8} xl={6}><Metric label="总用户" value={dashboard?.total_users ?? 0} /></Col>
          <Col xs={24} md={8} xl={6}><Metric label="活跃用户" value={dashboard?.active_users ?? 0} /></Col>
          <Col xs={24} md={8} xl={6}><Metric label="余额池" value={fmtMoney(dashboard?.balance_total)} /></Col>
          <Col xs={24} md={8} xl={6}><Metric label="累计毛利" value={fmtMoney(dashboard?.gross_profit)} /></Col>
        </Row>
      </Card>

      <Card title="资源健康">
        <Row gutter={[16, 16]}>
          <Col xs={24} md={8} xl={6}><Metric label="渠道" value={`${dashboard?.healthy_channels ?? 0}/${dashboard?.total_channels ?? 0}`} /></Col>
          <Col xs={24} md={8} xl={6}><Metric label="模型" value={`${dashboard?.enabled_models ?? 0}/${dashboard?.total_models ?? 0}`} /></Col>
          <Col xs={24} md={8} xl={6}><Metric label="活跃密钥" value={dashboard?.active_api_keys ?? 0} /></Col>
          <Col xs={24} md={8} xl={6}><Metric label="审计日志" value={auditCount ?? "--"} /></Col>
        </Row>
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={12}>
          <Card title="近 7 日扣费趋势">
            <MiniBars data={daily.map((item) => ({ label: item.label, value: Number(item.charge) }))} />
          </Card>
        </Col>
        <Col xs={24} xl={6}>
          <Card title="Top 模型">
            <Table size="small" rowKey="key" columns={topColumns} dataSource={models} pagination={false} />
          </Card>
        </Col>
        <Col xs={24} xl={6}>
          <Card title="Top 渠道">
            <Table size="small" rowKey="key" columns={topColumns} dataSource={channels} pagination={false} />
          </Card>
        </Col>
      </Row>
    </Space>
  );
}
