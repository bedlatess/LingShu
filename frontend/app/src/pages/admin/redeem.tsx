import React from "react";
import { useTranslation } from "react-i18next";
import type { RedeemCode, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, DataTable, Dialog, Input, PageHeader, Pagination, toast } from "@lingshu/ui";
import { fmtMoney, formatDateMinute, runWrite, statusVariant, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function RedeemPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [items, setItems] = React.useState<RedeemCode[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [form, setForm] = React.useState({ amount: "100", count: 1, batch_name: "", max_uses: 1 });
  const [created, setCreated] = React.useState<RedeemCode[]>([]);

  async function refresh() {
    const result = await api.listRedeemCodes(pager.page, pager.limit);
    setItems(result.items);
    setPager((prev) => ({ ...prev, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function create(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      const result = await api.createRedeemCodes(form);
      setCreated(result.items);
      toast.success(t("redeem.generateSuccess"));
      await refresh();
    }, t("redeem.generateFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("redeem.eyebrow")} title={t("redeem.title")} description={t("redeem.description")} />
      <Card>
        <CardContent className="p-5">
          <form className="grid gap-3 md:grid-cols-5" onSubmit={create}>
            <Input value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} placeholder={t("redeem.amountPlaceholder")} />
            <Input type="number" value={form.count} onChange={(e) => setForm({ ...form, count: Number(e.target.value) })} />
            <Input value={form.batch_name} onChange={(e) => setForm({ ...form, batch_name: e.target.value })} placeholder={t("redeem.batchPlaceholder")} />
            <Input type="number" value={form.max_uses} onChange={(e) => setForm({ ...form, max_uses: Number(e.target.value) })} />
            <Button type="submit">{t("redeem.generate")}</Button>
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={items}
        rowKey={(row) => row.id}
        columns={[
          { key: "code_prefix", title: t("redeem.table.code"), render: (row) => row.code ?? row.code_prefix },
          { key: "batch_name", title: t("redeem.table.batch") },
          { key: "amount", title: t("common.amount"), render: (row) => fmtMoney(row.amount) },
          { key: "used_count", title: t("redeem.table.used"), render: (row) => `${row.used_count}/${row.max_uses}` },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "created_at", title: t("common.createdAt"), render: (row) => formatDateMinute(row.created_at) },
          { key: "actions", title: t("common.actions"), render: (row) => <Button size="sm" variant="destructive" onClick={() => runWrite(async () => { await api.disableRedeemCode(row.id); await refresh(); }, t("redeem.disableFailed"))}>{t("common.disable")}</Button> }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
      <Dialog open={created.length > 0} title={t("redeem.newCodes")} onClose={() => setCreated([])} footer={<Button onClick={() => setCreated([])}>{t("common.close")}</Button>}>
        <div className="grid gap-2">{created.map((item) => <code className="rounded-md bg-[var(--bg-subtle)] p-2 text-xs" key={item.id}>{item.code ?? item.code_prefix}</code>)}</div>
      </Dialog>
    </div>
  );
}
