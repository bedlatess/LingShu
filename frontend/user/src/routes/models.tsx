import React from "react";
import { Boxes, Image, MessageSquareText, Zap } from "lucide-react";
import type { UserModelConfig } from "@lingshu/shared/user-types";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { zhBillingMode, zhStatus, zhType } from "@/lib/i18n";

export function ModelsPage() {
  const { api } = useAuth();
  const [models, setModels] = React.useState<UserModelConfig[]>([]);

  React.useEffect(() => {
    api.userModels().then((result) => setModels(result.items));
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="可用模型" title="模型列表" description="以下模型已为你的账户开放，可直接通过平台密钥调用。" />
      {models.length === 0 ? (
        <EmptyState title="暂无可用模型" description="管理员启用模型并绑定渠道后，这里会展示可用模型。" />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {models.map((model) => (
            <Card key={model.id} className="glass overflow-hidden transition-all hover:-translate-y-1 hover:border-primary/35">
              <CardContent className="p-5">
                <div className="mb-5 flex items-start justify-between gap-3">
                  <div className="grid h-11 w-11 place-items-center rounded-lg bg-primary/10 text-primary">{iconFor(model.type)}</div>
                  <Badge>{zhBillingMode(model.billing_mode)}</Badge>
                </div>
                <h3 className="text-lg font-semibold">{model.public_name}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{model.group || "默认分组"} · {zhType(model.type)}</p>
                <div className="mt-5 grid gap-2 rounded-lg border border-white/10 bg-white/[0.035] p-4 text-sm">
                  <Meta label="计费方式" value={zhBillingMode(model.billing_mode)} />
                  <Meta label="状态" value={zhStatus(model.status)} />
                </div>
              </CardContent>
            </Card>
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
      <strong>{value}</strong>
    </div>
  );
}
