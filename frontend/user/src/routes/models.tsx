import React from "react";
import { Boxes, Image, MessageSquareText, Zap } from "lucide-react";
import type { UserModelPrice } from "@lingshu/shared";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";

export function ModelsPage() {
  const { api } = useAuth();
  const [models, setModels] = React.useState<UserModelPrice[]>([]);

  React.useEffect(() => {
    api.userModels().then((result) => setModels(result.items));
  }, [api]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Models" title="模型广场和实际单价" description="展示基准成本乘倍率后的实际扣费口径。管理员调倍率后，新请求按新倍率结算。" />
      {models.length === 0 ? (
        <EmptyState title="暂无可用模型" description="管理员启用模型并绑定渠道后，这里会展示价格。" />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {models.map((model) => (
            <Card key={model.id} className="glass overflow-hidden transition-all hover:-translate-y-1 hover:border-primary/35">
              <CardContent className="p-5">
                <div className="mb-5 flex items-start justify-between gap-3">
                  <div className="grid h-11 w-11 place-items-center rounded-lg bg-primary/10 text-primary">{iconFor(model.type)}</div>
                  <Badge>{model.billing_mode}</Badge>
                </div>
                <h3 className="text-lg font-semibold">{model.public_name}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{model.group || "default"} · {model.type}</p>
                <div className="mt-5 grid gap-2 rounded-lg border border-white/10 bg-white/[0.035] p-4 text-sm">
                  {model.billing_mode === "per_call" ? (
                    <Price label="每次调用" value={model.call_unit_price} />
                  ) : (
                    <>
                      <Price label="输入 / 1K" value={model.input_unit_price} />
                      <Price label="输出 / 1K" value={model.output_unit_price} />
                    </>
                  )}
                  <Price label="倍率" value={`${model.rate_multiplier}x`} raw />
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

function Price({ label, value, raw }: { label: string; value: string; raw?: boolean }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <strong>{raw ? value : formatMoney(value)}</strong>
    </div>
  );
}
