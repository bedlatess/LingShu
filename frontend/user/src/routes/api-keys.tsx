import React from "react";
import { Check, Copy, KeyRound, Plus, Trash2 } from "lucide-react";
import type { APIKey } from "@lingshu/shared";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";

export function ApiKeysPage() {
  const { api } = useAuth();
  const [items, setItems] = React.useState<APIKey[]>([]);
  const [name, setName] = React.useState("");
  const [plaintext, setPlaintext] = React.useState("");
  const [copied, setCopied] = React.useState(false);

  async function refresh() {
    const result = await api.userAPIKeys();
    setItems(result.items);
  }

  React.useEffect(() => {
    refresh();
  }, []);

  async function createKey(event: React.FormEvent) {
    event.preventDefault();
    const result = await api.createUserAPIKey({ name });
    setPlaintext(result.plaintext);
    setName("");
    await refresh();
  }

  async function copyKey() {
    await navigator.clipboard.writeText(plaintext);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  async function disableKey(id: string) {
    await api.updateUserAPIKey(id, { status: "disabled" });
    await refresh();
  }

  async function deleteKey(id: string) {
    await api.deleteUserAPIKey(id);
    await refresh();
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow="API Keys" title="自助创建和管理平台 Key" description="Key 明文只在创建时显示一次。后续只展示脱敏值，避免泄露。" />
      {plaintext ? (
        <Card className="border-primary/35 bg-primary/10">
          <CardContent className="flex flex-col gap-3 p-5 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="text-sm font-semibold text-primary">新 Key 仅显示一次</p>
              <code className="mt-2 block break-all rounded-md bg-background/70 p-3 text-sm">{plaintext}</code>
            </div>
            <Button onClick={copyKey} variant="secondary">{copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}复制</Button>
          </CardContent>
        </Card>
      ) : null}

      <Card className="glass">
        <CardHeader><CardTitle>创建 API Key</CardTitle></CardHeader>
        <CardContent>
          <form className="flex flex-col gap-3 sm:flex-row" onSubmit={createKey}>
            <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="例如：production-router" required />
            <Button type="submit"><Plus className="h-4 w-4" />创建</Button>
          </form>
        </CardContent>
      </Card>

      <Card className="glass">
        <CardHeader><CardTitle>Key 列表</CardTitle></CardHeader>
        <CardContent className="grid gap-3">
          {items.length === 0 ? (
            <EmptyState title="还没有 API Key" description="创建一个 Key 后，即可把 OpenAI SDK 的 base_url 指向 LingShu。" />
          ) : (
            items.map((item) => (
              <div key={item.id} className="flex flex-col gap-3 rounded-lg border border-white/10 bg-white/[0.035] p-4 md:flex-row md:items-center md:justify-between">
                <div className="min-w-0">
                  <div className="flex items-center gap-2"><KeyRound className="h-4 w-4 text-primary" /><strong>{item.name}</strong><Badge>{item.status}</Badge></div>
                  <p className="mt-2 break-all font-mono text-sm text-muted-foreground">{item.mask}</p>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" onClick={() => disableKey(item.id)}>停用</Button>
                  <Button variant="destructive" size="icon" onClick={() => deleteKey(item.id)} title="删除"><Trash2 className="h-4 w-4" /></Button>
                </div>
              </div>
            ))
          )}
        </CardContent>
      </Card>
    </div>
  );
}
