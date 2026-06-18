import React from "react";
import { Gift, Ticket } from "lucide-react";
import type { UserLedgerRecord } from "@lingshu/shared/user-types";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { zhLedgerType } from "@/lib/i18n";
import { toast } from "sonner";

export function RedeemPage() {
  const { api, refreshMe } = useAuth();
  const [code, setCode] = React.useState("");
  const [message, setMessage] = React.useState("");
  const [ledger, setLedger] = React.useState<UserLedgerRecord[]>([]);

  async function refresh() {
    const result = await api.userLedger();
    setLedger(result.items);
  }

  React.useEffect(() => {
    refresh();
  }, []);

  async function redeem(event: React.FormEvent) {
    event.preventDefault();
    setMessage("");
    try {
      const result = await api.redeem(code);
      setCode("");
      setMessage(`兑换成功，入账 ${formatMoney(result.amount)}`);
      toast.success(`兑换成功，入账 ${formatMoney(result.amount)} 元`);
      await Promise.all([refresh(), refreshMe()]);
    } catch (err) {
      const message = err instanceof Error ? err.message : "兑换失败";
      toast.error("兑换失败：" + message);
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow="充值兑换" title="兑换码充值" description="请输入管理员提供的兑换码进行充值。" />
      <Card className="glass overflow-hidden">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[1fr_0.75fr]">
          <form className="grid content-start gap-4" onSubmit={redeem}>
            <div className="grid h-12 w-12 place-items-center rounded-lg bg-primary/10 text-primary"><Ticket className="h-6 w-6" /></div>
            <h2 className="text-2xl font-semibold">输入兑换码</h2>
            <Input value={code} onChange={(event) => setCode(event.target.value)} placeholder="LS-XXXX-XXXX" required />
            {message ? <p className="rounded-md border border-primary/30 bg-primary/10 px-3 py-2 text-sm text-primary">{message}</p> : null}
            <Button type="submit"><Gift className="h-4 w-4" />兑换</Button>
          </form>
          <div className="rounded-lg border border-white/10 bg-white/[0.035] p-5">
            <p className="text-sm text-muted-foreground">兑换成功后余额即时到账。</p>
          </div>
        </CardContent>
      </Card>

      <Card className="glass">
        <CardHeader><CardTitle>最近账本</CardTitle></CardHeader>
        <CardContent className="grid gap-3">
          {ledger.length ? ledger.slice(0, 8).map((item) => (
            <div key={`${item.type}-${item.created_at}`} className="flex items-center justify-between rounded-lg border border-white/10 bg-white/[0.035] p-3">
              <div><p className="text-sm font-medium">{zhLedgerType(item.type)}</p><p className="text-xs text-muted-foreground">{item.remark}</p></div>
              <strong className="text-sm">{formatMoney(item.amount)}</strong>
            </div>
          )) : <EmptyState title="暂无账本" description="兑换或调用后，账本记录会展示在这里。" />}
        </CardContent>
      </Card>
    </div>
  );
}
