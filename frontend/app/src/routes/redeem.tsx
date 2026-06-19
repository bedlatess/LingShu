import React from "react";
import { Ticket } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardHeader, CardTitle, EmptyState, Input, PageHeader, toast } from "@lingshu/ui";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";

export function RedeemPage() {
  const { t } = useTranslation("redeem");
  const { api, refreshMe } = useAuth();
  const [code, setCode] = React.useState("");
  const [lastAmount, setLastAmount] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setLoading(true);
    try {
      const result = await api.redeem(code.trim());
      setLastAmount(result.amount);
      setCode("");
      await refreshMe();
      toast.success(t("success"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("failed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow={t("eyebrow")} title={t("title")} description={t("description")} />
      <Card>
        <CardHeader><CardTitle>{t("cardTitle")}</CardTitle></CardHeader>
        <CardContent>
          <form className="grid gap-4 sm:grid-cols-[1fr_auto]" onSubmit={submit}>
            <Input value={code} onChange={(event) => setCode(event.target.value)} placeholder={t("placeholder")} required />
            <Button type="submit" disabled={loading}>{loading ? t("submitting") : t("submit")}</Button>
          </form>
        </CardContent>
      </Card>
      {lastAmount ? (
        <Card>
          <CardContent className="flex items-center gap-4 p-5">
            <div className="grid h-12 w-12 place-items-center rounded-md border border-border bg-[var(--clay-soft)] text-[var(--clay)]"><Ticket className="h-5 w-5" /></div>
            <div>
              <p className="font-serif text-xl font-semibold">{t("received", { amount: formatMoney(lastAmount) })}</p>
              <p className="text-sm text-muted-foreground">{t("receivedHint")}</p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <EmptyState title={t("emptyTitle")} description={t("emptyDescription")} icon={<Ticket className="h-5 w-5" />} />
      )}
    </div>
  );
}
