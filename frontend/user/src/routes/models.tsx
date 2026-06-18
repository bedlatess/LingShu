import React from "react";
import { Boxes, Image, MessageSquareText, Search, Zap } from "lucide-react";
import type { UserModelConfig } from "@lingshu/shared/user-types";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { zhBillingMode, zhStatus, zhType } from "@/lib/i18n";

export function ModelsPage() {
  const { api } = useAuth();
  const [models, setModels] = React.useState<UserModelConfig[]>([]);
  const [keyword, setKeyword] = React.useState("");

  React.useEffect(() => {
    api.userModels().then((result) => setModels(result.items));
  }, [api]);

  const filtered = React.useMemo(() => {
    const kw = keyword.trim().toLowerCase();
    if (!kw) return models;
    return models.filter((model) =>
      [model.public_name, model.group, zhType(model.type)]
        .filter(Boolean)
        .some((field) => field.toLowerCase().includes(kw))
    );
  }, [models, keyword]);

  const groups = React.useMemo(() => {
    const map = new Map<string, UserModelConfig[]>();
    for (const model of filtered) {
      const name = model.group?.trim() || "默认分组";
      const list = map.get(name) ?? [];
      list.push(model);
      map.set(name, list);
    }
    return Array.from(map.entries()).sort((a, b) => a[0].localeCompare(b[0], "zh"));
  }, [filtered]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="可用模型" title="模型列表" description="以下模型已为你的账户开放，可直接通过平台密钥调用。" />
      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input className="pl-9" value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索模型名称或分组" />
      </div>
      {filtered.length === 0 ? (
        <EmptyState title="暂无可用模型" description={keyword ? "没有匹配的模型，试试其他关键字。" : "管理员启用模型并绑定渠道后，这里会展示可用模型。"} />
      ) : (
        <div className="grid gap-8">
          {groups.map(([groupName, items]) => (
            <section key={groupName} className="grid gap-4">
              <div className="flex items-center gap-3">
                <h2 className="text-base font-semibold">{groupName}</h2>
                <Badge>{items.length}</Badge>
              </div>
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {items.map((model) => (
                  <Card key={model.id} className="overflow-hidden transition-colors hover:border-[var(--border-strong)]">
                    <CardContent className="p-5">
                      <div className="mb-5 flex items-start justify-between gap-3">
                        <div className="grid h-11 w-11 place-items-center rounded-lg border border-border bg-[var(--bg-subtle)] text-primary">{iconFor(model.type)}</div>
                        <Badge>{zhBillingMode(model.billing_mode)}</Badge>
                      </div>
                      <h3 className="font-serif text-lg font-semibold">{model.public_name}</h3>
                      <p className="mt-1 text-sm text-muted-foreground">{zhType(model.type)}</p>
                      <div className="mt-5 grid gap-2 rounded-lg border border-border bg-[var(--bg-subtle)] p-4 text-sm">
                        <Meta label="计费方式" value={zhBillingMode(model.billing_mode)} />
                        {model.billing_mode === "per_call" ? (
                          <Meta label="单次价格" value={formatUnitPrice(model.call_unit_price)} />
                        ) : (
                          <>
                            <Meta label="输入 / 1M" value={formatUnitPrice(model.input_unit_price)} />
                            <Meta label="输出 / 1M" value={formatUnitPrice(model.output_unit_price)} />
                          </>
                        )}
                        <Meta label="状态" value={zhStatus(model.status)} />
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}

function iconFor(type: string) {
  if (type === "image") return <Image className="h-5 w-5" />;
  if (type === "embedding") return <Boxes className="h-5 w-5" />;
  if (type === "video") return <Zap className="h-5 w-5" />;
  return <MessageSquareText className="h-5 w-5" />;
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <strong className="font-serif">{value}</strong>
    </div>
  );
}

function formatUnitPrice(value?: string) {
  return Number(value ?? 0) > 0 ? formatMoney(value) : "-";
}
