import React from "react";
import { useTranslation } from "react-i18next";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Announcement, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, DataTable, Dialog, Input, PageHeader, Pagination, Switch, Textarea, toast } from "@lingshu/ui";
import { formatDateMinute, runWrite, statusVariant, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function AnnouncementsPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [items, setItems] = React.useState<Announcement[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [form, setForm] = React.useState({ title: "", content: "", status: "published", priority: 0, pinned: false });
  const [editing, setEditing] = React.useState<Announcement | null>(null);
  const [editForm, setEditForm] = React.useState({ title: "", content: "", status: "published", priority: 0, pinned: false });

  async function refresh() {
    const result = await api.listAnnouncements(pager.page, pager.limit);
    setItems(result.items);
    setPager((prev) => ({ ...prev, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function create(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      await api.createAnnouncement(form);
      toast.success(t("announcements.publishSuccess"));
      setForm({ title: "", content: "", status: "published", priority: 0, pinned: false });
      await refresh();
    }, t("announcements.publishFailed"));
  }

  async function saveEdit(event: React.FormEvent) {
    event.preventDefault();
    if (!editing) return;
    await runWrite(async () => {
      await api.updateAnnouncement(editing.id, editForm);
      toast.success(t("announcements.updateSuccess"));
      setEditing(null);
      await refresh();
    }, t("announcements.updateFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("announcements.eyebrow")} title={t("announcements.title")} description={t("announcements.description")} />
      <Card>
        <CardContent className="grid gap-3 p-5">
          <form className="grid gap-3" onSubmit={create}>
            <Input placeholder={t("announcements.titlePlaceholder")} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
            <Textarea placeholder={t("announcements.contentPlaceholder")} value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} required />
            <label className="flex items-center gap-2 text-sm"><Switch checked={form.pinned} onCheckedChange={(pinned) => setForm({ ...form, pinned })} />{t("announcements.pinned")}</label>
            <Button type="submit">{t("announcements.publish")}</Button>
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={items}
        rowKey={(row) => row.id}
        columns={[
          { key: "title", title: t("announcements.table.title") },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "pinned", title: t("announcements.table.pinned"), render: (row) => row.pinned ? t("common.yes") : t("common.no") },
          { key: "created_at", title: t("common.time"), render: (row) => formatDateMinute(row.created_at) },
          { key: "actions", title: t("common.actions"), render: (row) => <div className="flex gap-2"><Button size="sm" variant="secondary" onClick={() => { setEditing(row); setEditForm({ title: row.title, content: row.content, status: row.status, priority: row.priority, pinned: row.pinned }); }}>{t("common.edit")}</Button><Button size="sm" variant="destructive" onClick={() => runWrite(async () => { await api.deleteAnnouncement(row.id); await refresh(); }, t("announcements.deleteFailed"))}>{t("common.delete")}</Button></div> }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
      <Card><CardContent className="prose p-5"><ReactMarkdown remarkPlugins={[remarkGfm]}>{form.content || t("announcements.previewEmpty")}</ReactMarkdown></CardContent></Card>
      <Dialog open={Boolean(editing)} title={editing ? t("announcements.editTitle", { title: editing.title }) : t("announcements.editFallback")} onClose={() => setEditing(null)}>
        <form className="grid gap-4" onSubmit={saveEdit}>
          <Input value={editForm.title} onChange={(e) => setEditForm({ ...editForm, title: e.target.value })} required />
          <Textarea value={editForm.content} onChange={(e) => setEditForm({ ...editForm, content: e.target.value })} required />
          <label className="flex items-center gap-2 text-sm"><Switch checked={editForm.pinned} onCheckedChange={(pinned) => setEditForm({ ...editForm, pinned })} />{t("announcements.pinned")}</label>
          <div className="flex justify-end gap-2"><Button variant="secondary" type="button" onClick={() => setEditing(null)}>{t("common.cancel")}</Button><Button type="submit">{t("common.save")}</Button></div>
        </form>
      </Dialog>
    </div>
  );
}
