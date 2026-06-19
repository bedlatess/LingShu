import React from "react";
import { BookOpen, Copy, KeyRound, Link as LinkIcon, Pencil, Plus, Trash2 } from "lucide-react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { APIKey, CreatedAPIKey } from "@lingshu/shared";

import { Badge, Button, Card, CardContent, CardHeader, CardTitle, DataTable, Dialog, EmptyState, Input, PageHeader, toast } from "@lingshu/ui";
import { copyText } from "@/lib/clipboard";
import { useAuth } from "@/providers/auth";
import { trStatus } from "@/lib/i18n";
import { useSiteInfo } from "@/providers/site-info";

export function ApiKeysPage() {
  const { t, i18n } = useTranslation("keys");
  const { api } = useAuth();
  const { siteInfo } = useSiteInfo();
  const [items, setItems] = React.useState<APIKey[]>([]);
  const [name, setName] = React.useState("");
  const [allowAll, setAllowAll] = React.useState(true);
  const [selectedEndpoints, setSelectedEndpoints] = React.useState<string[]>([]);
  const [created, setCreated] = React.useState<CreatedAPIKey | null>(null);
  const [editing, setEditing] = React.useState<APIKey | null>(null);
  const [editName, setEditName] = React.useState("");
  const [editAllowAll, setEditAllowAll] = React.useState(true);
  const [editEndpoints, setEditEndpoints] = React.useState<string[]>([]);
  const baseURL = React.useMemo(() => normalizedBaseURL(siteInfo?.api_base_url), [siteInfo?.api_base_url]);

  async function refresh() {
    const result = await api.userAPIKeys();
    setItems(result.items);
  }

  React.useEffect(() => {
    refresh();
  }, [api]);

  async function createKey(event: React.FormEvent) {
    event.preventDefault();
    const result = await api.createUserAPIKey({ name: name || t("defaultName"), allowed_endpoints: allowAll ? [] : selectedEndpoints });
    setCreated(result);
    setName("");
    setAllowAll(true);
    setSelectedEndpoints([]);
    toast.success(t("createSuccess"));
    await refresh();
  }

  function startEdit(item: APIKey) {
    setEditing(item);
    setEditName(item.name);
    setEditAllowAll(!item.allowed_endpoints?.length);
    setEditEndpoints(item.allowed_endpoints ?? []);
  }

  async function saveEdit(event: React.FormEvent) {
    event.preventDefault();
    if (!editing) return;
    await api.updateUserAPIKey(editing.id, { name: editName, allowed_endpoints: editAllowAll ? [] : editEndpoints });
    toast.success(t("updateSuccess"));
    setEditing(null);
    await refresh();
  }

  async function removeKey(id: string) {
    if (!window.confirm(t("deleteConfirm"))) return;
    await api.deleteUserAPIKey(id);
    toast.success(t("deleteSuccess"));
    await refresh();
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} />
      <Card>
        <CardHeader><CardTitle>{t("createTitle")}</CardTitle></CardHeader>
        <CardContent>
          <form className="grid gap-4" onSubmit={createKey}>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Input value={name} onChange={(event) => setName(event.target.value)} placeholder={t("namePlaceholder")} />
              <Button type="submit"><Plus className="h-4 w-4" />{t("createAction")}</Button>
            </div>
            <EndpointPicker allowAll={allowAll} endpoints={selectedEndpoints} onAllowAllChange={setAllowAll} onEndpointsChange={setSelectedEndpoints} />
          </form>
        </CardContent>
      </Card>
      <DataTable
        data={items}
        rowKey={(row) => row.id}
        empty={<EmptyState title={t("emptyTitle")} description={t("emptyDescription")} action={t("emptyAction")} onAction={() => document.querySelector<HTMLInputElement>("input")?.focus()} icon={<KeyRound className="h-5 w-5" />} />}
        columns={[
          { key: "name", title: t("table.name") },
          { key: "mask", title: t("table.mask"), render: (row) => <code className="font-mono text-xs">{row.mask}</code> },
          { key: "allowed_endpoints", title: t("table.endpoints"), render: (row) => <EndpointSummary endpoints={row.allowed_endpoints ?? []} /> },
          { key: "status", title: t("table.status"), render: (row) => (
            <span className="inline-flex items-center gap-2">
              <span className={`h-2 w-2 rounded-full ${row.status === "active" ? "bg-[var(--success)]" : "bg-muted-foreground"}`} />
              <Badge variant={row.status === "active" ? "success" : "muted"}>{trStatus(row.status)}</Badge>
            </span>
          ) },
          { key: "created_at", title: t("table.createdAt"), render: (row) => new Date(row.created_at).toLocaleString(i18n.resolvedLanguage === "zh" ? "zh-CN" : "en-US", { hour12: false }) },
          { key: "actions", title: t("table.actions"), render: (row) => (
            <div className="flex flex-wrap gap-2">
              <Button variant="secondary" size="sm" onClick={async () => {
                if (await copyText(baseURL)) toast.success(t("copyBaseURLSuccess"));
              }}><LinkIcon className="h-4 w-4" />{t("table.copyBaseURL")}</Button>
              <Button asChild variant="secondary" size="sm"><Link to="/docs"><BookOpen className="h-4 w-4" />{t("table.docs")}</Link></Button>
              <Button variant="secondary" size="sm" onClick={() => startEdit(row)}><Pencil className="h-4 w-4" />{t("table.edit")}</Button>
              <Button variant="destructive" size="sm" onClick={() => removeKey(row.id)}><Trash2 className="h-4 w-4" />{t("table.delete")}</Button>
            </div>
          ) }
        ]}
      />
      <Dialog
        open={Boolean(editing)}
        title={t("edit.title")}
        onClose={() => setEditing(null)}
      >
        <form className="grid gap-4" onSubmit={saveEdit}>
          <Input value={editName} onChange={(event) => setEditName(event.target.value)} placeholder={t("namePlaceholder")} />
          <EndpointPicker allowAll={editAllowAll} endpoints={editEndpoints} onAllowAllChange={setEditAllowAll} onEndpointsChange={setEditEndpoints} />
          <div className="flex justify-end gap-2">
            <Button type="button" variant="secondary" onClick={() => setEditing(null)}>{t("edit.cancel")}</Button>
            <Button type="submit">{t("edit.save")}</Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={Boolean(created)}
        title={t("dialog.title")}
        onClose={() => setCreated(null)}
        footer={<Button onClick={() => setCreated(null)}>{t("dialog.saved")}</Button>}
      >
        <p className="mb-3 text-sm leading-6 text-muted-foreground">{t("dialog.description")}</p>
        <div className="flex items-center gap-2 rounded-md border border-border bg-[var(--bg-subtle)] p-2">
          <code className="min-w-0 flex-1 break-all font-mono text-xs">{created?.plaintext}</code>
          <Button
            size="icon"
            variant="secondary"
            onClick={async () => {
              if (created?.plaintext && await copyText(created.plaintext)) toast.success(t("dialog.copied"));
            }}
            aria-label={t("dialog.ariaCopy")}
          >
            <Copy className="h-4 w-4" />
          </Button>
        </div>
      </Dialog>
    </div>
  );
}

function normalizedBaseURL(configured?: string) {
  const value = configured?.trim();
  if (value) return value.replace(/\/$/, "");
  return window.location.origin;
}

const endpointOptions = [
  { value: "/v1/chat/completions", labelKey: "chat" },
  { value: "/messages", labelKey: "messages" },
  { value: "/v1/embeddings", labelKey: "embeddings" },
  { value: "/v1/models", labelKey: "models" }
] as const;

function EndpointPicker({ allowAll, endpoints, onAllowAllChange, onEndpointsChange }: { allowAll: boolean; endpoints: string[]; onAllowAllChange: (value: boolean) => void; onEndpointsChange: (value: string[]) => void }) {
  const { t } = useTranslation("keys");
  function toggle(endpoint: string, checked: boolean) {
    if (checked) {
      onEndpointsChange(Array.from(new Set([...endpoints, endpoint])));
    } else {
      onEndpointsChange(endpoints.filter((item) => item !== endpoint));
    }
  }
  return (
    <fieldset className="grid gap-3 rounded-md border border-border bg-[var(--bg-subtle)] p-3">
      <legend className="px-1 text-sm font-medium text-foreground">{t("endpoints.title")}</legend>
      <label className="flex min-h-9 cursor-pointer items-center gap-2 text-sm">
        <input type="checkbox" checked={allowAll} onChange={(event) => onAllowAllChange(event.target.checked)} />
        {t("endpoints.all")}
      </label>
      {!allowAll ? (
        <div className="grid gap-2 sm:grid-cols-2">
          {endpointOptions.map((option) => (
            <label key={option.value} className="flex min-h-9 cursor-pointer items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm">
              <input type="checkbox" checked={endpoints.includes(option.value)} onChange={(event) => toggle(option.value, event.target.checked)} />
              <span>{t(`endpoints.${option.labelKey}`)}</span>
              <code className="ml-auto font-mono text-[11px] text-muted-foreground">{option.value}</code>
            </label>
          ))}
        </div>
      ) : null}
      <p className="text-xs leading-5 text-muted-foreground">{t("endpoints.hint")}</p>
    </fieldset>
  );
}

function EndpointSummary({ endpoints }: { endpoints: string[] }) {
  const { t } = useTranslation("keys");
  if (!endpoints.length) {
    return <Badge variant="muted">{t("endpoints.all")}</Badge>;
  }
  return (
    <div className="flex max-w-sm flex-wrap gap-1">
      {endpoints.map((endpoint) => <Badge key={endpoint} variant="muted">{endpoint}</Badge>)}
    </div>
  );
}
