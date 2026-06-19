import React from "react";
import { useTranslation } from "react-i18next";
import { Link, useParams } from "react-router-dom";
import type { Channel, ChannelDetail, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, DataTable, Dialog, Input, PageHeader, Pagination, Select, StatCard, toast } from "@lingshu/ui";
import { Activity, Clock, RadioTower } from "lucide-react";
import { formatDateMinute, providerOptions, runWrite, statusVariant, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function ChannelsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [channels, setChannels] = React.useState<Channel[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [form, setForm] = React.useState({ name: "", provider_type: "openai", base_url: "", api_key: "", status: "enabled", weight: 1 });
  const [editing, setEditing] = React.useState<Channel | null>(null);
  const [editForm, setEditForm] = React.useState({ name: "", provider_type: "openai", base_url: "", api_key: "", status: "enabled", weight: 1 });

  async function refresh() {
    const result = await api.listChannels(pager.page, pager.limit);
    setChannels(result.items);
    setPager((prev) => ({ ...prev, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function create(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      await api.createChannel({ ...form, timeout_seconds: 60, rpm_limit: 0, concurrency_limit: 0, fail_threshold: 3 });
      toast.success(t("channels.createSuccess"));
      setForm({ name: "", provider_type: "openai", base_url: "", api_key: "", status: "enabled", weight: 1 });
      await refresh();
    }, t("channels.createFailed"));
  }

  async function saveEdit(event: React.FormEvent) {
    event.preventDefault();
    if (!editing) return;
    await runWrite(async () => {
      await api.updateChannel(editing.id, editForm);
      toast.success(t("channels.updateSuccess"));
      setEditing(null);
      await refresh();
    }, t("channels.updateFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("channels.eyebrow")} title={t("channels.title")} description={t("channels.description")} />
      <Card>
        <CardContent className="p-5">
          <form className="grid gap-3 xl:grid-cols-[1fr_160px_1.3fr_1fr_100px_auto]" onSubmit={create}>
            <Input placeholder={t("channels.namePlaceholder")} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
            <Select value={form.provider_type} onChange={(e) => setForm({ ...form, provider_type: e.target.value })}>{providerOptions.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</Select>
            <Input placeholder={t("channels.baseURL")} value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} required />
            <Input placeholder={t("channels.apiKey")} value={form.api_key} onChange={(e) => setForm({ ...form, api_key: e.target.value })} required />
            <Input type="number" value={form.weight} onChange={(e) => setForm({ ...form, weight: Number(e.target.value) })} />
            <Button type="submit">{t("common.create")}</Button>
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={channels}
        rowKey={(row) => row.id}
        columns={[
          { key: "name", title: t("common.name"), render: (row) => <Link className="text-[var(--clay)] hover:underline" to={`/admin/channels/${row.id}`}>{row.name}</Link> },
          { key: "provider_type", title: t("channels.table.provider") },
          { key: "base_url", title: t("channels.table.address") },
          { key: "bound_count", title: t("channels.table.bound") },
          { key: "last_success_at", title: t("channels.table.lastSuccess"), render: (row) => formatDateMinute(row.last_success_at) },
          { key: "health", title: t("channels.table.health"), render: (row) => <Badge variant={statusVariant(row.health)}>{row.health}</Badge> },
          {
            key: "actions",
            title: t("common.actions"),
            render: (row) => (
              <div className="flex gap-2">
                <Button size="sm" variant="secondary" onClick={() => { setEditing(row); setEditForm({ name: row.name, provider_type: row.provider_type, base_url: row.base_url, api_key: "", status: row.status, weight: row.weight }); }}>{t("common.edit")}</Button>
                <Button size="sm" variant="secondary" onClick={() => runWrite(async () => { const result = await api.testChannel(row.id, row.base_url); result.ok ? toast.success(t("channels.testPassed", { latency: result.latency_ms })) : toast.error(result.message); }, t("channels.testFailed"))}>{t("common.test")}</Button>
                <Button size="sm" variant="destructive" onClick={() => runWrite(async () => { await api.deleteChannel(row.id); await refresh(); }, t("channels.deleteFailed"))}>{t("common.delete")}</Button>
              </div>
            )
          }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
      <Dialog open={Boolean(editing)} title={editing ? t("channels.editTitle", { name: editing.name }) : t("channels.editFallback")} onClose={() => setEditing(null)}>
        <form className="grid gap-4" onSubmit={saveEdit}>
          <Input placeholder={t("channels.namePlaceholder")} value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} required />
          <Select value={editForm.provider_type} onChange={(e) => setEditForm({ ...editForm, provider_type: e.target.value })}>{providerOptions.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</Select>
          <Input placeholder={t("channels.baseURL")} value={editForm.base_url} onChange={(e) => setEditForm({ ...editForm, base_url: e.target.value })} required />
          <Input placeholder={t("channels.newApiKeyPlaceholder")} value={editForm.api_key} onChange={(e) => setEditForm({ ...editForm, api_key: e.target.value })} />
          <Select value={editForm.status} onChange={(e) => setEditForm({ ...editForm, status: e.target.value })}><option value="enabled">{t("common.enabled")}</option><option value="disabled">{t("common.disabled")}</option></Select>
          <Input type="number" value={editForm.weight} onChange={(e) => setEditForm({ ...editForm, weight: Number(e.target.value) })} />
          <div className="flex justify-end gap-2"><Button variant="secondary" type="button" onClick={() => setEditing(null)}>{t("common.cancel")}</Button><Button type="submit">{t("common.save")}</Button></div>
        </form>
      </Dialog>
    </div>
  );
}

export function ChannelDetailPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const { id } = useParams();
  const [detail, setDetail] = React.useState<ChannelDetail | null>(null);
  React.useEffect(() => { if (id) api.getChannelDetail(id).then(setDetail); }, [api, id]);
  if (!detail) return <PageHeader title={t("channels.detailTitle")} description={t("channels.loadingDetail")} />;
  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("channels.detailTitle")} title={detail.channel.name} description={detail.channel.base_url} />
      <section className="grid gap-4 md:grid-cols-3">
        <StatCard label={t("channels.requests")} value={detail.stats.requests} hint={t("channels.failures", { count: detail.stats.failures })} icon={Activity} />
        <StatCard label={t("channels.averageLatency")} value={detail.stats.average_latency} hint={t("channels.milliseconds")} icon={Clock} />
        <StatCard label={t("channels.boundModels")} value={detail.models.length} hint={detail.channel.provider_type} icon={RadioTower} />
      </section>
      <DataTable
        data={detail.models}
        rowKey={(row) => row.id}
        columns={[
          { key: "model_name", title: t("channels.table.platformModel") },
          { key: "upstream_model_name", title: t("channels.table.upstreamModel") },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "created_at", title: t("channels.table.boundAt"), render: (row) => formatDateMinute(row.created_at) }
        ]}
      />
    </div>
  );
}
