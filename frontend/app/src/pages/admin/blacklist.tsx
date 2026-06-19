import React from "react";
import { Ban, ShieldCheck } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { AccessBlacklistEntry, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle, DataTable, Field, Input, PageHeader, Pagination, Select, Switch, Textarea, toast } from "@lingshu/ui";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { formatDateMinute, runWrite, statusVariant, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

type FormState = {
  kind: "ip" | "cidr" | "device";
  scope: "login" | "gateway" | "all";
  value: string;
  reason: string;
  permanent: boolean;
  expires_at: string;
};

const defaults: FormState = {
  kind: "ip",
  scope: "all",
  value: "",
  reason: "",
  permanent: false,
  expires_at: ""
};

export function BlacklistPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [items, setItems] = React.useState<AccessBlacklistEntry[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [filters, setFilters] = React.useState({ kind: "", scope: "", active: "true", q: "" });
  const [form, setForm] = React.useState<FormState>(defaults);
  const [pendingCreate, setPendingCreate] = React.useState<FormState | null>(null);
  const [pendingRelease, setPendingRelease] = React.useState<AccessBlacklistEntry | null>(null);

  async function refresh() {
    const result = await api.listBlacklist(pager.page, pager.limit, filters);
    setItems(result.items);
    setPager((prev) => ({ ...prev, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setPendingCreate({ ...form });
  }

  async function createRule(input: FormState) {
    await runWrite(async () => {
      await api.createBlacklistEntry({
        kind: input.kind,
        scope: input.scope,
        value: input.value.trim(),
        reason: input.reason.trim(),
        permanent: input.permanent,
        expires_at: input.expires_at ? new Date(input.expires_at).toISOString() : undefined
      });
      toast.success(t("blacklist.createSuccess"));
      setForm(defaults);
      setPendingCreate(null);
      await refresh();
    }, t("blacklist.createFailed"));
  }

  async function release(entry: AccessBlacklistEntry) {
    await runWrite(async () => {
      await api.releaseBlacklistEntry(entry.id);
      toast.success(t("blacklist.releaseSuccess"));
      setPendingRelease(null);
      await refresh();
    }, t("blacklist.releaseFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader
        eyebrow={t("blacklist.eyebrow")}
        title={t("blacklist.title")}
        description={t("blacklist.description")}
      />

      <div className="grid gap-5 xl:grid-cols-[380px_1fr]">
        <Card>
          <CardHeader><CardTitle>{t("blacklist.createTitle")}</CardTitle></CardHeader>
          <CardContent>
            <form className="grid gap-4" onSubmit={submit}>
              <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-1">
                <Field label={t("blacklist.kind")}>
                  <Select value={form.kind} onChange={(event) => setForm({ ...form, kind: event.target.value as FormState["kind"] })}>
                    <option value="ip">IP</option>
                    <option value="cidr">CIDR</option>
                    <option value="device">{t("blacklist.device")}</option>
                  </Select>
                </Field>
                <Field label={t("blacklist.scope")}>
                  <Select value={form.scope} onChange={(event) => setForm({ ...form, scope: event.target.value as FormState["scope"] })}>
                    <option value="all">{t("blacklist.scopes.all")}</option>
                    <option value="login">{t("blacklist.scopes.login")}</option>
                    <option value="gateway">{t("blacklist.scopes.gateway")}</option>
                  </Select>
                </Field>
              </div>
              <Field label={t("blacklist.value")} hint={t("blacklist.valueHint")}>
                <Input value={form.value} onChange={(event) => setForm({ ...form, value: event.target.value })} required />
              </Field>
              <Field label={t("blacklist.reason")}>
                <Textarea value={form.reason} onChange={(event) => setForm({ ...form, reason: event.target.value })} required />
              </Field>
              <div className="flex items-center justify-between gap-4 rounded-md border border-border bg-[var(--bg-subtle)] px-3 py-2">
                <div>
                  <p className="text-sm font-medium text-foreground">{t("blacklist.permanent")}</p>
                  <p className="text-xs text-muted-foreground">{t("blacklist.permanentHint")}</p>
                </div>
                <Switch checked={form.permanent} onCheckedChange={(checked) => setForm({ ...form, permanent: checked })} />
              </div>
              {!form.permanent ? (
                <Field label={t("blacklist.expiresAt")} hint={t("blacklist.expiresHint")}>
                  <Input type="datetime-local" value={form.expires_at} onChange={(event) => setForm({ ...form, expires_at: event.target.value })} />
                </Field>
              ) : null}
              <Button type="submit"><Ban className="h-4 w-4" />{t("blacklist.create")}</Button>
            </form>
          </CardContent>
        </Card>

        <div className="grid gap-4">
          <Card>
            <CardContent className="grid gap-3 p-5 md:grid-cols-[1fr_140px_140px_120px_auto]">
              <Input placeholder={t("blacklist.search")} value={filters.q} onChange={(event) => setFilters({ ...filters, q: event.target.value })} />
              <Select value={filters.kind} onChange={(event) => setFilters({ ...filters, kind: event.target.value })}>
                <option value="">{t("blacklist.allKinds")}</option>
                <option value="ip">IP</option>
                <option value="cidr">CIDR</option>
                <option value="device">{t("blacklist.device")}</option>
              </Select>
              <Select value={filters.scope} onChange={(event) => setFilters({ ...filters, scope: event.target.value })}>
                <option value="">{t("blacklist.allScopes")}</option>
                <option value="all">{t("blacklist.scopes.all")}</option>
                <option value="login">{t("blacklist.scopes.login")}</option>
                <option value="gateway">{t("blacklist.scopes.gateway")}</option>
              </Select>
              <Select value={filters.active} onChange={(event) => setFilters({ ...filters, active: event.target.value })}>
                <option value="true">{t("blacklist.activeOnly")}</option>
                <option value="false">{t("blacklist.releasedOnly")}</option>
                <option value="">{t("blacklist.allStatus")}</option>
              </Select>
              <Button onClick={() => { setPager((prev) => ({ ...prev, page: 1 })); void refresh(); }}>{t("common.filter")}</Button>
            </CardContent>
          </Card>

          <DataTable
            data={items}
            rowKey={(row) => row.id}
            emptyTitle={t("blacklist.emptyTitle")}
            emptyDescription={t("blacklist.emptyDescription")}
            columns={[
              { key: "kind", title: t("blacklist.table.kind"), render: (row) => <Badge variant="muted">{row.kind}</Badge> },
              { key: "value", title: t("blacklist.table.value"), render: (row) => <code className="font-mono text-xs">{row.value}</code> },
              { key: "scope", title: t("blacklist.table.scope"), render: (row) => t(`blacklist.scopes.${row.scope}`) },
              { key: "source", title: t("blacklist.table.source"), render: (row) => <Badge variant={row.source === "auto" ? "warning" : "info"}>{t(`blacklist.sources.${row.source}`)}</Badge> },
              { key: "active", title: t("blacklist.table.status"), render: (row) => <Badge variant={statusVariant(row.active ? "active" : "disabled")}>{row.active ? t("blacklist.active") : t("blacklist.released")}</Badge> },
              { key: "expires_at", title: t("blacklist.table.expiresAt"), render: (row) => row.expires_at ? formatDateMinute(row.expires_at) : t("blacklist.never") },
              { key: "created_at", title: t("blacklist.table.createdAt"), render: (row) => formatDateMinute(row.created_at) },
              {
                key: "actions",
                title: t("blacklist.table.actions"),
                render: (row) => row.active ? (
                  <Button variant="secondary" size="sm" onClick={() => setPendingRelease(row)}>
                    <ShieldCheck className="h-4 w-4" />{t("blacklist.release")}
                  </Button>
                ) : "-"
              }
            ]}
          />
          <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
        </div>
      </div>

      <ConfirmDialog
        open={Boolean(pendingCreate)}
        title={t("blacklist.confirmCreateTitle")}
        description={t("blacklist.confirmCreateDescription", { value: pendingCreate?.value ?? "" })}
        confirmText={t("blacklist.create")}
        cancelText={t("common.cancel")}
        intent="danger"
        onCancel={() => setPendingCreate(null)}
        onConfirm={() => pendingCreate ? createRule(pendingCreate) : undefined}
      />

      <ConfirmDialog
        open={Boolean(pendingRelease)}
        title={t("blacklist.confirmReleaseTitle")}
        description={t("blacklist.confirmReleaseDescription", { value: pendingRelease?.value ?? "" })}
        confirmText={t("blacklist.release")}
        cancelText={t("common.cancel")}
        intent="danger"
        onCancel={() => setPendingRelease(null)}
        onConfirm={() => pendingRelease ? release(pendingRelease) : undefined}
      />
    </div>
  );
}
