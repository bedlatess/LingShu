import React from "react";
import { Check, Copy, KeyRound, Pencil, Plus, RotateCcw, Trash2 } from "lucide-react";
import type { APIKey } from "@lingshu/shared/user-types";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { zhStatus } from "@/lib/i18n";
import { copyText } from "@/lib/clipboard";
import { toast } from "sonner";

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
    try {
      const result = await api.createUserAPIKey({ name });
      setPlaintext(result.plaintext);
      setName("");
      toast.success("API 密钥已创建");
      await refresh();
    } catch (err) {
      const message = err instanceof Error ? err.message : "创建失败";
      toast.error(`创建失败：${message}`);
    }
  }

  async function copyKey() {
    const ok = await copyText(plaintext);
    if (ok) {
      setCopied(true);
      toast.success("已复制到剪贴板");
      setTimeout(() => setCopied(false), 1500);
    } else {
      toast.error("复制失败，请手动选择复制");
    }
  }

  async function disableKey(id: string) {
    try {
      await api.updateUserAPIKey(id, { status: "disabled" });
      toast.success("已停用");
      await refresh();
    } catch (err) {
      const message = err instanceof Error ? err.message : "操作失败";
      toast.error(`操作失败：${message}`);
    }
  }

  async function renameKey(id: string, currentName: string) {
    const nextName = window.prompt("请输入新的密钥名称", currentName);
    if (!nextName || nextName === currentName) return;
    try {
      await api.updateUserAPIKey(id, { name: nextName });
      toast.success("API 密钥已重命名");
      await refresh();
    } catch (err) {
      const message = err instanceof Error ? err.message : "操作失败";
      toast.error(`操作失败：${message}`);
    }
  }

  async function enableKey(id: string) {
    try {
      await api.updateUserAPIKey(id, { status: "active" });
      toast.success("已重新启用");
      await refresh();
    } catch (err) {
      const message = err instanceof Error ? err.message : "操作失败";
      toast.error(`操作失败：${message}`);
    }
  }

  async function deleteKey(id: string) {
    try {
      await api.deleteUserAPIKey(id);
      toast.success("已删除");
      await refresh();
    } catch (err) {
      const message = err instanceof Error ? err.message : "删除失败";
      toast.error(`删除失败：${message}`);
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow="API 密钥" title="自助创建和管理平台密钥" description="密钥明文只在创建时展示一次。后续只展示脱敏值，请妥善保存。" />
      {plaintext ? (
        <Card className="border-primary/35 bg-primary/10">
          <CardContent className="flex flex-col gap-3 p-5 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="text-sm font-semibold text-primary">新密钥仅显示一次</p>
              <code className="mt-2 block break-all rounded-md bg-background/70 p-3 text-sm">{plaintext}</code>
            </div>
            <Button onClick={copyKey} variant="secondary">
              {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              复制
            </Button>
          </CardContent>
        </Card>
      ) : null}

      <Card className="glass">
        <CardHeader><CardTitle>创建 API 密钥</CardTitle></CardHeader>
        <CardContent>
          <form className="flex flex-col gap-3 sm:flex-row" onSubmit={createKey}>
            <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="例如：生产环境" required />
            <Button type="submit"><Plus className="h-4 w-4" />创建</Button>
          </form>
        </CardContent>
      </Card>

      <Card className="glass">
        <CardHeader><CardTitle>密钥列表</CardTitle></CardHeader>
        <CardContent className="grid gap-3">
          {items.length === 0 ? (
            <EmptyState title="还没有 API 密钥" description="创建一个密钥后，即可在 SDK 中配置平台接入地址。" />
          ) : (
            items.map((item) => (
              <div key={item.id} className="flex flex-col gap-3 rounded-lg border border-white/10 bg-white/[0.035] p-4 md:flex-row md:items-center md:justify-between">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <KeyRound className="h-4 w-4 text-primary" />
                    <strong>{item.name}</strong>
                    <Badge>{zhStatus(item.status)}</Badge>
                  </div>
                  <p className="mt-2 break-all font-mono text-sm text-muted-foreground">{item.mask}</p>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="icon" onClick={() => renameKey(item.id, item.name)} title="重命名"><Pencil className="h-4 w-4" /></Button>
                  {item.status === "disabled" ? (
                    <Button variant="outline" size="icon" onClick={() => enableKey(item.id)} title="重新启用"><RotateCcw className="h-4 w-4" /></Button>
                  ) : (
                    <Button variant="outline" onClick={() => disableKey(item.id)}>停用</Button>
                  )}
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
