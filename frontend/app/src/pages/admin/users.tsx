import React from "react";
import { useTranslation } from "react-i18next";
import { Link, useParams } from "react-router-dom";
import type { APIKey, GatewayLog, LedgerRecord, User, createAPI } from "@lingshu/shared";
import { Badge, Button, Card, CardContent, DataTable, Dialog, EmptyState, Input, PageHeader, Pagination, Select, StatCard, toast } from "@lingshu/ui";
import { Activity, KeyRound, WalletCards } from "lucide-react";
import { errText, exportCSV, fmtMoney, formatDateMinute, runWrite, statusVariant, type Pager } from "./admin-page-utils";

type AdminAPI = ReturnType<typeof createAPI>;

export function UsersPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const [users, setUsers] = React.useState<User[]>([]);
  const [pager, setPager] = React.useState<Pager>({ page: 1, limit: 20, total: 0 });
  const [form, setForm] = React.useState({ username: "", email: "", password: "", role: "user" as "user" | "admin" });
  const [balanceTarget, setBalanceTarget] = React.useState<User | null>(null);
  const [balanceForm, setBalanceForm] = React.useState({ amount: "", remark: "" });

  async function refresh(page = pager.page) {
    const result = await api.listUsers(page, pager.limit);
    setUsers(result.items);
    setPager((prev) => ({ ...prev, page, total: result.total }));
  }

  React.useEffect(() => { refresh(); }, [api, pager.page, pager.limit]);

  async function create(event: React.FormEvent) {
    event.preventDefault();
    await runWrite(async () => {
      await api.createUser(form);
      toast.success(t("users.createSuccess"));
      setForm({ username: "", email: "", password: "", role: "user" });
      await refresh(1);
    }, t("users.createFailed"));
  }

  async function adjustBalance(event: React.FormEvent) {
    event.preventDefault();
    if (!balanceTarget) return;
    await runWrite(async () => {
      await api.adjustUserBalance(balanceTarget.id, balanceForm);
      toast.success(t("users.adjustSuccess"));
      setBalanceTarget(null);
      setBalanceForm({ amount: "", remark: "" });
      await refresh();
    }, t("users.adjustFailed"));
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("users.eyebrow")} title={t("users.title")} description={t("users.description")} />
      <Card>
        <CardContent className="p-5">
          <form className="grid gap-3 lg:grid-cols-[1fr_1fr_1fr_140px_auto]" onSubmit={create}>
            <Input placeholder={t("users.username")} value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} required />
            <Input placeholder={t("users.email")} value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
            <Input placeholder={t("users.initialPassword")} type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={8} />
            <Select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value as "user" | "admin" })}><option value="user">{t("users.normalUser")}</option><option value="admin">{t("users.admin")}</option></Select>
            <Button type="submit">{t("common.create")}</Button>
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={users}
        rowKey={(row) => row.id}
        columns={[
          { key: "username", title: t("common.user"), render: (row) => <Link className="text-[var(--clay)] hover:underline" to={`/admin/users/${row.id}`}>{row.username}</Link> },
          { key: "email", title: t("users.email") },
          { key: "role", title: t("common.type"), render: (row) => row.role === "admin" ? t("users.admin") : t("users.normalUser") },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status === "active" ? t("common.active") : t("common.banned")}</Badge> },
          { key: "balance", title: t("users.balance"), render: (row) => fmtMoney(row.balance) },
          {
            key: "actions",
            title: t("common.actions"),
            render: (row) => (
              <div className="flex flex-wrap gap-2">
                <Button size="sm" variant="secondary" asChild><Link to={`/admin/users/${row.id}`}>{t("common.details")}</Link></Button>
                <Button size="sm" variant="secondary" onClick={() => setBalanceTarget(row)}>{t("users.balanceAction")}</Button>
                {row.status === "active" ? <Button size="sm" variant="destructive" onClick={() => runWrite(async () => { await api.banUser(row.id); toast.success(t("users.banSuccess")); await refresh(); }, t("users.banFailed"))}>{t("common.banned")}</Button> : null}
              </div>
            )
          }
        ]}
      />
      <Pagination page={pager.page} limit={pager.limit} total={pager.total} onChange={(page) => setPager((prev) => ({ ...prev, page }))} />
      <Dialog open={Boolean(balanceTarget)} title={balanceTarget ? t("users.adjustBalanceFor", { name: balanceTarget.username }) : t("users.adjustBalance")} onClose={() => setBalanceTarget(null)}>
        <form className="grid gap-4" onSubmit={adjustBalance}>
          <label className="grid gap-2 text-sm">{t("common.amount")}<Input value={balanceForm.amount} onChange={(e) => setBalanceForm({ ...balanceForm, amount: e.target.value })} placeholder={t("users.amountHelp")} required /></label>
          <label className="grid gap-2 text-sm">{t("users.remark")}<Input value={balanceForm.remark} onChange={(e) => setBalanceForm({ ...balanceForm, remark: e.target.value })} placeholder={t("users.remarkPlaceholder")} required /></label>
          <div className="flex justify-end gap-2"><Button variant="secondary" type="button" onClick={() => setBalanceTarget(null)}>{t("common.cancel")}</Button><Button type="submit">{t("users.confirmAdjust")}</Button></div>
        </form>
      </Dialog>
    </div>
  );
}

export function UserDetailPage({ api }: { api: AdminAPI }) {
  const { t } = useTranslation("admin");
  const { id } = useParams();
  const [user, setUser] = React.useState<User | null>(null);
  const [summary, setSummary] = React.useState<{ total_charge: string; total_recharge: string } | null>(null);
  const [apiKeys, setAPIKeys] = React.useState<APIKey[]>([]);
  const [logs, setLogs] = React.useState<GatewayLog[]>([]);
  const [ledger, setLedger] = React.useState<LedgerRecord[]>([]);
  const [error, setError] = React.useState("");

  async function refresh() {
    if (!id) return;
    try {
      const [userItem, keyList, logList, ledgerList, summaryItem] = await Promise.all([api.getUser(id), api.adminUserAPIKeys(id, 1, 20), api.adminUserLogs(id, 1, 20), api.adminUserLedger(id, 1, 20), api.adminUserSummary(id)]);
      setUser(userItem);
      setAPIKeys(keyList.items);
      setLogs(logList.items);
      setLedger(ledgerList.items);
      setSummary(summaryItem);
    } catch (err) {
      setError(errText(err));
    }
  }

  React.useEffect(() => { refresh(); }, [api, id]);

  if (error) return <EmptyState title={t("users.loadingFailed")} description={error} />;
  if (!user) return <EmptyState title={t("users.loadingUser")} description={t("users.loadingUserDesc")} />;

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("users.detailEyebrow")} title={user.username} description={t("users.userId", { id: user.id })} action={<Button variant="secondary" onClick={() => exportCSV(`user-${user.id}-logs.csv`, logs)}>{t("users.exportRequests")}</Button>} />
      <section className="grid gap-4 md:grid-cols-3">
        <StatCard label={t("users.balance")} value={fmtMoney(user.balance)} hint={user.status} icon={WalletCards} />
        <StatCard label={t("users.totalCharge")} value={fmtMoney(summary?.total_charge)} hint={t("users.ledgerSummary")} icon={Activity} />
        <StatCard label="API Key" value={apiKeys.length} hint={t("users.latest20")} icon={KeyRound} />
      </section>
      <DataTable
        data={apiKeys}
        rowKey={(row) => row.id}
        columns={[
          { key: "name", title: t("common.name") },
          { key: "mask", title: t("users.key") },
          { key: "status", title: t("common.status"), render: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
          { key: "created_at", title: t("common.createdAt"), render: (row) => formatDateMinute(row.created_at) }
        ]}
      />
      <DataTable
        data={ledger}
        rowKey={(row, i) => `${row.created_at}-${i}`}
        columns={[
          { key: "type", title: t("common.type") },
          { key: "amount", title: t("common.amount"), render: (row) => fmtMoney(row.amount) },
          { key: "balance_after", title: t("users.balanceAfter"), render: (row) => fmtMoney(row.balance_after) },
          { key: "base_cost", title: t("common.cost"), render: (row) => fmtMoney(row.base_cost) },
          { key: "rate_multiplier", title: t("common.multiplier") },
          { key: "created_at", title: t("common.time"), render: (row) => formatDateMinute(row.created_at) }
        ]}
      />
      <DataTable
        data={logs}
        rowKey={(row) => row.request_id}
        columns={[
          { key: "request_id", title: t("common.request") },
          { key: "model_id", title: t("common.model") },
          { key: "status", title: t("common.status") },
          { key: "base_cost", title: t("common.cost"), render: (row) => fmtMoney(row.base_cost) },
          { key: "charge", title: t("common.charge"), render: (row) => fmtMoney(row.charge) },
          { key: "created_at", title: t("common.time"), render: (row) => formatDateMinute(row.created_at) }
        ]}
      />
    </div>
  );
}
