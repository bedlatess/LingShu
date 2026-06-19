import React from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, Input, toast } from "@lingshu/ui";
import { useAuth } from "@/providers/auth";

export function LoginPage() {
  const { t } = useTranslation("auth");
  const { login, token } = useAuth();
  const location = useLocation();
  const [loginName, setLoginName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");

  if (token) return <Navigate to={(location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/dashboard"} replace />;

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      await login(loginName, password);
      toast.success(t("success"));
    } catch (err) {
      const text = err instanceof Error ? err.message : t("failed");
      setError(text);
      toast.error(text);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="grid min-h-screen place-items-center bg-background px-4 py-10">
      <div className="grid w-full max-w-5xl gap-8 lg:grid-cols-[1fr_420px] lg:items-center">
        <section className="space-y-6">
          <div className="inline-flex rounded-md border border-border bg-card px-3 py-1 text-xs text-muted-foreground">{t("badge")}</div>
          <div className="space-y-4">
            <h1 className="max-w-2xl font-serif text-5xl font-semibold leading-tight text-foreground">{t("title")}</h1>
            <p className="max-w-xl text-base leading-7 text-muted-foreground">{t("description")}</p>
          </div>
          <div className="grid max-w-xl gap-3 text-sm text-muted-foreground sm:grid-cols-3">
            <div className="rounded-lg border border-border bg-card p-4">{t("featureUnified")}</div>
            <div className="rounded-lg border border-border bg-card p-4">{t("featureBilling")}</div>
            <div className="rounded-lg border border-border bg-card p-4">{t("featureRoles")}</div>
          </div>
        </section>
        <Card>
          <CardContent className="p-6">
            <form className="grid gap-4" onSubmit={submit}>
              <div>
                <h2 className="font-serif text-2xl font-semibold">{t("loginTitle")}</h2>
                <p className="mt-2 text-sm text-muted-foreground">{t("loginDescription")}</p>
              </div>
              {error ? <p className="rounded-md border border-[var(--danger)]/30 bg-[var(--danger-soft)] px-3 py-2 text-sm text-[var(--danger)]">{error}</p> : null}
              <label className="grid gap-2 text-sm">
                {t("username")}
                <Input value={loginName} onChange={(event) => setLoginName(event.target.value)} autoComplete="username" required />
              </label>
              <label className="grid gap-2 text-sm">
                {t("password")}
                <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="current-password" required />
              </label>
              <Button type="submit" disabled={loading}>{loading ? t("submitting") : t("submit")}</Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
