import { useEffect, useState } from "react";
import { Button, Card, Input, Space, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { createAPI, type AuditLog } from "@lingshu/shared";

import { errText, exportCSV, type Pager, tablePagination } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function AuditPage({ api, auditColumns, initialLogs, pager, setPager }: { api: AdminAPI; auditColumns: ColumnsType<AuditLog>; initialLogs: AuditLog[]; pager: Pager; setPager: React.Dispatch<React.SetStateAction<Pager>> }) {
  const [logs, setLogs] = useState<AuditLog[]>(initialLogs);
  const [actorID, setActorID] = useState("");
  const [action, setAction] = useState("");
  const [targetType, setTargetType] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [loading, setLoading] = useState(false);

  async function loadAuditLogs() {
    setLoading(true);
    try {
      const result = await api.listAuditLogs(pager.page, pager.limit, {
        actor_id: actorID,
        action,
        target_type: targetType,
        from,
        to
      });
      setLogs(result.items);
      setPager((prev) => ({ ...prev, total: result.total }));
    } catch (err) {
      message.error(`加载审计日志失败: ${errText(err)}`);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    setLogs(initialLogs);
  }, [initialLogs]);

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title="审计日志筛选">
        <Space wrap>
          <Input placeholder="操作者 ID" value={actorID} onChange={(event) => setActorID(event.target.value)} style={{ width: 220 }} />
          <Input placeholder="动作" value={action} onChange={(event) => setAction(event.target.value)} style={{ width: 180 }} />
          <Input placeholder="对象类型" value={targetType} onChange={(event) => setTargetType(event.target.value)} style={{ width: 180 }} />
          <Input type="date" aria-label="开始日期" value={from} onChange={(event) => setFrom(event.target.value)} />
          <Input type="date" aria-label="结束日期" value={to} onChange={(event) => setTo(event.target.value)} />
          <Button type="primary" onClick={loadAuditLogs}>查询</Button>
          <Button onClick={() => { setActorID(""); setAction(""); setTargetType(""); setFrom(""); setTo(""); }}>重置</Button>
          <Button onClick={() => exportCSV("audit-logs.csv", logs)}>导出 CSV</Button>
        </Space>
      </Card>
      <Card title="审计日志"><Table rowKey="id" loading={loading} columns={auditColumns} dataSource={logs} pagination={tablePagination(pager, setPager)} /></Card>
    </Space>
  );
}
