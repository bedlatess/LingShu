import React from "react";
import { useTranslation } from "react-i18next";
import type { AuditLog, createAPI } from "@lingshu/shared";
import { Button, Card, CardContent, DataTable, Input, PageHeader, Pagination, toast } from "@lingshu/ui";
import { formatDateMinute, runWrite, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function AuditPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [items, setItems] = React.useState<AuditLog[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [filters, setFilters] = React.useState({ actor_id: "", action: "", target_type: "" });

  async function refresh() {
    const result = await api.listAuditLogs(pager.page, pager.limit, filters);
    setItems(result.items);
    setPager((prev) => ({ ...prev, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  return (
    <div className="page-grid">
      <PageHeader
        eyebrow={t("audit.eyebrow")}
        title={t("audit.title")}
        description={t("audit.description")}
        action={<Button variant="destructive" onClick={() => runWrite(async () => { const result = await api.cleanupAuditLogs(180); toast.success(t("audit.cleanupSuccess", { count: result.deleted })); await refresh(); }, t("audit.cleanupFailed"))}>{t("audit.cleanup")}</Button>}
      />
      <Card>
        <CardContent className="grid gap-3 p-5 md:grid-cols-4">
          <Input placeholder={t("audit.actorPlaceholder")} value={filters.actor_id} onChange={(event) => setFilters({ ...filters, actor_id: event.target.value })} />
          <Input placeholder={t("audit.actionPlaceholder")} value={filters.action} onChange={(event) => setFilters({ ...filters, action: event.target.value })} />
          <Input placeholder={t("audit.targetPlaceholder")} value={filters.target_type} onChange={(event) => setFilters({ ...filters, target_type: event.target.value })} />
          <Button onClick={refresh}>{t("common.filter")}</Button>
        </CardContent>
      </Card>
      <DataTable
        data={items}
        rowKey={(row) => row.id}
        columns={[
          { key: "action", title: t("audit.table.action") },
          { key: "target_type", title: t("audit.table.target") },
          { key: "target_id", title: t("audit.table.targetId") },
          { key: "actor_id", title: t("audit.table.actor") },
          { key: "ip", title: t("audit.table.ip") },
          { key: "created_at", title: t("audit.table.createdAt"), render: (row) => formatDateMinute(row.created_at) }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
    </div>
  );
}
