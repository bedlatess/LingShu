import React from "react";
import { useTranslation } from "react-i18next";
import { Link, useParams } from "react-router-dom";
import type { ModelConfig, ModelDetail, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, DataTable, Dialog, Input, PageHeader, Pagination, Select, StatCard, toast } from "@lingshu/ui";
import { Activity, CircleDollarSign, RadioTower } from "lucide-react";
import { fmtMoney, formatDateMinute, modelDefaults, normalizeModelPayload, runWrite, statusVariant, type Pager } from "./admin-page-utils";
import { ModelForm } from "./model-form";

type AdminAPI = ReturnType<typeof createAPI>;

export function ModelsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [models, setModels] = React.useState<ModelConfig[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [form, setForm] = React.useState(modelDefaults);
  const [editing, setEditing] = React.useState<ModelConfig | null>(null);
  const [editForm, setEditForm] = React.useState<Omit<ModelConfig, "id">>(modelDefaults);

  async function refresh() {
    const result = await api.listModels(pager.page, pager.limit);
    setModels(result.items);
    setPager((prev) => ({ ...prev, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function create(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      await api.createModel(normalizeModelPayload(form) as Omit<ModelConfig, "id">);
      toast.success(t("models.createSuccess"));
      setForm(modelDefaults);
      await refresh();
    }, t("models.createFailed"));
  }

  async function saveEdit(event: React.FormEvent) {
    event.preventDefault();
    if (!editing) return;
    await runWrite(async () => {
      await api.updateModel(editing.id, normalizeModelPayload(editForm) as Omit<ModelConfig, "id">);
      toast.success(t("models.updateSuccess"));
      setEditing(null);
      await refresh();
    }, t("models.updateFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("models.eyebrow")} title={t("models.title")} description={t("models.description")} />
      <Card>
        <CardContent className="grid gap-3 p-5">
          <form className="grid gap-3 xl:grid-cols-4" onSubmit={create}>
            <Input placeholder={t("models.publicName")} value={form.public_name} onChange={(e) => setForm({ ...form, public_name: e.target.value })} required />
            <Select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}><option value="chat">{t("models.types.chat")}</option><option value="embedding">{t("models.types.embedding")}</option><option value="image">{t("models.types.image")}</option></Select>
            <Select value={form.billing_mode} onChange={(e) => setForm({ ...form, billing_mode: e.target.value })}><option value="token">{t("models.billing.token")}</option><option value="per_call">{t("models.billing.per_call")}</option></Select>
            <Button type="submit">{t("models.createModel")}</Button>
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={models}
        rowKey={(row) => row.id}
        columns={[
          { key: "public_name", title: t("common.model"), render: (row) => <Link className="text-[var(--clay)] hover:underline" to={`/admin/models/${row.id}`}>{row.public_name}</Link> },
          { key: "type", title: t("common.type") },
          { key: "billing_mode", title: t("common.charge") },
          { key: "input_price_per_1k", title: t("models.inputPrice"), render: (row) => fmtMoney(row.input_price_per_1k) },
          { key: "output_price_per_1k", title: t("models.outputPrice"), render: (row) => fmtMoney(row.output_price_per_1k) },
          { key: "rate_multiplier", title: t("common.multiplier") },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "actions", title: t("common.actions"), render: (row) => <div className="flex gap-2"><Button size="sm" variant="secondary" onClick={() => { setEditing(row); setEditForm(row); }}>{t("common.edit")}</Button><Button size="sm" variant="destructive" onClick={() => runWrite(async () => { await api.deleteModel(row.id); await refresh(); }, t("models.deleteFailed"))}>{t("common.delete")}</Button></div> }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
      <Dialog open={Boolean(editing)} title={editing ? t("models.editTitle", { name: editing.public_name }) : t("models.editFallback")} onClose={() => setEditing(null)}>
        <form className="grid gap-4" onSubmit={saveEdit}>
          <ModelForm value={editForm} onChange={setEditForm} />
          <div className="flex justify-end gap-2"><Button variant="secondary" type="button" onClick={() => setEditing(null)}>{t("common.cancel")}</Button><Button type="submit">{t("common.save")}</Button></div>
        </form>
      </Dialog>
    </div>
  );
}

export function ModelDetailPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const { id } = useParams();
  const [detail, setDetail] = React.useState<ModelDetail | null>(null);
  React.useEffect(() => { if (id) api.getModelDetail(id).then(setDetail); }, [api, id]);
  if (!detail) return <PageHeader title={t("models.detailTitle")} description={t("models.loadingDetail")} />;
  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("models.detailTitle")} title={detail.model.public_name} description={t("models.modelId", { id: detail.model.id })} />
      <section className="grid gap-4 md:grid-cols-3">
        <StatCard label={t("common.request")} value={detail.stats.requests} hint={t("channels.failures", { count: detail.stats.failures })} icon={Activity} />
        <StatCard label={t("common.charge")} value={fmtMoney(detail.stats.charge)} hint={t("models.userSide")} icon={CircleDollarSign} />
        <StatCard label={t("models.boundChannels")} value={detail.channels.length} hint={t("models.schedulableUpstream")} icon={RadioTower} />
      </section>
      <DataTable
        data={detail.channels}
        rowKey={(row) => row.id}
        columns={[
          { key: "channel_name", title: t("common.channel") },
          { key: "provider_type", title: t("models.table.provider") },
          { key: "upstream_model_name", title: t("models.table.upstreamModel") },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "created_at", title: t("models.table.boundAt"), render: (row) => formatDateMinute(row.created_at) }
        ]}
      />
    </div>
  );
}
