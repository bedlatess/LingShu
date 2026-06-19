import React from "react";
import { useTranslation } from "react-i18next";
import type { APIKey, User, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, DataTable, Dialog, Input, PageHeader, Pagination, Select, toast } from "@lingshu/ui";
import { copyText, formatDateMinute, runWrite, statusVariant, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function ApiKeysPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [keys, setKeys] = React.useState<APIKey[]>([]);
  const [users, setUsers] = React.useState<User[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [form, setForm] = React.useState({ user_id: "", name: "" });
  const [createdKey, setCreatedKey] = React.useState("");

  async function refresh() {
    const [keyList, userList] = await Promise.all([api.listAPIKeys(pager.page, pager.limit), api.listUsers(1, 100)]);
    setKeys(keyList.items);
    setUsers(userList.items);
    setPager((prev) => ({ ...prev, total: keyList.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function create(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      const result = await api.createAPIKey(form);
      setCreatedKey(result.plaintext);
      toast.success(t("apiKeys.createSuccess"));
      await refresh();
    }, t("apiKeys.createFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("apiKeys.eyebrow")} title={t("apiKeys.title")} description={t("apiKeys.description")} />
      <Card>
        <CardContent className="p-5">
          <form className="grid gap-3 md:grid-cols-[1fr_1fr_auto]" onSubmit={create}>
            <Select value={form.user_id} onChange={(event) => setForm({ ...form, user_id: event.target.value })} required>
              <option value="">{t("apiKeys.chooseUser")}</option>
              {users.map((user) => <option key={user.id} value={user.id}>{user.username}</option>)}
            </Select>
            <Input placeholder={t("apiKeys.namePlaceholder")} value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
            <Button type="submit">{t("common.create")}</Button>
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={keys}
        rowKey={(row) => row.id}
        columns={[
          { key: "name", title: t("apiKeys.table.name") },
          { key: "user_id", title: t("apiKeys.table.user") },
          { key: "mask", title: t("apiKeys.table.key") },
          { key: "status", title: t("apiKeys.table.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "created_at", title: t("apiKeys.table.createdAt"), render: (row) => formatDateMinute(row.created_at) },
          {
            key: "actions",
            title: t("common.actions"),
            render: (row) => (
              <div className="flex gap-2">
                {row.status === "active" ? <Button size="sm" variant="secondary" onClick={() => runWrite(async () => { await api.disableAPIKey(row.id); await refresh(); }, t("apiKeys.disableFailed"))}>{t("common.disable")}</Button> : null}
                <Button size="sm" variant="destructive" onClick={() => runWrite(async () => { await api.deleteAPIKey(row.id); await refresh(); }, t("apiKeys.deleteFailed"))}>{t("common.delete")}</Button>
              </div>
            )
          }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
      <Dialog
        open={Boolean(createdKey)}
        title={t("apiKeys.newKeyTitle")}
        onClose={() => setCreatedKey("")}
        footer={<Button onClick={() => setCreatedKey("")}>{t("common.close")}</Button>}
      >
        <div className="flex gap-2">
          <code className="min-w-0 flex-1 break-all rounded-md bg-[var(--bg-subtle)] p-3 text-xs">{createdKey}</code>
          <Button onClick={async () => { if (await copyText(createdKey)) toast.success(t("common.copied")); }}>{t("common.copy")}</Button>
        </div>
      </Dialog>
    </div>
  );
}
